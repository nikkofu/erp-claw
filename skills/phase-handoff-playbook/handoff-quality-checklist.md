# Handoff Quality Checklist

Use before marking handoff complete.

## Critical

- Active `worktree_path`, `branch`, `head_sha` are explicit
- `origin/main` and phase branch remote SHA are explicit
- Fresh verification commands were run in this session
- Verification results are explicit PASS/FAIL (not implied)
- Completed and unfinished work are separated
- Tomorrow kickoff commands are concrete and runnable
- First task tomorrow is one smallest actionable task
- Remaining tasks include acceptance criteria
- Merge safety status is explicit: `safe_to_merge` or `blocked` with reason

If any critical item is missing, handoff is incomplete.

## Important

- Open PR and release tag status are captured (if relevant)
- Risks are explicit and tied to unresolved scope
- Parallel-safe tasks are marked only when write scope is disjoint
- "Do not do next" traps are listed
- Reflection includes at least one concrete process improvement

## Strong Positive Signals

- New contributor can start in under 10 minutes
- Next session can run commands without re-deriving context
- At least one known failure mode is prevented by the handoff text
