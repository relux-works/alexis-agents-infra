# TASK-260625-uwtoga Results

Added global workflow reinforcement for stop-the-line forced-fit failures.

## Changes

- `relux-agents-infra/.instructions/INSTRUCTIONS_WORKFLOW.md` now includes `Stop-The-Line: No Forced Fits`.
- The rule requires agents to stop when a clean solution no longer holds, document constraints/evidence/options, and ask for a decision or mark the task `blocked`.
- The rule explicitly rejects using tests, stubs, flags, priority rules, or mock-only behavior to make an invalid model appear valid.

## Verification

- `agents-infra setup global --source-dir ~/src/relux-works/relux-agents-infra` passed and rendered `~/.codex/AGENTS.md`.
- `agents-infra doctor global` passed.
- `git diff --check` passed.
- Installed runtime contains the new section in `~/.agents/.instructions/INSTRUCTIONS_WORKFLOW.md` and `~/.codex/AGENTS.md`.
