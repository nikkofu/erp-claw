# Handoff Quality Checklist

Use this before declaring a handoff complete.

## Critical

- active `worktree path`, `branch`, and `head sha` are explicit
- `origin/main` and phase branch remote sha are explicit
- merge safety status is explicit (`safe_to_merge` or `blocked` with reason)
- fresh verification command was run in the active worktree
- verification result is stated explicitly
- completed work is separated from unfinished work
- tomorrow's first task is obvious
- remaining tasks have acceptance criteria
- remaining tasks have dependency order
- parallel-safe tasks are marked as such

If any critical item is missing, the handoff is incomplete.

## Important

- high-risk gaps are listed explicitly
- "do not do next" traps are listed explicitly
- reflection includes at least one process failure and one architecture lesson
- project outputs are saved under `docs/phase-handoff-playbook/`
- reusable guidance stays in `skills/phase-handoff-playbook/`
- branch-specific notes are not mixed into the reusable skill text

## Strong Signals Of A Good Handoff

- a new contributor can identify where to start in under 10 minutes
- a future agent can split remaining work into safe parallel slices without rereading the whole repo
- the handoff prevents at least one known failure mode from repeating

## Strong Signals Of A Bad Handoff

- "continue Phase 1" is the only next step
- tasks are still epic-sized
- tests are assumed from memory
- wrong branch or wrong worktree could be used without anyone noticing
- the document reads like a diary instead of an operational artifact
