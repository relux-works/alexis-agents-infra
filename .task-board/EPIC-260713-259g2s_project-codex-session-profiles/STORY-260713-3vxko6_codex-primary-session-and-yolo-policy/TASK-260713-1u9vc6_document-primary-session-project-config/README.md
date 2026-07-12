# Document primary session project configuration

## Description
Document the agents-infra-owned primary Codex session schema and operator workflow without claiming ownership of task-board spawn selection.

## Scope
Update README.md, SKILL.md, CLI usage/examples, and the mandatory Tools documentation section after implementation. Document the exact [agents.codex.primary_session] TOML, ancestor discovery, per-field precedence, explicit CLI/profile behavior, yolo safety boundary, setup flags, clearing, print-config, doctor, native fallback, .codex/config.toml coexistence, and troubleshooting. Cross-link the task-board spawn-ceiling contract but do not duplicate its model-rank or resolver policy.

## Acceptance Criteria
Documentation contains runnable setup, manual TOML, print-config, doctor, and launch examples; states that absent config preserves native Codex resolution; states that persistent yolo affects only agents-infra Codex primary launches and never task-board spawns; explains nearest-field precedence and explicit false; explains explicit --profile suppression and .codex/config.toml coexistence; and matches CLI help and tested output exactly.
