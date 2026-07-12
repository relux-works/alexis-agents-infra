# TASK-260713-25bqi7 — Target Primary Session Rollout Evidence

## Outcome

Both target project runtimes now resolve the same primary Codex session:

```toml
[agents.codex.primary_session]
model = 'gpt-5.6-terra'
reasoning_effort = 'xhigh'
yolo_mode = true
```

Target config checksum in both roots: `697e00adfbe526c19c63a82b36bc9c51231fe2df8b9ef966fb4e4a373634cff3`.

## Commands

The accepted source was installed through the canonical source-of-truth flow before configuring project runtimes:

```bash
AGENTS_INFRA_SKIP_LLDB_MCP=1 ./setup.sh

agents-infra setup local "$HOME/src/relux-works/relux-agents-infra" \
  --source-dir "$HOME/.agents" \
  --codex-primary-model gpt-5.6-terra \
  --codex-primary-reasoning-effort xhigh \
  --codex-yolo-mode=true

agents-infra setup local "$HOME/src/relux-works/skill-project-management" \
  --source-dir "$HOME/.agents" \
  --codex-primary-model gpt-5.6-terra \
  --codex-primary-reasoning-effort xhigh \
  --codex-yolo-mode=true
```

Effective state was checked from each project root with:

```bash
agents-infra doctor local "$PWD"
agents-infra codex --print-config
.local/bin/agents-infra setup local "$PWD"
```

The last command is the project-local runtime launcher and was run without profile flags to prove that a subsequent setup keeps `project-config.toml` byte-identical.

## Sanitized Before / After Evidence

| Check | Before | After |
| --- | --- | --- |
| relux-agents-infra project config | absent | `$HOME/src/relux-works/relux-agents-infra/.agents/.configs/project-config.toml` |
| skill-project-management project config | absent | `$HOME/src/relux-works/skill-project-management/.agents/.configs/project-config.toml` |
| model | native | `gpt-5.6-terra`, target-local source |
| reasoning effort | native | `xhigh`, target-local source |
| yolo | `false` from default | `true`, target-local source |
| effective enabled MCP | none | none |
| project-local `.codex/config.toml` | absent | absent; global config remains authoritative |

For both roots, doctor reports:

```text
codex_primary_config_valid: true
codex_primary_model: gpt-5.6-terra
codex_primary_reasoning_effort: xhigh
codex_primary_yolo_mode: true
```

Each `_source` field is the corresponding target-local `.agents/.configs/project-config.toml` path. `agents-infra codex --print-config` reports all three project values as `applied` and renders:

```text
codex_args:
  - "--model"
  - "gpt-5.6-terra"
  - "-c"
  - "model_reasoning_effort=\"xhigh\""
  - "--dangerously-bypass-approvals-and-sandbox"
```

Machine assertions count exactly one native danger flag in each rendered `codex_args` section.

## Preservation Evidence

- relux-agents-infra `task-board.config.json`: unchanged SHA-256 `630796bf453ba8438aeed7cff0632cd0508e08c599ccd62ea5bb7e31432faefc`.
- skill-project-management `task-board.config.json`: unchanged SHA-256 `fb4b7764335191a2673e441c14646c32ddf1edee5a1719224342d8ab00aa7d25`.
- Neither task-board config contains a primary-session, Codex-primary, or yolo field; the sibling `spawn.max_parallel = 20` state remains intact.
- Global MCP registry remained `040d701bf0c28cb073da80e37af3e95bab2b3c01a9d421645235704e2ab1a608`.
- Ancestor project MCP registry remained `76a9e712db7a603dd368e3f03c8020c930dfaed4b914cd7f29398c338ac6305c`.
- Effective enabled MCP composition is `(none)` before and after in both roots.
- A no-profile project-local setup retained each project config byte-for-byte.
- skill-project-management's original tracked `AGENTS.md` is byte-identical in `.agents/.instructions/AGENTS.project.md`; setup added only the three expected generated/source marker lines to the rendered root entrypoint.

The sibling repository had active board agents during rollout. `tools/board-cli/cmd/root.go` and already-dirty spawn/query files changed after the before-state snapshot. Those unrelated concurrent changes were retained; no whole-worktree hash is claimed as attribution evidence. This anomaly is recorded in `LOGBOOK.md`.

## Validation

Run from `tools/agents-infra` after configuration:

```text
go test ./...                  PASS (0.488s / 1.293s)
go vet ./...                   PASS
gofmt -l                       PASS (no output)
go mod tidy -diff              PASS (no diff)
go build -trimpath             PASS
git diff --check               PASS in both repositories
task-board validate            PASS in relux-agents-infra
```

No production behavior was added in this configuration-only task, so no new source test file was required. Existing setup, launch-plan, doctor, and print-config tests passed, and real-root assertions cover the requested rollout behavior.

skill-project-management board validation reports five tracked orphan resource files and one orphan resource directory dated February–April 2026. They predate this rollout, setup did not touch `.task-board`, and they were left unchanged as out-of-scope board debt.

## Handoff

Configuration and evidence are ready for review.
