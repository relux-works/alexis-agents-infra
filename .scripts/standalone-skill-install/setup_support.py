#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import re
import shutil
import subprocess
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Callable, Optional


MANIFEST_FILENAME = ".skill-install.json"
CATALOG_RELATIVE_PATH = Path("locales") / "metadata.json"
SUPPORTED_BASE_LOCALES = ("en", "ru")
SUPPORTED_LOCALE_MODES = ("en", "ru", "en-ru", "ru-en")
REQUIRED_LOCALE_KEYS = (
    "description",
    "display_name",
    "short_description",
    "default_prompt",
    "local_prefix",
    "triggers",
)
OPENAI_YAML_FIELD_TEMPLATE = r"^(\s*{key}:\s*)(.*)$"
FRONTMATTER_KEY_RE = re.compile(r"^(?P<key>[A-Za-z0-9_-]+):(.*)$")
MARKDOWN_HEADING_RE = re.compile(r"^#{1,6}\s+.*$", re.MULTILINE)
MODULES_HEADING_RE = re.compile(r"^(?P<level>#{1,6})\s+Modules\s*$", re.MULTILINE)
SKILL_TRIGGERS_INCLUDE_NAME = "INSTRUCTIONS_SKILL_TRIGGERS.md"
GLOBAL_AGENTS_ENTRYPOINT = Path(".agents") / ".instructions" / "AGENTS.md"
GLOBAL_TRIGGER_INSTRUCTIONS = Path(".agents") / ".instructions" / SKILL_TRIGGERS_INCLUDE_NAME
LOCAL_PROJECT_AGENTS_ENTRYPOINT = Path("AGENTS.md")
LOCAL_PROJECT_TESTING_MODULE = Path(".agents") / ".instructions" / "INSTRUCTIONS_TESTING.md"
LOCAL_PROJECT_TESTING_MODULE_REF = "@.agents/.instructions/INSTRUCTIONS_TESTING.md"
MANAGED_TRIGGER_SECTION_START = "<!-- standalone-skill-install:managed-triggers:start -->"
MANAGED_TRIGGER_SECTION_END = "<!-- standalone-skill-install:managed-triggers:end -->"
MANAGED_TRIGGER_ENTRY_PREFIX = "<!-- standalone-skill-install:managed-trigger-entry "
MANAGED_TRIGGER_ENTRY_SUFFIX = " -->"
CopyIgnore = shutil.ignore_patterns(
    ".git",
    ".venv",
    "__pycache__",
    ".DS_Store",
    "*.pyc",
    MANIFEST_FILENAME,
)


class SetupError(RuntimeError):
    pass


@dataclass(frozen=True)
class LocaleSelection:
    mode: str
    primary_locale: str
    secondary_locale: Optional[str]


@dataclass(frozen=True)
class InstallResult:
    skill_name: str
    install_mode: str
    source_dir: Path
    runtime_dir: Path
    install_root: Path
    claude_link: Path
    codex_link: Path
    locale_mode: str


@dataclass(frozen=True)
class TriggerInstructionEntry:
    skill_name: str
    triggers: list[str]


LOCAL_TESTING_INSTRUCTIONS_TEXT = """# Testing & Refactoring

## Testing

* Use **Swift Testing** framework, not XCTest.
* Tests must be in **Swift**, not ObjC.

---

## Refactoring workflow

When refactoring (e.g., ObjC → Swift):

1. **Write tests first** (if none exist).
   * Test coverage must be **high for the code being refactored** (not the whole project):
     * target **~80%+** at minimum;
     * **prefer 100%** where practical.

2. **Refactor code.**

3. **Run tests** to verify nothing broke.
"""


def yaml_quote(value: str) -> str:
    return json.dumps(value, ensure_ascii=False)


def unique_strings(values: list[str]) -> list[str]:
    seen: set[str] = set()
    results: list[str] = []
    for value in values:
        cleaned = value.strip()
        if not cleaned:
            continue
        key = cleaned.lower()
        if key in seen:
            continue
        seen.add(key)
        results.append(cleaned)
    return results


def escape_markdown_table_cell(value: str) -> str:
    return value.replace("|", r"\|").replace("\n", " ").strip()


def default_trigger_instructions_document() -> str:
    return (
        "# Skill Triggers\n\n"
        "Automatic skill activation rules. When these topics come up, load the matching skill first.\n"
    )


