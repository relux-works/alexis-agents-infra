## Status
to-review

## Assigned To
codex-inline

## Created
2026-07-07T13:09:29Z

## Last Update
2026-07-07T13:12:25Z

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
(empty)

## Notes
Started inline per user request. Scope: add macOS/Homebrew LLVM lldb-mcp provisioning to scripts/setup.sh and document the behavior. Existing LLDB MCP registry/code changes are already dirty in this checkout and will be preserved.
Implemented macOS lldb-mcp bootstrap in scripts/setup.sh. Source repo: /Users/alexis/src/relux-works/alexis-agents-infra. Setup command verified: ./scripts/setup.sh. Evidence logs: /Users/alexis/src/videocall/ios/.temp/tool-readiness/lldb-mcp-agents-infra-setup-source-02.log and -03.log. Go tests: /Users/alexis/src/videocall/ios/.temp/tool-readiness/lldb-mcp-agents-infra-go-test-02.log. The hook installs Homebrew llvm if needed, writes /opt/homebrew/bin/lldb-mcp managed wrapper, and supports AGENTS_INFRA_SKIP_LLDB_MCP=1.

## Precondition Resources
(none)

## Outcome Resources
(none)
