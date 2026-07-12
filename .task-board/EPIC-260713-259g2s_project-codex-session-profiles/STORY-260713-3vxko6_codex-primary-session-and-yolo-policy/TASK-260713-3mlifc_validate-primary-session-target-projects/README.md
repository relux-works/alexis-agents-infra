# Validate primary session target projects

## Description
Independently validate the implemented agents-infra primary-session workflow and the ownership boundary in disposable copies of both target repositories.

## Scope
Use task-scoped copies of relux-agents-infra and skill-project-management. Exercise local setup set/update/clear, ancestor composition, MCP coexistence, print-config, doctor, explicit CLI/profile precedence, yolo true/false, invalid config failure, and no-config compatibility. Confirm no source checkout or task-board.config.json is modified and no task-board spawn policy is consumed. Report reproducible defects; do not patch production behavior.

## Acceptance Criteria
Outcome records exact commands and results for both repository copies. Each copy proves setup preservation, nearest per-field provenance, child false yolo masking, exactly one danger flag when enabled, no danger flag when disabled, explicit CLI/profile rules, invalid-config fail-fast, and native no-config behavior. It verifies task-board.config.json is irrelevant to primary launch and .agents project TOML contains no spawn ceilings. Relevant Go tests and diff checks pass or defects are routed with evidence.