def render_trigger_instruction_row(entry: TriggerInstructionEntry) -> str:
    trigger_cell = escape_markdown_table_cell(", ".join(entry.triggers))
    skill_cell = f"`{entry.skill_name}`"
    action_cell = f"Load `{entry.skill_name}` via Skill tool before proceeding."
    return f"| {trigger_cell} | {skill_cell} | {action_cell} |\n"


def render_managed_trigger_section(entries: list[TriggerInstructionEntry]) -> str:
    lines = [
        "## Standalone Skills\n\n",
        "Managed entries for standalone skills installed outside alexis-agents-infra.\n\n",
        f"{MANAGED_TRIGGER_SECTION_START}\n",
        "| Triggers | Skill | Action |\n",
        "|----------|-------|--------|\n",
    ]
    for entry in sorted(entries, key=lambda item: item.skill_name.lower()):
        lines.append(
            f"{MANAGED_TRIGGER_ENTRY_PREFIX}"
            f"{json.dumps({'skill_name': entry.skill_name, 'triggers': entry.triggers}, ensure_ascii=False)}"
            f"{MANAGED_TRIGGER_ENTRY_SUFFIX}\n"
        )
        lines.append(render_trigger_instruction_row(entry))
    lines.append(f"{MANAGED_TRIGGER_SECTION_END}\n")
    return "".join(lines)


def parse_managed_trigger_section(text: str) -> list[TriggerInstructionEntry]:
    start = text.find(MANAGED_TRIGGER_SECTION_START)
    end = text.find(MANAGED_TRIGGER_SECTION_END)
    if start == -1 or end == -1 or end < start:
        return []

    entries: list[TriggerInstructionEntry] = []
    parsed_from_comments = False
    body = text[start + len(MANAGED_TRIGGER_SECTION_START) : end]
    for raw_line in body.splitlines():
        line = raw_line.strip()
        if line.startswith(MANAGED_TRIGGER_ENTRY_PREFIX) and line.endswith(MANAGED_TRIGGER_ENTRY_SUFFIX):
            payload_text = line[len(MANAGED_TRIGGER_ENTRY_PREFIX) : -len(MANAGED_TRIGGER_ENTRY_SUFFIX)].strip()
            try:
                payload = json.loads(payload_text)
            except json.JSONDecodeError:
                continue
            skill_name = payload.get("skill_name")
            triggers = payload.get("triggers")
            if not isinstance(skill_name, str) or not skill_name.strip():
                continue
            if not isinstance(triggers, list):
                continue
            normalized_triggers = unique_strings([item for item in triggers if isinstance(item, str)])
            entries.append(
                TriggerInstructionEntry(
                    skill_name=skill_name.strip(),
                    triggers=normalized_triggers,
                )
            )
            parsed_from_comments = True
            continue
        if not line.startswith("|"):
            continue
        if line.startswith("| Triggers ") or line.startswith("|----------"):
            continue
        parts = [part.strip() for part in line.split("|")[1:-1]]
        if len(parts) != 3:
            continue
        triggers_cell, skill_cell, _ = parts
        skill_name = skill_cell.strip().strip("`")
        if not skill_name:
            continue
        triggers = unique_strings(
            [item.replace(r"\|", "|").strip() for item in triggers_cell.split(",")]
        )
        entries.append(TriggerInstructionEntry(skill_name=skill_name, triggers=triggers))
    if parsed_from_comments:
        deduped: dict[str, TriggerInstructionEntry] = {}
        for entry in entries:
            deduped[entry.skill_name] = entry
        return list(deduped.values())
    return entries


def replace_or_append_managed_trigger_section(text: str, entries: list[TriggerInstructionEntry]) -> str:
    section = render_managed_trigger_section(entries)
    start = text.find(MANAGED_TRIGGER_SECTION_START)
    end = text.find(MANAGED_TRIGGER_SECTION_END)
    if start != -1 and end != -1 and end >= start:
        end += len(MANAGED_TRIGGER_SECTION_END)
        remainder = text[end:]
        if remainder.startswith("\n"):
            end += 1
        return text[:start] + section + text[end:]

    if not text.strip():
        return default_trigger_instructions_document() + "\n" + section
    return text.rstrip() + "\n\n" + section


