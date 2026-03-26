# Phase Handoff Playbook

Use this workflow when pausing a non-trivial delivery wave so the next session can continue without rediscovery.

## Quick Start

1. Scaffold a dated handoff document:
   - `./scripts/new_phase_handoff.sh <topic-slug>`
   - compatibility alias: `./scripts/phase_handoff_new.sh <topic-slug>`
2. Fill in the generated markdown under `docs/phase-handoff-playbook/`.
3. Run the quality gate:
   - `./scripts/phase_handoff_check.sh <handoff-doc-path>`
   - or `make handoff-check HANDOFF_DOC=<handoff-doc-path>`

## Purpose

- capture exact execution snapshot (`worktree`, `branch`, `head sha`, remote refs)
- provide fresh verification evidence from the current session
- separate completed and unfinished scope
- define a smallest next task with acceptance criteria
- declare merge safety (`safe_to_merge` / `blocked`)

## Output Location Convention

Keep project-specific artifacts only in this folder:

- `docs/phase-handoff-playbook/YYYY-MM-DD-<topic>-handoff.md`
- `docs/phase-handoff-playbook/YYYY-MM-DD-<phase>-next-work-checklist.md`

Keep reusable process assets under:

- `skills/phase-handoff-playbook/`

## Required Handoff Content

- exact worktree path, branch, and `head sha`
- `origin/main` sha and phase branch remote sha
- fresh verification command and result
- completed vs unfinished scope
- tomorrow kickoff steps
- smallest-next-task checklist with acceptance criteria
- explicit merge safety status (`safe_to_merge` or `blocked` with reason)
- risks and "do not do next" traps

## Resume Next Session

At the start of the next session:

1. Run `skills/phase-handoff-playbook/resume-checklist.md`.
2. Optionally run `./scripts/phase_resume_from_latest.sh` for a quick startup summary.

If assumptions drifted (branch, worktree, SHAs, verification result), update the handoff before coding.
