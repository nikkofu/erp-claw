# Phase 1 Wave 1A Governance Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver the first executable Governance Core slice with real policy-rule evaluation and persistent audit/policy storage/query seams, without widening scope into command-pipeline/API rewiring.

**Architecture:** Add minimal governance domain contracts in `internal/platform/policy` and `internal/platform/audit`, then back them with a single Postgres repository implementation and one migration. Keep policy evaluation repository-driven and deterministic (tenant + command + actor matching with fallback) and keep audit querying simple (tenant-scoped list with optional filters and limit).

**Tech Stack:** Go 1.24, PostgreSQL SQL migrations, `database/sql`, Go test (`testing` package)

---

## Scope and Boundaries

- Write scope is limited to:
  - `docs/superpowers/plans/2026-03-25-phase1-wave1a-governance-core.md`
  - `internal/platform/policy/*`
  - `internal/platform/audit/*`
  - `internal/infrastructure/persistence/postgres/policy_audit_repository.go`
  - `internal/infrastructure/persistence/postgres/policy_audit_repository_test.go`
  - `migrations/000003_phase1_governance_core.up.sql`
  - `migrations/000003_phase1_governance_core.down.sql`
  - `test/integration/governance_core_test.go`
- Do not modify existing control-plane catalog files currently being worked on by other agents.
- Keep this wave to storage contracts + repository-backed policy/audit seams; avoid router/command pipeline wiring changes unless needed for tests.

## File Map (Planned)

- Create: `migrations/000003_phase1_governance_core.up.sql`
- Create: `migrations/000003_phase1_governance_core.down.sql`
- Create: `internal/platform/policy/rule.go`
- Create: `internal/platform/policy/rule_repository.go`
- Create: `internal/platform/policy/rule_service.go`
- Create: `internal/platform/policy/rule_evaluator_test.go`
- Create: `internal/platform/policy/rule_service_test.go`
- Create: `internal/platform/audit/event_store.go`
- Create: `internal/platform/audit/service.go`
- Create: `internal/platform/audit/service_test.go`
- Create: `internal/infrastructure/persistence/postgres/policy_audit_repository.go`
- Create: `internal/infrastructure/persistence/postgres/policy_audit_repository_test.go`
- Create: `test/integration/governance_core_test.go`
- Modify: `internal/platform/audit/model.go`
- Modify: `internal/platform/audit/recorder.go`

### Task 1: Migration Contract for Governance Core Tables

**Files:**
- Test: `test/integration/governance_core_test.go`
- Create: `migrations/000003_phase1_governance_core.up.sql`
- Create: `migrations/000003_phase1_governance_core.down.sql`

- [ ] **Step 1: Write failing migration contract test**

```go
func TestGovernanceCoreMigrationContainsPolicyRuleAndAuditEventTables(t *testing.T) {
    data, _ := os.ReadFile("../../migrations/000003_phase1_governance_core.up.sql")
    sql := strings.ToLower(string(data))
    requireContains(t, sql, "create table if not exists policy_rule")
    requireContains(t, sql, "create table if not exists audit_event")
}
```

- [ ] **Step 2: Run test to verify RED**

Run: `go test ./test/integration -run TestGovernanceCoreMigrationContainsPolicyRuleAndAuditEventTables -count=1`
Expected: FAIL because migration file/table definitions do not exist yet.

- [ ] **Step 3: Add migration up/down files**

```sql
create table if not exists policy_rule (... tenant_id text not null, ...);
create table if not exists audit_event (... tenant_id text not null, ...);
```

- [ ] **Step 4: Run test to verify GREEN**

Run: `go test ./test/integration -run TestGovernanceCoreMigrationContainsPolicyRuleAndAuditEventTables -count=1`
Expected: PASS.

### Task 2: Policy Rule Contracts + Real Evaluator (Repository-Backed)

**Files:**
- Create: `internal/platform/policy/rule.go`
- Create: `internal/platform/policy/rule_repository.go`
- Create: `internal/platform/policy/rule_service.go`
- Create: `internal/platform/policy/rule_evaluator_test.go`
- Create: `internal/platform/policy/rule_service_test.go`

- [ ] **Step 1: Write failing tests for evaluator behavior and service contract**

```go
func TestRuleEvaluatorMatchesTenantCommandActorWithPriority(t *testing.T) {}
func TestRuleEvaluatorFallsBackWhenNoRuleMatches(t *testing.T) {}
func TestRuleServiceUpsertAndListRequiresRepository(t *testing.T) {}
```

- [ ] **Step 2: Run policy package tests to verify RED**

Run: `go test ./internal/platform/policy -count=1`
Expected: FAIL due missing rule structs/repository/evaluator/service implementations.

- [ ] **Step 3: Implement minimal policy domain**

