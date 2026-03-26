---
name: phase-handoff-playbook
description: Use when pausing a multi-step or multi-agent delivery wave and you need a verified handoff, a smallest-next-task checklist, lessons learned, and continuation guidance for the next session or another project.
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

## When to Use

- End of day on a large branch
- After parallel agent work
- Before switching worktrees or branches
- Before asking another person or model to continue
- When the repo has partial progress, dirty state, and unfinished scope

Do not use this for a tiny single-file change with no unfinished work.

## Required Outputs

- One project-specific handoff markdown file under `docs/phase-handoff-playbook/`
- Fresh verification evidence
- A smallest-next-task checklist with acceptance criteria
- Experience summary and deep reflection
- Reusable guidance/templates remain under `skills/phase-handoff-playbook/`

Optional but strongly recommended:

- A verification matrix for the completed wave
- A risk register with ownerless gaps called out explicitly
- A “do not do next” section to prevent immediate regression
- Prompt-ready text for future parallel continuation

## Procedure

1. Verify actual state
   - Record worktree path and branch.
   - Check both active worktree and root repo status if multiple worktrees exist.
   - Run the full relevant verification command fresh.
   - Do not trust memory or agent reports.
   - If there are long-running background processes, confirm whether they matter to the handoff.

2. Inventory completed vs unfinished work
   - Read status docs, plans, and the touched code.
   - Separate “already landed” from “still planned”.
   - Call out any path drift, plan drift, or root/worktree confusion explicitly.

3. Split remaining work into independent slices
   - Group by bounded context or subsystem.
   - Keep write scopes disjoint if future parallel execution is likely.
   - Identify the true bottleneck and the risky areas.

4. Break each slice into smallest meaningful tasks
   - Each task should be directly actionable in one sitting.
   - Give every task:
     - task id
     - why it matters
     - acceptance criteria
     - suggested files
     - dependency note
     - whether it is safe to parallelize

5. Write tomorrow kickoff steps
   - Include the exact branch/worktree to enter.
   - Include the first verification command.
   - Include the first 3-8 tasks to do in order.
   - Save the resulting project artifact under `docs/phase-handoff-playbook/`.
   - Make the first task operationally obvious.

6. Write experience summary and deep reflection
   - What worked
   - What actually accelerated progress
   - What failed
   - What must be banned next time

7. Capture reusable process
   - If the handoff method exposed repeatable patterns, create or update a skill.
   - Keep the skill generic enough to reuse in other repos.
   - Keep prompts, templates, checklists, and examples reusable.
   - Do not move project-specific retrospectives or branch-specific notes into `skills/`.

8. Run the handoff quality gate
   - Use `handoff-quality-checklist.md`.
   - If any critical item is not satisfied, the handoff is not complete.

## Parallel Continuation Pattern

Use this when the next session will likely resume with multiple agents.

1. Split work by disjoint write scope
2. Identify which slice is the bottleneck
3. Write one owner statement per slice
4. Predeclare “do not touch” files
5. Predeclare verification command per slice
6. Save prompt-ready task summaries in the handoff doc

Never hand off “keep going on Phase 1” without file boundaries. That is not delegation. It is context leakage.

## Reusable Assets

Use these companion files:

- `handoff-template.md`
- `handoff-quality-checklist.md`
- `parallel-prompts.md`
- `resume-checklist.md`
- `examples/feature-wave-handoff-example.md`
- `examples/parallel-platform-wave-example.md`

## Common Mistakes

- Trusting agent “DONE” messages without fresh verification
- Forgetting to state the exact worktree path
- Letting root repo and worktree drift silently
- Writing tasks that are still too large to start immediately
- Updating code but not updating phase-status docs
- Putting project outputs in `skills/` instead of `docs/phase-handoff-playbook/`
- Writing a backlog without acceptance criteria
- Writing priorities without dependency order
- Handing off parallel work without explicit write scopes
- Omitting residual risks because tests are green
- Ending the session without reflections, so the same mistakes repeat

## Quick Output Shape

Use the companion template `handoff-template.md` from this skill directory, then copy/adapt it into:

- `docs/phase-handoff-playbook/YYYY-MM-DD-<topic>-handoff.md`
- or `docs/phase-handoff-playbook/YYYY-MM-DD-<phase>-next-work-checklist.md`

At minimum, the handoff doc should contain:

- current snapshot
- fresh verification
- completed today
- unfinished work
- next-day startup steps
- smallest-next-task checklist
- lessons learned
- deep reflection

Before finishing, walk the document against `handoff-quality-checklist.md`.

At the start of the next session, use `resume-checklist.md`.

## Naming Convention

Recommended output names:

- `docs/phase-handoff-playbook/YYYY-MM-DD-<phase>-next-work-checklist.md`
- `docs/phase-handoff-playbook/YYYY-MM-DD-<topic>-handoff.md`
- `docs/phase-handoff-playbook/YYYY-MM-DD-<topic>-retrospective.md`

Use lowercase kebab-case. Put the date first so files sort chronologically.

## Heuristics

- If scope is still unclear, the handoff is not done.
- If tomorrow’s first task is not obvious, the handoff is not done.
- If a new contributor could reopen the branch and continue in 10 minutes, the handoff is good.
- If the output cannot be reused by another company/team/project without rewriting the skill text, the skill is too project-specific.
- If the handoff does not reduce startup friction tomorrow, it is documentation theater.
