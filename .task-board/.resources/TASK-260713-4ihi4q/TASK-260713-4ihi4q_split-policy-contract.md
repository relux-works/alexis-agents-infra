# Split Project Session and Board Spawn Policy

- Architecture task: `TASK-260713-190sng` — Design split project session and board spawn policy
- Parent: `STORY-260713-3h3ajh` — Split project session and spawn policies
- Repositories:
  - `/Users/alexis/src/relux-works/relux-agents-infra`
  - `/Users/alexis/src/relux-works/skill-project-management`
- Status: implementation contract for review

## 1. Decision

Project policy is split by runtime ownership:

| Concern | Source of truth | Consumer | Must not own |
| --- | --- | --- | --- |
| Primary Codex model, reasoning effort, and persistent yolo choice | `.agents/.configs/project-config.toml` | `agents-infra codex`, `agents-infra setup local`, `agents-infra doctor local`, `agents-infra codex --print-config` | Spawn ceilings, task-board model ranks, spawn manifests |
| Explicit spawned-agent ceilings for Codex and Claude | `task-board.config.json` | `task-board spawn`, `project_config()`, `models()`, tracked run manifests | Primary-session model, primary reasoning, yolo, agents-infra TOML |

The abandoned design under `task-board.config.json -> agents.*.primary_session`
is not part of the product contract. It was not released and is not a fallback
or precedence source.

`task-board codex` keeps only its existing app-server delivery and
`codex.session.auto_continue` behavior. This story adds no task-board primary
model, primary reasoning, profile, or yolo resolution.

## 2. agents-infra primary-session TOML

### 2.1 Exact schema

```toml
[mcp]
enabled_servers = ["figma"]

[agents.codex.primary_session]
model = "gpt-5.6-terra"
reasoning_effort = "xhigh"
yolo_mode = false
```

The table is optional. When present, it must contain at least one supported
field.

| Field | Type | Required | Structural rule |
| --- | --- | --- | --- |
| `model` | TOML string | No | Trimmed value must be non-empty |
| `reasoning_effort` | TOML string | No | Trimmed value must be non-empty |
| `yolo_mode` | TOML boolean | No | Must be a real boolean; quoted values are invalid |

agents-infra validates shape and types, not task-board model ranks or spawn
compatibility. Codex remains the semantic authority for model availability and
model/effort compatibility.

### 2.2 Discovery and composition

1. Start at the filesystem root and walk to the invocation working directory.
2. At each ancestor, inspect `.agents/.configs/project-config.toml`.
3. Do not treat `~/.agents/.configs/project-config.toml` as project opt-in.
4. Parse each discovered file once and retain its absolute path as provenance.
5. Preserve the existing MCP behavior: enabled servers remain the ordered
   union with per-source provenance.
6. Compose primary-session scalars independently. A nearer present value wins.
7. Boolean presence is explicit: child `yolo_mode = false` overrides inherited
   `true`; omission inherits.
8. An invalid discovered file fails before launch, render, doctor success, or
   setup rewrite. It is never treated as absent.

### 2.3 Launch precedence

Model and reasoning use this order, per dimension:

1. Explicit Codex CLI selection passed through `agents-infra codex`:
   `--model`/`-m`, exact top-level `-c model=...`, or exact top-level
   `-c model_reasoning_effort=...`.
2. Effective project value from `[agents.codex.primary_session]`.
3. Codex-native project/profile/user/system/default resolution. An unconfigured
   dimension is omitted from generated argv.

Additional rules:

- An explicit `--profile` suppresses both project model and project reasoning.
  Explicit model or reasoning values used with the profile still pass through.
- Equal duplicate explicit values normalize to one effective override.
  Conflicting explicit values for one dimension fail before `exec`.
- Project values are rendered as Codex CLI/config overrides only after this
  normalization.
- agents-infra does not read or copy Codex global defaults.

### 2.4 Yolo safety semantics

Yolo is an independent boolean decision:

1. Explicit wrapper aliases `-d`, `--danger`, or `--yolo`, or the explicit
   native dangerous flag, opt the invocation in.
