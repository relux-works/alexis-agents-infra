# TASK-260708-3pgrdj Continuation Guide

Previous developer run `RUN-260708-4d98e6` was cancelled after partial edits because it stopped making progress before verification and handoff.

Continue from the current dirty worktree. Do not revert unrelated existing changes, especially the LLDB MCP setup changes already present before this task.

Focus:
- Inspect current Safari MCP changes.
- Fix any incorrect test expectations or documentation wording.
- Run `gofmt` on touched Go files.
- Run `go test ./...` from `tools/agents-infra`.
- Run `git diff --check`.
- Run `./setup.sh` to sync installed runtime.
- Smoke `agents-infra codex --print-config` from a temp project with `enabled_servers = ["safari"]`.
- Verify `/Users/alexis/.codex/AGENTS.md` lists `figma`, `lldb`, and `safari`.
- Check all applicable checklist items, attach a task-scoped outcome resource, and move the task to `to-review`.

Avoid long re-research. The official WebKit source is already attached in the task precondition resource.
