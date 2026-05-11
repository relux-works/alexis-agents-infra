## Status
done

## Assigned To
codex

## Created
2026-05-11T20:41:50Z

## Last Update
2026-05-11T20:56:46Z

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
(empty)

## Notes
Added Figma MCP endpoint to source-managed .configs/codex-config.toml and README. Ran agents-infra setup global --source-dir /Users/alexis/src/relux-works/alexis-agents-infra. Verified codex mcp list/get figma and fresh codex exec Figma whoami/get_metadata for node 496:33323. Current already-running chat session still needs restart for direct MCP tools.
Correction: Figma MCP was removed from global Codex defaults. Final design keeps global config MCP-free and enables Figma only through project-local [codex.mcp] opt-in plus .local/bin/codex-local launcher.

## Precondition Resources
(none)

## Outcome Resources
(none)
