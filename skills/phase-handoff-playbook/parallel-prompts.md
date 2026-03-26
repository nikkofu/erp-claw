# Parallel Prompt Pack

Reusable fragments for safe multi-agent continuation.

## Worker Slice Prompt

```text
Continue this isolated slice in:
- worktree: <ABS_WORKTREE_PATH>
- branch: <BRANCH>

You are not alone in the repo. Do not revert others' changes.
Stay strictly inside your owned write scope.

Goal:
- <ONE OUTCOME>

Owned write scope:
- <ABS_PATH_1>
- <ABS_PATH_2>

Do not modify:
- <ABS_PATH_3>
- <ABS_PATH_4>

Validation:
- <EXACT COMMAND>

Return format:
- DONE / DONE_WITH_CONCERNS / NEEDS_CONTEXT / BLOCKED
- changed files
- RED/GREEN evidence
- residual risks
```

## Handoff Reviewer Prompt

```text
Review this handoff for operational readiness (not writing style).

Check:
- explicit worktree/branch/sha?
- explicit remote refs and drift?
- fresh verification with outcomes?
- completed vs unfinished split?
- first task tomorrow obvious?
- task list small and actionable?
- merge safety call present?

Output:
1) findings by severity
2) open questions
3) short readiness verdict
```

## Resume Prompt

```text
Resume from:
- <HANDOFF_DOC_PATH>

Start with:
1. verify worktree + branch
2. fetch and check remote refs
3. rerun verification commands

Then propose exactly one smallest safe next task.
If assumptions are stale, state what changed before coding.
```
