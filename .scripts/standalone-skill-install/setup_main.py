#!/usr/bin/env python3
from __future__ import annotations

import argparse
import sys
from pathlib import Path

from setup_support import (
    SUPPORTED_LOCALE_MODES,
    InstallResult,
    SetupError,
    perform_install,
)


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="setup.sh",
        description="Install a standalone skill into global or project-local agent environments.",
    )
    subparsers = parser.add_subparsers(dest="mode", required=True)

    locale_help = (
        "Locale mode for installed metadata. "
        f"Supported: {', '.join(SUPPORTED_LOCALE_MODES)}. "
        "Required on first install, optional on reruns when an install manifest already exists."
    )

    global_parser = subparsers.add_parser("global", help="Install symlinks into ~/.claude/skills and ~/.codex/skills")
    global_parser.add_argument("--locale", help=locale_help)

    local_parser = subparsers.add_parser(
        "local",
        help="Copy the skill into <repo>/.skills/<skill-name> and link project-local agent dirs to that copy",
    )
    local_parser.add_argument("repo_path", help="Path to the git repository root or any path inside that repository")
    local_parser.add_argument("--locale", help=locale_help)

    return parser


def print_result(result: InstallResult) -> None:
    print(f"Installed {result.skill_name}")
    print(f"  Source: {result.source_dir}")
    print(f"  Locale: {result.locale_mode}")
    if result.install_mode == "global":
        print(f"  Managed copy: {result.runtime_dir}")
    else:
        print(f"  Project copy: {result.runtime_dir}")
    print(f"  Claude skill link: {result.claude_link}")
    print(f"  Codex skill link: {result.codex_link}")


def main(argv: list[str] | None = None) -> None:
    parser = build_parser()
    args = parser.parse_args(argv)
    current_skill_dir = Path(__file__).resolve().parent.parent

    try:
        if args.mode == "global":
            result = perform_install(
                source_dir=current_skill_dir,
                install_mode="global",
                requested_locale=args.locale,
            )
        else:
            result = perform_install(
                source_dir=current_skill_dir,
                install_mode="local",
                requested_locale=args.locale,
                repo_root=Path(args.repo_path).expanduser(),
            )
    except SetupError as exc:
        print(str(exc), file=sys.stderr)
        raise SystemExit(1) from exc

    print_result(result)


if __name__ == "__main__":
    main()
