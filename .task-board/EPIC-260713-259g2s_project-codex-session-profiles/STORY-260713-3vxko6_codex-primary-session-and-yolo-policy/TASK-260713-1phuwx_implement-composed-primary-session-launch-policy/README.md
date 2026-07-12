# Implement composed primary session launch policy

## Description
Parse, validate, compose, and apply the agents-infra-owned Codex primary-session TOML policy before launching Codex.

## Scope
Add a typed project-config parser or focused project-config module and update tools/agents-infra/internal/infra/codex_launch.go with tests. Read [agents.codex.primary_session] from every discovered project config in root-to-leaf order while continuing the current MCP union behavior. Compose model, reasoning_effort, and yolo_mode independently; a nearer present field wins and false is a present yolo value. Explicit --model or -m and exact top-level -c model or model_reasoning_effort values win by dimension; explicit --profile suppresses project model and effort. Explicit -d, --danger, or --yolo and effective yolo_mode=true produce exactly one --dangerously-bypass-approvals-and-sandbox. Do not import task-board catalogs, ranks, ceilings, or manifests.

## Acceptance Criteria
Absent primary config and absent explicit selection leave model, effort, and danger argv unchanged. Valid parent and child TOML compose per field with nearest provenance and child yolo_mode=false masks parent true. Empty strings, wrong TOML types, duplicate conflicting values, and malformed tables fail before exec with source path and field. Explicit model or effort overrides only that dimension; explicit --profile suppresses both project dimensions but not the independent yolo decision. Project or explicit yolo emits exactly one native dangerous flag; false or absent emits none. Existing MCP composition tests remain green and focused launch tests cover interactive and exec forwarding.
