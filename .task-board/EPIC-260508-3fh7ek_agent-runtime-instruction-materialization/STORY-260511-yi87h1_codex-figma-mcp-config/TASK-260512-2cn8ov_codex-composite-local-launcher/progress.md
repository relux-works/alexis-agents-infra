## Status
done

## Assigned To
codex

## Created
2026-05-12T08:01:49Z

## Last Update
2026-05-12T08:14:59Z

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
(empty)

## Notes
Implemented global agents-infra codex wrapper. It composes project configs upward from cwd, resolves MCP definitions from project/global registries, renders provenance for config sources, supports --print-config and -d yolo shorthand, and keeps codex-local as a compatibility shim.
Verification: go test ./... passed (.temp/agents-infra-go-test-codex-wrapper-01.log); git diff --check passed (.temp/git-diff-check-codex-wrapper-01.log); Tap2Cash agents-infra codex --print-config shows figma from app-local project-config and registry; /tmp shows no enabled MCP; live codex exec smoke returned aagrigore1@mts.ru via figma/whoami.

## Precondition Resources
(none)

## Outcome Resources
(none)
