# Attachments

Use this contract when the runtime materializes user-provided files for the agent.

## Contract

- Manifest file name: `agents-attachments-manifest.json`
- Environment variable: `AGENTS_ATTACHMENTS_MANIFEST`
- Default project-local fallback: `.temp/agents-attachments-manifest.json`

The manifest may be either:

- a JSON array of attachment objects
- or an object with top-level key `attachments`

Each attachment object should contain:

- `id` — stable attachment identifier
- `name` — original file name
- `mime_type` — media type such as `image/png`
- `size_bytes` — byte size
- `sha256` — optional content hash
- `local_path` — absolute local path to the materialized file

Optional fields:

- `source`
- `created_at`
- `metadata`

## Agent behavior

- Treat `local_path` as the source of truth for file access.
- Treat incoming attachments as read-only inputs.
- Prefer the helper CLI `agents-attachments` over ad hoc JSON parsing when practical.
- If a tool needs a file path, pass `local_path` directly.

Examples:

```bash
agents-attachments materialize
agents-attachments list
agents-attachments path screenshot.png
task-board resource add TASK-123 "$(agents-attachments path screenshot.png)" --type precondition
```

## Codex bootstrap path

If the runtime did not pre-materialize files but Codex rollout history is available, the helper can bootstrap a local manifest from the current thread:

```bash
agents-attachments materialize
```

Behavior:

- uses `CODEX_THREAD_ID` by default
- locates the matching `~/.codex/sessions/**/rollout-*.jsonl`
- extracts `input_image` payloads from user turns
- writes files under `.temp/agents-attachments/`
- writes `.temp/agents-attachments-manifest.json`

Overrides:

- `agents-attachments materialize --thread-id ...`
- `agents-attachments materialize --session /abs/path/to/rollout.jsonl`
- `agents-attachments materialize --out-dir .temp/custom-attachments`
- `agents-attachments materialize --manifest .temp/custom-manifest.json`

## Spawned agents

When a parent process spawns a child agent, it should:

- propagate `AGENTS_ATTACHMENTS_MANIFEST`
- allow access to the directory that contains the materialized files
- keep child agents on the same manifest unless there is an explicit reason to narrow scope

## Runtime boundary

This repo defines the contract and helper tooling. It does not itself ingest chat attachments.

The runtime/launcher is responsible for:

- extracting files from the host conversation layer
- materializing them on disk
- writing `agents-attachments-manifest.json`
- exporting `AGENTS_ATTACHMENTS_MANIFEST`
