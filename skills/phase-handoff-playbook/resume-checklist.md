# Resume Checklist

Use at session start before coding.

## 1) Reconfirm Context

- confirm intended worktree path
- confirm branch name
- run `git status -sb`
- confirm the handoff file path being resumed

## 2) Reconfirm Remote Drift

- run `git fetch origin --prune`
- record `origin/main` SHA
- record phase branch remote SHA
- check if new merges landed since handoff

## 3) Re-run Verification

- rerun the same verification command(s) from handoff
- if results differ, stop and update handoff assumptions first

## 4) Rebuild Focus

- restate first task in one sentence
- restate acceptance criteria
- restate in-scope files and out-of-scope files

## 5) Parallel Safety

- only parallelize tasks with disjoint write scope
- identify bottleneck slice first
- avoid overlapping router/container edits in parallel

## Stop Conditions

Stop and refresh handoff before coding if:

- branch/worktree mismatch
- verification now failing
- remote main changed and impacts assumptions
- documented first task is no longer smallest safe step
