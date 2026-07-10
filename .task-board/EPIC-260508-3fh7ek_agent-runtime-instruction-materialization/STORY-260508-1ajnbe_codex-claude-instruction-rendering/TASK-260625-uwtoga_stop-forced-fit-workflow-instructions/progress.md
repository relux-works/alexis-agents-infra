## Status
done

## Assigned To
codex-inline

## Created
2026-06-25T13:32:19Z

## Last Update
2026-07-10T08:54:13Z

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Add no-forced-fit workflow guardrail to shared instructions
- [x] Preserve existing unrelated dirty changes
- [x] Add skill-agnostic source-fix rule for reusable skill/tool contract gaps
- [x] Clarify autonomous completion versus objective external blockers
- [x] Add canonical iOS Bluetooth pre-permission example
- [x] Require persisted escalation evidence alternatives and exact decision
- [x] Add renderer coverage and reinstall verification

## Notes
2026-07-02: Added skill-agnostic source-fix rule to .instructions/INSTRUCTIONS_SKILLS.md. Refreshed global runtime and VideoCall local runtime with agents-infra setup. Rule intentionally avoids enumerating individual skills/tools.
2026-07-10: Reopened by explicit human request for a foundational global policy pass. This remains distinct from project-management orchestrator review-loop changes.
Expanded Stop-The-Line to autonomous-by-default with objective external blocker criteria, persisted escalation evidence, exact decision requirements, and the iOS Bluetooth pre-permission canonical example. Regression fixtures and temp runtime rendering verified.
Canonical ./setup.sh completed successfully. Installed Codex and Claude runtime outputs contain the foundational policy and Bluetooth example; source/runtime drift checks pass. Ready for review.
Independent review accepted. Verified autonomous-by-default Stop-The-Line policy, objective external blocker boundary, persisted constraint/evidence/failed assumption/attempts/options/tradeoffs/exact decision packet, canonical iOS Bluetooth pre-permission powered-state example, Codex and Claude materialization, uncached full and focused tests, vet, board validation, and diff hygiene. Review logs: .temp/review-TASK-260710-381t1q-TASK-260625-uwtoga/.

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260625-uwtoga_results.md](file://TASK-260625-uwtoga/TASK-260625-uwtoga_results.md) — Global workflow guardrail implementation notes
- [TASK-260625-uwtoga_skill-source-rule.md](file://TASK-260625-uwtoga/TASK-260625-uwtoga_skill-source-rule.md) — Skill-agnostic source-fix instruction update and setup evidence
- [foundational-stop-the-line-policy-outcome.md](file://TASK-260625-uwtoga/foundational-stop-the-line-policy-outcome.md) — Foundational autonomy and forced-fit policy implementation evidence
