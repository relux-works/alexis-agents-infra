# Codex primary session and yolo policy

## Description
Own project-scoped primary Codex model, reasoning effort, and persistent yolo choice in agents-infra independently of task-board spawn policy.

## Scope
Extend .agents/.configs/project-config.toml, the agents-infra Codex launch planner, local setup, print-config, doctor, documentation, and validation. Preserve current MCP composition and existing .codex/config.toml modes. Do not add or consume task-board spawn ceilings.

## Acceptance Criteria
The exact TOML contract is [agents.codex.primary_session] with optional non-empty model and reasoning_effort strings plus optional boolean yolo_mode. Ancestor project configs compose root to leaf with nearest per-field precedence and explicit false overriding inherited true. Explicit Codex CLI model or effort values override project values by dimension, explicit --profile suppresses project model and effort, and yolo adds exactly one native dangerous flag only when explicitly requested or effectively true. Setup preserves unrelated TOML, print-config and doctor expose effective values and provenance, absent config preserves native behavior, documentation and disposable two-repository validation pass.
