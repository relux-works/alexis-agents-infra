---
name: pdf
description: Render Markdown and HTML documents to PDF using pandoc plus WeasyPrint with shared themes and a reproducible CLI workflow. Use when creating research papers, one-pagers, printable reports, or any artifact that should end up as a PDF.
triggers:
  - pdf
  - render pdf
  - export pdf
  - markdown to pdf
  - html to pdf
  - weasyprint
  - pandoc pdf
  - printable artifact
  - pdf theme
  - сгенерировать pdf
  - сделать pdf
  - экспорт в pdf
  - тема для pdf
---

# pdf

Use this skill when the output should be a real PDF artifact rather than just markdown.

This skill provides:

- a reusable renderer script: `scripts/render-pdf.sh`
- a minimal HTML template: `assets/template.html5`
- shared CSS themes in `assets/themes/`
- an optional workstation bootstrap path via the `alexis-agents-infra` source repo:
  - `./setup.sh --with-pdf-tools`
  - `./.scripts/setup-pdf-tools.sh`

## Toolchain

Preferred stack:

- `pandoc` for Markdown -> HTML normalization
- `weasyprint` for HTML -> PDF rendering
- `pdftotext` for validation that text survived rendering

From the `alexis-agents-infra` source repo, install them with:

```bash
./setup.sh --with-pdf-tools
```

Preflight-check the toolchain with:

```bash
./.scripts/setup-pdf-tools.sh --check
```

Or install/check only the PDF toolchain without reinstalling the whole runtime:

```bash
./.scripts/setup-pdf-tools.sh
./.scripts/setup-pdf-tools.sh --check
```

## Quick Start

From this skill directory, render a markdown file with the default theme:

```bash
scripts/render-pdf.sh notes/report.md -o .temp/report.pdf
```

Render with an explicit theme and title:

```bash
scripts/render-pdf.sh notes/report.md \
  -o artifacts/report.pdf \
  --theme report-clean \
  --title "Quarterly Research Report"
```

List available themes:

```bash
scripts/render-pdf.sh --list-themes
```

The renderer can be called from the source repo or via the installed runtime
skill link such as `~/.agents/skills/pdf/scripts/render-pdf.sh`. It resolves
its template and bundled themes relative to the script location.

## Themes

Current built-in themes:

- `prose-classic` — serif, print-like, good for papers and long-form prose
- `report-clean` — cleaner sans-serif report style for short memos and executive summaries

If the task needs a custom look:

- start from one of the built-in CSS files in `assets/themes/`
- keep page size and margins explicit via `@page`
- prefer stable system fonts or fonts already installed by the environment
- avoid depending on remote assets

## Validation Workflow

After rendering:

1. Check the file exists and has non-trivial size.
2. Run `pdfinfo output.pdf` to confirm page count and metadata.
3. Run `pdftotext output.pdf -` and inspect the first section to confirm text was not dropped.
4. If the render fails in a clean shell, re-run `./.scripts/setup-pdf-tools.sh --check` and confirm `pandoc` plus `weasyprint` are visible on `PATH`.

Store generated PDFs in project artifact locations such as `.temp/` unless the repo has a more specific package/output convention.

## Notes

- The renderer resolves relative images and links from the source markdown directory.
- If the PDF looks truncated, inspect the markdown source first, especially tables and raw HTML blocks.
- For official filings or strict submission formats, treat the built-in themes as a baseline and adapt a dedicated theme for that format.
