# Generic attachment and image intake

## Description
Provide runtime-agnostic, board-agnostic contracts and helper tooling for safely materializing, staging, normalizing, inspecting, and handing off user-provided attachments and images across Codex and Claude workflows.

## Scope
The generic agents-attachments manifest contract, explicit local file paths, read-only source handling, project-local .temp staging, common image normalization including HEIC, visual inspection guidance, bounded OCR fallback, privacy redaction, spawned-agent handoff, setup/install linkage, documentation, and tests. No task-board API, board resource, epic, task ID, or status dependency may be required.

## Acceptance Criteria
Agents can follow one documented workflow for manifest-backed attachments or explicit paths; originals remain read-only; unsupported image formats can be normalized with platform-aware tooling into project-local scratch; direct visual inspection is preferred and OCR fallback is bounded; sensitive identifiers are redacted; spawned agents can receive the same inputs without a board; helpers are installed by agents-infra setup and covered by tests; public docs and the source skill describe the contract.
