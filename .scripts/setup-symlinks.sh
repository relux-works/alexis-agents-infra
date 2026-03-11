#!/bin/zsh
#
# setup-symlinks.sh
# Internal helper used by setup.sh.
# Refreshes symlinks from ~/.claude/ and ~/.codex/ to ~/.agents/ and installs helper CLIs.
#

set -e

export PATH="/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:${PATH:-}"

AGENTS_DIR="${AGENTS_DIR:-$HOME/.agents}"
CLAUDE_DIR="${CLAUDE_DIR:-$HOME/.claude}"
CODEX_DIR="${CODEX_DIR:-$HOME/.codex}"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
AGENTS_SKILLS_DIR="$AGENTS_DIR/skills"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo "${RED}[ERROR]${NC} $1"; }

ensure_repo_skill_links() {
    mkdir -p "$AGENTS_SKILLS_DIR"

    # Promote the repo itself as a skill entry.
    create_symlink "$AGENTS_DIR" "$AGENTS_SKILLS_DIR/alexis-agents-infra"

    # Promote bundled skill-creator into the public skills registry.
    if [ -d "$AGENTS_DIR/.skills/skill-creator" ]; then
        create_symlink "$AGENTS_DIR/.skills/skill-creator" "$AGENTS_SKILLS_DIR/skill-creator"
    fi
}

# Backup and remove existing file/dir, then create symlink
create_symlink() {
    local target="$1"
    local link="$2"
    local backup_suffix=".bak.$(date +%Y%m%d%H%M%S)"

    if [ -L "$link" ]; then
        # Already a symlink - remove it
        rm "$link"
        log_info "Removed existing symlink: $link"
    elif [ -e "$link" ]; then
        # Exists but not a symlink - backup
        mv "$link" "${link}${backup_suffix}"
        log_warn "Backed up existing: $link -> ${link}${backup_suffix}"
    fi

    ln -s "$target" "$link"
    log_info "Created symlink: $link -> $target"
}

# ============================================================================
# Claude Code setup
# ============================================================================

setup_claude() {
    log_info "=== Setting up Claude Code symlinks ==="

    # Create ~/.claude if it doesn't exist
    mkdir -p "$CLAUDE_DIR"
    mkdir -p "$CLAUDE_DIR/skills"

    # 1. Instructions directory
    create_symlink "$AGENTS_DIR/.instructions" "$CLAUDE_DIR/instructions"

    # 2. CLAUDE.md entry point (update to point to new location)
    cat > "$CLAUDE_DIR/CLAUDE.md" << 'EOF'
# Global Claude Instructions

Load all global instructions:

@~/.agents/.instructions/INSTRUCTIONS.md
EOF
    log_info "Updated $CLAUDE_DIR/CLAUDE.md"

    # 3. Skills - link from ~/.agents/skills/ registry
    for skill_dir in "$AGENTS_SKILLS_DIR"/*; do
        if [ -d "$skill_dir" ] || [ -L "$skill_dir" ]; then
            skill_name=$(basename "$skill_dir")
            create_symlink "$skill_dir" "$CLAUDE_DIR/skills/$skill_name"
        fi
    done

    # 4. Settings (permissions, model, plugins)
    create_symlink "$AGENTS_DIR/.configs/claude-settings.json" "$CLAUDE_DIR/settings.json"

    log_info "=== Claude Code setup complete ==="
}

# ============================================================================
# Codex CLI setup
# ============================================================================

setup_codex() {
    log_info "=== Setting up Codex CLI symlinks ==="

    # Create ~/.codex if it doesn't exist
    mkdir -p "$CODEX_DIR"
    mkdir -p "$CODEX_DIR/skills"
    mkdir -p "$CODEX_DIR/rules"

    # 1. AGENTS.md - main instructions file
    create_symlink "$AGENTS_DIR/.instructions/AGENTS.md" "$CODEX_DIR/AGENTS.md"

    # 2. Skills - create symlinks for each skill (skip .system)
    for skill_dir in "$AGENTS_SKILLS_DIR"/*; do
        if [ -d "$skill_dir" ] || [ -L "$skill_dir" ]; then
            skill_name=$(basename "$skill_dir")
            # Don't overwrite system skills
            if [ ! -d "$CODEX_DIR/skills/.system/$skill_name" ]; then
                create_symlink "$skill_dir" "$CODEX_DIR/skills/$skill_name"
            else
                log_warn "Skipping $skill_name - exists in .system"
            fi
        fi
    done

    # 3. Config
    create_symlink "$AGENTS_DIR/.configs/codex-config.toml" "$CODEX_DIR/config.toml"

    # 4. Rules
    for rule_file in "$AGENTS_DIR/.rules"/*; do
        if [ -f "$rule_file" ]; then
            rule_name=$(basename "$rule_file")
            create_symlink "$rule_file" "$CODEX_DIR/rules/$rule_name"
        fi
    done

    log_info "=== Codex CLI setup complete ==="
}

# ============================================================================
# Shared helper CLI setup
# ============================================================================

setup_helpers() {
    log_info "=== Setting up shared helper CLIs ==="

    mkdir -p "$BIN_DIR"

    create_symlink "$AGENTS_DIR/.scripts/agents-attachments" "$BIN_DIR/agents-attachments"

    log_info "=== Shared helper CLI setup complete ==="
}

# ============================================================================
# Verification
# ============================================================================

verify_setup() {
    log_info "=== Verifying setup ==="

    echo ""
    echo "$AGENTS_DIR structure:"
    ls -la "$AGENTS_DIR"

    echo ""
    echo "$CLAUDE_DIR symlinks:"
    ls -la "$CLAUDE_DIR/instructions" 2>/dev/null || echo "  instructions: NOT FOUND"
    ls -la "$CLAUDE_DIR/settings.json" 2>/dev/null || echo "  settings.json: NOT FOUND"
    ls -la "$CLAUDE_DIR/skills" 2>/dev/null | head -5 || echo "  skills: NOT FOUND"

    echo ""
    echo "$CODEX_DIR symlinks:"
    ls -la "$CODEX_DIR/AGENTS.md" 2>/dev/null || echo "  AGENTS.md: NOT FOUND"
    ls -la "$CODEX_DIR/config.toml" 2>/dev/null || echo "  config.toml: NOT FOUND"
    ls -la "$CODEX_DIR/skills" 2>/dev/null | head -5 || echo "  skills: NOT FOUND"

    echo ""
    echo "Shared helper CLIs:"
    ls -la "$BIN_DIR/agents-attachments" 2>/dev/null || echo "  agents-attachments: NOT FOUND"

    log_info "=== Verification complete ==="
}

# ============================================================================
# Main
# ============================================================================

main() {
    echo ""
    log_info "Starting agent symlink setup from $AGENTS_DIR..."
    echo ""

    # Check that source exists
    if [ ! -d "$AGENTS_DIR" ]; then
        log_error "$AGENTS_DIR does not exist!"
        exit 1
    fi

    ensure_repo_skill_links
    echo ""
    setup_claude
    echo ""
    setup_codex
    echo ""
    setup_helpers
    echo ""
    verify_setup

    echo ""
    log_info "Done! $AGENTS_DIR is now the configured source of truth."
    echo ""
}

main "$@"
