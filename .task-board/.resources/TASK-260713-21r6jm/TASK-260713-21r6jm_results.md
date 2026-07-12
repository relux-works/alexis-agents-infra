# TASK-260713-21r6jm — Expose primary session diagnostics

## Implementation

- Extended `CodexLaunchPlan` with invocation-level primary-session resolution evidence for model, reasoning effort, and yolo mode.
- The launch resolver now produces both generated argv and the matching diagnostic decision in one precedence path.
- `RenderCodexLaunchPlan` reports every discovered project config path, effective values, effective/project sources, applied versus explicit CLI/profile suppression, wrapper yolo expansion, and final `codex_args`.
- `Doctor` reuses `loadCompositeProjectConfig`, returns errors instead of dropping invalid configuration, and preserves composed MCP plus config-shadowing diagnostics.
- `doctor local` emits the required stable snake-case primary fields. Configured `yolo_mode = false` retains its file source; absent false uses source `default`.
- Invalid project TOML emits `codex_primary_config_valid: false`, omits partial primary values, returns nonzero, and includes the exact source path and field.
- Environment-backed bearer-token values are not read or rendered; a regression test sets a sentinel secret and proves it is absent from diagnostics.
- README diagnostics were updated in the source repository.

## Test coverage

Tests cover configured parent/child composition, inherited per-field provenance, explicit model/effort selection, explicit profile suppression, wrapper yolo expansion, configured false, absent defaults, home config exclusion, syntactically malformed TOML, wrong field types, MCP preservation, config shadowing, final argv, and non-launching print-config behavior.

Coverage from the final combined workspace:

- `internal/infra`: 80.5% statements
- `RenderCodexLaunchPlan`: 93.5%
- primary resolution/render helpers: 92.9–100%
- `Doctor`: 85.3%
- `runDoctor`: 89.6%

## Verification

- `gofmt -l`: clean
- `go test ./... -count=1`: pass
- `go test -race ./... -count=1`: pass
- `go vet ./...`: pass
- native macOS build: pass
- `GOOS=linux GOARCH=amd64 go build`: pass
- `GOOS=windows GOARCH=amd64 go build`: pass
- `go mod tidy -diff`: clean
- configured print-config binary smoke: pass without launching Codex
- configured and absent doctor binary smokes: pass with exact source semantics
- invalid doctor binary smoke: expected exit code 1 with exact path/field
- `git diff --check`: clean
- `task-board validate`: pass

## Workspace note

The shared worktree already contained the accepted composed-policy implementation from `TASK-260713-1phuwx` and concurrently received the setup implementation from `TASK-260713-4ihi4q`. Those changes were preserved. Final Go validation was repeated against the integrated shared snapshot after the setup implementation's last source-code write.

No new diagnostics-specific architectural decision or anomaly required a separate logbook entry; the accepted split-policy/precedence entry remains authoritative.
