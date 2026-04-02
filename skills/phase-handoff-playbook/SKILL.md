---
name: phase-handoff-playbook
description: Use when pausing a multi-step or multi-agent delivery wave and you need a verified, restartable handoff artifact with concrete next actions.
---

# Phase Handoff Playbook

## Overview

End-of-wave handoff is not a diary entry. It is a control artifact.

Core principle: verify first, snapshot second, backlog third, reflection fourth.

This skill directory contains only reusable process assets. Project-specific outputs created by this skill belong under `docs/phase-handoff-playbook/`, not under `skills/`.

## Asset Boundary

Keep these boundaries strict:

- `skills/phase-handoff-playbook/`
  - reusable process guidance
  - reusable prompt fragments
  - reusable quality gates
  - reusable templates
- `docs/phase-handoff-playbook/`
  - branch-specific handoffs
  - project-specific retrospectives
  - next-work checklists
  - execution summaries

If a file contains branch names, real paths, current blockers, or project-specific backlog items, it belongs in `docs/phase-handoff-playbook/`.

## Purpose

If tomorrow's first command and first task are not obvious, the handoff is incomplete.

## Hard Rules

1. Do not claim handoff complete without fresh verification evidence.
2. Do not hand off without explicit `worktree path + branch + head sha`.
3. Do not skip remote refs (`origin/main` and phase branch remote sha).
4. Do not skip merge safety status (`safe_to_merge` / `blocked` + reason).
5. Do not hand off parallel tasks without disjoint write scopes.

## When To Use

- ending a day on a non-trivial branch
- pausing after multi-commit or multi-agent execution
- switching worktree/branch ownership
- preparing merge or release readiness review

Do not use this for a tiny single-file change with no unfinished work.

## Required Outputs

- one handoff document under `docs/phase-handoff-playbook/`
- explicit snapshot: worktree, branch, head sha, remote refs
- explicit fresh verification evidence
- completed vs unfinished scope split
- smallest-next-task checklist with acceptance criteria
- clear residual risks and "do not do next" traps

## Workflow

1. Verify actual state:
   - active worktree path
   - active branch
   - active worktree `git status -sb`
   - root repo `git status -sb` if multiple worktrees exist
   - fresh verification command output (`go test ./...` by default)
2. Record snapshot:
   - `head sha`
   - `origin/main sha`
   - phase branch remote sha
   - merge safety status
3. Separate completed work from remaining work.
4. Split remaining work into independent slices with disjoint write scopes where possible.
5. Break remaining work into smallest actionable tasks with dependencies.
6. Write tomorrow kickoff steps and first task.
7. Run the handoff quality gate before finishing.

## Parallel Continuation Pattern

Use this when the next session will likely resume with multiple agents.

1. Split work by disjoint write scope.
2. Identify which slice is the bottleneck.
3. Write one owner statement per slice.
4. Predeclare "do not touch" files.
5. Predeclare verification command per slice.
6. Save prompt-ready task summaries in the handoff doc.

Never hand off "keep going on Phase 1" without file boundaries.

## Project Integration

This repository includes command-line helpers:

- `./scripts/new_phase_handoff.sh <topic-slug>`
- `./scripts/phase_handoff_check.sh <handoff-doc-path>`
- `make handoff-check HANDOFF_DOC=<handoff-doc-path>`

Use these to avoid ad-hoc handoffs.

## Reusable Assets

Use these companion files:

- `handoff-template.md`
- `handoff-quality-checklist.md`
- `resume-checklist.md`
- `parallel-prompts.md`
- `examples/feature-wave-handoff-example.md`
- `examples/parallel-platform-wave-example.md`

## Quick Output Shape

Use `handoff-template.md`, then copy/adapt it into:

- `docs/phase-handoff-playbook/YYYY-MM-DD-<topic>-handoff.md`
- or `docs/phase-handoff-playbook/YYYY-MM-DD-<phase>-next-work-checklist.md`

At minimum, the handoff doc should contain:

- current snapshot
- fresh verification
- completed today
- unfinished work
- next-day startup steps
- smallest-next-task checklist
- reflection

Before finishing, walk the document against `handoff-quality-checklist.md`.