```go
type Rule struct { TenantID, ID, CommandName, ActorID string; Decision Decision; Priority int; Active bool }
type RuleRepository interface { UpsertRule(context.Context, Rule) (Rule, error); ListRules(context.Context, RuleFilter) ([]Rule, error) }
type RuleEvaluator struct { Rules RuleRepository; Fallback Decision }
```

- [ ] **Step 4: Run policy tests to verify GREEN**

Run: `go test ./internal/platform/policy -count=1`
Expected: PASS.

### Task 3: Audit Event Persistence + Query Domain Seams

**Files:**
- Modify: `internal/platform/audit/model.go`
- Modify: `internal/platform/audit/recorder.go`
- Create: `internal/platform/audit/event_store.go`
- Create: `internal/platform/audit/service.go`
- Create: `internal/platform/audit/service_test.go`

- [ ] **Step 1: Write failing tests for append/query service behavior**

```go
func TestServiceRecordStoresEventAndListReturnsTenantScopedEvents(t *testing.T) {}
func TestServiceListAppliesLimitAndFilters(t *testing.T) {}
```

- [ ] **Step 2: Run audit package tests to verify RED**

Run: `go test ./internal/platform/audit -count=1`
Expected: FAIL due missing query/store contracts and service.

- [ ] **Step 3: Implement minimal audit contracts and in-memory query support**

```go
type Query struct { TenantID, CommandName, ActorID string; Limit int }
type EventStore interface { Append(context.Context, Record) (Record, error); List(context.Context, Query) ([]Record, error) }
```

- [ ] **Step 4: Run audit tests to verify GREEN**

Run: `go test ./internal/platform/audit -count=1`
Expected: PASS.

### Task 4: Postgres Policy/Audit Repository Baseline

**Files:**
- Create: `internal/infrastructure/persistence/postgres/policy_audit_repository.go`
- Create: `internal/infrastructure/persistence/postgres/policy_audit_repository_test.go`

- [ ] **Step 1: Write failing repository contract tests**

```go
var _ policy.RuleRepository = (*PolicyAuditRepository)(nil)
var _ audit.EventStore = (*PolicyAuditRepository)(nil)
func TestNewPolicyAuditRepositoryRejectsNilDB(t *testing.T) {}
```

- [ ] **Step 2: Run repository tests to verify RED**

Run: `go test ./internal/infrastructure/persistence/postgres -run 'Test(NewPolicyAuditRepositoryRejectsNilDB|PolicyAuditRepository)' -count=1`
Expected: FAIL due missing repository.

- [ ] **Step 3: Implement repository with minimal SQL**

```go
func (r *PolicyAuditRepository) UpsertRule(ctx context.Context, rule policy.Rule) (policy.Rule, error) { ... }
func (r *PolicyAuditRepository) ListRules(ctx context.Context, filter policy.RuleFilter) ([]policy.Rule, error) { ... }
func (r *PolicyAuditRepository) Append(ctx context.Context, record audit.Record) (audit.Record, error) { ... }
func (r *PolicyAuditRepository) List(ctx context.Context, q audit.Query) ([]audit.Record, error) { ... }
```

- [ ] **Step 4: Run repository tests to verify GREEN**

Run: `go test ./internal/infrastructure/persistence/postgres -run 'Test(NewPolicyAuditRepositoryRejectsNilDB|PolicyAuditRepository)' -count=1`
Expected: PASS.

### Task 5: Targeted Integration Contract for Governance Core Seams

**Files:**
- Modify: `test/integration/governance_core_test.go`

- [ ] **Step 1: Write failing integration test proving real evaluator + audit query seam**

```go
func TestGovernanceCorePolicyAndAuditSeams(t *testing.T) {
    // insert rule in in-memory store, evaluate command, record event, query event list
}
```

- [ ] **Step 2: Run integration test to verify RED**

Run: `go test ./test/integration -run TestGovernanceCorePolicyAndAuditSeams -count=1`
Expected: FAIL before missing components are implemented.

- [ ] **Step 3: Implement/finish minimal plumbing in policy+audit packages**

```go
// ensure RuleEvaluator + audit Service compose cleanly without router/pipeline rewiring
```

- [ ] **Step 4: Run targeted and full verification**

Run: `go test ./internal/platform/policy ./internal/platform/audit ./internal/infrastructure/persistence/postgres ./test/integration -count=1`
Expected: PASS for new slice.

### Task 6: Closeout Evidence

**Files:**
- Modify: `docs/superpowers/plans/2026-03-25-phase1-wave1a-governance-core.md` (checkbox state update if desired)

- [ ] **Step 1: Capture exact command outputs for all RED/GREEN cycles**
- [ ] **Step 2: Summarize files changed and residual Phase 1 Wave 1A work**
- [ ] **Step 3: Commit this slice (if requested by owner)**
