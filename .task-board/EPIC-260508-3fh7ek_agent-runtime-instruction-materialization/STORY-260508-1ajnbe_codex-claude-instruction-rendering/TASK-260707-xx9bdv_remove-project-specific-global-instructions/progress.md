## Status
to-review

## Assigned To
codex

## Created
2026-07-07T10:12:37Z

## Last Update
2026-07-07T10:27:29Z

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Preserve source and installed global instruction diffs under task temp artifacts
- [x] Remove project-specific Tap2Cash/x-platform-airdrop workflow rule from global source instructions
- [x] Delete stale installed global instruction directory and run agents-infra setup global
- [x] Verify global instructions no longer contain project-specific Swipe2Cash/Tap2Cash material

## Notes
Preserved source worktree diff, source/global project-specific hit lists, installed extra global instruction files, and untracked source files under ios/swipe2cash/.temp/global-instructions-reset-260707 before resetting installed global runtime.
Ran agents-infra setup global after deleting ~/.agents/.instructions and rendered global AGENTS files. Verification: agents-infra doctor global passed; global project-specific hit list is empty; installed extra instruction file list is empty; go test ./... passed from tools/agents-infra.
Follow-up review from skill-swift-relux bootstrap: kept reusable infra changes, removed the remaining swipe2cash fixture path from tools/agents-infra/main_test.go, reran go test ./... and agents-infra setup global, and verified global/project-specific hit checks are empty.

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260707-xx9bdv_global-instructions-reset-summary.md](file://TASK-260707-xx9bdv/TASK-260707-xx9bdv_global-instructions-reset-summary.md) — Global instruction reset and preserved valuable diff summary
