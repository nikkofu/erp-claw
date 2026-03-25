# Agentic AI-Native ERP Platform Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Phase 0/1 platform foundation for the Go-based AI-native ERP: runtime entrypoints, local infrastructure via `docker-compose.yml`, bootstrap/config, tenant routing, HTTP middleware, persistence seams, eventing seams, and the first policy/audit execution pipeline.

**Architecture:** This plan intentionally covers the first executable sub-project from the approved ERP spec rather than the entire ERP at once. The implementation should produce a platform-oriented modular skeleton with strong package boundaries so later business domains can attach without reworking the runtime, tenant routing, or execution model.

**Tech Stack:** Go 1.24, Gin, PostgreSQL, Redis, NATS JetStream, MinIO, OpenTelemetry, Docker Compose, Make, golang-migrate, Testify

---

**Spec Reference:** `docs/superpowers/specs/2026-03-25-agentic-ai-native-erp-design.md`

**Scope Note:** The approved design spans multiple independent subsystems. This plan covers only the first shippable slice:

- repository and runtime skeleton
- local infrastructure contract
- process entrypoints
- platform bootstrap and configuration
- tenant routing and request context propagation
- persistence and migration seams
- event bus, worker, and scheduler seams
- policy and audit command pipeline skeleton

**Workspace Note:** The current workspace snapshot is not a git repository. Commit steps are included because the implementation flow should use small commits once the repository is initialized or checked out under version control.

## File Structure Map

The implementation produced by this plan should create or modify the following structure.

```text
erp-claw/
  cmd/
    api-server/main.go
    agent-gateway/main.go
    worker/main.go
    scheduler/main.go
    migrate/main.go
  configs/
    local/app.yaml
    local/docker.env
  internal/
    bootstrap/
      app.go
      config.go
      container.go
      runtime.go
    platform/
      audit/
        model.go
        recorder.go
      eventbus/
        bus.go
        nats.go
        memory.go
      iam/
        actor.go
      policy/
        decision.go
        evaluator.go
        static.go
      runtime/
        request_context.go
      tenant/
        cell_route.go
        resolver.go
      health/
        service.go
    application/
      shared/
        command.go
        pipeline.go
        transaction.go
    interfaces/
      http/
        router/
          admin.go
          platform.go
          workspace.go
          integration.go
          health.go
          router.go
        middleware/
          request_id.go
          logging.go
          tenant.go
          auth.go
          audit.go
        presenter/
          response.go
    infrastructure/
      persistence/
        postgres/
          db.go
          migrate.go
          tx.go
      cache/
        redis/
          client.go
      messaging/
        nats/
          client.go
      storage/
        minio/
          client.go
      observability/
        otel/
          setup.go
  migrations/
    000001_init_platform_tables.up.sql
    000001_init_platform_tables.down.sql
  test/
    integration/
      api_health_test.go
      tenant_resolution_test.go
      command_pipeline_test.go
  docker-compose.yml
  Makefile
  go.mod
  go.sum
  README.md
```

### Task 1: Initialize the Go Module and Runtime Skeleton

**Files:**
- Create: `go.mod`
- Create: `README.md`
- Create: `cmd/api-server/main.go`
- Create: `cmd/agent-gateway/main.go`
- Create: `cmd/worker/main.go`
- Create: `cmd/scheduler/main.go`
- Create: `cmd/migrate/main.go`
- Create: `internal/bootstrap/runtime.go`
- Test: `internal/bootstrap/runtime_test.go`

- [ ] **Step 1: Write the failing bootstrap test**

Create `internal/bootstrap/runtime_test.go` with a first contract test that proves the runtime role metadata is wired correctly:

```go
package bootstrap

import "testing"

func TestRuntimeRoleString(t *testing.T) {
    if APIServerRole.String() != "api-server" {
        t.Fatalf("expected api-server, got %q", APIServerRole.String())
    }
}
```

- [ ] **Step 2: Run the test to verify the package does not exist yet**

Run:

```bash
go test ./internal/bootstrap -run TestRuntimeRoleString -v
```

Expected:

- FAIL with missing package or undefined identifiers

- [ ] **Step 3: Create the module and runtime role definitions**

Create `go.mod` with the initial module and dependencies:

