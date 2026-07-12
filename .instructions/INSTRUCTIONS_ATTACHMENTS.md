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
mkdir -p .temp/attachments
cp "$(agents-attachments path screenshot.png)" .temp/attachments/
agents-attachments stage-images screenshot.png --out-dir .temp/image-intake
```

## Image intake and inspection

Use this workflow for user-provided screenshots, photos, scans, SIM/package photos,
and other image evidence. It is board-agnostic and works with explicit local
paths or the generic attachments manifest.

Stage first, inspect staged files:

```bash
agents-attachments stage-images ./photo.heic screenshot.png --out-dir .temp/image-intake
agents-attachments stage-images --all --manifest .temp/agents-attachments-manifest.json --out-dir .temp/image-intake
```

Behavior:

- accepts explicit local paths and manifest references by `id`, `name`, or local file name
- treats originals as read-only inputs and writes staged files under `--out-dir`
  (default `.temp/image-intake`)
- writes and prints a machine-readable mapping JSON
  (default `<out-dir>/image-stage-map.json`)
- copies normal image formats as-is
- normalizes HEIC/HEIF to PNG with macOS `sips` when available, then falls back
  to ImageMagick (`magick`, then `convert`) on other platforms or when `sips`
  is unavailable
- fails clearly if HEIC/HEIF conversion is required but no converter is available
- redacts ICCID/IMSI/key-like strings from persisted source labels and staged
  filenames while keeping content hashes for auditability

Inspection rules:

- Prefer direct visual inspection through the runtime's vision or image-viewing
  capability before OCR.
- Use OCR only as a bounded fallback when direct inspection cannot answer the
  question. Do not run unbounded OCR/upscale/retry loops; keep fallback attempts
  narrow and task-scoped.
- Tie every observation to a staged filename and mapping entry.
- Record per-file evidence, confidence (`high`, `medium`, `low`), uncertainty,
  and any redactions applied.
- Redact ICCID, IMSI, QR payloads, activation codes, tokens, keys, passwords,
  and similar secrets before writing notes, docs, task resources, logs, or
  human-facing output. Do not persist raw sensitive identifiers extracted from
  images.

Child-agent handoff:

- Pass the staged image directory and mapping JSON path.
- Propagate `AGENTS_ATTACHMENTS_MANIFEST` only when the child also needs the
  original generic manifest.
- Do not require a board, board resource, task ID, or status field for image
  intake.

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
