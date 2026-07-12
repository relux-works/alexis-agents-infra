# TASK-260713-3mlifc — Primary Session Target Validation

Status: ready for review

## Outcome

The implemented agents-infra primary-session workflow passed the independent
disposable-copy validation matrix in both target repositories. No reproducible
primary-session behavior defect was found.

| Target copy | Assertions | Result |
| --- | ---: | --- |
| `relux-agents-infra` | 117 | PASS |
| `skill-project-management` | 117 | PASS |

The tests used the current dirty implementation state from both source working
trees, copied before execution. Production behavior was not patched.

## Isolation and Build

Source revisions at copy time:

- `relux-agents-infra`: `072e95c1c78c997666d8b787f4146361242b134a`
- `skill-project-management`: `6317d70493007f97917783ff8c94b2a0a2999b59`

Disposable copies:

- `/Users/alexis/src/relux-works/relux-agents-infra/.temp/TASK-260713-3mlifc/copies/relux-agents-infra`
- `/Users/alexis/src/relux-works/relux-agents-infra/.temp/TASK-260713-3mlifc/copies/skill-project-management`

The launcher used `/tmp/TASK-260713-3mlifc-primary-session` as a symlinked
discovery root so the source checkout's outer `.agents` configuration could not
contaminate ancestor resolution. `HOME` was task-local for every invocation.

Exact preparation commands:

```bash
rsync -a --exclude='.git/' --exclude='.temp/' --exclude='.task-board/' \
  /Users/alexis/src/relux-works/relux-agents-infra/ \
  .temp/TASK-260713-3mlifc/copies/relux-agents-infra/

rsync -a --exclude='.git/' --exclude='.temp/' --exclude='.task-board/' \
  /Users/alexis/src/relux-works/skill-project-management/ \
  .temp/TASK-260713-3mlifc/copies/skill-project-management/

cd .temp/TASK-260713-3mlifc/copies/relux-agents-infra/tools/agents-infra
go build -mod=readonly -trimpath \
  -o /Users/alexis/src/relux-works/relux-agents-infra/.temp/TASK-260713-3mlifc/bin/agents-infra .
```

Exact matrix entrypoints:

```bash
.temp/TASK-260713-3mlifc/validate-primary-session.zsh \
  relux-agents-infra relux-agents-infra

.temp/TASK-260713-3mlifc/validate-primary-session.zsh \
  skill-project-management skill-project-management
```

The attached per-target transcripts contain every fully expanded command,
stdout/stderr, and exit status. The attached harness contains the corresponding
assertions and controlled fixture bodies.

## Behavioral Evidence Per Copy

The same matrix ran against both copies.

### Local setup

- A full `setup local` without primary flags preserved the controlled
  `project-config.toml` byte-for-byte while syncing the local runtime.
- Set created model, reasoning, and a real boolean yolo field while preserving
  `[mcp]`, comments, and an unrelated table.
- Partial update preserved the omitted model, changed reasoning, and retained
  explicit `yolo_mode = false` presence.
- Clear removed only `[agents.codex.primary_session]`.
- Invalid child TOML made setup return nonzero and left the invalid bytes
  unchanged, proving fail-fast atomicity.

Representative expanded commands, with the target path changed per copy:

```bash
env HOME=/tmp/TASK-260713-3mlifc-primary-session/home/<target> \
  /tmp/TASK-260713-3mlifc-primary-session/bin/agents-infra \
  setup local /tmp/TASK-260713-3mlifc-primary-session/copies/<target> \
  --source-dir /tmp/TASK-260713-3mlifc-primary-session/runtime-source

env HOME=/tmp/TASK-260713-3mlifc-primary-session/home/<target> \
  /tmp/TASK-260713-3mlifc-primary-session/bin/agents-infra \
  setup local /tmp/TASK-260713-3mlifc-primary-session/copies/<target> \
  --source-dir /tmp/TASK-260713-3mlifc-primary-session/runtime-source \
  --no-sync --codex-primary-model primary-parent-<target> \
  --codex-primary-reasoning-effort high --codex-yolo-mode=true

env HOME=/tmp/TASK-260713-3mlifc-primary-session/home/<target> \
  /tmp/TASK-260713-3mlifc-primary-session/bin/agents-infra \
  setup local /tmp/TASK-260713-3mlifc-primary-session/copies/<target> \
  --source-dir /tmp/TASK-260713-3mlifc-primary-session/runtime-source \
  --no-sync --clear-codex-primary-session
```

### Ancestor composition, MCP, and diagnostics

The parent fixture supplied:

- model `primary-parent-<target>`;
- reasoning `high`;
- `yolo_mode = true`;
- MCP `figma`.

The child fixture supplied:

- reasoning `medium`;
- explicit `yolo_mode = false`;
- MCP `safari`;
- no model.

`codex --print-config` and `doctor local CHILD` proved:

- model inherited from the parent with the parent file as provenance;
- reasoning selected from the nearer child with the child file as provenance;
- child false masked inherited true and retained the child source;
- MCP composition was ordered `figma,safari`, and final argv contained both
  MCP overrides;
- doctor emitted the stable validity/value/source fields.

### CLI/profile and yolo rules

- Parent project yolo emitted exactly one native danger flag.
- Child false emitted no danger flag.
- Project true plus `-d`, `--danger`, `--yolo`, and the native flag still
  normalized to exactly one final native flag.