```go
module github.com/nikkofu/erp-claw

go 1.24

require (
    github.com/gin-gonic/gin v1.10.1
    github.com/stretchr/testify v1.10.0
)
```

Create `internal/bootstrap/runtime.go`:

```go
package bootstrap

type RuntimeRole string

const (
    APIServerRole    RuntimeRole = "api-server"
    AgentGatewayRole RuntimeRole = "agent-gateway"
    WorkerRole       RuntimeRole = "worker"
    SchedulerRole    RuntimeRole = "scheduler"
    MigrateRole      RuntimeRole = "migrate"
)

func (r RuntimeRole) String() string {
    return string(r)
}
```

Create one minimal `main.go` per runtime under `cmd/...` that imports `internal/bootstrap` and prints or starts the named role through a shared bootstrap API stub. Keep the implementation minimal and identical in shape.

- [ ] **Step 4: Re-run the test and a compile check**

Run:

```bash
go test ./internal/bootstrap -run TestRuntimeRoleString -v
go test ./cmd/... -run TestDoesNotExist
```

Expected:

- `TestRuntimeRoleString` PASS
- all command packages compile

- [ ] **Step 5: Add a minimal README describing runtime roles**

Document:

- what the project is
- the five runtime entrypoints
- that third-party dependencies are started through `docker-compose.yml`

- [ ] **Step 6: Commit**

Run:

```bash
git add go.mod README.md cmd internal/bootstrap/runtime.go internal/bootstrap/runtime_test.go
git commit -m "chore: initialize go runtime skeleton"
```

### Task 2: Add Local Infrastructure Contract with Docker Compose and Make Targets

**Files:**
- Create: `docker-compose.yml`
- Create: `configs/local/docker.env`
- Create: `Makefile`
- Modify: `README.md`
- Test: `test/integration/compose_contract_test.go`

- [ ] **Step 1: Write the failing infrastructure contract test**

Create `test/integration/compose_contract_test.go`:

```go
package integration

import (
    "os"
    "strings"
    "testing"
)

func TestDockerComposeIncludesRequiredServices(t *testing.T) {
    data, err := os.ReadFile("../../docker-compose.yml")
    if err != nil {
        t.Fatalf("read compose: %v", err)
    }

    required := []string{"postgres", "redis", "nats", "minio", "otel-collector", "prometheus", "grafana"}
    content := string(data)
    for _, service := range required {
        if !strings.Contains(content, service+":") {
            t.Fatalf("expected service %q in compose file", service)
        }
    }
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./test/integration -run TestDockerComposeIncludesRequiredServices -v
```

Expected:

- FAIL because `docker-compose.yml` does not exist

- [ ] **Step 3: Create the compose contract and Make targets**

Create `docker-compose.yml` with these services:

- `postgres`
- `redis`
- `nats`
- `minio`
- `otel-collector`
- `prometheus`
- `grafana`

Use named volumes for stateful services and publish predictable local ports.

Create `configs/local/docker.env` with defaults such as:

```env
POSTGRES_DB=erp_claw
POSTGRES_USER=erp
POSTGRES_PASSWORD=erp
REDIS_PORT=6379
NATS_PORT=4222
MINIO_ROOT_USER=minio
MINIO_ROOT_PASSWORD=minio123
```

Create `Makefile` targets:

- `infra-up`
- `infra-down`
- `infra-reset`
- `api`
- `gateway`
- `worker`
- `scheduler`
- `test`
- `test-integration`

- [ ] **Step 4: Re-run the test and validate compose syntax**

Run:

```bash
go test ./test/integration -run TestDockerComposeIncludesRequiredServices -v
docker compose config >/tmp/erp-claw-compose.rendered.yaml
```

Expected:

- test PASS
- compose config exits `0`

- [ ] **Step 5: Update the README with local startup instructions**

Document:

- `make infra-up`
- how to inspect service ports
- that Go services run outside Compose in local development

- [ ] **Step 6: Commit**

Run:

```bash
git add docker-compose.yml configs/local/docker.env Makefile README.md test/integration/compose_contract_test.go
git commit -m "chore: add local infrastructure contract"
```

### Task 3: Build Bootstrap Configuration and Dependency Container

**Files:**
- Create: `configs/local/app.yaml`
- Create: `internal/bootstrap/config.go`
- Create: `internal/bootstrap/container.go`
- Create: `internal/platform/health/service.go`
- Test: `internal/bootstrap/config_test.go`

