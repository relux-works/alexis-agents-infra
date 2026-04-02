# Skills & Plugins

## Using Skills (IMPORTANT)

**Before performing any technical task, check if a relevant skill exists.**

* Search `agents/skills/`, `.claude/skills/`, `.codex/skills/` for applicable skills.
* If a skill covers the task — **read and follow it**, don't improvise.
* Skills exist to ensure consistent patterns; ignoring them leads to mistakes.

Example: before extracting screenshots, check if `ios-ui-validation` skill has instructions for this.

---

## Skill Locations

| Tool | Global | Project-local |
|------|--------|---------------|
| Claude Code | `~/.claude/skills/` | `.claude/skills/` |
| Codex CLI | `~/.codex/skills/` | `.codex/skills/` |

---

## Editing Shared Skills & Instructions (CRITICAL)

* `~/.agents/`, `~/.claude/`, and `~/.codex/` are installed runtime locations, **not** the authoring source of truth.
* **Never edit `~/.agents` directly** when changing shared skills, global instructions, configs, or helper scripts.
* First identify the real source repo:
  * shared global infra/instructions/config → `alexis-agents-infra`
  * separately versioned skill/tool → that skill's own source repo
* If the source repo is not obvious, stop and find it before editing anything.
* Make the change in the source repo, then sync/install it into `~/.agents` with the repo's setup/install flow.
* Treat direct edits inside `~/.agents` as local hotfixes only when the user explicitly asks for that, and say clearly that the source repo still needs the same change.

---

## Creating Skills (Our Pattern)

**Always use the `agents/skills/` pattern:**

1. Create the actual skill in `agents/skills/<skill-name>/`
2. Symlink to `.claude/skills/` and `.codex/skills/`

```bash
# Create skill
mkdir -p agents/skills/my-skill
# ... add SKILL.md, references/, assets/, etc.

# Symlink for Claude Code
mkdir -p .claude/skills
ln -s ../../agents/skills/my-skill .claude/skills/my-skill

# Symlink for Codex CLI
mkdir -p .codex/skills
ln -s ../../agents/skills/my-skill .codex/skills/my-skill
```

**Why this pattern:**
* `agents/skills/` is visible in Finder (no dot prefix)
* Single source of truth, works for both Claude Code and Codex CLI
* Skills are checked into git via `agents/`

For global skills, create in `~/agents/skills/` and symlink to `~/.claude/skills/` and `~/.codex/skills/`.

If a standalone skill repo needs an install flow with localized installed
metadata, use the canonical vendorable helper from:

```text
alexis-agents-infra/.scripts/standalone-skill-install/
```

Vendor that helper into the standalone skill repo. Do not create a runtime
dependency on `alexis-agents-infra`.

---

## Skill Structure & Rules

* Agent may create custom skills when there is a **clear, justified need** for specialized workflows.
* Use the `skill-creator` skill to scaffold new skills if available.

#### Skill structure

```
skill-name/
├── SKILL.md              # required: frontmatter + instructions
├── scripts/              # optional: executable code
├── references/           # optional: docs/schemas
└── assets/               # optional: templates/resources
```

#### Skill creation rules

* Skills must be **focused and narrow** (one clear purpose).
* Keep `SKILL.md` under 500 lines; split into references if needed.
* Test any included scripts before committing.
* Document skill purpose, triggers, and usage clearly in `description` field.
* Prefer explicit `triggers:` in `SKILL.md` frontmatter for natural-language activation.
* If install-time localization is used, localize both `description` and `triggers`.

#### Evolving skills with tooling

* When using a skill and **creating ad-hoc tooling** (scripts, commands, patterns) to accomplish the task:
  * If the tooling is **reusable across projects** — **add it to the skill immediately**.
  * Include it in `scripts/` or document it in `SKILL.md`.
  * This ensures the next time the skill is used, the tooling is already available.
* Skills should **accumulate useful tooling** over time, not just document patterns.

### Plugins

* **Plugins** bundle multiple extensions (commands, agents, skills, hooks, MCP servers).
* Agent may use plugins from official/community marketplaces when appropriate:
  * Install via `/plugin install @namespace/plugin-name`
  * Prefer official Anthropic plugins for standard workflows (code review, feature dev, etc.)
  * Document installed plugins in `.temp/plugins.md` with rationale.

---

## Codex CLI

* **Skills** use the same structure as Claude Code, but different paths (`.codex/skills/`).
* Codex also checks `$CWD/../.codex/skills` and `$REPO_ROOT/.codex/skills`.
* Priority: CWD > repo root > user (`~/.codex/skills/`) > system.
* No plugin system in Codex CLI — use skills only.
