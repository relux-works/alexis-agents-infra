# Standalone Skill Install Pattern

Canonical install/localization helper for standalone skill repositories under
`~/agents/skills/*`.

This pattern exists for skills that are authored outside `alexis-agents-infra`
but still need a consistent installation contract:

- global install into `${XDG_DATA_HOME:-~/.local/share}/agents/skills/<skill-name>`
- project-local install into `<repo>/.skills/<skill-name>`
- local testing instructions provisioned into `<repo>/.agents/.instructions/INSTRUCTIONS_TESTING.md`
- project-root `AGENTS.md` updated to reference `@.agents/.instructions/INSTRUCTIONS_TESTING.md`
- runtime links in `.claude/skills/` and `.codex/skills/` pointing at the
  installed copy instead of the source checkout
- install-time localization for `SKILL.md` and `agents/openai.yaml`

## Files

- `setup_support.py` — canonical helper library
- `setup_main.py` — generic CLI entrypoint used by vendored copies

## Minimal Skill Repo Layout

```text
skill-repo/
├── setup.sh
├── scripts/
│   ├── setup_main.py
│   └── setup_support.py
├── SKILL.md
├── agents/openai.yaml
└── locales/metadata.json
```

Minimal `setup.sh` wrapper:

```sh
#!/usr/bin/env sh
set -eu
exec python3 "$(dirname "$0")/scripts/setup_main.py" "$@"
```

## Vendoring Rules

- Standalone skill repos should vendor these files into their own `scripts/`
  directory.
- Do not create a runtime dependency on this repo from standalone skills.
- When the canonical helper changes, update vendored copies in downstream skill
  repos deliberately.

## Usage

- `./setup.sh global --locale <mode>`
- `./setup.sh local /path/to/repo --locale <mode>`

After install:

- global mode registers skill triggers in the shared global instructions
- local mode writes `.agents/.instructions/INSTRUCTIONS_TESTING.md`
- local mode updates the repo-root `AGENTS.md` so `Modules` includes
  `@.agents/.instructions/INSTRUCTIONS_TESTING.md`

## Localization Contract

The helper expects `locales/metadata.json` with these keys per locale:

- `description`
- `display_name`
- `short_description`
- `default_prompt`
- `local_prefix`
- `triggers`

Supported locale modes:

- `en`
- `ru`
- `en-ru`
- `ru-en`

For dual modes:

- `description` becomes bilingual in the selected order
- `triggers` become an ordered, deduplicated union in the selected order
- `display_name`, `short_description`, and `default_prompt` use the primary locale