- [ ] **Step 1: Write the failing config-loading test**

Create `internal/bootstrap/config_test.go`:

```go
package bootstrap

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
    cfg, err := LoadConfig("testdata/does-not-need-to-exist-yet")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if cfg.HTTP.Port != 8080 {
        t.Fatalf("expected HTTP port 8080, got %d", cfg.HTTP.Port)
    }
}
```

- [ ] **Step 2: Run the test to confirm it fails**

Run:

```bash
go test ./internal/bootstrap -run TestLoadConfigDefaults -v
```

Expected:

- FAIL with undefined `LoadConfig` or config types

- [ ] **Step 3: Implement configuration and dependency container**

Create `internal/bootstrap/config.go` with:

- `Config` root struct
- nested `HTTP`, `Database`, `Redis`, `NATS`, `ObjectStorage`, `Telemetry`
- `LoadConfig(path string) (Config, error)`
- sensible local defaults when config file is absent

Minimal shape:

```go
type Config struct {
    Env   string
    HTTP  HTTPConfig
    Store StoreConfig
}
```

Create `configs/local/app.yaml` with matching local values.

Create `internal/bootstrap/container.go` with a lightweight dependency container:

```go
type Container struct {
    Config Config
    Health *health.Service
}
```

Create `internal/platform/health/service.go` exposing liveness and readiness booleans.

- [ ] **Step 4: Run tests and compile the bootstrap package**

Run:

```bash
go test ./internal/bootstrap -run TestLoadConfigDefaults -v
go test ./internal/platform/health -run TestDoesNotExist
```

Expected:

- config test PASS
- health package compiles

- [ ] **Step 5: Document config precedence in the README**

Describe:

- file config
- env overrides
- local profile conventions

- [ ] **Step 6: Commit**

Run:

```bash
git add configs/local/app.yaml internal/bootstrap/config.go internal/bootstrap/container.go internal/platform/health/service.go internal/bootstrap/config_test.go README.md
git commit -m "feat: add bootstrap config and container"
```

### Task 4: Add API Server Router, Health Endpoints, and Response Presenter

**Files:**
- Create: `internal/interfaces/http/router/router.go`
- Create: `internal/interfaces/http/router/health.go`
- Create: `internal/interfaces/http/router/admin.go`
- Create: `internal/interfaces/http/router/platform.go`
- Create: `internal/interfaces/http/router/workspace.go`
- Create: `internal/interfaces/http/router/integration.go`
- Create: `internal/interfaces/http/presenter/response.go`
- Modify: `cmd/api-server/main.go`
- Test: `test/integration/api_health_test.go`

- [ ] **Step 1: Write the failing API health test**

Create `test/integration/api_health_test.go`:

```go
package integration

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestHealthRoutes(t *testing.T) {
    h := router.New()
    req := httptest.NewRequest(http.MethodGet, "/api/platform/v1/health/livez", nil)
    rec := httptest.NewRecorder()

    h.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./test/integration -run TestHealthRoutes -v
```

Expected:

- FAIL because router package does not exist

- [ ] **Step 3: Implement the router and presenter**

Create `internal/interfaces/http/presenter/response.go`:

```go
package presenter

import "github.com/gin-gonic/gin"

func OK(c *gin.Context, data any) {
    c.JSON(200, gin.H{
        "data": data,
        "meta": gin.H{"request_id": c.GetString("request_id")},
    })
}
```

Create grouped route registrars:

- `admin.go`
- `platform.go`
- `workspace.go`
- `integration.go`
- `health.go`

Create `router.go` returning a configured `*gin.Engine`.

Wire `/api/platform/v1/health/livez` and `/api/platform/v1/health/readyz`.

Update `cmd/api-server/main.go` to:

- load config
- build container
- build router
- listen on configured port

- [ ] **Step 4: Re-run the health test and a package-level compile**

Run:

```bash
go test ./test/integration -run TestHealthRoutes -v
go test ./internal/interfaces/http/... -run TestDoesNotExist
```

Expected:

- health test PASS
- HTTP interface packages compile

- [ ] **Step 5: Add a smoke run command to the README**

Document:

```bash
go run ./cmd/api-server
curl -sS http://127.0.0.1:8080/api/platform/v1/health/livez
```

- [ ] **Step 6: Commit**

