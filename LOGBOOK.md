# Flight Logbook

> Institutional memory. Concise, factual, high-signal.
> Newest entries first. One block per insight.

## 2026-07-10

### 1155 — Board-Agnostic Image Intake
- DECISION: Image intake belongs in `.scripts/agents-attachments` as `stage-images`, not in board-specific scripts or resources.
- DECISION: Source-to-staged mappings persist redacted source labels plus content hashes; raw ICCID/IMSI/key-like labels are not written into staged filenames.
- SCOPE: `.scripts/agents-attachments`, `.instructions/INSTRUCTIONS_ATTACHMENTS.md`, `SKILL.md`, `README.md`, `tests/test_agents_attachments.py`.