- Explicit wrapper danger over child false emitted exactly one flag and showed
  the wrapper source.
- `--profile fast` suppressed project model and reasoning independently but did
  not suppress yolo.
- Exact top-level `-c model=...` and
  `-c model_reasoning_effort=...` won over project values even with a profile.
- Equal duplicate model/reasoning selections normalized to one override.
- Conflicting model and reasoning selections both failed before render/exec.

### Invalid config and native no-config behavior

The invalid child fixture used quoted `yolo_mode = "false"`.

- `codex --print-config`, `doctor local`, actual `codex` launch, and setup all
  returned nonzero with the exact config path and
  `agents.codex.primary_session.yolo_mode` field.
- Doctor reported `codex_primary_config_valid: false` without partial values.
- A task-local fake Codex sentinel proved the invalid launch failed before exec.

With `.agents/.configs/project-config.toml` absent and the target's real
`task-board.config.json` still present:

- model and reasoning remained Codex-native;
- yolo was false from `default`;
- no generated model, reasoning, MCP, or danger argument appeared;
- fake Codex received exactly `exec`, `native`.

An additional nested project contained only this task-board policy:

```json
{
  "mode": "local",
  "local": {"board_dir": ".task-board"},
  "spawn": {
    "max_parallel": 7,
    "ceilings": {
      "codex": {
        "model": "task-board-only-<target>",
        "reasoning_effort": "max"
      }
    }
  }
}
```

The primary launcher still reported no discovered project TOML, kept both
dimensions native, emitted no dangerous flag, did not print the task-board path
or ceiling model, and left the JSON byte-identical. This directly proves the
task-board spawn ceiling is not a primary-session source.

Both root and child `.agents/.configs/project-config.toml` fixtures were checked
to contain no spawn/ceiling member. Source-level search found agents-infra
references to `task-board.config.json` only in install/scrub exclusion logic,
not policy loading. Task-board primary-session strings occur only in legacy
rejection diagnostics and negative tests, not an accepted policy structure.

## Go Tests, Vet, Build, and Formatting

agents-infra disposable copy:

```bash
cd .temp/TASK-260713-3mlifc/copies/relux-agents-infra/tools/agents-infra
go test -mod=readonly ./...
go vet -mod=readonly ./...
go build -mod=readonly ./...
rg --files -g '*.go' -0 | xargs -0 gofmt -l
```

Result: all packages passed (`1.875s` and `4.449s` in the recorded run), vet and
build passed, and `gofmt -l` was empty.

skill-project-management disposable copy ran the following in each of
`pkg/board`, `pkg/remoteconfig`, `tools/board-cli`, `tools/board-server`, and
`tools/board-tui`:

```bash
go test -mod=readonly ./...
go vet -mod=readonly ./...
go build -mod=readonly ./...
```

Result: all five modules passed. `tools/board-cli/cmd` passed in `15.470s` and
`tools/board-server/internal/api` passed in `18.981s`; full package output is in
the evidence archive. The 42 changed/new Go files in the current implementation
diff produced empty scoped `gofmt -l` output.

Whole-repository `gofmt -l` also exposed unrelated baseline drift in three
unchanged files:

- `pkg/board/consistency_test.go`
- `pkg/board/markdown/progress.go`
- `pkg/board/board.go`

Those files are outside the current implementation diff and were not patched,
as this task explicitly forbids production fixes. `go vet` and `go build` still
pass for every module.

## Source Checkout and Concurrency Evidence

Every mutating validation command targeted only the task-scoped disposable
copies through `/tmp/TASK-260713-3mlifc-primary-session`. The harness restored
each copy's original project config and proved each copy's
`task-board.config.json` byte-identical before/after.

The initial full-source checksum comparison was intentionally retained and did
not compare equal because other tracked workers changed the shared source
checkouts during this run:

- At `2026-07-13T15:56:19+0300`, both source `task-board.config.json` files
  acquired the Terra/xhigh spawn ceiling owned by active
  `TASK-260713-2qfcpb` (Configure target spawn ceilings).
- The infra source instructions and logbook changed under
  `TASK-260713-1ripj2` (Remove yolo from Codex aliases), which reached
  `to-review` during this run.
- Prebuilt task-board binaries in the second source checkout were refreshed by
  concurrent work shortly after the initial snapshot.

These concurrent deltas are recorded in the evidence archive and were neither
reverted nor attributed to this validation. A new post-concurrency source
manifest and task-board config hash were taken before handoff; the handoff
comparison passed for the complete infra checkout. In skill-project-management,
only concurrently rebuilt `tools/board-cli/task-board` and
`tools/board-tui/task-board-tui` binaries changed; the source-only manifest and
`task-board.config.json` hash remained stable. `git diff --check` passed in both
source checkouts. Board-resource/status writes are excluded from these manifests.

## Evidence Files

- `TASK-260713-3mlifc_results.md` — this report.
- `TASK-260713-3mlifc_validation-harness.zsh` — executable smoke matrix.
- `TASK-260713-3mlifc_relux-agents-infra-transcript.log` — exact expanded
  commands, outputs, and exit codes for the first copy.
- `TASK-260713-3mlifc_skill-project-management-transcript.log` — exact expanded
  commands, outputs, and exit codes for the second copy.
- `TASK-260713-3mlifc_evidence.tar.gz` — assertion logs, Go logs, ownership
  search, manifests, recovery notes, and command-level smoke logs.
