# TASK-260713-4ihi4q — Local Primary-Session Setup Results

## Implementation

- Added local setup flags `--codex-primary-model`,
  `--codex-primary-reasoning-effort`, `--codex-yolo-mode=true|false`, and
  `--clear-codex-primary-session`.
- Preserved explicit boolean presence, so supplied `false` differs from an
  omitted yolo flag.
- Validated every discovered ancestor project config before setup side effects;
  invalid TOML and field/type errors report the exact path and field.
- Added focused TOML surgery for `[agents.codex.primary_session]`: existing
  values are replaced in place, missing supplied fields are inserted, omitted
  fields remain untouched, and clear removes only the target table/header and
  its supported field lines.
- Preserved MCP config, unrelated tables, inline/standalone comments, newline
  style, and existing `.codex/config.toml` mode behavior.
- Excluded `.configs/project-config.toml` from source sync so setup without
  primary flags keeps project state byte-identical, even if a source-side file
  appears accidentally.
- Added same-directory temporary-file write, file sync, original mode
  preservation, and a concurrent-change guard. Final replacement is now
  platform-backed: POSIX `os.Rename` on supported Unix targets,
  [`github.com/natefinch/atomic.ReplaceFile`](https://pkg.go.dev/github.com/natefinch/atomic#ReplaceFile)
  (`MoveFileExW` with replace/write-through flags) on Windows, and fail-closed
  behavior on targets without a documented atomic primitive.
- Added final-replacement failure injection plus Windows-only behavioral tests
  for successful replacement and delete-locked failure preservation.
- Rejected primary-session policy flags in global mode and rejected clear plus
  set combinations before sync or mutation.
- Rejected set and clear requests when the local target resolves to the ignored
  global `~/.agents/.configs/project-config.toml` path, including symlink
  aliases, before setup creates or changes any destination path.
- Strengthened path identity checks for missing final files by comparing the
  nearest existing filesystem ancestor plus the unresolved suffix; Windows
  comparisons remain case-insensitive.
- Fixed the documented path-first CLI form: Go flag parsing previously stopped
  at positional `PROJECT_DIR`, so trailing flags could remain unparsed.

## Files

- `tools/agents-infra/internal/infra/project_config_setup.go`
- `tools/agents-infra/internal/infra/project_config_setup_test.go`
- `tools/agents-infra/internal/infra/project_config_replace_posix.go`
- `tools/agents-infra/internal/infra/project_config_replace_windows.go`
- `tools/agents-infra/internal/infra/project_config_replace_unsupported.go`
- `tools/agents-infra/internal/infra/project_config_replace_windows_test.go`
- `tools/agents-infra/internal/infra/infra.go`
- `tools/agents-infra/main.go`
- `tools/agents-infra/setup_test.go`
- `tools/agents-infra/go.mod`
- `tools/agents-infra/go.sum`
- `README.md`
- `LOGBOOK.md`

The shared checkout also contains accepted prerequisite launch-policy changes
and concurrent diagnostics changes owned by their separate board tasks. Those
changes were preserved and validated together; this task did not reset, stage,
or commit any file.

## Verification

- `go test ./... -count=1`: pass.
- `go test -race ./... -count=1`: pass.
- `go vet ./...`: pass.
- `go test ./... -coverprofile=...`: pass, `80.6%` statement
  coverage.
- `go build ./...`: pass on host `darwin/arm64`.
- `GOOS=linux GOARCH=amd64 go build ./...`: pass.
- `GOOS=windows GOARCH=amd64 go build ./...`: pass.
- `GOOS=windows GOARCH=arm64 go build ./...`: pass.
- Windows amd64 and arm64 `go test -c ./internal/infra`: pass, including the
  Windows-only success and delete-lock failure-preservation tests.
- Focused final-replacement failure test: pass; staged bytes were complete,
  the existing destination stayed byte-identical, and the temporary file was
  removed after the reported error.
- `go mod verify`: pass.
- `go mod tidy -diff`: clean.
- `gofmt -l .`: clean.
- `git diff --check`: clean.
- `task-board validate`: pass.
- Negative ownership assertion: no spawn ceiling/rank/manifest policy in
  `tools/agents-infra`.
- Template assertion: no shared `.configs/project-config.toml` exists in the
  source repository.

Final command logs are under `.temp/TASK-260713-4ihi4q/` and are attached as a
task-scoped archive.

## Disposable Target-Repository Smokes

Used detached worktrees under `.temp/TASK-260713-4ihi4q/smoke-rework2/`, then
removed them after evidence capture:

| Target | Baseline | Result |
| --- | --- | --- |
| `relux-agents-infra` | `072e95c1c78c997666d8b787f4146361242b134a` | Windows-atomic rework rerun: set, no-flag preserve, partial update, doctor/print-config, clear passed |
| `skill-project-management` | `6317d70493007f97917783ff8c94b2a0a2999b59` | Windows-atomic rework rerun: set, no-flag preserve, partial update, doctor/print-config, clear passed |

For both targets:

- Initial set produced model `gpt-5.6-terra`, effort `xhigh`, and explicit
  `yolo_mode = false` with file provenance in doctor and print-config.
- Re-running setup without primary flags retained each target's exact
  project-config hash (`2b6806fa...` for relux-agents-infra and
  `28d3edb9...` for skill-project-management).
- Model-only update to `gpt-5.6-sol` retained effort `xhigh` and yolo `false`.
- Clear returned doctor provenance to model/effort `native` and yolo `default`.
- A separate `HOME == target` smoke returned a path/field error and created no
  `.agents` destination.
- Source status hashes were identical before and after both disposable smokes;
  unrelated dirty work in both repositories was preserved.

## Review Notes

- The setup mutator intentionally requires an explicit
  `[agents.codex.primary_session]` table when updating an already-present
  primary policy. This prevents a comment-destroying rewrite of semantically
  equivalent deeply nested inline-table syntax.
- A local Windows runtime was unavailable. `brew install --cask wine-stable`
  was stopped when its dependency requested sudo; the partial cask record was
  removed. A task-local archive was then killed by macOS Gatekeeper, and no
  protection was bypassed. Final state confirms neither Wine nor GStreamer is
  installed. The review requirement is covered by the Windows-only behavioral
  tests rather than a claimed local Windows execution.
- No unresolved blocker or forced-fit condition remains.
