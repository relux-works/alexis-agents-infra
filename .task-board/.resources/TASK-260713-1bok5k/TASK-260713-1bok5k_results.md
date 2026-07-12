# Provider-specific primary session policy

- Added independent `[agents.claude.primary_session]` model policy with ancestor provenance, native `--model` launch injection, and explicit CLI precedence.
- Preserved the existing Codex model, reasoning-effort, yolo, MCP, and task-board spawn-ceiling contracts.
- Added Claude-only safe setup mutation/clear behavior, print-config and doctor diagnostics, docs, and Go coverage.
- Rolled out `claude-opus-4-6` to relux-agents-infra and skill-project-management while preserving Codex `gpt-5.6-terra` / `xhigh` / `yolo=true`.

Validation:
- `go test ./...`
- `go vet ./...`
- `go build .`
- Claude and Codex `--print-config` plus `doctor local` smokes for both target repos.

Logs: `.temp/TASK-260713-1bok5k/`