Run:

```bash
git add cmd/api-server/main.go internal/interfaces/http/router internal/interfaces/http/presenter test/integration/api_health_test.go README.md
git commit -m "feat: add api server router and health endpoints"
```

### Task 5: Implement Request Context and Tenant Resolution Middleware

**Files:**
- Create: `internal/platform/runtime/request_context.go`
- Create: `internal/platform/tenant/cell_route.go`
- Create: `internal/platform/tenant/resolver.go`
- Create: `internal/platform/iam/actor.go`
- Create: `internal/interfaces/http/middleware/request_id.go`
- Create: `internal/interfaces/http/middleware/logging.go`
- Create: `internal/interfaces/http/middleware/tenant.go`
- Create: `internal/interfaces/http/middleware/auth.go`
- Create: `internal/interfaces/http/middleware/audit.go`
- Modify: `internal/interfaces/http/router/router.go`
- Test: `test/integration/tenant_resolution_test.go`

- [ ] **Step 1: Write the failing tenant resolution test**

Create `test/integration/tenant_resolution_test.go`:

```go
package integration

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestTenantResolutionFromHeader(t *testing.T) {
    h := router.New()
    req := httptest.NewRequest(http.MethodGet, "/api/platform/v1/health/livez", nil)
    req.Header.Set("X-Tenant-ID", "tenant-a")

    rec := httptest.NewRecorder()
    h.ServeHTTP(rec, req)

    if rec.Header().Get("X-Tenant-ID") != "tenant-a" {
        t.Fatalf("expected tenant header to round-trip through middleware")
    }
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./test/integration -run TestTenantResolutionFromHeader -v
```

Expected:

- FAIL because tenant middleware is not installed

- [ ] **Step 3: Implement request and tenant context objects**

Create `internal/platform/runtime/request_context.go`:

```go
package runtime

type RequestContext struct {
    RequestID string
    TenantID  string
    ActorID   string
    TraceID   string
}
```

Create `internal/platform/tenant/cell_route.go` and `resolver.go`:

```go
package tenant

type CellRoute struct {
    TenantID     string
    Isolation    string
    DatabaseDSN  string
    CachePrefix  string
    StoragePrefix string
}

type Resolver interface {
    Resolve(tenantID string) (CellRoute, error)
}
```

Add middleware for:

- request ID injection
- request logging
- tenant resolution from header
- placeholder auth actor injection
- audit metadata injection

Install the middleware stack in `router.go`.

- [ ] **Step 4: Re-run the test and add one package compile**

Run:

```bash
go test ./test/integration -run TestTenantResolutionFromHeader -v
go test ./internal/platform/tenant -run TestDoesNotExist
```

Expected:

- tenant resolution test PASS
- tenant package compiles

- [ ] **Step 5: Add a short architecture note in the README**

Document:

- request path
- request ID
- tenant header in local mode
- actor placeholder behavior until IAM is implemented

- [ ] **Step 6: Commit**

Run:

```bash
git add internal/platform/runtime internal/platform/tenant internal/platform/iam internal/interfaces/http/middleware internal/interfaces/http/router/router.go test/integration/tenant_resolution_test.go README.md
git commit -m "feat: add request context and tenant middleware"
```

### Task 6: Add PostgreSQL, Redis, NATS, and MinIO Infrastructure Clients plus Migration Skeleton

**Files:**
- Create: `internal/infrastructure/persistence/postgres/db.go`
- Create: `internal/infrastructure/persistence/postgres/tx.go`
- Create: `internal/infrastructure/persistence/postgres/migrate.go`
- Create: `internal/infrastructure/cache/redis/client.go`
- Create: `internal/infrastructure/messaging/nats/client.go`
- Create: `internal/infrastructure/storage/minio/client.go`
- Create: `migrations/000001_init_platform_tables.up.sql`
- Create: `migrations/000001_init_platform_tables.down.sql`
- Modify: `cmd/migrate/main.go`
- Test: `internal/infrastructure/persistence/postgres/db_test.go`

- [ ] **Step 1: Write the failing database bootstrap test**

Create `internal/infrastructure/persistence/postgres/db_test.go`:

```go
package postgres

import "testing"

func TestConfigValidationRejectsEmptyDSN(t *testing.T) {
    _, err := New(Config{})
    if err == nil {
        t.Fatal("expected error for empty DSN")
    }
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./internal/infrastructure/persistence/postgres -run TestConfigValidationRejectsEmptyDSN -v
```

Expected:

- FAIL because package or constructor does not exist

- [ ] **Step 3: Implement infrastructure clients and migration seam**

Create `db.go`:

```go
package postgres

import (
    "database/sql"
    "errors"
)

type Config struct {
    DSN string
}

func New(cfg Config) (*sql.DB, error) {
    if cfg.DSN == "" {
        return nil, errors.New("postgres dsn required")
    }
    return sql.Open("pgx", cfg.DSN)
}
```

Create:

- `tx.go` with a transaction helper interface for application handlers
- `redis/client.go` with config validation and constructor
- `nats/client.go` with config validation and constructor
- `minio/client.go` with config validation and constructor

Add initial migration SQL for platform control tables:

- `tenant`
- `tenant_cell`
- `audit_log`
- `agent_session`
- `agent_task`
- `outbox`

Wire `cmd/migrate/main.go` to load config and apply migrations.

- [ ] **Step 4: Run tests and package compile checks**

Run:

```bash
go test ./internal/infrastructure/persistence/postgres -run TestConfigValidationRejectsEmptyDSN -v
go test ./internal/infrastructure/... -run TestDoesNotExist
```

Expected:

- DB config test PASS
- infrastructure packages compile

- [ ] **Step 5: Validate the migration file shape**

Run:

```bash
rg "create table" migrations/000001_init_platform_tables.up.sql
```

Expected:

- output includes the initial platform tables listed above

- [ ] **Step 6: Commit**

Run:

```bash
git add internal/infrastructure migrations cmd/migrate/main.go
git commit -m "feat: add infrastructure clients and migration skeleton"
```

### Task 7: Add Event Bus, Worker, and Scheduler Seams

**Files:**
- Create: `internal/platform/eventbus/bus.go`
- Create: `internal/platform/eventbus/memory.go`
- Create: `internal/platform/eventbus/nats.go`
- Modify: `cmd/worker/main.go`
- Modify: `cmd/scheduler/main.go`
- Test: `internal/platform/eventbus/memory_test.go`

- [ ] **Step 1: Write the failing in-memory event bus test**

Create `internal/platform/eventbus/memory_test.go`:

```go
package eventbus

import (
    "context"
    "testing"
)

func TestMemoryBusPublish(t *testing.T) {
    bus := NewMemory()
    err := bus.Publish(context.Background(), Event{
        Topic: "tenant.created",
        Payload: map[string]any{"tenant_id": "tenant-a"},
    })
    if err != nil {
        t.Fatalf("publish: %v", err)
    }
}
```

- [ ] **Step 2: Run the test to confirm failure**

Run:

```bash
go test ./internal/platform/eventbus -run TestMemoryBusPublish -v
```

Expected:

- FAIL because the bus contract is missing

- [ ] **Step 3: Implement the bus contract and role skeletons**

Create `bus.go`:

```go
package eventbus

import "context"

type Event struct {
    Topic       string
    TenantID    string
    Correlation string
    Payload     any
}

type Bus interface {
    Publish(ctx context.Context, evt Event) error
}
```

Create `memory.go` with a simple in-memory bus implementation used in tests.

Create `nats.go` with a `Bus` implementation backed by NATS JetStream.

Update `cmd/worker/main.go` to:

- load config
- connect dependencies
- log startup
- leave TODO-free placeholder loops for outbox polling

Update `cmd/scheduler/main.go` similarly with a ticker-driven skeleton.

- [ ] **Step 4: Run tests and compile checks**

Run:

```bash
go test ./internal/platform/eventbus -run TestMemoryBusPublish -v
go test ./cmd/worker ./cmd/scheduler -run TestDoesNotExist
```

Expected:

- event bus test PASS
- worker and scheduler compile

- [ ] **Step 5: Add one integration note to the README**

Document:

- worker consumes outbox in later tasks
- scheduler emits time-based commands rather than mutating data directly

- [ ] **Step 6: Commit**

Run:

```bash
git add internal/platform/eventbus cmd/worker/main.go cmd/scheduler/main.go README.md
git commit -m "feat: add event bus and runtime worker skeleton"
```

### Task 8: Implement Policy, Audit, Transaction, and Command Pipeline Foundations

