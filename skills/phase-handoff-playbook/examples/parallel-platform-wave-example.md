# Example: Parallel Platform Wave Handoff

This example shows how to hand off a wave that was developed by multiple agents in parallel.

## Snapshot

- worktree path: `/workspace/.worktrees/platform-wave`
- branch: `feature/platform-wave`
- fresh verification command:
  - `GOCACHE=$(pwd)/.cache/go-build go test ./... -count=1`
- fresh verification result:
  - all packages passing
- completed today:
  - governance slice
  - runtime slice
  - capability slice
  - reliability slice
- still unfinished:
  - tenant/iam expansion
  - registry wiring
  - workflow baseline
  - live database integration

## Safe Parallel Slices For Next Session

| Slice | Owner Focus | Write Scope | Verification |
| --- | --- | --- | --- |
| governance-next | policy/audit lifecycle gaps | `internal/platform/policy/*`, `internal/application/governance/*` | targeted governance tests |
| runtime-read-side | runtime query/replay seams | `internal/application/runtime/*`, `internal/interfaces/ws/*` | targeted runtime tests |
| capability-next | tool catalog and bindings | `internal/domain/capability/*`, `internal/application/capability/*` | targeted capability tests |
| live-db | repository round-trip tests | `internal/infrastructure/persistence/*`, `test/integration/*` | targeted repo/integration tests |

## Tomorrow Kickoff

1. confirm this is still the active worktree
2. rerun the recorded verification command
3. check for root/worktree drift
4. pick one bottleneck slice first
5. only then dispatch parallel workers with disjoint write scopes

## Smallest Next Tasks

| ID | Priority | Task | Why | Acceptance | Suggested files | Dependency | Parallel-safe |
| --- | --- | --- | --- | --- | --- | --- | --- |
| T10 | P0 | Reconfirm root vs worktree drift | Prevent writing to the wrong tree | explicit conclusion written down | repo status only | none | no |
| T11 | P1 | Add tenant role migration contract tests | Enables real IAM expansion | RED then GREEN | `test/integration/*iam*` | T10 | yes |
| T12 | P1 | Add tool catalog migration contract tests | Enables capability expansion | RED then GREEN | `test/integration/*tool*` | T10 | yes |
| T13 | P1 | Add runtime replay query contract tests | Enables workspace read side | RED then GREEN | `internal/application/runtime/*` | T10 | yes |
| T14 | P1 | Add live repo round-trip test harness | Enables repository truth checks | compose-backed tests pass | `test/integration/*` | T10 | yes |

## Risk Register

- highest risk:
  - repository tests still compile-only, not live-db-backed
- coordination risk:
  - runtime read-side and transport wiring can overlap if not scoped tightly
- documentation risk:
  - phase status docs can go stale after one more wave unless updated immediately

## Do Not Do Tomorrow

- do not dispatch parallel tasks that modify the same router or container files
- do not trust old green test results without rerunning them
- do not call the wave complete just because the current slices pass
