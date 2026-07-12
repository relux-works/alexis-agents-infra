# Review Verdict: Changes Requested

Route: `to-dev`.

## Finding

P2 — Reject a local target whose primary-session path is the ignored global path. `prepareCodexPrimarySessionSetup` writes `layout.AgentsDir/.configs/project-config.toml` without checking whether it equals `$HOME/.agents/.configs/project-config.toml` (project_config_setup.go:65-103). The shared resolver deliberately excludes that exact path from project discovery (project_config.go:70-74; codex_launch.go:139-146).

Reproducer: set `HOME` and the local setup target to the same temporary directory, then run `agents-infra setup local TARGET --source-dir REPO --codex-primary-model gpt-review-home`. Setup exits 0 and reports that it updated `TARGET/.agents/.configs/project-config.toml`. With the same HOME, `doctor local TARGET` reports model source `native`, and `codex --print-config` reports no project configs, absent project model, and no model argv. Evidence: `.temp/TASK-260713-4ihi4q/reviewer-home-collision.log`.

This violates the safe explicit setup contract: a successful persisted choice must be consumed by the project resolver, while the canonical discovery rule says the global runtime path must never count as project opt-in.

## Required Rework

- Before sync or mutation, reject a set or clear request when the target project-config path is path-equivalent to the ignored global project-config path. Return the target path and `agents.codex.primary_session` field in the error. Use path comparison robust enough for supported platforms and aliases.
- Add a focused test with `HOME == target` proving the command fails before side effects and does not create or change the global project-config file. Cover both set and clear semantics as appropriate.
- Re-run focused setup tests, full tests, race, vet, format, tidy diff, and the documented target-repository smoke.

## Passing Evidence

All other reviewed setup behavior matches the contract: omitted-field merge, explicit false, clear-only target-table edits, comments/MCP/unrelated-table preservation, invalid-input byte preservation, guarded atomic replacement, global-mode rejection, source-template exclusion, and both target-repository smokes. Independent reviewer runs passed focused tests, full `go test ./...`, `go test -race ./...`, `go vet`, `gofmt`, `go mod tidy -diff`, `git diff --check`, and `task-board validate`.