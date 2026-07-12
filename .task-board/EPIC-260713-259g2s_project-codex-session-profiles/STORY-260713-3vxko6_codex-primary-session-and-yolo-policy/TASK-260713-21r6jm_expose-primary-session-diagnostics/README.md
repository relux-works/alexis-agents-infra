# Expose primary session diagnostics

## Description
Expose the effective agents-infra primary-session policy, application decision, and per-field provenance through Codex render and local doctor surfaces.

## Scope
Extend CodexLaunchPlan and RenderCodexLaunchPlan plus Doctor and its CLI output. Reuse the same composite project-config resolver as launch; do not reparse or invent a second precedence implementation. agents-infra codex --print-config renders effective model, reasoning effort, yolo mode, whether each project value was applied or suppressed, and exact source path. agents-infra doctor local reports codex_primary_config_valid, codex_primary_model and _source, codex_primary_reasoning_effort and _source, codex_primary_yolo_mode and _source. Do not print secrets or environment values.

## Acceptance Criteria
Print-config shows all discovered project config paths, effective values, sources, CLI/profile suppression, wrapper expansion, and final Codex argv without launching. Doctor local returns the exact snake-case fields for configured and absent policy and exits nonzero with source path/field for invalid TOML instead of silently omitting it. Effective false yolo and default false are distinguishable by provenance. Diagnostics match the launch plan in parent/child, explicit CLI, explicit profile, and no-config tests; existing MCP and config-shadowing fields remain intact.
