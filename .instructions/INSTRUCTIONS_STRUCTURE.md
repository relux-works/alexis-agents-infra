# Project Structure & Environment

## Shell & Environment

* Shell: **zsh**
* Shell framework: **oh-my-zsh**

---

## Project Structure

* `.temp/` — temporary utilities, logs, debugging artifacts (**gitignored**).
  * Use task-scoped subdirectories when possible, especially for worktrees and smoke artifacts: `.temp/<TASK-ID>/...`
  * Temporary `git worktree` checkouts should live under `.temp/`, not beside the main repo checkout.
* `.scripts/` — reusable scripts and utilities for the project.
* `diagrams/` (project root) — all diagrams live here.
  * Use subfolders when diagrams start to split by module or purpose (keep hierarchy clean and obvious).
