## Status
done

## Assigned To
codex

## Created
2026-06-15T10:29:41Z

## Last Update
2026-06-15T10:46:41Z

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
(empty)

## Notes
Starting investigation. User reports that opening a project with no MCP project TOML appears to print MCP startup attempts on stdout. Will inspect config rendering and Codex launcher behavior, then reproduce in a temp no-config project.
Reproduced: agents-infra codex --print-config from a no-config project under ~ still enabled figma because ancestor scan treated ~/.agents/.configs/project-config.toml as a project-local config. /tmp reproduction correctly had enabled_mcp=(none). Fixing launcher to skip home-level ~/.agents project-config while preserving ~/.agents MCP registry as definitions only.
Fixed and verified. Root cause: agents-infra codex ancestor scan treated ~/.agents/.configs/project-config.toml as project-local opt-in for every project under $HOME. Fix: BuildCodexLaunchPlan now skips the home-level ~/.agents project-config path while still using ~/.agents/.configs/codex-mcp-servers.toml as the global definitions registry. Setup global now removes stale ~/.agents/.configs/project-config.toml. scripts/setup.sh now installs the binary through temp file + mv to avoid the observed SIGKILL-after-overwrite executable state. Verification logs: .temp/BUG-260615-9u5jvq-go-test-04.log, .temp/BUG-260615-9u5jvq-print-config-repo-04.log, .temp/BUG-260615-9u5jvq-print-config-with-mcp-01.log, .temp/BUG-260615-9u5jvq-setup-02.log.
Review pass started after user request. Checking changed launcher/setup logic, tests, docs wording, and installed-runtime behavior.
Review result: found and fixed a regression-test gap. The original new test used start=t.TempDir(), outside fake home, so it would not have exercised ~/.agents ancestor scanning. Updated it to start under fake home/project/subdir. Re-ran go test ./..., git diff --check, no-config print-config, opt-in print-config, and codex mcp list; all pass. No remaining review findings.

## Precondition Resources
(none)

## Outcome Resources
- [BUG-260615-9u5jvq_mcp-root-cause.md](file://BUG-260615-9u5jvq/BUG-260615-9u5jvq_mcp-root-cause.md) — MCP no-config reproduction and fix summary
