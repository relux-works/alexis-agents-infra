#!/usr/bin/env zsh
#
# setup.sh
# Installs this repo into ~/.agents and refreshes Claude/Codex symlinks.
#

set -euo pipefail

export PATH="/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:${PATH:-}"

SOURCE_DIR="$(cd "$(dirname "$0")" && pwd)"
AGENTS_DIR="${AGENTS_DIR:-$HOME/.agents}"
CLAUDE_DIR="${CLAUDE_DIR:-$HOME/.claude}"
CODEX_DIR="${CODEX_DIR:-$HOME/.codex}"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"

red()   { print -P "%F{red}$1%f" }
green() { print -P "%F{green}$1%f" }
yellow(){ print -P "%F{yellow}$1%f" }

show_usage() {
  cat <<EOF
Usage: ./setup.sh [options]

Options:
  --agents-dir PATH   Install destination (default: ~/.agents)
  --no-sync           Skip repo sync; only refresh symlinks/helpers in destination
  --help, -h          Show this help
EOF
}

NO_SYNC=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --agents-dir)
      AGENTS_DIR="$2"
      shift 2
      ;;
    --no-sync)
      NO_SYNC=1
      shift
      ;;
    --help|-h)
      show_usage
      exit 0
      ;;
    *)
      red "Unknown option: $1"
      show_usage
      exit 1
      ;;
  esac
done

warn_dirty_agents_repo() {
  if [[ "$SOURCE_DIR" == "$AGENTS_DIR" ]]; then
    return
  fi
  if git -C "$AGENTS_DIR" rev-parse --show-toplevel >/dev/null 2>&1; then
    local dirty
    dirty="$(git -C "$AGENTS_DIR" status --short 2>/dev/null || true)"
    if [[ -n "$dirty" ]]; then
      yellow "WARNING: $AGENTS_DIR has local changes."
      yellow "Setup will overwrite destination files that also exist in the source repo."
    fi
  fi
}

build_rsync_excludes() {
  local exclude_file="$1"

  : > "$exclude_file"

  if ! git -C "$AGENTS_DIR" rev-parse --show-toplevel >/dev/null 2>&1; then
    return
  fi

  local dirty_paths
  dirty_paths="$(
    {
      git -C "$AGENTS_DIR" diff --name-only 2>/dev/null || true
      git -C "$AGENTS_DIR" diff --name-only --cached 2>/dev/null || true
      git -C "$AGENTS_DIR" ls-files --others --exclude-standard 2>/dev/null || true
    } | awk 'NF' | sort -u
  )"

  if [[ -z "$dirty_paths" ]]; then
    return
  fi

  yellow "Preserving dirty destination paths during sync:"
  local dirty_path
  while IFS= read -r dirty_path; do
    [[ -z "$dirty_path" ]] && continue
    print -- "$dirty_path" >> "$exclude_file"
    yellow "  - $dirty_path"
  done <<< "$dirty_paths"
}

copy_tree_fallback() {
  local exclude_file="$1"
  local rel src dst

  green "rsync not found; using portable copy fallback"

  while IFS= read -r rel; do
    [[ -z "$rel" ]] && continue
    rel="${rel#./}"

    if [[ -s "$exclude_file" ]] && grep -Fqx -- "$rel" "$exclude_file"; then
      continue
    fi

    src="$SOURCE_DIR/$rel"
    dst="$AGENTS_DIR/$rel"

    if [[ -d "$src" && ! -L "$src" ]]; then
      if [[ -e "$dst" && ! -d "$dst" ]]; then
        rm -rf "$dst"
      fi
      mkdir -p "$dst"
      continue
    fi

    mkdir -p "$(dirname "$dst")"
    if [[ -e "$dst" || -L "$dst" ]]; then
      rm -rf "$dst"
    fi
    cp -PR "$src" "$dst"
  done < <(
    cd "$SOURCE_DIR"
    find . \
      \( -path './.git' -o -path './.git/*' -o -name '.DS_Store' -o -name '.skill-lock.json' \) -prune \
      -o -mindepth 1 -print
  )
}

sync_repo() {
  mkdir -p "$AGENTS_DIR"

  if [[ "$SOURCE_DIR" == "$AGENTS_DIR" ]]; then
    green "Source repo already lives at $AGENTS_DIR"
    return
  fi

  warn_dirty_agents_repo

  green "Syncing repo -> $AGENTS_DIR"
  local exclude_file
  exclude_file="$(mktemp)"
  build_rsync_excludes "$exclude_file"
  if command -v rsync >/dev/null 2>&1; then
    rsync -a \
      --exclude '.git/' \
      --exclude '.DS_Store' \
      --exclude '.skill-lock.json' \
      --exclude-from "$exclude_file" \
      "$SOURCE_DIR/" "$AGENTS_DIR/"
  else
    copy_tree_fallback "$exclude_file"
  fi
  rm -f "$exclude_file"
}

run_installed_setup() {
  local installed_script="$AGENTS_DIR/.scripts/setup-symlinks.sh"
  if [[ ! -x "$installed_script" ]]; then
    chmod +x "$installed_script"
  fi

  green "Refreshing Claude/Codex/helper symlinks..."
  AGENTS_DIR="$AGENTS_DIR" \
  CLAUDE_DIR="$CLAUDE_DIR" \
  CODEX_DIR="$CODEX_DIR" \
  BIN_DIR="$BIN_DIR" \
  "$installed_script"
}

main() {
  print ""
  green "=== alexis-agents-infra setup ==="
  print ""

  if [[ "$NO_SYNC" -eq 0 ]]; then
    sync_repo
  else
    yellow "Skipping repo sync (--no-sync)"
  fi

  run_installed_setup

  print ""
  green "=== Done ==="
  print ""
}

main "$@"
