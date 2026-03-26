# Phase 1 Wave 1D Reliability And Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver the first executable reliability slice for Phase 1: a real outbox polling/publish baseline with explicit repository and service boundaries, plus migration hardening needed for retry-aware state transitions.

**Architecture:** Keep runtime wiring minimal and local to worker bootstrap while pushing behavior into focused outbox abstractions. Use a repository boundary for DB selection/state transitions and a dispatcher service boundary for publish orchestration so worker runtime remains a thin loop. Observability setup remains intentionally deferred to a later Wave 1D slice after outbox baseline is executable and test-backed.

**Tech Stack:** Go 1.25, PostgreSQL, NATS JetStream event bus seam, golang-migrate SQL migrations, Go standard testing

---

## Scope And Boundaries

**In scope for this wave slice:**
- Outbox repository methods for selecting publishable records and state transitions (`published` / retry requeue).
- Outbox dispatcher service with one-iteration processing seam (`ProcessOnce`) and worker polling wiring.
- Reliability-hardening migration that adds minimal retry/error metadata and polling index support.
- Tests for dispatcher behavior and migration contract.

**Explicitly out of scope for this slice:**
- Full dead-letter queue lifecycle and max-attempt exhaustion policy.
- Cross-worker distributed lock manager beyond DB row-lock semantics.
- Full OpenTelemetry provider/bootstrap wiring for all runtimes.
- Consumer-side idempotency framework.

## File Structure Map

```text
erp-claw/
  cmd/
    worker/
      main.go                                      (modify)
  internal/
    application/
      shared/
        outbox/
          dispatcher.go                            (create)
          dispatcher_test.go                       (create)
          publisher.go                             (create)
          types.go                                 (create)
    infrastructure/
      persistence/
        postgres/
          outbox_repository.go                     (create)
          outbox_repository_test.go                (create)
  migrations/
    000006_phase1_reliability_hardening.up.sql    (create)
    000006_phase1_reliability_hardening.down.sql  (create)
  test/
    integration/
      outbox_reliability_test.go                  (create)
```

## Task 1: Define Outbox Service Boundaries (TDD)

**Files:**
- Create: `internal/application/shared/outbox/types.go`
- Create: `internal/application/shared/outbox/publisher.go`
- Create: `internal/application/shared/outbox/dispatcher.go`
- Create: `internal/application/shared/outbox/dispatcher_test.go`

- [ ] **Step 1: Write failing dispatcher tests for success and retry flows**

Test cases:
- `ProcessOnce` publishes each fetched message and marks them published on success.
- `ProcessOnce` marks message for retry when publish fails and continues processing.
- `ProcessOnce` returns repository fetch error when fetching publishable items fails.

- [ ] **Step 2: Run dispatcher tests to verify RED**

Run:

```bash
go test ./internal/application/shared/outbox -v
```

Expected:
- FAIL due to missing package/types/dispatcher implementation.

- [ ] **Step 3: Implement minimal outbox abstractions and dispatcher**

Implementation requirements:
- `Message` model with DB-facing fields required by this slice.
- `Repository` interface with:
  - fetch publishable batch
  - mark published
  - mark for retry (with next available timestamp and error detail)
- `Publisher` interface with a single publish operation.
- `Dispatcher` with `ProcessOnce(ctx)` that:
  - fetches batch
  - publishes each message
  - updates status based on publish outcome

- [ ] **Step 4: Re-run dispatcher tests to verify GREEN**

Run:

```bash
go test ./internal/application/shared/outbox -v
```

Expected:
- PASS for all new dispatcher tests.

## Task 2: Add PostgreSQL Outbox Repository Baseline

**Files:**
- Create: `internal/infrastructure/persistence/postgres/outbox_repository.go`
- Create: `internal/infrastructure/persistence/postgres/outbox_repository_test.go`

- [ ] **Step 1: Write failing repository constructor contract test**

Coverage:
- constructor rejects nil DB
- repository satisfies outbox repository interface

- [ ] **Step 2: Run repository test to verify RED**

Run:

```bash
go test ./internal/infrastructure/persistence/postgres -run TestNewOutboxRepositoryRejectsNilDB -v
```

Expected:
- FAIL because repository constructor does not exist yet.

- [ ] **Step 3: Implement repository methods**

Methods:
- `FetchPublishable(ctx, limit, now)` using `status='pending'`, `available_at <= now`, deterministic ordering, and `FOR UPDATE SKIP LOCKED`.
- `MarkPublished(ctx, id, publishedAt)` updates status/timestamp and clears transient error fields.
- `MarkForRetry(ctx, id, nextAvailableAt, reason)` requeues with error detail.

Reliability baseline rules:
- increment attempts only on claim/fetch path.
- preserve at-least-once publish semantics.
- keep SQL localized to repository.

- [ ] **Step 4: Re-run repository tests**

Run:

```bash
go test ./internal/infrastructure/persistence/postgres -run TestNewOutboxRepositoryRejectsNilDB -v
```

Expected:
- PASS.

## Task 3: Wire Worker Polling Loop To Dispatcher Seam

**Files:**
- Modify: `cmd/worker/main.go`

- [ ] **Step 1: Add failing integration seam test for migration contract**

Add `test/integration/outbox_reliability_test.go` migration contract checks first (file absent = failing state):
- asserts 000006 migration exists.
- asserts reliability columns/index statements exist for outbox.

- [ ] **Step 2: Run migration contract test to verify RED**

Run:

```bash
go test ./test/integration -run TestOutboxReliabilityMigrationContract -v
```

Expected:
- FAIL until migration is created.