2. Otherwise effective `yolo_mode = true` opts the invocation in.
3. Effective `false` or no configured value emits no dangerous flag.

When enabled, the launch argv contains exactly one
`--dangerously-bypass-approvals-and-sandbox`. `--profile` does not suppress
yolo. Project yolo applies only to invocations through `agents-infra codex`; it
does not propagate to Claude, `task-board spawn`, task-board manifests, or
child-run ceiling policy.

The default is safe: absence means `false`. A child project can explicitly
mask an inherited unsafe choice with `yolo_mode = false`.

### 2.5 Setup contract

The supported mutation surface is:

```bash
agents-infra setup local /path/to/project \
  --codex-primary-model gpt-5.6-terra \
  --codex-primary-reasoning-effort xhigh \
  --codex-yolo-mode=false

agents-infra setup local /path/to/project \
  --clear-codex-primary-session
```

- No primary flags: preserve project-config TOML unchanged.
- Set flags update only supplied fields in
  `[agents.codex.primary_session]`; explicit false is distinct from omission.
- Clear removes only the primary-session table.
- Clear plus any set flag is invalid.
- These flags are local-only; global setup rejects them.
- Existing `[mcp]`, unrelated tables, existing project Codex config mode, and
  user content remain intact.
- Parse and write failures are atomic and leave the original file unchanged.
- The source repository must not add a shared `.configs/project-config.toml`
  template that would overwrite per-project state during sync.

### 2.6 Render and doctor evidence

`agents-infra codex --print-config` is non-launching and must show:

- every discovered project-config path;
- effective model, effort, and yolo values;
- source path per configured field;
- explicit CLI or profile suppression state;
- wrapper yolo expansion source;
- exact final Codex argv.

`agents-infra doctor local PROJECT` reuses the launch resolver and emits these
stable fields:

```text
codex_primary_config_valid: true
codex_primary_model: gpt-5.6-terra
codex_primary_model_source: /abs/project/.agents/.configs/project-config.toml
codex_primary_reasoning_effort: xhigh
codex_primary_reasoning_effort_source: /abs/project/.agents/.configs/project-config.toml
codex_primary_yolo_mode: false
codex_primary_yolo_mode_source: /abs/project/.agents/.configs/project-config.toml
```

Absent strings render empty with source `native`; absent yolo renders `false`
with source `default`. Invalid config makes doctor return nonzero with exact
path and field. Diagnostics must not print tokens, credential values, or
environment contents.

## 3. task-board spawn-ceiling JSON

### 3.1 Exact schema

```json
{
  "mode": "local",
  "local": {
    "board_dir": ".task-board"
  },
  "spawn": {
    "max_parallel": 20,
    "ceilings": {
      "codex": {
        "model": "gpt-5.6-terra",
        "reasoning_effort": "max"
      },
      "claude": {
        "model": "claude-sonnet-4-6"
      }
    }
  },
  "codex": {
    "session": {
      "auto_continue": {
        "enabled": false
      }
    }
  }
}
```

Typed shape:

```go
type SpawnConfig struct {
    MaxParallel int                  `json:"max_parallel"`
    Ceilings    *SpawnCeilingsConfig `json:"ceilings,omitempty"`
}

type SpawnCeilingsConfig struct {
    Codex  *SpawnCeilingConfig `json:"codex,omitempty"`
    Claude *SpawnCeilingConfig `json:"claude,omitempty"`
}

type SpawnCeilingConfig struct {
    Model           string `json:"model,omitempty"`
    ReasoningEffort string `json:"reasoning_effort,omitempty"`
}
```

Structural rules:

- Omitted or null `spawn.ceilings` means no ceiling policy.
- Unknown keys are rejected inside `spawn.ceilings` and its agent objects.
- A present Codex object contains model, reasoning effort, or both.
- A present Claude object requires model and forbids reasoning effort.
- Present strings are trimmed and non-empty.
- Codex effort vocabulary is
  `minimal < low < medium < high < xhigh < max < ultra`.
