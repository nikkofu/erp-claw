# Handoff Quality Checklist

Use this before declaring a handoff complete.

## Critical

- Fresh verification command was run in the active worktree
- Verification result is stated explicitly
- Active worktree path is written explicitly
- Branch name is written explicitly
- Root repo vs worktree drift is called out if present
- Completed work is separated from unfinished work
- Tomorrow’s first task is obvious
- Remaining tasks have acceptance criteria
- Remaining tasks have dependency order
- Parallel-safe tasks are marked as such

If any critical item is missing, the handoff is incomplete.

## Important

- High-risk gaps are listed explicitly
- “Do not do next” traps are listed explicitly
- Reflection includes at least one process failure and one architecture lesson
- Project outputs are saved under `docs/phase-handoff-playbook/`
- Reusable guidance stays in `skills/phase-handoff-playbook/`
- Branch-specific notes are not mixed into the reusable skill text

## Strong Signals Of A Good Handoff

- A new contributor can identify where to start in under 10 minutes
- A future agent can split remaining work into safe parallel slices without rereading the whole repo
- The handoff prevents at least one known failure mode from repeating

## Strong Signals Of A Bad Handoff

- “Continue Phase 1” is the only next step
- Tasks are still epic-sized
- Tests are assumed from memory
- The wrong branch or wrong worktree could be used without anyone noticing
- The document reads like a diary instead of an operational artifact