**Files:**
- Create: `internal/platform/policy/decision.go`
- Create: `internal/platform/policy/evaluator.go`
- Create: `internal/platform/policy/static.go`
- Create: `internal/platform/audit/model.go`
- Create: `internal/platform/audit/recorder.go`
- Create: `internal/application/shared/command.go`
- Create: `internal/application/shared/transaction.go`
- Create: `internal/application/shared/pipeline.go`
- Test: `test/integration/command_pipeline_test.go`

- [ ] **Step 1: Write the failing command pipeline test**

Create `test/integration/command_pipeline_test.go`:

```go
package integration

import (
    "context"
    "testing"

    "github.com/nikkofu/erp-claw/internal/application/shared"
    "github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestCommandPipelineRejectsDeniedPolicy(t *testing.T) {
    pipe := shared.NewPipeline(
        shared.PipelineDeps{
            Policy: policy.StaticEvaluator(policy.DecisionDeny),
        },
    )

    err := pipe.Execute(context.Background(), shared.Command{
        Name: "CreateTenant",
    })
    if err == nil {
        t.Fatal("expected denied policy to fail the command")
    }
}
```

- [ ] **Step 2: Run the test to confirm failure**

Run:

```bash
go test ./test/integration -run TestCommandPipelineRejectsDeniedPolicy -v
```

Expected:

- FAIL because shared command pipeline and policy contracts do not exist

- [ ] **Step 3: Implement policy and audit contracts plus pipeline**

Create `internal/platform/policy/decision.go`:

```go
package policy

type Decision string

const (
    DecisionAllow            Decision = "ALLOW"
    DecisionAllowWithGuard   Decision = "ALLOW_WITH_GUARD"
    DecisionRequireApproval  Decision = "REQUIRE_APPROVAL"
    DecisionDeny             Decision = "DENY"
)
```

Create `evaluator.go` with:

```go
type Evaluator interface {
    Evaluate(ctx context.Context, input Input) (Decision, error)
}
```

Create `static.go` with a deterministic test implementation.

Create `internal/platform/audit/model.go` and `recorder.go` with:

- `Record` model
- `Recorder` interface
- in-memory or noop recorder for bootstrap use

Create `internal/application/shared/command.go`:

```go
package shared

type Command struct {
    Name     string
    TenantID string
    ActorID  string
    Payload  any
}
```

Create `transaction.go` with:

- transaction manager interface
- noop transaction manager for tests

Create `pipeline.go` implementing:

- policy evaluation
- transaction boundary
- handler invocation
- audit recording

- [ ] **Step 4: Re-run the test and add one package compile**

Run:

```bash
go test ./test/integration -run TestCommandPipelineRejectsDeniedPolicy -v
go test ./internal/application/shared ./internal/platform/policy ./internal/platform/audit -run TestDoesNotExist
```

Expected:

- command pipeline test PASS
- shared application and platform packages compile

- [ ] **Step 5: Add one README section for the command pipeline**

Document:

- commands enter application handlers
- policy executes before mutation
- audit executes around mutation
- later domain modules will plug into this same pipeline

- [ ] **Step 6: Commit**

Run:

```bash
git add internal/platform/policy internal/platform/audit internal/application/shared test/integration/command_pipeline_test.go README.md
git commit -m "feat: add policy audit and command pipeline foundation"
```

### Task 9: Wire Agent Gateway and Workspace Event Streaming Skeleton

**Files:**
- Modify: `cmd/agent-gateway/main.go`
- Create: `internal/interfaces/ws/workspace_gateway.go`
- Create: `internal/platform/runtime/workspace_event.go`
- Test: `internal/interfaces/ws/workspace_gateway_test.go`

- [ ] **Step 1: Write the failing workspace gateway test**

Create `internal/interfaces/ws/workspace_gateway_test.go`:

```go
package ws

import "testing"

func TestGatewayRegistersWorkspaceChannel(t *testing.T) {
    g := NewWorkspaceGateway()
    if g == nil {
        t.Fatal("expected workspace gateway")
    }
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./internal/interfaces/ws -run TestGatewayRegistersWorkspaceChannel -v
```

Expected:

- FAIL because the workspace gateway package does not exist

- [ ] **Step 3: Implement the workspace gateway skeleton**