- Structural loading does not know model IDs, agent ownership, ranks, or
  model-specific effort support.
- Invalid ceiling config fails closed. It does not reuse the fail-open
  `spawn.max_parallel` fallback.
- No `primary_session`, primary model, primary reasoning, or yolo field exists
  anywhere in task-board policy.

### 3.2 Config discovery

Task-board uses `TASK_BOARD_CONFIG` when explicitly set; otherwise it uses the
project-local `task-board.config.json` resolved for the command working
directory. `project_config()` reports the effective path. There is no ancestor
TOML composition and no read of agents-infra project config.

### 3.3 Selection invariant and preflight order

Every spawn remains explicit:

- `--role`, `--background`, and `--model` are required;
- Codex also requires `--reasoning-effort`;
- Claude rejects reasoning effort;
- a ceiling constrains a complete request and never supplies a missing agent,
  model, effort, role, or execution mode.

Preflight order:

1. Require background mode and complete explicit selection.
2. Perform current intrinsic agent/model/effort validation and retain uncapped
   forward-compatibility warnings.
3. Load and structurally validate the complete `spawn.ceilings` subtree.
4. Map only the selected agent ceiling into the CLI-owned semantic resolver.
5. Validate selected-agent ceiling registration, ownership, rank, and effort
   support; compare and resolve the request.
6. Only after successful resolution load task/role/resources or mutate board
   state, write prompts/logs, or register a `RUN-...`.
7. Pass only resolved values into the launcher and executable manifest fields.

Structural errors in either agent object fail the config load. Catalog-semantic
invalidity for the unselected agent does not block the selected spawn, but
`project_config()` validates both agents before reporting policy.

### 3.4 Model rank policy

Ranks are task-board repository policy, not vendor benchmarks, model-name
inference, registry display order, price claims, or agents-infra data.

| Codex rank | Model |
| ---: | --- |
| 10 | `gpt-5.1-codex-mini` |
| 20 | `gpt-5.2` |
| 30 | `gpt-5.1-codex-max` |
| 40 | `gpt-5.2-codex` |
| 50 | `gpt-5.3-codex` |
| 60 | `gpt-5.3-codex-spark` |
| 70 | `gpt-5.4-mini` |
| 80 | `gpt-5.4` |
| 90 | `gpt-5.5` |
| 100 | `gpt-5.6-luna` |
| 110 | `gpt-5.6-terra` |
| 120 | `gpt-5.6-sol` |

| Claude rank | Model |
| ---: | --- |
| 10 | `claude-haiku-4-5-20251001` |
| 20 | `claude-sonnet-4-6` |
| 30 | `claude-opus-4-6` |

`ModelInfo.PolicyRank` is the single rank source and `models(agent=...)`
exposes it. Newly registered or unknown models remain usable under current
uncapped warning rules, but are incomparable while a selected-agent ceiling is
active until ranked.

### 3.5 Deterministic resolver

For explicit request `R` and optional selected-agent ceiling `C`:

1. No ceiling: return `R` unchanged and preserve current warnings.
2. Active ceiling: require request and configured ceiling models used for
   comparison to be known, owned by the selected agent, and ranked.
3. If `C.model` exists and `rank(R.model) > rank(C.model)`, resolve model to
   exactly `C.model`; otherwise retain `R.model`.
4. For Codex, start the effort upper bound at requested effort. Lower it when
   `C.reasoning_effort` is lower.
5. Choose the greatest effort supported by the resolved model at or below that
   bound.
6. If none exists, fail before side effects.
7. Record every changed dimension and ordered reason.

Example: `gpt-5.6-sol/ultra` under a model-only `gpt-5.6-luna` ceiling resolves
to `gpt-5.6-luna/max` because Luna does not support `ultra`.

Stable reason codes:

| Code | Meaning |
| --- | --- |
| `project_spawn_model_ceiling` | Requested model rank exceeded configured model rank |
| `project_spawn_reasoning_effort_ceiling` | Requested effort exceeded configured effort rank |
| `resolved_model_effort_compatibility` | Resolved model required a lower compatible effort |

