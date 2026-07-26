# Claude Guidelines for Vault

How Claude should work on this repository.

## Overview

Vault is a minimal, cross-platform file manager: search-first navigation, zero chrome,
keyboard-first, instant everything. It aims to replace Finder (macOS) and Windows File Explorer.
Stack: **Wails v3** (Go backend + native webview) + **React 18 + TypeScript + Tailwind** frontend.

**Status:** pre-implementation. See `docs/STATUS.md` for build-order progress.

**Primary source of truth:** `.cursor/rules/` (authoritative) and `docs/SPEC.md` (design intent —
product vision, layout, screens, non-goals, performance targets). Claude follows the cursor rules
directly.

## Architecture — UI-agnostic core, per-OS adapters

Keep file-system logic (list, search, sort, copy/move/delete) behind a small, platform-neutral
interface. OS-specific behavior (trash/recycle-bin semantics, path quirks, native dialogs,
file-watching APIs) lives in one adapter per OS. The shared UI talks only to the core interface,
never to a platform adapter directly.

**Reference:** `.cursor/rules/architecture.mdc`

## Testing — TDD, isolated file-system tests

- **TDD** for behavior changes: failing test first → implement → refactor.
- File-operation tests run against an isolated temp directory / fake, never the host's real file
  system. Cover destructive paths explicitly (missing file, permission denied, file in use,
  cross-volume move).
- Co-located tests; keep touched areas well covered (aim ≥80%).

**Reference:** `.cursor/rules/unit-tests-tdd.mdc`

## Response discipline

Code-first, minimal prose; patch-based edits (changed lines + a little context); no summaries of a
visible diff; no affirmations. Facts only, no validation, small phrases, bullet facts over prose.
**Reference:** `.cursor/rules/core.mdc`, `token-optimization.mdc`, `minimal-communication.mdc`.

## Git workflow

Claude **prepares** commits; the human **executes** them. Read-only inspection only
(`git status`/`diff`/`log`); draft a `why`-focused Conventional-Commits message; provide copy-paste
`git add`/`git commit`. Never run `git commit`/`git add`(staging)/`push --force`/skip hooks.
**Reference:** `.cursor/rules/git-commits.mdc`

## When to apply rules

| Situation | Rule File (`.cursor/rules/`) |
|-----------|------------------------------|
| Every request | `core.mdc`, `token-optimization.mdc`, `minimal-communication.mdc` |
| Any development task | `agentic-loop.mdc` (understand→locate→plan→implement→validate→self-review) |
| Core/UI/platform-adapter boundary changes | `architecture.mdc` |
| Behavior changes | `unit-tests-tdd.mdc` (TDD, coverage) |
| Commits/PRs | `git-commits.mdc` |

## Feature development checklist

1. Failing unit test (TDD).
2. Implement in the platform-neutral core; keep OS-specific code in its adapter.
3. Wire the UI to the core interface, not to an adapter directly.
4. Validate: build, lint, and run tests for the chosen stack.
5. No scope creep — only the files the task needs.

## Questions?

1. `.cursor/rules/` — authoritative rules
2. This file — architecture and workflow overview
3. `docs/SPEC.md` — full design spec, screens, non-goals
4. `docs/STATUS.md` — build-order progress
5. `README.md` — project goals and quickstart commands
