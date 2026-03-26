# Phase 1 Control Plane Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first real Phase 1 control-plane slice: tenant catalog, organization/user/role catalog, agent profile registry, and the first Admin API endpoints backed by PostgreSQL.

**Architecture:** This slice extends the existing platform skeleton without disturbing the current runtime seams. The implementation introduces a real `internal/domain` layer for control-plane entities, adds application-level command/query handlers, persists the control-plane catalog in PostgreSQL, and exposes explicit Admin API actions through Gin. The focus is catalog truth and governance metadata, not yet full policy evaluation, approvals, or workspace execution.

**Tech Stack:** Go 1.25, Gin, PostgreSQL, golang-migrate, standard library testing, existing platform bootstrap/runtime packages

---

**Spec References:**

- `docs/superpowers/specs/2026-03-25-agentic-ai-native-erp-design.md`
- `docs/phase-1-coverage-status.md`
- `docs/phase-roadmap-overview.md`

**Scope Note:** This plan intentionally covers only the first executable Phase 1 control-plane slice. It does not include:

- dynamic policy rules engine
- approval/workflow runtime
- workspace protocol
- tool/model registry runtime
- supply-chain business domains

## File Structure Map

The implementation produced by this plan should create or modify the following structure.

```text
erp-claw/
  cmd/
    api-server/main.go
  internal/
    bootstrap/
      container.go
    domain/
      controlplane/
        tenant.go
        tenant_test.go
        iam.go
        iam_test.go
        agent_profile.go
        agent_profile_test.go
        repository.go
    application/
      controlplane/
        command/
          create_tenant.go
          create_user.go
          create_agent_profile.go
        query/
          list_tenants.go
          list_users.go
          list_agent_profiles.go
    interfaces/
      http/
        router/
          admin.go
        presenter/
          response.go
    infrastructure/
      persistence/
        postgres/
          controlplane_repository.go
          controlplane_repository_test.go
  migrations/
    000002_init_control_plane_catalog.up.sql
    000002_init_control_plane_catalog.down.sql
  test/
    integration/
      admin_control_plane_test.go
  README.md
```

### Task 1: Add Domain Model for Tenant, IAM, and Agent Profile Catalog

**Files:**

- Create: `internal/domain/controlplane/tenant.go`
- Create: `internal/domain/controlplane/tenant_test.go`
- Create: `internal/domain/controlplane/iam.go`
- Create: `internal/domain/controlplane/iam_test.go`
- Create: `internal/domain/controlplane/agent_profile.go`
- Create: `internal/domain/controlplane/agent_profile_test.go`
- Create: `internal/domain/controlplane/repository.go`

- [ ] **Step 1: Write the failing tenant domain test**

Create `internal/domain/controlplane/tenant_test.go`:

```go
package controlplane

import "testing"

func TestNewTenantRejectsEmptyCode(t *testing.T) {
	_, err := NewTenant("", "ERP Claw")
	if err == nil {
		t.Fatal("expected empty tenant code to fail")
	}
}
```

- [ ] **Step 2: Write the failing IAM domain test**

Create `internal/domain/controlplane/iam_test.go`:

```go
package controlplane

import "testing"

func TestNewUserRejectsEmptyEmail(t *testing.T) {
	_, err := NewUser("tenant-a", "", "Ada")
	if err == nil {
		t.Fatal("expected empty email to fail")
	}
}
```

- [ ] **Step 3: Write the failing agent profile test**

Create `internal/domain/controlplane/agent_profile_test.go`:

```go
package controlplane

import "testing"

func TestNewAgentProfileRequiresModel(t *testing.T) {
	_, err := NewAgentProfile("tenant-a", "planner", "")
	if err == nil {
		t.Fatal("expected empty model to fail")
	}
}
```

- [ ] **Step 4: Run the domain tests to verify RED**

Run:

```bash
go test ./internal/domain/controlplane -v
```

Expected:

- FAIL because the package and constructors do not exist yet

- [ ] **Step 5: Implement minimal domain entities and repository interfaces**

Create:

- `tenant.go` with `Tenant` plus `NewTenant(code, name string) (Tenant, error)`
- `iam.go` with `Organization`, `User`, `Role`, and `NewUser(tenantID, email, displayName string) (User, error)`
- `agent_profile.go` with `AgentProfile` and `NewAgentProfile(tenantID, name, model string) (AgentProfile, error)`
- `repository.go` with interfaces:
  - `TenantRepository`
  - `UserRepository`
  - `AgentProfileRepository`

Rules:

- use explicit validation
- keep domain packages free of SQL/Gin/transport concerns
- use string IDs for now to keep the first slice small

- [ ] **Step 6: Re-run the domain tests**

Run:

```bash
go test ./internal/domain/controlplane -v
```

Expected:

- PASS for all three tests

### Task 2: Add Control-Plane Catalog Migration

**Files:**