### 3.6 Output and durable audit

Constrained text:

```text
Selection: codex gpt-5.6-sol/ultra -> gpt-5.6-luna/max (project spawn ceiling: spawn.ceilings.codex)
```

`task-board spawn --json` adds camel-case selection evidence:

```json
{
  "runId": "RUN-260713-example",
  "selection": {
    "agent": "codex",
    "requested": {
      "model": "gpt-5.6-sol",
      "reasoningEffort": "ultra"
    },
    "resolved": {
      "model": "gpt-5.6-luna",
      "reasoningEffort": "max"
    },
    "ceiling": {
      "model": "gpt-5.6-luna",
      "source": "project_config",
      "configPath": "/project/task-board.config.json",
      "configKey": "spawn.ceilings.codex"
    },
    "constrained": true,
    "adjustments": [
      {
        "field": "model",
        "reasonCode": "project_spawn_model_ceiling",
        "from": "gpt-5.6-sol",
        "to": "gpt-5.6-luna",
        "limit": "gpt-5.6-luna"
      },
      {
        "field": "reasoningEffort",
        "reasonCode": "resolved_model_effort_compatibility",
        "from": "ultra",
        "to": "max",
        "limit": "gpt-5.6-luna"
      }
    ]
  }
}
```

The manifest retains backward-compatible executable top-level `model` and
`reasoning_effort` as resolved values and stores the same audit in snake case.
Restart and reroute successors copy the complete audit and resolved executable
pair without re-reading a changed config.

Unconstrained text remains unchanged. JSON and manifests still include
requested/resolved equality, `constrained: false`, and no ceiling object.
Claude omits reasoning fields.

### 3.7 Query contract

`project_config().spawn` retains existing tracking and concurrency fields and
adds:

```json
{
  "ceilings": {
    "contract_version": "spawn-ceilings-v1",
    "configured": true,
    "config_path": "/project/task-board.config.json",
    "codex": {
      "configured": true,
      "model": "gpt-5.6-terra",
      "reasoning_effort": "max",
      "config_key": "spawn.ceilings.codex"
    },
    "claude": {
      "configured": true,
      "model": "claude-sonnet-4-6",
      "config_key": "spawn.ceilings.claude"
    }
  }
}
```

Absent policy reports `configured: false`. `models(agent=...)` exposes
`policyRank`. Local queries remain local in remote-board mode. Invalid policy
returns an error before any partial query result.

Stable validation reasons are:

- `invalid_spawn_ceiling_config`;
- `spawn_model_incomparable_under_ceiling`;
- `spawn_no_supported_effort_under_ceiling`;
- `unsupported_legacy_agent_policy`.

No policy error contains a run ID or leaves task/run artifacts.

## 4. Migration and backward compatibility

### agents-infra

- No primary table preserves current Codex argv and native config precedence.
- Existing `[mcp]` configuration and ancestor union behavior remain intact.
- Existing custom or managed `.codex/config.toml` remains supported. No
  automatic migration occurs. `doctor` reports both project-primary
  provenance and existing shadowing state.
- Operators may intentionally move model/effort to project-config TOML and use
  existing `--codex-config=global` separately; setup does not silently remove
  `.codex/config.toml`.

### task-board

- No `spawn.ceilings` preserves all current spawn behavior.
- `mode`, `local`, `remote`, `spawn.max_parallel`, and
  `codex.session.auto_continue` stay compatible.
- The unreleased top-level `agents` prototype is not auto-migrated or ignored.
  Fail with `unsupported_legacy_agent_policy` and tell the operator to move:
  - primary model/reasoning/yolo to
    `[agents.codex.primary_session]` in project-config TOML;
  - Codex/Claude ceilings to `spawn.ceilings` in task-board JSON.
- Setup preserves valid `spawn.ceilings` and fails without overwrite for
  invalid or legacy prototype JSON.

## 5. Cross-repository development graph

