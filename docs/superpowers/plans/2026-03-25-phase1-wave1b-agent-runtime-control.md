# Phase 1 Wave 1B Agent Runtime Control Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver the first executable Agent Runtime Control slice with real session/task runtime state transitions, persistence contracts over existing `agent_session` and `agent_task` tables, and a minimal workspace stream event baseline.

**Architecture:** Keep the slice narrow and testable. Domain layer owns status/state rules, application layer orchestrates transitions and event append, and Postgres repository persists metadata with existing tables plus additive safety constraints in `000004`. WebSocket changes stay minimal by using the existing workspace gateway as an event sink without protocol redesign.

**Tech Stack:** Go 1.25, standard library tests, PostgreSQL migrations, existing workspace gateway seam

---

**Spec Reference:** `docs/superpowers/specs/2026-03-25-agentic-ai-native-erp-design.md`

**Scope:** This plan intentionally implements only Wave 1B first slice:
- explicit session/task runtime contracts
- state transition guardrails
- Postgres read/write repository for metadata
- minimal workspace event append path

## File Structure Map

```text
docs/superpowers/plans/
  2026-03-25-phase1-wave1b-agent-runtime-control.md

internal/domain/agentruntime/
  session.go
  session_test.go
  task.go
  task_test.go
  repository.go

internal/application/agentruntime/
  service.go
  service_test.go

internal/platform/runtime/
  workspace_event.go (extend with status event constants + timestamp)

internal/interfaces/ws/
  workspace_gateway.go (minimal sink seam)
  workspace_gateway_test.go

internal/infrastructure/persistence/postgres/
  agent_runtime_repository.go
  agent_runtime_repository_test.go

migrations/
  000004_phase1_agent_runtime_control.up.sql
  000004_phase1_agent_runtime_control.down.sql

test/integration/
  agent_runtime_control_test.go
```

### Task 1: Define Domain Runtime Contracts with Transition Rules

**Files:**
- Create: `internal/domain/agentruntime/session.go`
- Create: `internal/domain/agentruntime/session_test.go`
- Create: `internal/domain/agentruntime/task.go`
- Create: `internal/domain/agentruntime/task_test.go`
- Create: `internal/domain/agentruntime/repository.go`

- [ ] **Step 1: Write failing session transition tests**

Add tests for:
- creating session requires tenant ID + session key
- valid transition `open -> closed`
- invalid transition `closed -> open`

- [ ] **Step 2: Run tests to verify RED**

Run:
```bash
go test ./internal/domain/agentruntime -run Session -v
```
Expected:
- FAIL with missing package/types or undefined functions

- [ ] **Step 3: Write failing task transition tests**

Add tests for:
- creating task requires tenant ID + task type
- valid transitions `pending -> running -> succeeded`
- invalid transition from terminal status back to running

- [ ] **Step 4: Run tests to verify RED**

Run:
```bash
go test ./internal/domain/agentruntime -run Task -v
```
Expected:
- FAIL with missing status transition behavior

- [ ] **Step 5: Implement minimal domain objects and repository interfaces**

Implement:
- `SessionStatus` and `TaskStatus` constants
- constructors with required-field validation
- transition methods returning explicit errors
- repository interfaces for session/task create/get/update operations

- [ ] **Step 6: Re-run domain tests to verify GREEN**

Run:
```bash
go test ./internal/domain/agentruntime -v
```
Expected:
- PASS

### Task 2: Add Application Service for Runtime State Changes + Event Append

**Files:**
- Create: `internal/application/agentruntime/service.go`
- Create: `internal/application/agentruntime/service_test.go`
- Modify: `internal/platform/runtime/workspace_event.go`

- [ ] **Step 1: Write failing service tests**

Add tests for:
- nil dependencies rejected
- `StartTask` allows `pending -> running`
- `CompleteTask` rejects non-running tasks
- successful completion appends one workspace runtime event

- [ ] **Step 2: Run tests to verify RED**

Run:
```bash
go test ./internal/application/agentruntime -v
```
Expected:
- FAIL due missing service implementation

- [ ] **Step 3: Implement minimal runtime service**

Implement a service that:
- loads task from repository
- applies domain transition rule
- persists status update
- appends one `WorkspaceEvent` for task status changes through an injected appender interface

Keep event envelope minimal:
- event type
- tenant/session/task IDs
- payload with status
- timestamp field

- [ ] **Step 4: Re-run service tests to verify GREEN**

Run:
```bash
go test ./internal/application/agentruntime -v
```
Expected:
- PASS

### Task 3: Implement Postgres Repository for Session/Task Metadata

**Files:**
- Create: `internal/infrastructure/persistence/postgres/agent_runtime_repository.go`
- Create: `internal/infrastructure/persistence/postgres/agent_runtime_repository_test.go`

- [ ] **Step 1: Write failing repository tests**

Add tests for:
- constructor rejects nil DB
- compile-time interface conformance to domain repository contracts

- [ ] **Step 2: Run tests to verify RED**

Run:
```bash
go test ./internal/infrastructure/persistence/postgres -run AgentRuntime -v
```
Expected:
- FAIL due missing repository type/methods

- [ ] **Step 3: Implement repository methods**

Implement operations:
- `CreateSession`
- `GetSessionByTenantAndKey`
- `UpdateSessionStatus`
- `CreateTask`
- `GetTaskByID`
- `UpdateTaskStatus`

Use existing columns from `agent_session` / `agent_task` and keep SQL statements straightforward.

- [ ] **Step 4: Re-run repository tests to verify GREEN**

Run:
```bash
go test ./internal/infrastructure/persistence/postgres -run AgentRuntime -v
```
Expected:
- PASS

### Task 4: Add Minimal Workspace Stream Baseline + Migration Guardrails

**Files:**
- Modify: `internal/interfaces/ws/workspace_gateway.go`
- Modify: `internal/interfaces/ws/workspace_gateway_test.go`
- Create: `migrations/000004_phase1_agent_runtime_control.up.sql`
- Create: `migrations/000004_phase1_agent_runtime_control.down.sql`
- Create: `test/integration/agent_runtime_control_test.go`

- [ ] **Step 1: Write failing tests**

Add tests for:
- workspace gateway accepts runtime event append through minimal sink seam
- migration `000004` contains status-check guardrails and runtime query indexes

- [ ] **Step 2: Run tests to verify RED**

Run:
```bash
go test ./internal/interfaces/ws -run Runtime -v
go test ./test/integration -run AgentRuntimeControl -v
```
Expected:
- FAIL until sink seam/migration files exist

- [ ] **Step 3: Implement minimal gateway seam + migration SQL**

Implement:
- one method in workspace gateway to append runtime event by delegating to `Broadcast`
- additive SQL in `000004`:
  - check constraints for session/task status values
  - indexes supporting tenant/session/task status lookups

- [ ] **Step 4: Re-run targeted tests to verify GREEN**

Run:
```bash
go test ./internal/interfaces/ws -run Runtime -v
go test ./test/integration -run AgentRuntimeControl -v
```
Expected:
- PASS

### Task 5: Wave Slice Verification

**Files:**
- Validate only files in Wave 1B write scope were changed

- [ ] **Step 1: Run targeted package tests**

Run:
```bash
go test ./internal/domain/agentruntime -v
go test ./internal/application/agentruntime -v
go test ./internal/infrastructure/persistence/postgres -run AgentRuntime -v
go test ./internal/interfaces/ws -run Runtime -v
go test ./test/integration -run AgentRuntimeControl -v
```

Expected:
- all PASS

- [ ] **Step 2: Record concrete test output in handoff**

Include exact command + pass/fail summary in the final report.
