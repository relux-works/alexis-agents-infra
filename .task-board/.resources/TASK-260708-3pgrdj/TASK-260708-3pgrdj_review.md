# TASK-260708-3pgrdj Review

Verdict: done.

## Findings

No blocking findings.

## Evidence

- Shared registry defines `figma`, `lldb`, and `safari`; Safari uses `/Applications/Safari Technology Preview.app/Contents/MacOS/safaridriver` with `args = ["--mcp"]`: `.configs/codex-mcp-servers.toml:1`.
- Project instructions list `figma`, `lldb`, and `safari` as project-local MCP definitions and state that agents-infra managed MCP servers are not globally enabled: `.instructions/INSTRUCTIONS_TOOLS.md:34`.
- README documents project-local opt-in, stdio `command`/`args`, Safari Technology Preview prerequisites, and external agents setting: `README.md:281`.
- Repo `SKILL.md` documents the same MCP policy and Safari prerequisites: `SKILL.md:120`.
- Launcher supports URL and stdio MCP definitions, validates mutually exclusive `url`/`command`, and emits `mcp_servers.<name>.command` plus `args` config overrides: `tools/agents-infra/internal/infra/codex_launch.go:40`.
- Tests cover stdio MCP servers, Safari opt-in without registry-only global enablement, `--print-config` output, and setup registry sync: `tools/agents-infra/internal/infra/codex_launch_test.go:94`, `tools/agents-infra/main_test.go:43`, `tools/agents-infra/internal/infra/infra_test.go:328`.

## Verification Run By Reviewer

- `cd tools/agents-infra && go test -count=1 ./...` passed; log `.temp/TASK-260708-3pgrdj/review-go-test-01.log`.
- `cd tools/agents-infra && go vet ./...` passed; log `.temp/TASK-260708-3pgrdj/review-go-vet-01.log`.
- `git diff --check` passed; log `.temp/TASK-260708-3pgrdj/review-git-diff-check-01.log`.
- `gofmt -l` over touched Go files returned no files; log `.temp/TASK-260708-3pgrdj/review-gofmt-check-01.log`.
- `./setup.sh` passed and installed `/Users/alexis/.local/bin/agents-infra`; log `.temp/TASK-260708-3pgrdj/review-setup-01.log`.
- Installed `/Users/alexis/.codex/AGENTS.md` contains the supported MCP list with `figma`, `lldb`, and `safari`.
- Installed-runtime smoke from `.temp/TASK-260708-3pgrdj/review-smoke` with `enabled_servers = ["safari"]` emitted:
  - `mcp_servers.safari.command="/Applications/Safari Technology Preview.app/Contents/MacOS/safaridriver"`
  - `mcp_servers.safari.args=["--mcp"]`
  - log `.temp/TASK-260708-3pgrdj/review-print-config-safari-01.log`.
- `rg -n "safari|enabled_servers|mcp_servers" /Users/alexis/.codex/config.toml` returned no matches, so the review did not find a global Codex MCP enablement.

## Notes

No logbook entry was needed from review; no non-obvious regression, anomaly, or reusable architecture decision was discovered.