The repositories have separate local boards, so dependencies are enforced
inside each board and cross-board ownership is linked by this resource and
task notes. No fake cross-board filesystem dependency is created.

### relux-agents-infra board

1. `TASK-260713-1phuwx` — Implement composed primary session launch policy.
2. In parallel after task 1:
   - `TASK-260713-4ihi4q` — Configure primary session through local setup.
   - `TASK-260713-21r6jm` — Expose primary session diagnostics.
3. In parallel after setup and diagnostics:
   - `TASK-260713-1u9vc6` — Document primary session project configuration.
   - `TASK-260713-3mlifc` — Validate primary session target projects.

### skill-project-management board

1. `TASK-260713-190sng` — this architecture contract and review gate.
2. In parallel after architecture acceptance:
   - `TASK-260713-2kgake` — Remove cancelled sidecar agent-policy prototype.
   - `TASK-260713-2ojkoi` — Implement spawn ceiling ranks and resolver.
3. `TASK-260713-2pfc45` — Implement spawn ceiling config foundation, after
   prototype cleanup.
4. In parallel after the required foundation/resolver dependencies:
   - `TASK-260713-32fjiy` — Integrate spawn ceiling preflight and audit.
   - `TASK-260713-21zqeu` — Expose spawn ceiling policy queries.
   - `TASK-260713-qbxqfa` — Preserve spawn ceilings during setup.
5. In parallel after implementation surfaces:
   - `TASK-260713-3a76xg` — Validate task-board spawn ceiling policy.
   - `TASK-260713-3jpzwj` — Document task-board spawn ceiling policy.

## 6. Cancelled prototype disposition

- `RUN-260713-35f507` was cancelled by an operator directive after the
  ownership correction.
- `TASK-260713-33b28f` is closed and unassigned. Its partial diff is not
  accepted implementation evidence.
- The run introduced or modified only the observed prototype surface:
  - `pkg/remoteconfig/agents.go`;
  - `pkg/remoteconfig/agents_test.go`;
  - top-level `Agents` and load validation hunks in
    `pkg/remoteconfig/provider.go`;
  - save validation hunk in `pkg/remoteconfig/sidecar_manager.go`.
- `TASK-260713-2kgake` removes only those changes without destructive reset and
  proves unrelated dirty work remains intact.
- `TASK-260713-1t6qlq` is closed and moved to the agents-infra board graph.
- `TASK-260713-21rsgl` is closed because its run was cancelled before producing
  an outcome; this contract and agents-infra tasks replace it.
- Accepted historical `TASK-260713-2sgb9a` remains done for audit history, but
  its combined task-board ownership contract is superseded by this resource.
- Closed broad `TASK-260713-3f4m7e` remains closed; nothing is reopened or
  deleted.

## 7. Completeness and required verification

No unresolved product or architecture decision remains. Implementation and
review must cover:

- agents-infra TOML parsing, ancestor composition, CLI/profile precedence,
  explicit false, yolo deduplication, setup atomicity, render, doctor, docs,
  and disposable two-repository smokes;
- task-board JSON structure, legacy error, rank catalog, resolver, selected-agent
  semantic scope, no-side-effect preflight, text/JSON diagnostics, manifest and
  successor durability, queries, setup preservation, docs, and independent
  validation;
- a negative ownership assertion in both repositories: agents-infra contains
  no spawn ceilings/ranks/manifests, and task-board contains no primary-session
  model/reasoning/yolo policy.

Canonical source anchors inspected for this design:

- agents-infra:
  `tools/agents-infra/internal/infra/codex_launch.go`,
  `tools/agents-infra/internal/infra/infra.go`,
  `tools/agents-infra/main.go`, `README.md`, `SKILL.md`;
- task-board:
  `pkg/remoteconfig`, `tools/board-cli/internal/spawn`,
  `tools/board-cli/cmd/spawn*.go`, `tools/board-cli/cmd/auth.go`,
  `scripts/setup.sh`, `.specs/project-agent-selection-policy.md`.
