# TASK-260708-3pgrdj Implementation Guide

Official source:
- https://webkit.org/blog/18136/introducing-the-safari-mcp-server-for-web-developers/

WebKit contract to implement:
- Safari MCP server was introduced with Safari Technology Preview 247.
- It is a stdio MCP server launched through Safari Technology Preview's `safaridriver`.
- Codex command from WebKit:
  `codex mcp add safari-mcp-stp -- "/Applications/Safari Technology Preview.app/Contents/MacOS/safaridriver" --mcp`
- Equivalent config:
  `command = "/Applications/Safari Technology Preview.app/Contents/MacOS/safaridriver"`
  `args = ["--mcp"]`
- Prerequisites:
  - Install Safari Technology Preview.
  - Enable `Safari Settings > Advanced > Show features for web developers`.
  - Enable `Safari Settings > Developer > Enable remote automation and external agents`.

Implementation expectations:
- Add a shared MCP registry entry with key `safari`.
- Keep it project-local opt-in only; do not add global enablement.
- Update `.instructions/INSTRUCTIONS_TOOLS.md` supported MCP list to include `safari`.
- Update README/SKILL MCP sections with Safari Technology Preview prerequisites.
- Add or update focused tests/smoke coverage so `agents-infra codex --print-config` emits:
  - `mcp_servers.safari.command="/Applications/Safari Technology Preview.app/Contents/MacOS/safaridriver"`
  - `mcp_servers.safari.args=["--mcp"]`
- Preserve existing dirty LLDB/Figma changes and do not revert unrelated work.
- Run `./setup.sh` and verify installed `/Users/alexis/.codex/AGENTS.md` includes the supported MCP list.