def validate_global_agents_entrypoint(global_agents_path: Path) -> None:
    try:
        agents_text = global_agents_path.read_text(encoding="utf-8")
    except FileNotFoundError as exc:
        raise SetupError(
            "Global agents instructions are missing. "
            f"Expected {global_agents_path}. Run agents-infra setup global first."
        ) from exc

    for line in agents_text.splitlines():
        stripped = line.strip()
        if not stripped.startswith("@"):
            continue
        if Path(stripped[1:].strip()).name == SKILL_TRIGGERS_INCLUDE_NAME:
            return

    raise SetupError(
        "Global AGENTS.md does not include the skill trigger module. "
        f"Add a reference to {SKILL_TRIGGERS_INCLUDE_NAME} in {global_agents_path}."
    )


def register_global_skill_triggers(skill_name: str, triggers: list[str]) -> None:
    global_agents_path = Path.home() / GLOBAL_AGENTS_ENTRYPOINT
    trigger_instructions_path = Path.home() / GLOBAL_TRIGGER_INSTRUCTIONS
    validate_global_agents_entrypoint(global_agents_path)

    if trigger_instructions_path.exists():
        text = trigger_instructions_path.read_text(encoding="utf-8")
    else:
        trigger_instructions_path.parent.mkdir(parents=True, exist_ok=True)
        text = ""

    existing_entries = {
        entry.skill_name: entry for entry in parse_managed_trigger_section(text)
    }
    existing_entries[skill_name] = TriggerInstructionEntry(
        skill_name=skill_name,
        triggers=unique_strings(triggers),
    )
    updated = replace_or_append_managed_trigger_section(
        text,
        list(existing_entries.values()),
    )
    trigger_instructions_path.write_text(updated, encoding="utf-8")


def ensure_local_testing_module(repo_root: Path) -> None:
    module_path = repo_root / LOCAL_PROJECT_TESTING_MODULE
    if module_path.exists():
        if module_path.is_dir():
            raise SetupError(f"Expected a file, found a directory: {module_path}")
        return
    module_path.parent.mkdir(parents=True, exist_ok=True)
    module_path.write_text(LOCAL_TESTING_INSTRUCTIONS_TEXT, encoding="utf-8")


def ensure_local_agents_modules_section(text: str, required_ref: str) -> str:
    match = MODULES_HEADING_RE.search(text)
    if match is None:
        if not text.strip():
            return f"## Modules\n\n{required_ref}\n"
        return text.rstrip() + f"\n\n## Modules\n\n{required_ref}\n"

    section_start = match.end()
    next_heading = MARKDOWN_HEADING_RE.search(text, section_start)
    section_end = next_heading.start() if next_heading else len(text)
    section_body = text[section_start:section_end]
    if required_ref in {line.strip() for line in section_body.splitlines()}:
        return text

    has_following_heading = next_heading is not None
    trimmed_body = section_body.rstrip("\n")
    suffix = "\n\n" if has_following_heading else "\n"
    if trimmed_body.strip():
        updated_body = f"{trimmed_body}\n{required_ref}{suffix}"
    else:
        updated_body = f"\n\n{required_ref}{suffix}"
    return text[:section_start] + updated_body + text[section_end:]


def ensure_local_agents_entrypoint(repo_root: Path) -> None:
    agents_path = repo_root / LOCAL_PROJECT_AGENTS_ENTRYPOINT
    if agents_path.exists() and agents_path.is_dir():
        raise SetupError(f"Expected a file, found a directory: {agents_path}")

    existing_text = ""
    if agents_path.exists():
        existing_text = agents_path.read_text(encoding="utf-8")
    updated_text = ensure_local_agents_modules_section(
        existing_text,
        LOCAL_PROJECT_TESTING_MODULE_REF,
    )
    agents_path.write_text(updated_text, encoding="utf-8")


def parse_locale_mode(value: str) -> LocaleSelection:
    normalized = value.strip().lower()
    if normalized not in SUPPORTED_LOCALE_MODES:
        supported = ", ".join(SUPPORTED_LOCALE_MODES)
        raise SetupError(f"Unsupported locale mode: {value}. Supported values: {supported}")

    if "-" in normalized:
        primary_locale, secondary_locale = normalized.split("-", 1)
        return LocaleSelection(
            mode=normalized,
            primary_locale=primary_locale,
            secondary_locale=secondary_locale,
        )

    return LocaleSelection(mode=normalized, primary_locale=normalized, secondary_locale=None)


