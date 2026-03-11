---
name: alexis-agents-infra
description: Shared agent infrastructure repo for Claude Code and Codex CLI. Use when updating global agent instructions, skills, symlink setup, tool configs, rules, or the generic agents attachments manifest contract and helper tooling.
triggers:
  - alexis-agents-infra
  - agents infra
  - agent infrastructure
  - shared agent config
  - global agent instructions
  - codex config
  - claude settings
  - setup symlinks
  - agents attachments manifest
  - attachments manifest
  - агентская инфра
  - конфиг агентов
  - настройки codex
  - настройки claude
---

# alexis-agents-infra

Shared source of truth for `~/.agents`, `~/.claude`, and `~/.codex`.

Use this repo when you need to:

- update global instructions in `.instructions/`
- add or adjust shared skills in `.skills/` or `skills/`
- change Codex or Claude configuration in `.configs/`
- update symlink/bootstrap logic in `.scripts/setup-symlinks.sh`
- use `setup.sh` to sync into `~/.agents` and refresh installed links
- maintain the generic `agents-attachments-manifest.json` contract and helper tooling

## Quick start

```bash
cd /path/to/alexis-agents-infra
./setup.sh
```

This repo is setup/configuration infrastructure, not the runtime that launches agent sessions.

## Attachments Contract

Incoming user files are modeled as a generic manifest, not as board-specific state.

- Manifest file name: `agents-attachments-manifest.json`
- Environment variable: `AGENTS_ATTACHMENTS_MANIFEST`
- Default project-local fallback: `.temp/agents-attachments-manifest.json`
- Helper CLI installed from this repo: `agents-attachments`
- Codex bootstrap helper: `agents-attachments materialize`

Runtime responsibilities:

- materialize incoming files to local disk
- write `agents-attachments-manifest.json`
- export `AGENTS_ATTACHMENTS_MANIFEST`
- propagate the same manifest/env into spawned child agents

This repo's responsibilities:

- define the contract in `.instructions/INSTRUCTIONS_ATTACHMENTS.md`
- ship the helper in `.scripts/agents-attachments`
- install/symlink the helper via `.scripts/setup-symlinks.sh`

## Key Paths

- `.instructions/` — global instruction modules
- `.configs/` — Codex/Claude config files
- `.rules/` — Codex rules
- `.scripts/` — setup and helper tooling
- `.skills/`, `skills/` — installable skills

## References

- [README.md](README.md)
- [.instructions/INSTRUCTIONS_ATTACHMENTS.md](.instructions/INSTRUCTIONS_ATTACHMENTS.md)
- [.scripts/agents-attachments](.scripts/agents-attachments)
