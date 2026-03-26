# Phase 1 Wave 1C Capability Governance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver the first executable Capability Governance slice with a tenant-scoped model catalog baseline that can be managed through application command/query seams and persisted in PostgreSQL.

**Architecture:** Keep this slice narrow and governance-focused. The domain layer defines the model-catalog entry invariants, the application layer exposes command/query handlers for creation and listing, and a Postgres repository backs the catalog with one additive migration. Do not widen scope into router/container wiring, tool runtime integration, or policy pipeline changes in this round.

**Tech Stack:** Go 1.25, PostgreSQL SQL migrations, `database/sql`, Go test (`testing` package)

---

## Implementation Note (Actual Landed Slice, 2026-03-25)

This wave was ultimately landed by aligning to the existing root-baseline shape instead of the earlier draft below. The executable baseline in the worktree uses:

- table name: `model_catalog_entries`
- domain shape: `ModelCatalogEntry{TenantID, EntryID, ModelKey, DisplayName, Provider, Status}`
- repository seam: `Save(...)` and `ListByTenant(...)`
- application handlers under `internal/application/capability`:
  - `create_model_catalog_entry.go`
  - `list_model_catalog_entries.go`

Actual landed file map:

- `internal/domain/capability/model_catalog_entry.go`
- `internal/domain/capability/model_catalog_entry_test.go`
- `internal/domain/capability/model_catalog_repository.go`
- `internal/application/capability/create_model_catalog_entry.go`
- `internal/application/capability/create_model_catalog_entry_test.go`
- `internal/application/capability/list_model_catalog_entries.go`
- `internal/application/capability/list_model_catalog_entries_test.go`
- `internal/application/capability/repository_fake_test.go`
- `internal/infrastructure/persistence/postgres/capability_repository.go`
- `internal/infrastructure/persistence/postgres/capability_repository_test.go`
- `migrations/000005_phase1_capability_governance.up.sql`
- `migrations/000005_phase1_capability_governance.down.sql`
- `test/integration/capability_governance_test.go`

When using this plan as future reference, treat the executable baseline above as the source of truth. The detailed checklist below remains useful as historical planning context, but its table/handler naming differs from the code that was actually landed.

## Scope And Boundaries

- Write scope is limited to:
  - `docs/superpowers/plans/2026-03-25-phase1-wave1c-capability-governance.md`
  - `internal/domain/capability/*`
  - `internal/application/capability/*`
  - `internal/infrastructure/persistence/postgres/capability_repository.go`
  - `internal/infrastructure/persistence/postgres/capability_repository_test.go`
  - `migrations/000005_phase1_capability_governance.up.sql`
  - `migrations/000005_phase1_capability_governance.down.sql`
  - `test/integration/capability_governance_test.go`
- Keep this wave to tenant-scoped model catalog only.
- Do not modify router, bootstrap container, worker, scheduler, workspace gateway, or command pipeline files.
- Do not add tool registry, plugin registry, quota, or feature-flag behavior in this round.

## File Map (Planned)

- Create: `migrations/000005_phase1_capability_governance.up.sql`
- Create: `migrations/000005_phase1_capability_governance.down.sql`
- Create: `internal/domain/capability/model_catalog.go`
- Create: `internal/domain/capability/model_catalog_test.go`
- Create: `internal/domain/capability/repository.go`
- Create: `internal/application/capability/command/upsert_model_catalog_entry.go`
- Create: `internal/application/capability/command/upsert_model_catalog_entry_test.go`
- Create: `internal/application/capability/query/list_model_catalog_entries.go`
- Create: `internal/application/capability/query/list_model_catalog_entries_test.go`
- Create: `internal/infrastructure/persistence/postgres/capability_repository.go`
- Create: `internal/infrastructure/persistence/postgres/capability_repository_test.go`
- Create: `test/integration/capability_governance_test.go`

### Task 1: Add Migration Contract For Tenant Model Catalog

**Files:**
- Test: `test/integration/capability_governance_test.go`
- Create: `migrations/000005_phase1_capability_governance.up.sql`
- Create: `migrations/000005_phase1_capability_governance.down.sql`

- [ ] **Step 1: Write failing migration contract test**

```go
func TestCapabilityGovernanceMigrationContainsModelCatalogTable(t *testing.T) {
    data, _ := os.ReadFile("../../migrations/000005_phase1_capability_governance.up.sql")
    sql := strings.ToLower(string(data))
    requireContains(t, sql, "create table if not exists tenant_model_catalog")
}
```

- [ ] **Step 2: Run test to verify RED**

Run: `go test ./test/integration -run TestCapabilityGovernanceMigrationContainsModelCatalogTable -count=1`
Expected: FAIL because the migration does not exist yet.

- [ ] **Step 3: Add migration up/down files**

```sql
create table if not exists tenant_model_catalog (
    tenant_id text not null,
    id text not null,
    provider text not null,
    model text not null,
    display_name text not null,
    enabled boolean not null default true,
    is_default boolean not null default false,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (tenant_id, id)
);
```

- [ ] **Step 4: Run test to verify GREEN**

Run: `go test ./test/integration -run TestCapabilityGovernanceMigrationContainsModelCatalogTable -count=1`
Expected: PASS.

### Task 2: Define Model Catalog Domain Contracts

