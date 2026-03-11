# ~/.agents/

Single source of truth for AI agent configurations, instructions, skills, and rules.

Works with:
- **Claude Code** (`~/.claude/`)
- **Codex CLI** (`~/.codex/`)

## Quick Start

```bash
# From the source repo
cd /path/to/alexis-agents-infra
./setup.sh
```

`setup.sh` syncs the repo into `~/.agents`, promotes `alexis-agents-infra` and
`skill-creator` into the public `~/.agents/skills/` registry, then refreshes
symlinks in `~/.claude/`, `~/.codex/`, and `~/.local/bin`.

## Structure

```
~/.agents/
├── .instructions/          # Global instructions (modular .md files)
│   ├── INSTRUCTIONS.md     # Entry point (loads all modules)
│   ├── AGENTS.md           # Entry point for Codex CLI
│   ├── INSTRUCTIONS_ATTACHMENTS.md
│   ├── INSTRUCTIONS_PLATFORM.md
│   ├── INSTRUCTIONS_STRUCTURE.md
│   ├── INSTRUCTIONS_TOOLS.md
│   ├── INSTRUCTIONS_SKILLS.md
│   ├── INSTRUCTIONS_DIAGRAMS.md
│   ├── INSTRUCTIONS_TESTING.md
│   ├── INSTRUCTIONS_WORKFLOW.md
│   ├── INSTRUCTIONS_DOCS.md
│   └── INSTRUCTIONS_STYLE.md
│
├── .skills/                # Skills for Claude Code & Codex CLI
│   ├── algorithmic-art/
│   ├── architecture-diagrams/
│   ├── brand-guidelines/
│   ├── canvas-design/
│   ├── doc-coauthoring/
│   ├── docx/
│   ├── frontend-design/
│   ├── internal-comms/
│   ├── ios-ui-validation/
│   ├── mcp-builder/
│   ├── pdf/
│   ├── pptx/
│   ├── skill-creator/
│   ├── slack-gif-creator/
│   ├── theme-factory/
│   ├── web-artifacts-builder/
│   ├── web-search/
│   ├── webapp-testing/
│   └── xlsx/
│
├── .scripts/               # Setup and utility scripts
│   ├── setup-symlinks.sh   # Internal symlink refresher used by setup.sh
│   └── agents-attachments  # Helper for agents-attachments-manifest.json
│
├── .configs/               # Tool configurations
│   ├── claude-settings.json    # Claude Code settings (reference)
│   └── codex-config.toml       # Codex CLI config
│
└── .rules/                 # Codex CLI rules
    └── default.rules       # Pre-approved commands
```

## Instructions

Modular instruction files in `.instructions/`:

| File | Purpose |
|------|---------|
| `INSTRUCTIONS.md` | Entry point for Claude Code |
| `AGENTS.md` | Entry point for Codex CLI |
| `INSTRUCTIONS_PLATFORM.md` | Target platform preferences (iOS > macOS) |
| `INSTRUCTIONS_STRUCTURE.md` | Project structure conventions |
| `INSTRUCTIONS_TOOLS.md` | Allowed CLI tools |
| `INSTRUCTIONS_SKILLS.md` | Skills system usage |
| `INSTRUCTIONS_DIAGRAMS.md` | C4/PlantUML diagram rules |
| `INSTRUCTIONS_TESTING.md` | Swift Testing, refactoring workflow |
| `INSTRUCTIONS_WORKFLOW.md` | Git, task tracking, logging |
| `INSTRUCTIONS_DOCS.md` | Documentation requirements |
| `INSTRUCTIONS_STYLE.md` | Communication style |

## Skills

Each skill follows the structure:

```
skill-name/
├── SKILL.md              # Required: frontmatter + instructions
├── scripts/              # Optional: executable code
├── references/           # Optional: docs/schemas
└── assets/               # Optional: templates/resources
```

### Available Skills

| Skill | Description |
|-------|-------------|
| `ios-ui-validation` | UI testing with screenshot validation, Page Object pattern |
| `skill-creator` | Scaffold new skills |
| `architecture-diagrams` | C4/PlantUML diagrams |
| `frontend-design` | Production-grade frontend interfaces |
| `docx` / `pdf` / `pptx` / `xlsx` | Office document manipulation |
| `webapp-testing` | Playwright-based web testing |
| `mcp-builder` | Build MCP servers |
| `web-search` | Web search integration |
| `canvas-design` | Visual art in PNG/PDF |
| `algorithmic-art` | p5.js generative art |
| `theme-factory` | Artifact styling toolkit |
| `brand-guidelines` | Anthropic brand colors/typography |
| `internal-comms` | Internal communications templates |
| `slack-gif-creator` | Animated GIFs for Slack |
| `doc-coauthoring` | Documentation co-authoring workflow |
| `web-artifacts-builder` | Multi-component HTML artifacts |

## Configs

### Claude Code (`claude-settings.json`)

Reference config with:
- Allowed tools (Bash, Read, Edit, Write, etc.)
- Default model: `opus`
- Enabled plugins: `swift-lsp`

### Codex CLI (`codex-config.toml`)

- Model: `gpt-5.2`
- Reasoning effort: `xhigh`
- Trusted projects list

## Attachments

This repo defines a generic agent attachment contract:

- manifest file name: `agents-attachments-manifest.json`
- env var: `AGENTS_ATTACHMENTS_MANIFEST`
- helper CLI: `agents-attachments`

The repo does not itself ingest chat attachments. A separate runtime or launcher
must materialize files locally, write the manifest, and export the env var before
starting the agent process.

For Codex sessions, the helper can bootstrap a local manifest from rollout
history when `CODEX_THREAD_ID` is available:

```bash
agents-attachments materialize
```

## Rules

`.rules/default.rules` — pre-approved Codex CLI commands:
- PlantUML download and rendering
- Temporary directory creation

## How It Works

After running `setup.sh`:

```
~/.agents/
├── skills/
│   ├── alexis-agents-infra -> ~/.agents
│   ├── skill-creator -> ~/.agents/.skills/skill-creator
│   └── ...

~/.claude/
├── CLAUDE.md           # Points to ~/.agents/.instructions/INSTRUCTIONS.md
├── instructions/ -> ~/.agents/.instructions/
└── skills/
    ├── alexis-agents-infra -> ~/.agents/skills/alexis-agents-infra
    ├── skill-creator/ -> ~/.agents/skills/skill-creator
    └── ...

~/.codex/
├── AGENTS.md -> ~/.agents/.instructions/AGENTS.md
├── config.toml -> ~/.agents/.configs/codex-config.toml
├── skills/
│   └── ... -> ~/.agents/skills/...
└── rules/
    └── default.rules -> ~/.agents/.rules/default.rules
```

## Adding New Skills

1. Create skill in `.skills/<skill-name>/`
2. Add `SKILL.md` with frontmatter
3. Run `setup-symlinks.sh` to propagate

Or use the `skill-creator` skill:

```
/skill-creator
```

## Updating Instructions

Edit files in `.instructions/`, changes apply immediately (symlinked).

## Git

This repo is version-controlled. Commit your changes:

```bash
cd ~/.agents
git add -A
git commit -m "Update skills/instructions"
git push
```
