#!/usr/bin/env zsh

set -euo pipefail

export PATH="/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:${PATH:-}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
AGENTS_DIR="${AGENTS_DIR:-$(cd "$SCRIPT_DIR/.." && pwd)}"
CLAUDE_DIR="${CLAUDE_DIR:-$HOME/.claude}"
CODEX_DIR="${CODEX_DIR:-$HOME/.codex}"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"

CLI="$BIN_DIR/agents-infra"
if [[ -x "$CLI" ]]; then
  exec "$CLI" refresh-links --agents-dir "$AGENTS_DIR" --claude-dir "$CLAUDE_DIR" --codex-dir "$CODEX_DIR" --bin-dir "$BIN_DIR"
fi

cd "$AGENTS_DIR/tools/agents-infra"
exec go run . refresh-links --agents-dir "$AGENTS_DIR" --claude-dir "$CLAUDE_DIR" --codex-dir "$CODEX_DIR" --bin-dir "$BIN_DIR"
