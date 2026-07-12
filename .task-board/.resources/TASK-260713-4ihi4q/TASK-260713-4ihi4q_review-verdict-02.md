# Review Verdict: Changes Requested (Cycle 2)

Route: `to-dev`.

## Finding

P2 — The supported Windows setup path does not satisfy the atomic-write acceptance criterion.

`writeProjectConfigAtomically` finishes with `os.Rename(temporaryPath, path)` at `tools/agents-infra/internal/infra/project_config_setup.go:529`. The Go API contract explicitly states that `os.Rename` is not atomic on non-Unix platforms, even within one directory. This repository supports Windows through `scripts/setup.ps1`, and README documents the same `agents-infra setup local` interface there. A `GOOS=windows` cross-build proves compilation only; it cannot establish atomic replacement behavior.

The implementation therefore cannot guarantee atomic project-config replacement on one supported platform. This is ordinary implementation rework, not a stop-the-line blocker.

## Required Rework

- Put final project-config replacement behind a platform-aware helper whose underlying primitive or library has a documented atomic-replace guarantee on each supported platform, including Windows.
- Preserve the existing target byte-for-byte on every reported failure and retain the current guarded same-directory temporary-file behavior, mode handling, and cleanup.
- Add Windows-specific replacement coverage or Windows runtime CI evidence. Cross-compilation alone is insufficient for this behavioral contract.
- Re-run focused/full/race/vet/format/tidy/build gates and the documented two-repository setup flow, then hand off to another reviewer cycle.

## Passing Evidence

- The previous HOME/global-path defect is fixed: exact HOME set/clear and a symlink alias fail before sync with target path plus `agents.codex.primary_session`, and no destination is created.
- Independent focused tests, full tests, race tests, vet, gofmt, `go mod tidy -diff`, and host/Linux/Windows builds pass.
- Independent black-box CLI smoke passes field merge, explicit false, comments/MCP/unrelated-table preservation, custom `.codex/config.toml` preservation, byte-identical no-flag setup, doctor/print-config provenance, partial update, and clear.
- A real permission-denied temporary-file creation returns a path/field error, leaves the original config hash unchanged, and leaves no temporary file.
- Global primary flags are rejected before destination creation.
- Developer evidence for both target repositories is present and internally consistent; the no-flag hashes match in both flows.
- `git diff --check`, `task-board validate`, the negative ownership check, and the no-source-template check pass.

Reviewer logs are attached separately as `TASK-260713-4ihi4q_review-02-logs.tar.gz`.
