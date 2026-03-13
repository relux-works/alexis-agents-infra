# Workflow

## Priority: Scope Control Over Speed

* **Task tracking and scope control always come first.** Never skip creating a task "because it's small" or "just a quick fix".
* If a board exists (`.task-board/`), **every code change must have a task on the board BEFORE implementation starts.** No exceptions for size — a one-line fix gets a task just like a multi-story epic.
* The workflow is: **trigger → skill → board → implement → review → reopen or close.** Never jump straight to implementation because it feels faster.
* Speed is a side effect of good process, not a substitute for it. A change done fast but untracked is worse than a change done properly.

---

## Version Control

* **Never commit or stage files automatically.**
* When work is ready to commit, stop and ask for review.
* **Never add Co-Authored-By lines** or any AI attribution to commits.
* When you need to work on multiple revisions, parallel fixes, or isolated experiments in the same repo, prefer **`git worktree`** over juggling branches in one checkout.
* Place temporary worktrees under the project's **`.temp/`** directory, not next to the main checkout.
* If the worktree is for a tracked task, place it under a task-scoped temp path using the task ID:
  * `.temp/<TASK-ID>/worktree/`
  * or `.temp/<TASK-ID>/<repo-name>-worktree/`
* This keeps the main checkout stable while making task-local scratch state easy to find and clean up.

---

## Task Tracking

* For projects with `project-management` skill active — use `task-board` CLI (epics, stories, tasks). Don't use `.temp/tasks.md`.
* For projects without a board — create a task plan in `.temp/tasks.md` before starting work.
* Track progress in the same file.
* Update/append to the existing plan — **don't create new task files each session**.
* Purpose: resume smoothly if the session breaks.

---

## Research & Knowledge Persistence

* **All research must go through the board and documentation.** Never keep research only in conversation context.
* If a board exists — store research findings in `artifacts/RESEARCH.md` inside the relevant element's directory. Link from parent element's notes.
* If no board — store in `.temp/` with descriptive names (`research-auth-flow.md`, `analysis-performance.md`).
* **Why:** Context windows collapse. If research lives only in the conversation, it's lost forever when the session resets. Files persist.
* Sub-agents doing research/analysis **must** write their findings to files before finishing.
* Reference research artifacts from task notes: `task-board progress notes TASK-XX "See artifacts/RESEARCH.md"`.

---

## Logging

* Store logs in `.temp/` with numbered naming:
  * `pod-lint-01.log`, `spm-build-01.log`, etc.
* Document log locations in your notes/tasks.
