# Parallel Prompt Pack

These are reusable prompt fragments for safely continuing a wave with subagents.

## Prompt 1: Parallel Worker Slice

Use for one isolated implementation slice.

```text
You are continuing a bounded slice in the repo at <WORKTREE_PATH>.

You are not alone in the codebase. Other workers may be editing other files.
Do not revert anyone else's changes.
Stay strictly inside your owned write scope.

Goal:
- <ONE CLEAR OUTCOME>

Owned write scope only:
- <ABSOLUTE_PATH_1>
- <ABSOLUTE_PATH_2>

Do not modify:
- <ABSOLUTE_PATH_OR_AREA_1>
- <ABSOLUTE_PATH_OR_AREA_2>

Implementation constraints:
- Use TDD
- Show RED before production edits
- Keep changes minimal and local

Validation to run before finishing:
- <EXACT TEST COMMAND>

Return format:
- Start with DONE / DONE_WITH_CONCERNS / NEEDS_CONTEXT / BLOCKED
- Summarize what changed
- List exact files changed
- Include RED/GREEN evidence
- State residual concerns
```

## Prompt 2: Handoff Reviewer

Use for a reviewer agent checking whether a handoff doc is actually operational.

```text
Review this handoff document for operational quality, not style.

Check:
- Is the active worktree and branch explicit?
- Is verification fresh and concrete?
- Are completed vs unfinished items clearly separated?
- Is tomorrow's first task obvious?
- Are remaining tasks small enough to start directly?
- Are parallel-safe tasks clearly marked?
- Are root/worktree drift and major risks called out?

Report:
- Findings first, ordered by severity
- Then open questions
- Then a short assessment
```

## Prompt 3: Resume From Handoff

Use at the start of the next session.

```text
Resume work from this handoff:
- <HANDOFF_DOC_PATH>

Start by:
1. verifying the stated worktree and branch
2. rerunning the stated verification command
3. checking whether the repo still matches the handoff assumptions

Then propose the smallest safe next task to execute.
If the handoff assumptions are stale, say exactly what changed before coding.
```
