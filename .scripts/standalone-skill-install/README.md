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

- `setup.sh` — thin shell wrapper for the canonical Python entrypoint
- `setup_support.py` — canonical helper library
- `setup_main.py` — generic CLI entrypoint used by vendored copies

## Minimal Skill Repo Layout

```text
skill-repo/
├── setup.sh
├── scripts/
│   ├── setup.sh
│   ├── setup_main.py
│   └── setup_support.py
├── SKILL.md
├── agents/openai.yaml
└── locales/metadata.json
```

## Vendoring Rules

- Standalone skill repos should vendor the full contents of this directory into
  their own `scripts/` directory.
- Do not create a runtime dependency on this repo from standalone skills.
- When the canonical helper changes, update vendored copies in downstream skill
  repos deliberately.

## Usage

- `./scripts/setup.sh global --locale <mode>`
- `./scripts/setup.sh local /path/to/repo --locale <mode>`

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

## Adding Localization To A Skill

To make a standalone skill compatible with localized installs through this
helper:

1. Keep the source repo in a stable base language, typically English, with a
   normal `SKILL.md` and `agents/openai.yaml`.
2. Add `locales/metadata.json` and define both `en` and `ru` entries with the
   required keys from the localization contract.
3. Put locale-specific trigger phrases into each locale's `triggers` list. The
   helper will render the selected locale for single-language installs and will
   build an ordered deduplicated trigger union for `en-ru` and `ru-en`.
4. Install with `./scripts/setup.sh global --locale <mode>` or
   `./scripts/setup.sh local /path/to/repo --locale <mode>` and verify the
   rendered installed copy, not the source checkout.

Minimal example:

```json
{
  "locales": {
    "en": {
      "description": "English description",
      "display_name": "skill-example",
      "short_description": "Short English summary",
      "default_prompt": "Use $skill-example in English.",
      "local_prefix": "[local] ",
      "triggers": ["example skill", "english trigger"]
    },
    "ru": {
      "description": "Русское описание",
      "display_name": "skill-example",
      "short_description": "Короткое русское описание",
      "default_prompt": "Используй $skill-example по-русски.",
      "local_prefix": "[локально] ",
      "triggers": ["пример скилла", "русский триггер"]
    }
  }
}
```

The shared helper owns the install-time behavior. Skill repos only provide the
localized metadata catalog.