def skill_data_home() -> Path:
    xdg_data_home = os.environ.get("XDG_DATA_HOME", "").strip()
    if xdg_data_home:
        return Path(xdg_data_home)
    return Path.home() / ".local" / "share"


def managed_global_install_dir(skill_name: str) -> Path:
    return skill_data_home() / "agents" / "skills" / skill_name


def install_manifest_path(skill_dir: Path) -> Path:
    return skill_dir / MANIFEST_FILENAME


def load_install_manifest(skill_dir: Path) -> Optional[dict[str, Any]]:
    path = install_manifest_path(skill_dir)
    if not path.exists():
        return None
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        raise SetupError(f"Invalid install manifest: {path}") from exc
    if not isinstance(payload, dict):
        raise SetupError(f"Expected JSON object in install manifest: {path}")
    return payload


def write_install_manifest(
    *,
    skill_dir: Path,
    skill_name: str,
    install_mode: str,
    locale_mode: str,
    source_dir: Path,
    runtime_dir: Path,
) -> None:
    selection = parse_locale_mode(locale_mode)
    payload = {
        "schema_version": 1,
        "skill_name": skill_name,
        "install_mode": install_mode,
        "locale_mode": selection.mode,
        "primary_locale": selection.primary_locale,
        "secondary_locale": selection.secondary_locale,
        "source_dir": str(source_dir),
        "runtime_dir": str(runtime_dir),
    }
    install_manifest_path(skill_dir).write_text(
        json.dumps(payload, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )


def resolve_source_dir(current_skill_dir: Path) -> Path:
    manifest = load_install_manifest(current_skill_dir)
    if manifest:
        candidate = manifest.get("source_dir")
        if isinstance(candidate, str) and candidate.strip():
            candidate_path = Path(candidate).expanduser()
            if candidate_path.exists():
                return candidate_path.resolve()
    return current_skill_dir.resolve()


def load_metadata_catalog(skill_dir: Path) -> dict[str, dict[str, Any]]:
    path = skill_dir / CATALOG_RELATIVE_PATH
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError as exc:
        raise SetupError(f"Missing localization catalog: {path}") from exc
    except json.JSONDecodeError as exc:
        raise SetupError(f"Invalid localization catalog: {path}") from exc

    locales = payload.get("locales")
    if not isinstance(locales, dict):
        raise SetupError(f"Localization catalog must contain a 'locales' object: {path}")

    normalized: dict[str, dict[str, Any]] = {}
    for locale in SUPPORTED_BASE_LOCALES:
        locale_payload = locales.get(locale)
        if not isinstance(locale_payload, dict):
            raise SetupError(f"Missing locale '{locale}' in localization catalog: {path}")
        normalized_locale: dict[str, Any] = {}
        for key in REQUIRED_LOCALE_KEYS:
            value = locale_payload.get(key)
            if key == "triggers":
                if not isinstance(value, list) or not value:
                    raise SetupError(f"Locale '{locale}' must define a non-empty trigger list in {path}")
                triggers: list[str] = []
                for item in value:
                    if not isinstance(item, str) or not item.strip():
                        raise SetupError(f"Locale '{locale}' contains an invalid trigger in {path}")
                    triggers.append(item)
                normalized_locale[key] = triggers
                continue
            if not isinstance(value, str) or not value:
                raise SetupError(f"Locale '{locale}' is missing string field '{key}' in {path}")
            normalized_locale[key] = value
        normalized[locale] = normalized_locale
    return normalized


def build_localized_metadata(skill_dir: Path, locale_mode: str, install_mode: str) -> dict[str, Any]:
    selection = parse_locale_mode(locale_mode)
    catalog = load_metadata_catalog(skill_dir)
    primary = catalog[selection.primary_locale]

    description = primary["description"]
    triggers = list(primary["triggers"])
    if selection.secondary_locale is not None:
        secondary = catalog[selection.secondary_locale]
        description = f"{description} / {secondary['description']}"
        triggers = unique_strings([*triggers, *secondary["triggers"]])

    display_name = primary["display_name"]
    short_description = primary["short_description"]
    default_prompt = primary["default_prompt"]

    if install_mode == "local":
        prefix = primary["local_prefix"]
        description = f"{prefix}{description}"
        display_name = f"{prefix}{display_name}"
        short_description = f"{prefix}{short_description}"

    return {
        "description": description,
        "display_name": display_name,
        "short_description": short_description,
        "default_prompt": default_prompt,
        "triggers": triggers,
    }


def parse_frontmatter_sections(skill_text: str) -> tuple[list[tuple[str, str]], str]:
    lines = skill_text.splitlines(keepends=True)
    if not lines or lines[0].strip() != "---":
        raise SetupError("SKILL.md must start with a YAML frontmatter block.")

    closing_index: int | None = None
    for index, line in enumerate(lines[1:], start=1):
        if line.strip() == "---":
            closing_index = index
            break
    if closing_index is None:
        raise SetupError("SKILL.md frontmatter is missing a closing --- line.")

    frontmatter_lines = lines[1:closing_index]
    body = "".join(lines[closing_index + 1 :])
    sections: list[tuple[str, str]] = []
    current_key: str | None = None
    current_lines: list[str] = []

    for line in frontmatter_lines:
        match = FRONTMATTER_KEY_RE.match(line)
        if match:
            if current_key is not None:
                sections.append((current_key, "".join(current_lines)))
            current_key = match.group("key")
            current_lines = [line]
            continue
        if current_key is None:
            raise SetupError("SKILL.md frontmatter contains content before the first key.")
        current_lines.append(line)

    if current_key is not None:
        sections.append((current_key, "".join(current_lines)))

    return sections, body


def render_triggers_block(triggers: list[str]) -> str:
    lines = ["triggers:\n"]
    for trigger in triggers:
        lines.append(f"  - {yaml_quote(trigger)}\n")
    return "".join(lines)


def replace_frontmatter_sections(skill_text: str, replacements: dict[str, str]) -> str:
    sections, body = parse_frontmatter_sections(skill_text)
    index_by_key = {key: index for index, (key, _) in enumerate(sections)}

    for key, rendered_value in replacements.items():
        if key in index_by_key:
            sections[index_by_key[key]] = (key, rendered_value)
            continue

        insert_at = len(sections)
        if key == "triggers" and "description" in index_by_key:
            insert_at = index_by_key["description"] + 1
        sections.insert(insert_at, (key, rendered_value))
        index_by_key = {section_key: index for index, (section_key, _) in enumerate(sections)}

    frontmatter = "".join(section for _, section in sections)
    return f"---\n{frontmatter}---\n{body}"


def render_skill_metadata(skill_dir: Path, locale_mode: str, install_mode: str) -> None:
    metadata = build_localized_metadata(skill_dir, locale_mode, install_mode)

    skill_md_path = skill_dir / "SKILL.md"
    skill_text = skill_md_path.read_text(encoding="utf-8")
    skill_text = replace_frontmatter_sections(
        skill_text,
        {
            "description": f"description: {yaml_quote(metadata['description'])}\n",
            "triggers": render_triggers_block(metadata["triggers"]),
        },
    )
    skill_md_path.write_text(skill_text, encoding="utf-8")

    openai_yaml_path = skill_dir / "agents" / "openai.yaml"
    yaml_text = openai_yaml_path.read_text(encoding="utf-8")
    for key in ("display_name", "short_description", "default_prompt"):
        pattern = re.compile(OPENAI_YAML_FIELD_TEMPLATE.format(key=key), flags=re.MULTILINE)
        yaml_text, count = pattern.subn(
            lambda match, value=metadata[key]: f"{match.group(1)}{yaml_quote(value)}",
            yaml_text,
            count=1,
        )
        if count != 1:
            raise SetupError(f"Could not update {key} in {openai_yaml_path}")
    openai_yaml_path.write_text(yaml_text, encoding="utf-8")


def sync_skill_copy(source_dir: Path, dest_dir: Path) -> None:
    if dest_dir.is_symlink() or dest_dir.is_file():
        dest_dir.unlink()
    elif dest_dir.exists():
        shutil.rmtree(dest_dir, ignore_errors=False)
    dest_dir.parent.mkdir(parents=True, exist_ok=True)
    shutil.copytree(source_dir, dest_dir, ignore=CopyIgnore)


def ensure_skill_link(link_value: str, target_path: Path) -> None:
    if target_path.is_symlink() or target_path.is_file():
        target_path.unlink()
    elif target_path.exists():
        raise SetupError(f"Refusing to replace existing directory: {target_path}")

    target_path.parent.mkdir(parents=True, exist_ok=True)
    os.symlink(link_value, target_path)


def run_bootstrap(skill_dir: Path) -> None:
    bootstrap_path = skill_dir / "scripts" / "bootstrap.sh"
    if not bootstrap_path.exists():
        return
    subprocess.run([str(bootstrap_path), "--quiet"], check=True)


def resolve_repo_root(path: Path) -> Path:
    completed = subprocess.run(
        ["git", "-C", str(path), "rev-parse", "--show-toplevel"],
        check=False,
        capture_output=True,
        text=True,
    )
    if completed.returncode != 0:
        raise SetupError(f"Local mode expects a git repository: {path}")
    return Path(completed.stdout.strip()).resolve()


def resolve_locale_mode(install_mode: str, runtime_dir: Path, requested_locale: Optional[str]) -> str:
    manifest = load_install_manifest(runtime_dir)
    existing_locale = manifest.get("locale_mode") if manifest else None
    if existing_locale is not None and not isinstance(existing_locale, str):
        raise SetupError(f"Invalid locale_mode in install manifest: {install_manifest_path(runtime_dir)}")

    if requested_locale:
        requested_mode = parse_locale_mode(requested_locale).mode
        if install_mode == "local" and existing_locale and existing_locale != requested_mode:
            raise SetupError(
                "Local install locale is project-fixed after the first install. "
                f"Expected {existing_locale}, got {requested_mode}."
            )
        return requested_mode

    if existing_locale:
        return parse_locale_mode(existing_locale).mode

    supported = ", ".join(SUPPORTED_LOCALE_MODES)
    raise SetupError(f"First {install_mode} install requires --locale <{supported}>.")


def perform_install(
    *,
    source_dir: Path,
    install_mode: str,
    requested_locale: Optional[str],
    repo_root: Optional[Path] = None,
    bootstrap_runner: Callable[[Path], None] = run_bootstrap,
) -> InstallResult:
    source_dir = resolve_source_dir(source_dir).resolve()
    skill_name = source_dir.name

    if install_mode == "global":
        install_root = Path.home()
        runtime_dir = managed_global_install_dir(skill_name)
        claude_link_value = str(runtime_dir)
        codex_link_value = str(runtime_dir)
    elif install_mode == "local":
        if repo_root is None:
            raise SetupError("Local install requires a repository path.")
        install_root = resolve_repo_root(repo_root)
        runtime_dir = install_root / ".skills" / skill_name
        claude_link_value = f"../../.skills/{skill_name}"
        codex_link_value = f"../../.skills/{skill_name}"
    else:
        raise SetupError(f"Unsupported install mode: {install_mode}")

    locale_mode = resolve_locale_mode(install_mode, runtime_dir, requested_locale)

    sync_skill_copy(source_dir, runtime_dir)
    render_skill_metadata(runtime_dir, locale_mode, install_mode)
    write_install_manifest(
        skill_dir=runtime_dir,
        skill_name=skill_name,
        install_mode=install_mode,
        locale_mode=locale_mode,
        source_dir=source_dir,
        runtime_dir=runtime_dir,
    )
    bootstrap_runner(runtime_dir)

    claude_link = install_root / ".claude" / "skills" / skill_name
    codex_link = install_root / ".codex" / "skills" / skill_name
    ensure_skill_link(claude_link_value, claude_link)
    ensure_skill_link(codex_link_value, codex_link)
    if install_mode == "global":
        metadata = build_localized_metadata(runtime_dir, locale_mode, install_mode)
        register_global_skill_triggers(skill_name, metadata["triggers"])
    if install_mode == "local":
        ensure_local_testing_module(install_root)
        ensure_local_agents_entrypoint(install_root)

    return InstallResult(
        skill_name=skill_name,
        install_mode=install_mode,
        source_dir=source_dir,
        runtime_dir=runtime_dir,
        install_root=install_root,
        claude_link=claude_link,
        codex_link=codex_link,
        locale_mode=locale_mode,
    )
