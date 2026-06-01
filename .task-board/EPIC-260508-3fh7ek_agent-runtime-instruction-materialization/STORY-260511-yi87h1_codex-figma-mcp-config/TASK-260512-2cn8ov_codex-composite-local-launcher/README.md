# TASK-260512-2cn8ov: codex-composite-local-launcher

## Description
Add agents-infra codex command that walks upward from cwd, composes project-local MCP opt-in configs, and launches Codex with the resulting -c overrides and yolo shorthand.

## Scope
Implement global agents-infra codex launcher with composite project-local MCP configuration, config provenance rendering, and compatibility with existing local setup.

## Acceptance Criteria
agents-infra codex walks upward from cwd and composes .agents/.configs/project-config.toml files; enabled MCP servers are resolved from local registries with global fallback; normal launch logs source provenance before exec; --print-config prints provenance without launching; -d expands to Codex yolo flag; existing codex-local launcher delegates to agents-infra codex; tests cover merge/provenance and setup compatibility.
