# TASK-260625-uwtoga Skill Source Rule Update

## Change

Added a skill-agnostic global instruction section to
`.instructions/INSTRUCTIONS_SKILLS.md`:

- do not build project-local workarounds around broken reusable skill/tool
  contracts;
- identify the real source repository for the reusable skill/tool;
- fix the source and run its setup/install flow;
- keep only temporary probes in project `.temp/`;
- document source path, setup command, and verification evidence in the normal
  tracked task flow.

The rule intentionally does not enumerate individual skills or tools.

## Setup

Global runtime was refreshed with:

```bash
agents-infra setup global --source-dir /Users/alexis/src/relux-works/alexis-agents-infra
```

VideoCall project-local runtime was refreshed with:

```bash
agents-infra setup local \
  -source-dir /Users/alexis/src/relux-works/alexis-agents-infra \
  -project-dir /Users/alexis/src/videocall/ios
```

## Verification

- `git diff --check -- .instructions/INSTRUCTIONS_SKILLS.md`: passed.
- `agents-infra doctor global`: passed.
- `agents-infra doctor local /Users/alexis/src/videocall/ios`: passed.
- Rendered files contain `Fix Reusable Workflow Contracts At Source`:
  - `/Users/alexis/.agents/.instructions/INSTRUCTIONS_SKILLS.md`
  - `/Users/alexis/.codex/AGENTS.md`
  - `/Users/alexis/src/videocall/ios/.agents/.instructions/INSTRUCTIONS_SKILLS.md`
  - `/Users/alexis/src/videocall/ios/.codex/AGENTS.md`
  - `/Users/alexis/src/videocall/ios/AGENTS.md`
