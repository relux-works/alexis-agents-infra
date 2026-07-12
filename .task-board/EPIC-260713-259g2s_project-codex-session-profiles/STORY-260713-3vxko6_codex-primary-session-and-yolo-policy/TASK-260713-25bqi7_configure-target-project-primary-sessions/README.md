# Configure target project primary sessions

## Description
Install and verify the requested Terra/xhigh/yolo primary Codex profile in both target projects

## Scope
After the primary policy, setup flags, and diagnostics are accepted, use the built agents-infra CLI to locally set [agents.codex.primary_session] in the relux-agents-infra and skill-project-management project runtimes. Set model gpt-5.6-terra, reasoning_effort xhigh, and yolo_mode true. Preserve existing project MCP config and unrelated local runtime state. This task configures actual project runtimes; it does not edit task-board spawn ceilings or shell aliases.

## Acceptance Criteria
Both target project roots have an effective primary profile with Terra/xhigh/yolo true through agents-infra setup local. agents-infra codex --print-config and doctor local show the requested values, source path, and exactly one native danger flag in the rendered primary Codex argv. MCP composition remains intact, no task-board.config.json primary field is added, and the outcome records commands plus before/after sanitized config evidence.
