# Resume Checklist

Use this at the start of the next session before writing code.

## Verify Context

- confirm the intended worktree path
- confirm the intended branch name
- inspect `git status --short` in the active worktree
- inspect root repo status if multiple worktrees exist
- confirm the handoff document path you are resuming from

## Verify Assumptions

- rerun the verification command recorded in the handoff
- compare current status to the handoff assumptions
- note any new files, rebases, merges, or drift
- check whether any old background process still matters

## Rebuild Focus

- restate the smallest next task in one sentence
- restate the acceptance criteria
- restate which files are in scope
- restate which files are explicitly out of scope

## Parallel Safety Check

- if multiple slices are being resumed, confirm write scopes are disjoint
- identify the bottleneck slice
- avoid starting parallel work on overlapping files

## Stop Conditions

Stop and update the handoff before coding if:

- the branch or worktree is not the one named in the handoff
- the verification command is now failing
- the documented next task is no longer the smallest safe step
- root/worktree drift has changed since the handoff was written
