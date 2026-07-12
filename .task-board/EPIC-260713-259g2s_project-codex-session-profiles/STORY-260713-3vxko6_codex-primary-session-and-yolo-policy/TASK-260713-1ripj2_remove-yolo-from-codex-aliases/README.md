# Remove yolo from Codex aliases

## Description
Move persistent yolo behavior from shell aliases to the verified agents-infra project profile

## Scope
Locate the canonical tracked owner of Codex aliases or shell functions that currently append -d, --danger, --yolo, or the native dangerous flag to agents-infra codex. After both target projects have the verified yolo_mode=true primary profile, remove only that redundant flag from the normal aliases. Preserve an explicit opt-in escape hatch for ad-hoc yolo if one already exists. Do not edit installed runtime copies, print secrets, or change aliases unrelated to Codex.

## Acceptance Criteria
The canonical alias source is identified and changed only if it contains the redundant yolo flag. A fresh zsh lookup proves the normal alias launches agents-infra codex without a danger argument, while from both configured project roots print-config still renders exactly one native danger flag from yolo_mode=true. Existing raw or explicit-danger entry points remain intentional and documented. The outcome names the source owner, exact smoke commands, and any no-op reason.
