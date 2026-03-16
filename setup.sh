#!/usr/bin/env zsh

set -euo pipefail

export PATH="/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:${PATH:-}"

SOURCE_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
TARGET="$BIN_DIR/agents-infra"

show_usage() {
  cat <<EOF
Usage: ./setup.sh [options]

Options:
  --bin-dir PATH      Install the agents-infra launcher into PATH
  --help, -h          Show this help

Behavior:
  1. Install/update the agents-infra launcher
  2. Run: agents-infra setup global
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --bin-dir)
      BIN_DIR="$2"
      shift 2
      ;;
    --help|-h)
      show_usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      show_usage >&2
      exit 1
      ;;
  esac
done

mkdir -p "$BIN_DIR"
cat > "$TARGET" <<EOF
#!/usr/bin/env sh
set -eu
export AGENTS_INFRA_SOURCE_DIR="$SOURCE_DIR"
cd "$SOURCE_DIR/tools/agents-infra"
exec go run . "\$@"
EOF
chmod +x "$TARGET"

echo "Installed agents-infra launcher: $TARGET"
echo "Running global setup via: $TARGET setup global"
"$TARGET" setup global
echo "Global setup finished."
echo "Project-local install:"
echo "  agents-infra setup local /path/to/project"
echo "Doctor:"
echo "  agents-infra doctor global"
