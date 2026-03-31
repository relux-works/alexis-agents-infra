#!/usr/bin/env zsh

set -euo pipefail

export PATH="/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:${PATH:-}"

MODE="install"

show_usage() {
  cat <<EOF
Usage: ./.scripts/setup-pdf-tools.sh [--check] [--help]

Options:
  --check       Only verify tool availability, do not install
  --help, -h    Show this help

Managed toolchain:
  - pandoc
  - weasyprint
  - poppler (for pdftotext/pdfinfo)
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check)
      MODE="check"
      shift
      ;;
    --help|-h)
      show_usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      show_usage >&2
      exit 1
      ;;
  esac
done

need_cmd() {
  command -v "$1" >/dev/null 2>&1
}

report_status() {
  local name="$1"
  local formula="$2"
  if need_cmd "$name"; then
    echo "OK: $name -> $(command -v "$name")"
  else
    echo "MISSING: $name (formula: $formula)"
  fi
}

maybe_install_formula() {
  local formula="$1"
  if brew list --versions "$formula" >/dev/null 2>&1; then
    echo "Already installed: $formula"
    return 0
  fi
  echo "Installing: $formula"
  brew install "$formula"
}

if [[ "$MODE" == "check" ]]; then
  report_status pandoc pandoc
  report_status weasyprint weasyprint
  report_status pdftotext poppler
  report_status pdfinfo poppler
  exit 0
fi

if ! need_cmd brew; then
  echo "Homebrew is required to install the PDF toolchain automatically." >&2
  echo "Install Homebrew first, or install these packages manually: pandoc, weasyprint, poppler" >&2
  exit 1
fi

maybe_install_formula pandoc
maybe_install_formula weasyprint
maybe_install_formula poppler

echo ""
echo "PDF toolchain status:"
report_status pandoc pandoc
report_status weasyprint weasyprint
report_status pdftotext poppler
report_status pdfinfo poppler