- [ ] **Step 3: Replace placeholder outbox poll with dispatcher-driven loop**

`cmd/worker/main.go` changes:
- construct postgres outbox repository.
- construct bus-backed outbox publisher.
- construct dispatcher with batch/retry defaults.
- make ticker loop invoke `dispatcher.ProcessOnce(ctx)` each iteration.

- [ ] **Step 4: Verify compile and tests for worker slice**

Run:

```bash
go test ./cmd/worker -v
```

Expected:
- PASS (or `[no test files]` with successful compile).

## Task 4: Add Reliability-Hardening Migration 000006

**Files:**
- Create: `migrations/000006_phase1_reliability_hardening.up.sql`
- Create: `migrations/000006_phase1_reliability_hardening.down.sql`
- Create: `test/integration/outbox_reliability_test.go`

- [ ] **Step 1: Implement migration after failing contract test**

`up.sql` minimal baseline:
- add `attempts` integer default 0.
- add `last_error` text nullable.
- add `processing_at` timestamptz nullable.
- add outbox polling index on `(status, available_at, id)`.

`down.sql` reverse:
- drop index if exists.
- drop added columns.

- [ ] **Step 2: Re-run migration contract test to verify GREEN**

Run:

```bash
go test ./test/integration -run TestOutboxReliabilityMigrationContract -v
```

Expected:
- PASS.

## Task 5: Verification Pass For This Slice

**Files:**
- Modify any files above only as needed for green tests/refactor.

- [ ] **Step 1: Run focused package tests**

```bash
go test ./internal/application/shared/outbox ./internal/infrastructure/persistence/postgres ./test/integration -run "Outbox|outbox|TestNewOutboxRepositoryRejectsNilDB" -v
```

- [ ] **Step 2: Run broader regression check for touched runtime areas**

```bash
go test ./cmd/worker ./internal/platform/eventbus -v
```

- [ ] **Step 3: Confirm no scope violations**

Manually verify edited files remain inside Wave 1D write scope.

---

## Remaining Wave 1D Work (After First Slice)

1. Add max-attempt and dead-letter semantics (`failed` terminal state, DLQ topic strategy, and operator replay path).
2. Add repository integration tests against live PostgreSQL (compose-backed or CI service container) for lock/claim semantics.
3. Add queue-depth and publish-latency metrics plus structured publish failure logs.
4. Introduce `internal/infrastructure/observability/*` runtime bootstrap seam (tracer/meter providers) and apply first worker instrumentation span around poll iteration.
5. Add idempotency strategy guidance for downstream consumers and include integration scenarios for duplicate publish recovery.

## Execution Status (2026-03-25, first executable slice)

- [x] Plan file created and scoped to Wave 1D write boundaries.
- [x] RED phase completed for:
  - outbox dispatcher unit tests
  - outbox repository constructor test
  - migration contract integration test
- [x] GREEN implementation completed for:
  - `internal/application/shared/outbox/*` dispatcher/publisher/types seams
  - `internal/infrastructure/persistence/postgres/outbox_repository.go`
  - `cmd/worker/main.go` dispatcher polling integration
  - `migrations/000006_phase1_reliability_hardening.*`
- [x] Targeted verification completed:
  - `go test ./internal/application/shared/outbox -v`
  - `go test ./test/integration -run TestOutboxReliabilityMigrationContract -v`
  - `go test ./cmd/worker -v`
- [ ] Full `./internal/infrastructure/persistence/postgres` package green is blocked by unrelated parallel-wave compile failures in `policy_audit_repository_test.go`; outbox repository was verified via file-scoped command:
  - `go test ./internal/infrastructure/persistence/postgres/outbox_repository.go ./internal/infrastructure/persistence/postgres/outbox_repository_test.go -v`

## Execution Status (2026-03-25, second executable slice)

- [x] Added max-attempt terminal-failure semantics in outbox dispatcher:
  - retry when `Attempts < MaxAttempts`
  - terminal `failed` transition when `Attempts >= MaxAttempts`
- [x] Extended outbox message/repository contract:
  - `Message.Attempts`
  - `Repository.MarkFailed(...)`
- [x] Updated Postgres outbox repository to:
  - fetch and return `attempts`
  - increment `attempts` on claim
  - support `MarkFailed` state transition
- [x] Added TDD coverage for terminal-failure path:
  - `TestDispatcherProcessOnceMarksFailedWhenMaxAttemptsReached`
  - retry-path test now asserts no false `failed` transition
- [x] Fresh verification after changes:
  - `go test ./internal/application/shared/outbox -v`
  - `go test ./internal/infrastructure/persistence/postgres -run TestNewOutboxRepositoryRejectsNilDB -v`
  - `go test ./cmd/worker -v`
  - `go test ./test/integration -run TestOutboxReliabilityMigrationContract -v`

## Execution Status (2026-03-25, third executable slice)

- [x] Added minimal observability seam under `internal/infrastructure/observability/*`:
  - `OutboxPollMetrics` interface
  - `InstrumentedOutboxProcessor` wrapper
  - `NoopOutboxPollMetrics` default implementation
- [x] Instrumented worker polling path via observability wrapper in `cmd/worker/main.go`.
- [x] TDD RED→GREEN completed for observability slice:
  - RED: `go test ./internal/infrastructure/observability -v` failed with undefined `NewInstrumentedOutboxProcessor`
  - GREEN: same command passes with 3 tests
- [x] Regression checks after instrumentation:
  - `go test ./cmd/worker -v`
  - `go test ./internal/application/shared/outbox -v`
  - `go test ./test/integration -run TestOutboxReliabilityMigrationContract -v`