Create `internal/platform/runtime/workspace_event.go`:

```go
package runtime

type WorkspaceEvent struct {
    Type      string
    TenantID  string
    SessionID string
    TaskID    string
    Payload   any
}
```

Create `internal/interfaces/ws/workspace_gateway.go` with:

- gateway struct
- channel registration skeleton
- event broadcast method stub
- comment explaining that actual WebSocket protocol and Agent task streaming land in later plans

Update `cmd/agent-gateway/main.go` to boot this gateway alongside config and container initialization.

- [ ] **Step 4: Re-run the test and compile the gateway**

Run:

```bash
go test ./internal/interfaces/ws -run TestGatewayRegistersWorkspaceChannel -v
go test ./cmd/agent-gateway -run TestDoesNotExist
```

Expected:

- workspace gateway test PASS
- agent-gateway compiles

- [ ] **Step 5: Update the README with runtime role purpose**

Add a short section for:

- `api-server`: HTTP APIs
- `agent-gateway`: workspace and agent event entry
- `worker`: async processing
- `scheduler`: timed orchestration

- [ ] **Step 6: Commit**

Run:

```bash
git add cmd/agent-gateway/main.go internal/interfaces/ws internal/platform/runtime/workspace_event.go README.md
git commit -m "feat: add agent gateway workspace skeleton"
```

### Task 10: Add End-to-End Local Verification and Developer Workflow

**Files:**
- Modify: `Makefile`
- Modify: `README.md`
- Create: `scripts/smoke_local.sh`
- Modify: `test/integration/api_health_test.go`

- [ ] **Step 1: Extend the health integration test into a local smoke target**

Update `test/integration/api_health_test.go` to include a real-http version gated behind an environment variable:

```go
func TestHealthRoutesLive(t *testing.T) {
    if os.Getenv("ERP_CLAW_LIVE_SMOKE") != "1" {
        t.Skip("live smoke disabled")
    }
    // hit http://127.0.0.1:8080/api/platform/v1/health/livez
}
```

- [ ] **Step 2: Run the current test suite before adding the script**

Run:

```bash
go test ./...
```

Expected:

- PASS for all current tests

- [ ] **Step 3: Add the smoke script and final Make targets**

Create `scripts/smoke_local.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

export ERP_CLAW_LIVE_SMOKE=1
go test ./test/integration -run TestHealthRoutesLive -v
```

Update `Makefile` with:

- `smoke`
- `migrate-up`
- `migrate-down`

Update `README.md` with the full local workflow:

1. `make infra-up`
2. `go run ./cmd/api-server`
3. `go run ./cmd/agent-gateway`
4. `make test`
5. `make smoke`

- [ ] **Step 4: Run verification commands**

Run:

```bash
go test ./...
shellcheck scripts/smoke_local.sh
```

Expected:

- unit and integration tests PASS
- smoke script passes shellcheck

- [ ] **Step 5: Perform a manual local startup check**

Run:

```bash
make infra-up
go run ./cmd/api-server &
sleep 2
curl -sS http://127.0.0.1:8080/api/platform/v1/health/livez
```

Expected:

- JSON response with `data.status` or equivalent healthy payload

- [ ] **Step 6: Commit**

Run:

```bash
git add Makefile README.md scripts/smoke_local.sh test/integration/api_health_test.go
git commit -m "chore: add local verification workflow"
```

## Review Checklist

Before executing this plan, confirm that the resulting scaffold satisfies all of the following:

- runtime entrypoints exist for all five process roles
- local infrastructure is managed by `docker-compose.yml`
- tenant routing is explicit in the request path
- command execution is policy-gated
- audit exists as a cross-cutting platform concern
- infrastructure packages stay below application and domain packages in dependency direction
- the initial migration creates only platform tables, not premature ERP business tables
- the Agent Gateway is present as a runtime seam without prematurely implementing the full protocol

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-03-25-platform-foundation-implementation-plan.md`.

Two execution options:

1. Subagent-Driven (recommended) - dispatch a fresh subagent per task, review between tasks, fast iteration
2. Inline Execution - execute tasks in this session using executing-plans, batch execution with checkpoints

In this session I cannot automatically use subagents unless you explicitly ask for delegated/subagent execution, so the practical default is inline execution. If you want, say either:

- `开始按计划内联执行`
- `用子代理执行`
