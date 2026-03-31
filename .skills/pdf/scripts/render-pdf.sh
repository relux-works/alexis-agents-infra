#!/usr/bin/env bash
set -euo pipefail

export PATH="/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:${PATH:-}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ASSETS_DIR="$SKILL_DIR/assets"
THEMES_DIR="$ASSETS_DIR/themes"
DEFAULT_THEME="prose-classic"
DEFAULT_TEMPLATE="$ASSETS_DIR/template.html5"

usage() {
  cat <<'EOF'
Render a Markdown or HTML file to PDF using pandoc + WeasyPrint.

Usage:
  render-pdf.sh <input> [-o output.pdf] [--theme NAME] [--css FILE] [--title TEXT]
  render-pdf.sh --list-themes

Defaults:
  theme      -> prose-classic
  template   -> assets/template.html5
  output.pdf -> next to the input file
EOF
}

list_themes() {
  find "$THEMES_DIR" -maxdepth 1 -type f -name '*.css' -print \
    | sed 's#^.*/##' \
    | sed 's/\.css$//' \
    | sort
}

INPUT=""
OUTPUT=""
CSS=""
TITLE=""
THEME="$DEFAULT_THEME"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -o|--output)
      OUTPUT="${2:-}"
      shift 2
      ;;
    --theme)
      THEME="${2:-}"
      shift 2
      ;;
    --css)
      CSS="${2:-}"
      shift 2
      ;;
    --title)
      TITLE="${2:-}"
      shift 2
      ;;
    --list-themes)
      list_themes
      exit 0
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    -*)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
    *)
      if [[ -n "$INPUT" ]]; then
        echo "Only one input file is supported" >&2
        exit 1
      fi
      INPUT="$1"
      shift
      ;;
  esac
done

if [[ -z "$INPUT" ]]; then
  usage >&2
  exit 1
fi

for cmd in pandoc weasyprint; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "Missing required tool: $cmd" >&2
    echo "Install the PDF toolchain with ./setup.sh --with-pdf-tools" >&2
    echo "Or verify it with ./.scripts/setup-pdf-tools.sh --check" >&2
    exit 1
  fi
done

if [[ -z "$CSS" ]]; then
  CSS="$THEMES_DIR/${THEME}.css"
fi

if [[ ! -f "$INPUT" ]]; then
  echo "Input file not found: $INPUT" >&2
  exit 1
fi

if [[ ! -f "$CSS" ]]; then
  echo "Theme/CSS file not found: $CSS" >&2
  echo "Available themes:" >&2
  list_themes >&2
  exit 1
fi

if [[ -z "$OUTPUT" ]]; then
  OUTPUT="${INPUT%.*}.pdf"
fi

mkdir -p "$(dirname "$OUTPUT")"

INPUT_ABS="$(cd "$(dirname "$INPUT")" && pwd)/$(basename "$INPUT")"
OUTPUT_ABS="$(cd "$(dirname "$OUTPUT")" && pwd)/$(basename "$OUTPUT")"
CSS_ABS="$(cd "$(dirname "$CSS")" && pwd)/$(basename "$CSS")"
INPUT_DIR="$(dirname "$INPUT_ABS")"

TMP_HTML="$(mktemp "${TMPDIR:-/tmp}/render-pdf.XXXXXX.html")"
cleanup() {
  rm -f "$TMP_HTML"
}
trap cleanup EXIT

PANDOC_ARGS=(
  "$INPUT_ABS"
  --standalone
  --template "$DEFAULT_TEMPLATE"
  -o "$TMP_HTML"
)

case "$INPUT_ABS" in
  *.md|*.markdown)
    PANDOC_ARGS=( "$INPUT_ABS" --from gfm --to html5 --standalone --template "$DEFAULT_TEMPLATE" -o "$TMP_HTML" )
    ;;
  *.html|*.htm)
    PANDOC_ARGS=( "$INPUT_ABS" --from html --to html5 --standalone --template "$DEFAULT_TEMPLATE" -o "$TMP_HTML" )
    ;;
  *)
    echo "Unsupported input type: $INPUT_ABS" >&2
    echo "Supported extensions: .md, .markdown, .html, .htm" >&2
    exit 1
    ;;
esac

if [[ -n "$TITLE" ]]; then
  PANDOC_ARGS+=(--metadata "title=$TITLE")
fi

pandoc "${PANDOC_ARGS[@]}"

weasyprint \
  --base-url "$INPUT_DIR" \
  --stylesheet "$CSS_ABS" \
  "$TMP_HTML" \
  "$OUTPUT_ABS"

echo "PDF ready: $OUTPUT_ABS"
