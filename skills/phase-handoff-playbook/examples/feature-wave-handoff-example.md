# Example: Feature Wave Handoff

This is a generic example, not tied to any specific repo.

## Snapshot

- worktree path: `/workspace/.worktrees/feature-a`
- branch: `feature/customer-notes`
- fresh verification command:
  - `GOCACHE=$(pwd)/.cache/go-build go test ./... -count=1`
- fresh verification result:
  - all tests passing
- completed today:
  - notes domain object
  - create/list handlers
  - repository baseline
  - migration contract test
- still unfinished:
  - update handler
  - delete/archive semantics
  - admin API wiring

## Tomorrow Kickoff

1. enter `/workspace/.worktrees/feature-a`
2. inspect worktree and root `git status --short`
3. rerun `go test ./... -count=1`
4. confirm no drift relative to this handoff
5. start with `archive note` domain rule, not the API

## Minimal Task Checklist

| ID | Priority | Task | Why | Acceptance | Suggested files | Dependency | Parallel-safe |
| --- | --- | --- | --- | --- | --- | --- | --- |
| T01 | P1 | Write failing tests for archive semantics | Locks expected behavior before API work | RED then GREEN tests for active -> archived | `internal/domain/notes/*` | none | yes |
| T02 | P1 | Implement archive transition in domain model | Domain rule should exist before handlers | domain tests pass | `internal/domain/notes/*` | T01 | yes |
| T03 | P1 | Add archive handler | Exposes behavior to app layer | handler tests pass | `internal/application/notes/*` | T02 | yes |
| T04 | P1 | Add repository support | Persists archive state | repo tests pass | `internal/infrastructure/persistence/*` | T03 | no |
| T05 | P2 | Wire admin route | Transport should be last | route integration test passes | `internal/interfaces/http/*` | T04 | no |

## Experience Summary

- what worked:
  - strict write-scope prompts for parallel workers
  - targeted tests before full test sweep
- what accelerated delivery:
  - splitting domain/app/repo concerns
- what slowed delivery:
  - one worker initially wrote to the wrong branch

## Deep Reflection

- biggest process mistake:
  - branch context was assumed, not stated
- biggest architecture lesson:
  - transport wiring should come after domain and repository contracts
- what must change next time:
  - always include absolute worktree path in parallel prompts

## Do Not Do Tomorrow

- do not start with route wiring
- do not parallelize tasks touching the same repository file
