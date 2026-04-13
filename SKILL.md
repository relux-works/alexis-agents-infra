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

Source repo for the shared agent infrastructure that installs into `~/.agents`, `~/.claude`, and `~/.codex`.

Do not edit `~/.agents` directly when changing shared instructions, configs, or skills.
Work in the source repo, then run `agents-infra setup global` or `./setup.sh`
to sync the installed runtime copy.

Use this repo when you need to:

- update global instructions in `.instructions/`
- add or adjust shared skills in `.skills/`
- change Codex or Claude configuration in `.configs/`
- update the Go CLI in `tools/agents-infra/`
- update symlink/bootstrap logic in `.scripts/setup-symlinks.sh`, `scripts/setup.sh`, `scripts/setup.ps1`, or `setup.sh`
- use `agents-infra setup global|local` to sync and refresh installed links
- maintain the generic `agents-attachments-manifest.json` contract and helper tooling

## Quick start

```bash
cd /path/to/alexis-agents-infra
./setup.sh
.\\setup.ps1

# Canonical interface after bootstrap
agents-infra setup global
agents-infra setup local /path/to/project
agents-infra doctor global
agents-infra doctor local /path/to/project
agents-infra version
```

This repo is setup/configuration infrastructure, not the runtime that launches agent sessions.
`~/.agents` is the installed destination, not the place to author shared changes.

`./setup.sh` and `.\setup.ps1` are bootstrap wrappers: they delegate into the
cross-platform scripts under `scripts/`, build the `agents-infra` binary with
embedded version metadata, install it into the user-local bin dir, write
install-state metadata, and then immediately run `agents-infra setup global`.

For project-local setup, install into the target repo so that:
- `.agents/` holds the actual installed runtime contents
- `.claude/` and `.codex/` are just thin shims/symlinks into `.agents`
- `.local/bin/` exposes helper CLIs for that local setup, including `agents-infra`

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
- `.skills/` — source-managed shared skills versioned in this repo
- `skills/` — external skills/tooling area in installed runtimes; not versioned by this repo

## References

- [README.md](README.md)
- [.instructions/INSTRUCTIONS_ATTACHMENTS.md](.instructions/INSTRUCTIONS_ATTACHMENTS.md)
- [.scripts/agents-attachments](.scripts/agents-attachments)
