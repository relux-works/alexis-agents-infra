from __future__ import annotations

import json
import os
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock


SCRIPTS_DIR = Path(__file__).resolve().parents[1] / ".scripts" / "standalone-skill-install"
if str(SCRIPTS_DIR) not in sys.path:
    sys.path.insert(0, str(SCRIPTS_DIR))

import setup_support as ss


class StandaloneSkillSetupSupportTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tempdir = tempfile.TemporaryDirectory()
        self.addCleanup(self.tempdir.cleanup)
        self.root = Path(self.tempdir.name)
        self.home = self.root / "home"
        self.home.mkdir(parents=True, exist_ok=True)
        self.xdg_data_home = self.root / "xdg-data"
        self.env_patcher = mock.patch.dict(
            os.environ,
            {
                "HOME": str(self.home),
                "XDG_DATA_HOME": str(self.xdg_data_home),
            },
            clear=False,
        )
        self.env_patcher.start()
        self.addCleanup(self.env_patcher.stop)

    def make_source_skill_dir(self) -> Path:
        skill_dir = self.root / "source" / "skill-sample"
        (skill_dir / "agents").mkdir(parents=True, exist_ok=True)
        (skill_dir / "locales").mkdir(parents=True, exist_ok=True)
        (skill_dir / "scripts").mkdir(parents=True, exist_ok=True)
        (skill_dir / ".git").mkdir(parents=True, exist_ok=True)
        (skill_dir / ".venv").mkdir(parents=True, exist_ok=True)
        (skill_dir / "__pycache__").mkdir(parents=True, exist_ok=True)

        (skill_dir / "SKILL.md").write_text(
            "---\n"
            "name: skill-sample\n"
            "description: >\n"
            "  English source description\n"
            "  on multiple lines\n"
            "triggers:\n"
            '  - "my open merge requests"\n'
            '  - "status mr"\n'
            "---\n\n"
            "# Sample Skill\n",
            encoding="utf-8",
        )
        (skill_dir / "agents" / "openai.yaml").write_text(
            'interface:\n'
            '  display_name: "English Display"\n'
            '  short_description: "English Short"\n'
            '  default_prompt: "Use $skill-sample in English."\n',
            encoding="utf-8",
        )
        (skill_dir / "locales" / "metadata.json").write_text(
            json.dumps(
                {
                    "locales": {
                        "en": {
                            "description": "English localized description",
                            "display_name": "English Display",
                            "short_description": "English Short",
                            "default_prompt": "Use $skill-sample in English.",
                            "local_prefix": "[local] ",
                            "triggers": [
                                "my open merge requests",
                                "status mr",
                                "gitlab merge request",
                            ],
                        },
                        "ru": {
                            "description": "Русское описание",
                            "display_name": "Русский Display",
                            "short_description": "Русский Short",
                            "default_prompt": "Используй $skill-sample по-русски.",
                            "local_prefix": "[локально] ",
                            "triggers": [
                                "мои открытые mr",
                                "status mr",
                            ],
                        },
                    }
                },
                ensure_ascii=False,
                indent=2,
            )
            + "\n",
            encoding="utf-8",
        )
        (skill_dir / ".git" / "config").write_text("", encoding="utf-8")
        (skill_dir / ".venv" / "marker").write_text("", encoding="utf-8")
        (skill_dir / "__pycache__" / "cache.pyc").write_text("", encoding="utf-8")
        return skill_dir

    def make_global_agents_runtime(self, *, include_trigger_module: bool = True) -> Path:
        instructions_dir = self.home / ".agents" / ".instructions"
        instructions_dir.mkdir(parents=True, exist_ok=True)

        lines = [
            "# Global Instructions\n",
            "\n",
            "## Modules\n",
            "\n",
            "@~/.agents/.instructions/INSTRUCTIONS_SKILLS.md\n",
        ]
        if include_trigger_module:
            lines.append("@~/.agents/.instructions/INSTRUCTIONS_SKILL_TRIGGERS.md\n")
        (instructions_dir / "AGENTS.md").write_text("".join(lines), encoding="utf-8")
        return instructions_dir

    def test_render_skill_metadata_dual_mode_merges_trigger_lists(self) -> None:
        skill_dir = self.make_source_skill_dir()

        ss.render_skill_metadata(skill_dir, "ru-en", "global")

        skill_text = (skill_dir / "SKILL.md").read_text(encoding="utf-8")
        openai_yaml = (skill_dir / "agents" / "openai.yaml").read_text(encoding="utf-8")

        self.assertIn('description: "Русское описание / English localized description"', skill_text)
        self.assertIn('  - "мои открытые mr"\n', skill_text)
        self.assertIn('  - "status mr"\n', skill_text)
        self.assertIn('  - "my open merge requests"\n', skill_text)
        self.assertIn('  - "gitlab merge request"\n', skill_text)
        self.assertEqual(skill_text.count('"status mr"'), 1)
        self.assertIn('display_name: "Русский Display"', openai_yaml)
        self.assertIn('short_description: "Русский Short"', openai_yaml)
        self.assertIn('default_prompt: "Используй $skill-sample по-русски."', openai_yaml)

    def test_perform_global_install_requires_locale_for_first_install(self) -> None:
        source_dir = self.make_source_skill_dir()

        with self.assertRaises(ss.SetupError) as exc:
            ss.perform_install(
                source_dir=source_dir,
                install_mode="global",
                requested_locale=None,
                bootstrap_runner=lambda _: None,
            )

        self.assertIn("First global install requires --locale", str(exc.exception))

    def test_perform_global_install_creates_managed_copy_and_ignores_local_artifacts(self) -> None:
        source_dir = self.make_source_skill_dir()
        instructions_dir = self.make_global_agents_runtime()

        first_result = ss.perform_install(
            source_dir=source_dir,
            install_mode="global",
            requested_locale="ru",
            bootstrap_runner=lambda _: None,
        )
        second_result = ss.perform_install(
            source_dir=source_dir,
            install_mode="global",
            requested_locale=None,
            bootstrap_runner=lambda _: None,
        )

        self.assertEqual(
            first_result.runtime_dir.resolve(),
            (self.xdg_data_home / "agents" / "skills" / "skill-sample").resolve(),
        )
        self.assertEqual(second_result.locale_mode, "ru")
        self.assertTrue((first_result.runtime_dir / ss.MANIFEST_FILENAME).exists())
        self.assertFalse((first_result.runtime_dir / ".git").exists())
        self.assertFalse((first_result.runtime_dir / ".venv").exists())
        self.assertFalse((first_result.runtime_dir / "__pycache__").exists())
        self.assertEqual(
            (self.home / ".claude" / "skills" / "skill-sample").resolve(),
            first_result.runtime_dir.resolve(),
        )
        self.assertEqual(
            (self.home / ".codex" / "skills" / "skill-sample").resolve(),
            first_result.runtime_dir.resolve(),
        )
        self.assertIn(
            'description: "Русское описание"',
            (first_result.runtime_dir / "SKILL.md").read_text(encoding="utf-8"),
        )
        trigger_doc = (instructions_dir / "INSTRUCTIONS_SKILL_TRIGGERS.md").read_text(encoding="utf-8")
        self.assertIn("## Standalone Skills", trigger_doc)
        self.assertIn("| мои открытые mr, status mr | `skill-sample` |", trigger_doc)
        self.assertEqual(
            trigger_doc.count("standalone-skill-install:managed-trigger-entry"),
            1,
        )

    def test_perform_local_install_uses_project_fixed_locale(self) -> None:
        source_dir = self.make_source_skill_dir()
        repo_root = self.root / "repo"
        repo_root.mkdir(parents=True, exist_ok=True)

        with mock.patch.object(ss, "resolve_repo_root", return_value=repo_root.resolve()):
            first_result = ss.perform_install(
                source_dir=source_dir,
                install_mode="local",
                requested_locale="ru",
                repo_root=repo_root,
                bootstrap_runner=lambda _: None,
            )

            with self.assertRaises(ss.SetupError) as exc:
                ss.perform_install(
                    source_dir=source_dir,
                    install_mode="local",
                    requested_locale="en",
                    repo_root=repo_root,
                    bootstrap_runner=lambda _: None,
                )

        self.assertEqual(
            first_result.runtime_dir.resolve(),
            (repo_root / ".skills" / "skill-sample").resolve(),
        )
        self.assertEqual(
            os.readlink(repo_root / ".claude" / "skills" / "skill-sample"),
            "../../.skills/skill-sample",
        )
        self.assertEqual(
            os.readlink(repo_root / ".codex" / "skills" / "skill-sample"),
            "../../.skills/skill-sample",
        )
        self.assertIn("project-fixed", str(exc.exception))
        rendered_skill = (first_result.runtime_dir / "SKILL.md").read_text(encoding="utf-8")
        rendered_yaml = (first_result.runtime_dir / "agents" / "openai.yaml").read_text(encoding="utf-8")
        self.assertIn('description: "[локально] Русское описание"', rendered_skill)
        self.assertIn('display_name: "[локально] Русский Display"', rendered_yaml)
        self.assertIn('short_description: "[локально] Русский Short"', rendered_yaml)
        self.assertIn('default_prompt: "Используй $skill-sample по-русски."', rendered_yaml)

    def test_perform_local_install_creates_root_agents_and_testing_module(self) -> None:
        source_dir = self.make_source_skill_dir()
        repo_root = self.root / "repo"
        repo_root.mkdir(parents=True, exist_ok=True)

        with mock.patch.object(ss, "resolve_repo_root", return_value=repo_root.resolve()):
            ss.perform_install(
                source_dir=source_dir,
                install_mode="local",
                requested_locale="ru",
                repo_root=repo_root,
                bootstrap_runner=lambda _: None,
            )

        agents_text = (repo_root / "AGENTS.md").read_text(encoding="utf-8")
        testing_text = (repo_root / ".agents" / ".instructions" / "INSTRUCTIONS_TESTING.md").read_text(
            encoding="utf-8"
        )
        self.assertIn("## Modules", agents_text)
        self.assertIn("@.agents/.instructions/INSTRUCTIONS_TESTING.md", agents_text)
        self.assertIn("# Testing & Refactoring", testing_text)

    def test_perform_local_install_adds_modules_section_when_missing(self) -> None:
        source_dir = self.make_source_skill_dir()
        repo_root = self.root / "repo"
        repo_root.mkdir(parents=True, exist_ok=True)
        (repo_root / "AGENTS.md").write_text(
            "# Repo Guide\n\n## Notes\n\nKeep this content.\n",
            encoding="utf-8",
        )

        with mock.patch.object(ss, "resolve_repo_root", return_value=repo_root.resolve()):
            ss.perform_install(
                source_dir=source_dir,
                install_mode="local",
                requested_locale="ru",
                repo_root=repo_root,
                bootstrap_runner=lambda _: None,
            )

        agents_text = (repo_root / "AGENTS.md").read_text(encoding="utf-8")
        self.assertIn("## Notes", agents_text)
        self.assertIn("Keep this content.", agents_text)
        self.assertIn("\n## Modules\n\n@.agents/.instructions/INSTRUCTIONS_TESTING.md\n", agents_text)

    def test_perform_local_install_appends_testing_ref_to_existing_modules(self) -> None:
        source_dir = self.make_source_skill_dir()
        repo_root = self.root / "repo"
        repo_root.mkdir(parents=True, exist_ok=True)
        (repo_root / "AGENTS.md").write_text(
            "# Repo Guide\n\n## Modules\n\n@foo/bar.md\n\n## Notes\n\nKeep this content.\n",
            encoding="utf-8",
        )

        with mock.patch.object(ss, "resolve_repo_root", return_value=repo_root.resolve()):
            ss.perform_install(
                source_dir=source_dir,
                install_mode="local",
                requested_locale="ru",
                repo_root=repo_root,
                bootstrap_runner=lambda _: None,
            )
            ss.perform_install(
                source_dir=source_dir,
                install_mode="local",
                requested_locale="ru",
                repo_root=repo_root,
                bootstrap_runner=lambda _: None,
            )

        agents_text = (repo_root / "AGENTS.md").read_text(encoding="utf-8")
        self.assertIn("@foo/bar.md", agents_text)
        self.assertIn("@.agents/.instructions/INSTRUCTIONS_TESTING.md", agents_text)
        self.assertEqual(agents_text.count("@.agents/.instructions/INSTRUCTIONS_TESTING.md"), 1)
        self.assertIn("## Notes", agents_text)

    def test_perform_global_install_requires_trigger_module_include(self) -> None:
        source_dir = self.make_source_skill_dir()
        self.make_global_agents_runtime(include_trigger_module=False)

        with self.assertRaises(ss.SetupError) as exc:
            ss.perform_install(
                source_dir=source_dir,
                install_mode="global",
                requested_locale="ru",
                bootstrap_runner=lambda _: None,
            )

        self.assertIn("INSTRUCTIONS_SKILL_TRIGGERS.md", str(exc.exception))
        self.assertIn("AGENTS.md", str(exc.exception))


if __name__ == "__main__":
    unittest.main()
