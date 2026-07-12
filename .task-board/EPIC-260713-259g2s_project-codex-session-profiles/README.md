# Project Codex session profiles

## Description
Project-scoped primary Codex selection and yolo launch configuration

## Scope
Extend agents-infra project-config.toml, launcher, setup, and doctor paths so a project can own its primary Codex model, reasoning effort, and yolo mode while preserving existing MCP configuration and ancestor composition.

## Acceptance Criteria
The configuration is documented, validated, composed deterministically, applied to agents-infra codex, visible in doctor and print-config, safe by default when absent, covered by tests, and usable from the two target projects.
