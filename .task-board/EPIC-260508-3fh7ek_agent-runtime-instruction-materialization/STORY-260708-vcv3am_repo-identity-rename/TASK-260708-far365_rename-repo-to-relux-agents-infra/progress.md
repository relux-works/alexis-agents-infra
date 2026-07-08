## Status
done

## Assigned To
codex

## Created
2026-07-08T09:37:15Z

## Last Update
2026-07-08T09:43:22Z

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] All source-controlled old repository identity references are renamed to relux-agents-infra where they describe this repository identity
- [x] Go module, imports, docs, setup metadata, generated instructions, and showcase metadata pass tests after rename
- [x] GitHub repository and local checkout folder are renamed to relux-agents-infra and origin remote points to the new URL
- [x] Repository-wide personal-name search is clean except intentionally external user/account paths if any are documented

## Notes
Renamed source identity, Go module/import path, docs, setup banners, skill link name, board references, GitHub repository slug, and local checkout folder to relux-agents-infra. Removed personal-name source occurrences and stale installed skill links. Verified with rg, go test, go vet, setup, doctor, and GitHub repo view.

## Precondition Resources
(none)

## Outcome Resources
(none)
