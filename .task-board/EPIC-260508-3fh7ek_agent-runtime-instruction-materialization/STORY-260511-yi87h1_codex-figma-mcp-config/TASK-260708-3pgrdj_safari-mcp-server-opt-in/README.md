# Add Safari MCP server opt-in

## Description
Add Apple WebKit Safari MCP server as a supported project-local Codex MCP option in relux-agents-infra. Preserve existing Figma and LLDB MCP behavior and current uncommitted LLDB-related changes. The shared registry must include a Safari Technology Preview stdio server using safaridriver --mcp, project instructions must list Safari alongside Figma and LLDB as a supported MCP, docs/skill text must explain prerequisites, and global setup must sync installed runtime artifacts.

## Scope
Touch only the shared MCP registry, MCP instructions/docs/skill text, focused tests if needed, and setup/install verification for relux-agents-infra. Do not enable Safari MCP globally; it must remain project-local opt-in via enabled_servers.

## Acceptance Criteria
A project can opt in with enabled_servers including safari; agents-infra codex --print-config emits mcp_servers.safari.command for Safari Technology Preview safaridriver and args [--mcp]; source instructions list figma, lldb, and safari as supported shared MCP definitions; README/SKILL document Safari Technology Preview prerequisites and external agents setting; setup syncs installed runtime; relevant Go tests, diff hygiene, and installed-runtime smoke checks pass.
