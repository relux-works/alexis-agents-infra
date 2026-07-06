# TASK-260625-uwtoga: stop-forced-fit-workflow-instructions

## Description
Add global workflow reinforcement that agents must stop and escalate when a task cannot be solved autonomously without forcing a fragile workaround. Use the Bluetooth permission/off-state failure as the pattern: after discovering an API or product constraint, the agent must document the constraint, present options, and ask/block instead of accumulating hacks and wasting time.

## Scope
Shared workflow instruction module only. Avoid dirty README.md/SKILL.md and Go source files already modified by other work.

## Acceptance Criteria
Global workflow instructions contain a clear no-forced-fit rule; the rule says to stop, document constraint/evidence/options, and ask or mark blocked when autonomous implementation cannot proceed without brittle hacks; it warns against using tests or stubs to justify an invalid product/API model.
