---
name: phase-handoff-playbook
description: Use when pausing a multi-step phase or wave and you need a verified, restartable handoff artifact with concrete next actions.
---

# Phase Handoff Playbook

## Purpose

Handoff is an execution-control artifact, not a narrative summary.

If tomorrow's first command and first task are not obvious, handoff is incomplete.

## Hard Rules

1. Do not claim handoff complete without fresh verification evidence.
2. Do not hand off without explicit `worktree path + branch + head sha`.
3. Do not mix completed work and unfinished work in one list.
4. Do not hand off parallel tasks without disjoint write scopes.
5. Do not skip merge-safety status (`safe_to_merge` / `blocked` + reason).

## When To Use

- End of day on any active phase branch
- Before switching to another branch/worktree
- Before asking another human/agent to continue
- Before merge window planning

For tiny one-file tasks with zero unfinished work, keep a short checkpoint note instead.

## Required Output

Write one handoff file under:

- `docs/phase-handoff-playbook/YYYY-MM-DD-<topic>-handoff.md`

The handoff must include:

- exact repository snapshot
- fresh verification commands + pass/fail
- completed vs unfinished split
- tomorrow kickoff commands
- smallest-next-task checklist with acceptance criteria
- merge safety status + release/PR state
- risks and "do not do next"

## Procedure

1. Snapshot
- Record `pwd`, active branch, `HEAD`, `origin/main`, and phase branch remote SHA.
- Record open PR status and latest release tag if relevant.

2. Verify
- Rerun the real verification commands in current session.
- Copy command strings and concise outcomes into handoff.

3. Separate work
- Completed today
- Still unfinished
- Blockers/dependencies

4. Build smallest next tasks
- Every task must include:
- id, priority, why, acceptance, file scope, dependency, parallel-safe flag

5. Tomorrow kickoff
- First commands only; make startup deterministic.
- Name the very first task to execute.

6. Quality gate
- Run `skills/phase-handoff-playbook/handoff-quality-checklist.md`.
- If any critical item fails, handoff is not done.

## Fast Mode (10 Minutes)

Use this only when time is tight:

1. Fill snapshot + verification + completed/unfinished.
2. Define 3-5 smallest tasks with acceptance.
3. Add merge safety and first task tomorrow.
4. Run checklist critical items only.

## Companion Assets

- `handoff-template.md`
- `handoff-quality-checklist.md`
- `resume-checklist.md`
- `parallel-prompts.md`

## Anti-Patterns

- "continue phase2" with no file scope
- stale test results from earlier session
- missing remote references (`origin/main` changed but not acknowledged)
- backlog items too large to start immediately
- handoff file exists but no deterministic kickoff steps