- Create: `migrations/000002_init_control_plane_catalog.up.sql`
- Create: `migrations/000002_init_control_plane_catalog.down.sql`

- [ ] **Step 1: Write the migration shape as a failing contract check**

Add a new test to `test/integration/admin_control_plane_test.go` with a minimal read of the migration file:

```go
package integration

import (
	"os"
	"strings"
	"testing"
)

func TestControlPlaneMigrationContainsCatalogTables(t *testing.T) {
	data, err := os.ReadFile("../../migrations/000002_init_control_plane_catalog.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	required := []string{
		"create table if not exists organization",
		"create table if not exists iam_user",
		"create table if not exists iam_role",
		"create table if not exists iam_user_role_binding",
		"create table if not exists agent_profile",
	}
	for _, needle := range required {
		if !strings.Contains(strings.ToLower(string(data)), needle) {
			t.Fatalf("expected migration to contain %q", needle)
		}
	}
}
```

- [ ] **Step 2: Run the migration contract test to verify RED**

Run:

```bash
go test ./test/integration -run TestControlPlaneMigrationContainsCatalogTables -v
```

Expected:

- FAIL because the migration file does not exist yet

- [ ] **Step 3: Implement the migration**

Create `000002_init_control_plane_catalog.up.sql` with tables:

- `organization`
- `iam_user`
- `iam_role`
- `iam_user_role_binding`
- `agent_profile`

Rules:

- all records must carry `tenant_id`
- use text IDs in this slice
- include only essential uniqueness constraints
- do not add unrelated business tables

Create the matching `down.sql` in strict reverse dependency order.

- [ ] **Step 4: Re-run the migration contract test**

Run:

```bash
go test ./test/integration -run TestControlPlaneMigrationContainsCatalogTables -v
```

Expected:

- PASS

### Task 3: Implement PostgreSQL Control-Plane Repositories

**Files:**

- Create: `internal/infrastructure/persistence/postgres/controlplane_repository.go`
- Create: `internal/infrastructure/persistence/postgres/controlplane_repository_test.go`

- [ ] **Step 1: Write the failing repository constructor test**

Create `internal/infrastructure/persistence/postgres/controlplane_repository_test.go`:

```go
package postgres

import "testing"

func TestNewControlPlaneRepositoryRejectsNilDB(t *testing.T) {
	_, err := NewControlPlaneRepository(nil)
	if err == nil {
		t.Fatal("expected nil db to fail")
	}
}
```

- [ ] **Step 2: Run the repository test to verify RED**

Run:

```bash
go test ./internal/infrastructure/persistence/postgres -run TestNewControlPlaneRepositoryRejectsNilDB -v
```

Expected:

- FAIL because the constructor does not exist yet

- [ ] **Step 3: Implement the repository**

Create `controlplane_repository.go` with:

- `ControlPlaneRepository` struct holding `*sql.DB`
- `NewControlPlaneRepository(db *sql.DB) (*ControlPlaneRepository, error)`
- methods:
  - `CreateTenant`
  - `ListTenants`
  - `CreateUser`
  - `ListUsers`
  - `CreateAgentProfile`
  - `ListAgentProfiles`

Rules:

- satisfy interfaces from `internal/domain/controlplane/repository.go`
- use plain SQL for this slice
- keep row mapping close to repository methods
- return deterministic ordering on list methods

- [ ] **Step 4: Re-run the constructor test and package compile check**

Run:

```bash
go test ./internal/infrastructure/persistence/postgres -run TestNewControlPlaneRepositoryRejectsNilDB -v
go test ./internal/infrastructure/persistence/postgres -run TestDoesNotExist
```

Expected:

- constructor test PASS
- package compile PASS

### Task 4: Add Control-Plane Application Commands and Queries

**Files:**

- Create: `internal/application/controlplane/command/create_tenant.go`
- Create: `internal/application/controlplane/command/create_user.go`
- Create: `internal/application/controlplane/command/create_agent_profile.go`
- Create: `internal/application/controlplane/query/list_tenants.go`
- Create: `internal/application/controlplane/query/list_users.go`
- Create: `internal/application/controlplane/query/list_agent_profiles.go`

- [ ] **Step 1: Write the failing application compile contract**

Extend `test/integration/admin_control_plane_test.go` with:

```go
package integration

import (
	"context"
	"testing"

	controlcommand "github.com/nikkofu/erp-claw/internal/application/controlplane/command"
)

func TestCreateTenantCommandRejectsEmptyCode(t *testing.T) {
	handler := controlcommand.CreateTenantHandler{}
	err := handler.Handle(context.Background(), controlcommand.CreateTenant{
		Code: "",
		Name: "Tenant A",
	})
	if err == nil {
		t.Fatal("expected empty code to fail")
	}
}
```

- [ ] **Step 2: Run the targeted test to verify RED**

Run:

```bash
go test ./test/integration -run TestCreateTenantCommandRejectsEmptyCode -v
```

Expected:

- FAIL because the application package does not exist yet

- [ ] **Step 3: Implement application handlers**

Create command and query handlers that:

- accept repository dependencies explicitly
- call domain constructors before persistence
- return validation/domain errors unchanged
- keep policy/audit hooks injectable later

Keep command handlers small and transport-agnostic.

- [ ] **Step 4: Re-run the targeted test and compile the new packages**

Run:

```bash
go test ./test/integration -run TestCreateTenantCommandRejectsEmptyCode -v
go test ./internal/application/controlplane/... -run TestDoesNotExist
```

Expected:

- targeted test PASS
- application packages compile

### Task 5: Expose Admin API for Tenant, User, and Agent Profile Catalog

**Files:**

- Modify: `internal/interfaces/http/router/admin.go`
- Modify: `internal/bootstrap/container.go`
- Test: `test/integration/admin_control_plane_test.go`

- [ ] **Step 1: Write the failing admin API test**

Extend `test/integration/admin_control_plane_test.go`:

```go
package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestAdminCreateTenantRoute(t *testing.T) {
	h := router.New(router.WithContainer(bootstrap.NewContainer(bootstrap.DefaultConfig())))
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/tenants", strings.NewReader(`{"code":"tenant-a","name":"Tenant A"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "platform-root")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run the admin API test to verify RED**

Run:

```bash
go test ./test/integration -run TestAdminCreateTenantRoute -v
```

Expected:

- FAIL because the route is not registered yet

- [ ] **Step 3: Wire the container and routes**

Update `internal/bootstrap/container.go` to hold:

- `ControlPlaneRepository`

Initialize it lazily or only when a DB is available through config wiring for this slice.

Update `internal/interfaces/http/router/admin.go` to add:

- `POST /api/admin/v1/tenants`
- `GET /api/admin/v1/tenants`
- `POST /api/admin/v1/users`
- `GET /api/admin/v1/users`
- `POST /api/admin/v1/agent-profiles`
- `GET /api/admin/v1/agent-profiles`

Rules:

- keep handlers thin
- bind DTOs locally in router file
- call application handlers only
- return `201` for create routes and `200` for list routes

- [ ] **Step 4: Re-run the admin API test**

Run:

```bash
go test ./test/integration -run TestAdminCreateTenantRoute -v
```

Expected:

- PASS

### Task 6: Wire API Runtime and Documentation

**Files:**

- Modify: `cmd/api-server/main.go`
- Modify: `README.md`

- [ ] **Step 1: Write the failing runtime compile expectation**

No new test file is needed here; the compile check is the contract.

- [ ] **Step 2: Update runtime wiring**

Update `cmd/api-server/main.go` so the runtime can construct the repository-backed container cleanly after config load.

Rules:

- do not embed SQL logic in `main.go`
- use the existing bootstrap assembly seam
- keep runtime startup readable and explicit

- [ ] **Step 3: Document the new control-plane slice**

Update `README.md` to describe:

- the new Admin API purpose
- the control-plane catalog entities in this slice
- that this is the first Phase 1 control-plane implementation, not the full control plane

- [ ] **Step 4: Run compile checks**

Run:

```bash
go test ./cmd/api-server -run TestDoesNotExist
```

Expected:

- PASS

### Task 7: Run Full Verification

**Files:**

- Verify only

- [ ] **Step 1: Run targeted package tests**

Run:

```bash
go test ./internal/domain/controlplane -v
go test ./internal/infrastructure/persistence/postgres -run TestNewControlPlaneRepositoryRejectsNilDB -v
go test ./test/integration -run 'Test(ControlPlaneMigrationContainsCatalogTables|CreateTenantCommandRejectsEmptyCode|AdminCreateTenantRoute)' -v
```

Expected:

- PASS

- [ ] **Step 2: Run full suite**

Run:

```bash
go test ./...
```

Expected:

- PASS

- [ ] **Step 3: Commit**

Run:

```bash
git add cmd internal migrations README.md test docs/superpowers/plans/2026-03-25-phase1-control-plane-core-implementation-plan.md
git commit -m "feat: add phase1 control plane core"
```

## Review Checklist

Before marking this slice complete, confirm all of the following:

- the project now has a real `internal/domain` layer
- tenant, user, role, and agent profile are persisted rather than remaining pure placeholders
- Admin API exposes explicit control-plane catalog actions
- application handlers stay transport-agnostic
- repository logic stays in infrastructure
- no supply-chain domain code is mixed into this slice
- the implementation remains a Phase 1 control-plane slice, not a premature full workflow/policy engine

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-03-25-phase1-control-plane-core-implementation-plan.md`.

Two execution options:

1. Subagent-Driven (recommended) - I dispatch a fresh subagent per task, review between tasks, fast iteration
2. Inline Execution - Execute tasks in this session using executing-plans, batch execution with checkpoints

This session already established a subagent-driven preference, so the default next step is to execute this plan with subagents unless redirected.