**Files:**
- Create: `internal/domain/capability/model_catalog.go`
- Create: `internal/domain/capability/model_catalog_test.go`
- Create: `internal/domain/capability/repository.go`

- [ ] **Step 1: Write failing domain tests**

Add tests for:
- creating an entry requires tenant ID
- creating an entry requires provider and model
- new entries default to `enabled=true`
- filter remains tenant-scoped

- [ ] **Step 2: Run tests to verify RED**

Run: `go test ./internal/domain/capability -count=1`
Expected: FAIL due missing types and constructors.

- [ ] **Step 3: Implement minimal domain model**

```go
type ModelCatalogEntry struct {
    TenantID    string
    ID          string
    Provider    string
    Model       string
    DisplayName string
    Enabled     bool
    IsDefault   bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type ModelCatalogFilter struct {
    TenantID string
    EnabledOnly bool
    Limit int
}

type ModelCatalogRepository interface {
    UpsertModelCatalogEntry(context.Context, ModelCatalogEntry) (ModelCatalogEntry, error)
    ListModelCatalogEntries(context.Context, ModelCatalogFilter) ([]ModelCatalogEntry, error)
}
```

- [ ] **Step 4: Run tests to verify GREEN**

Run: `go test ./internal/domain/capability -count=1`
Expected: PASS.

### Task 3: Add Capability Application Command/Query Seams

**Files:**
- Create: `internal/application/capability/command/upsert_model_catalog_entry.go`
- Create: `internal/application/capability/command/upsert_model_catalog_entry_test.go`
- Create: `internal/application/capability/query/list_model_catalog_entries.go`
- Create: `internal/application/capability/query/list_model_catalog_entries_test.go`

- [ ] **Step 1: Write failing application tests**

Add tests for:
- handler rejects nil repository/service dependency
- upsert handler maps command input into domain entry and returns stored entry
- list handler enforces tenant-scoped query input

- [ ] **Step 2: Run tests to verify RED**

Run:
- `go test ./internal/application/capability/command -count=1`
- `go test ./internal/application/capability/query -count=1`

Expected: FAIL due missing handlers and package files.

- [ ] **Step 3: Implement minimal handlers**

```go
type UpsertModelCatalogEntry struct { ... }
type UpsertModelCatalogEntryHandler struct { Entries capability.ModelCatalogRepository ... }

type ListModelCatalogEntries struct { ... }
type ListModelCatalogEntriesHandler struct { Entries capability.ModelCatalogRepository }
```

- [ ] **Step 4: Run tests to verify GREEN**

Run:
- `go test ./internal/application/capability/command -count=1`
- `go test ./internal/application/capability/query -count=1`

Expected: PASS.

### Task 4: Add PostgreSQL Capability Repository Baseline

**Files:**
- Create: `internal/infrastructure/persistence/postgres/capability_repository.go`
- Create: `internal/infrastructure/persistence/postgres/capability_repository_test.go`

- [ ] **Step 1: Write failing repository tests**

Coverage:
- constructor rejects nil DB
- compile-time interface conformance to `capability.ModelCatalogRepository`

- [ ] **Step 2: Run repository tests to verify RED**

Run: `go test ./internal/infrastructure/persistence/postgres -run 'Test(NewCapabilityRepositoryRejectsNilDB|CapabilityRepository)' -count=1`
Expected: FAIL due missing repository.

- [ ] **Step 3: Implement repository methods**

Implement:
- `UpsertModelCatalogEntry`
- `ListModelCatalogEntries`

SQL rules:
- tenant-scoped write/read only
- deterministic ordering by `is_default desc`, `display_name asc`, `id asc`
- simple upsert keyed by `(tenant_id, id)`

- [ ] **Step 4: Run repository tests to verify GREEN**

Run: `go test ./internal/infrastructure/persistence/postgres -run 'Test(NewCapabilityRepositoryRejectsNilDB|CapabilityRepository)' -count=1`
Expected: PASS.

### Task 5: Add Integration Contract For Capability Governance Seam

**Files:**
- Modify: `test/integration/capability_governance_test.go`

- [ ] **Step 1: Write failing integration seam test**

```go
func TestCapabilityGovernanceApplicationHandlersManageTenantModelCatalog(t *testing.T) {
    // use in-memory repository implementing capability.ModelCatalogRepository
    // upsert two entries for different tenants and assert list is tenant-scoped
}
```

- [ ] **Step 2: Run integration test to verify RED**

Run: `go test ./test/integration -run TestCapabilityGovernanceApplicationHandlersManageTenantModelCatalog -count=1`
Expected: FAIL before handlers/domain are implemented.

- [ ] **Step 3: Implement/finish minimal in-memory seam support in tests**

Keep production scope narrow. Test-local in-memory repository is acceptable.

- [ ] **Step 4: Run targeted verification**

Run:
- `go test ./internal/domain/capability ./internal/application/capability/... ./internal/infrastructure/persistence/postgres -run 'Capability|ModelCatalog|NewCapabilityRepositoryRejectsNilDB' -count=1`
- `go test ./test/integration -run TestCapabilityGovernance -count=1`

Expected: PASS.

### Task 6: Closeout Evidence

- [ ] **Step 1: Capture exact RED/GREEN commands and outcomes**
- [ ] **Step 2: Summarize files changed and residual Wave 1C work**
- [ ] **Step 3: Commit this slice if explicitly requested by owner**
