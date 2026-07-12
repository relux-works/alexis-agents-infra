# Configure primary session through local setup

## Description
Add a safe explicit local-setup interface for creating, updating, or clearing the project primary-session TOML table without overwriting unrelated configuration.

## Scope
Extend agents-infra setup local with --codex-primary-model, --codex-primary-reasoning-effort, --codex-yolo-mode=true|false, and --clear-codex-primary-session. Merge only [agents.codex.primary_session] in TARGET/.agents/.configs/project-config.toml through an atomic write. With no profile flags, preserve the file unchanged. Preserve [mcp], unknown unrelated tables, comments when practical, and existing .codex/config.toml mode behavior. Reject profile flags in global mode and reject conflicting clear-plus-set requests.

## Acceptance Criteria
Local setup with no primary flags preserves the existing project config. Each supplied field is set without erasing omitted existing fields; explicit yolo false is distinguishable from omission. Clear removes only [agents.codex.primary_session]. Invalid existing TOML, wrong types, empty supplied strings, conflicting flags, or write failure leave the original file byte-identical and return a field/path error. Global setup cannot persist project primary policy. Focused setup tests prove MCP and unrelated table preservation plus atomic failure behavior.
