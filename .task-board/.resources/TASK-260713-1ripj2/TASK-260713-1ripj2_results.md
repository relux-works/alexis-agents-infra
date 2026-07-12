# TASK-260713-1ripj2 — Alias cleanup results

## Source ownership

- Active alias owner: `/Users/alexis/.zshrc:134`.
- The file is a regular user-authored zsh startup file, not a symlink and not an installed agents-infra runtime copy.
- No separate tracked alias definition was found in the inspected home/source roots or the personal source repository.
- Shared tracked documentation owner: `.instructions/INSTRUCTIONS_TOOLS.md` in `relux-agents-infra`.

## Changes

- Changed `codexD='agents-infra codex -d'` to `codexD='agents-infra codex'` in `/Users/alexis/.zshrc`.
- Updated `.instructions/INSTRUCTIONS_TOOLS.md` so `codexD` is described as the normal wrapped launcher and `agents-infra codex -d` is the explicit ad-hoc full-trust escape hatch.
- Added the ownership and policy transition to `LOGBOOK.md`.
- Did not edit `~/.agents`, `~/.claude`, `~/.codex`, or any project-installed runtime copy.

## Escape hatches and no-op decisions

- Fresh zsh still resolves raw `codex` to `/opt/homebrew/bin/codex`.
- The existing explicit danger entry point remains `agents-infra codex -d`; no new alias was invented.
- No `codex-raw`, `codex-yolo`, or `codexY` shell entry point existed to preserve.
- There is no tracked alias-definition file to patch; the only tracked change for alias ownership is the shared documentation correction. This is the source-level no-op reason.
- No new automated test file was added because the changed behavior belongs to the personal shell startup file; the acceptance test is the real fresh-login-zsh smoke below.

## Exact smoke commands

Syntax and lookup:

```bash
zsh -n "$HOME/.zshrc"
zsh -lic 'alias codexD; whence -v codexD; whence -va codex 2>/dev/null | sed -n "1,3p"'
```

Result:

```text
codexD='agents-infra codex'
codexD is an alias for agents-infra codex
codex is /opt/homebrew/bin/codex
```

The following command was run once from each target root:

```bash
zsh -lic 'rendered=$(codexD --print-config) || exit $?; count=$(print -r -- "$rendered" | command rg -F -c -- "--dangerously-bypass-approvals-and-sandbox"); [[ "$count" == 1 ]] || { print -u2 -- "danger-count=$count"; exit 1; }; print -r -- "$rendered" | command rg -A 5 "^(cwd:|  yolo_mode:|wrapper_expansions:|codex_args:)"; print -- "danger-count=$count"'
```

Target roots:

```text
/Users/alexis/src/relux-works/relux-agents-infra
/Users/alexis/src/relux-works/skill-project-management
```

Both runs reported:

- `codexD` reached `agents-infra codex --print-config` from the caller root.
- `yolo_mode.effective_value: true` came from that root's `.agents/.configs/project-config.toml`.
- `wrapper_expansions` was `(none)`, proving the alias supplied no danger argument.
- Final `codex_args` contained exactly one `--dangerously-bypass-approvals-and-sandbox`.

Explicit-danger de-duplication was checked with:

```bash
zsh -lic 'rendered=$(agents-infra codex -d --print-config) || exit $?; argv=$(print -r -- "$rendered" | sed -n "/^codex_args:/,$ p"); count=$(print -r -- "$argv" | command rg -F -c -- "--dangerously-bypass-approvals-and-sandbox"); [[ "$count" == 1 ]] || { print -u2 -- "argv-danger-count=$count"; exit 1; }; print -- "argv-danger-count=$count"'
```

Result: `argv-danger-count=1`.

## Validation

Run from `tools/agents-infra` unless noted:

| Command | Result |
| --- | --- |
| `go test -count=1 ./...` | PASS; both packages passed |
| `go vet ./...` | PASS |
| `gofmt -l .` | PASS; no files listed |
| `go build ./...` | PASS |
| `git diff --check` (repository root) | PASS |
| `task-board validate` (repository root) | PASS; no issues found |

The worktree already contained sibling primary-session implementation changes. They were preserved and were not modified as part of this alias task.
