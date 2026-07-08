---
title: Agents Infra
summary: The shared configuration, instructions, skills, and rules behind our agents.
category: Agent infrastructure
featured: true
---

## What it is

The source of truth for how our AI agents are configured. One repository holds the
shared instructions, skills, rules, and tool configuration that install into both Claude
Code (`~/.claude/`) and Codex CLI (`~/.codex/`), with a bootstrap installer that syncs
the global runtime and per-project setup.

## Why it matters

Agentic development only compounds if the setup is reproducible. Instead of every
engineer hand-tuning their assistant, the whole team runs from one versioned baseline:
the same guardrails, the same skills, the same conventions. This is the substrate our
agentic enablement work is built on, and the reason our own agents behave consistently
across machines and projects.

## Who it is for

Teams standardizing how their engineers work with coding agents, who want configuration
as code rather than tribal knowledge.
