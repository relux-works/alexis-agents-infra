# Add provider-specific primary session model policies

## Description
Let a project select its primary Codex and Claude sessions independently, without letting one provider configuration affect the other.

## Scope
Extend the agents-infra project-config schema, Claude launch planner, setup mutation API, print-config, doctor output, documentation, tests, and local rollout for relux-agents-infra and skill-project-management. Preserve the current Codex contract and do not read or alter task-board spawn ceilings.

## Acceptance Criteria
Project TOML supports independent [agents.codex.primary_session] and [agents.claude.primary_session] tables. Codex retains model, reasoning_effort, and yolo_mode unchanged; Claude accepts only its own optional non-empty model string. Ancestor composition and field provenance apply independently per provider. agents-infra claude applies the effective project Claude model through its supported native launch argument, while an explicit Claude CLI model wins. Setup can set and clear only the Claude primary-session table without changing Codex, MCP, comments, or unrelated TOML. print-config and doctor expose Claude model validity, value, and source. Both target project local configs use Codex gpt-5.6-terra/xhigh/yolo=true plus Claude claude-opus-4-6. Focused and full relevant Go tests pass; docs explain the provider boundary and no cross-provider leakage.
