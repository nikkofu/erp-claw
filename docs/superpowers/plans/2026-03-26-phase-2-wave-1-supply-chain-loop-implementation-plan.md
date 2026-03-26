# Phase 2 Wave 1 Supply-Chain Loop Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first executable Phase 2 business slice: supplier/product/warehouse master data, purchase order creation and submission, approval progression, and admin HTTP endpoints to drive that workflow.

**Architecture:** Keep domain rules inside new `internal/domain/...` packages for master data, procurement, and approval. Use an application-level supply-chain service to orchestrate repository access and the existing command pipeline, then expose the slice through admin routes and a thin in-memory persistence adapter so the workflow is runnable before PostgreSQL-backed repositories are introduced.

**Tech Stack:** Go 1.25, Gin, existing command pipeline/audit/policy seams, in-memory repositories for Phase 2 Wave 1 runtime wiring, Testify-free standard-library tests

---

**Spec Reference:** `docs/superpowers/specs/2026-03-25-agentic-ai-native-erp-design.md`

**Coverage Reference:** `docs/phase-2-coverage-status.md`

**Scope Note:** This plan intentionally narrows Phase 2 to the `Wave 1` slice already recommended in the coverage doc:

- supplier, product, warehouse master data
- purchase order draft creation
- purchase order submission into approval
- approval approve / reject actions
- admin API for the above workflow
- purchase order detail query for workflow state inspection

**Out of Scope For This Plan:**

- receipt / inbound inventory posting
- stock ledger and available/reserved calculations
- payable bill generation
- sales order loop
- PostgreSQL repository implementation beyond forward-looking schema preparation

## File Structure Map

The implementation produced by this plan should create or modify the following structure.

```text
erp-claw/
  docs/
    superpowers/
      plans/
        2026-03-26-phase-2-wave-1-supply-chain-loop-implementation-plan.md
  internal/
    application/
      admin/
        supplychain/
          service.go
          service_test.go
          commands.go
    bootstrap/
      container.go
    domain/
      approval/
        approval.go
        repository.go
      masterdata/
        product.go
        supplier.go
        warehouse.go
        repository.go
      procurement/
        purchase_order.go
        repository.go
    infrastructure/
      persistence/
        memory/
          supplychain.go
    interfaces/
      http/
        router/
          admin.go
  migrations/
    000002_init_phase2_wave1_tables.up.sql
    000002_init_phase2_wave1_tables.down.sql
  test/
    integration/
      admin_supply_chain_test.go
```

### Task 1: Model Master Data, Purchase Orders, and Approval State Machines

**Files:**
- Create: `internal/domain/masterdata/supplier.go`
- Create: `internal/domain/masterdata/product.go`
- Create: `internal/domain/masterdata/warehouse.go`
- Create: `internal/domain/masterdata/repository.go`
- Create: `internal/domain/procurement/purchase_order.go`
- Create: `internal/domain/procurement/repository.go`
- Create: `internal/domain/approval/approval.go`
- Create: `internal/domain/approval/repository.go`
- Test: `internal/application/admin/supplychain/service_test.go`

- [ ] **Step 1: Write the failing workflow unit test**

Create `internal/application/admin/supplychain/service_test.go` with a test that exercises the intended business path:

```go
func TestServiceCreatesAndSubmitsPurchaseOrderForApproval(t *testing.T) {
	ctx := context.Background()
	svc := NewService(ServiceDeps{
		MasterData: memory.NewMasterDataRepository(),
		PurchaseOrders: memory.NewPurchaseOrderRepository(),
		Approvals: memory.NewApprovalRepository(),
		Pipeline: shared.NewPipeline(shared.PipelineDeps{}),
	})

	supplier, err := svc.CreateSupplier(ctx, CreateSupplierInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Code:     "SUP-001",
		Name:     "Acme Supply",
	})
	if err != nil {
		t.Fatalf("create supplier: %v", err)
	}

	product, err := svc.CreateProduct(ctx, CreateProductInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		SKU:      "SKU-001",
		Name:     "Copper Wire",
		Unit:     "roll",
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	warehouse, err := svc.CreateWarehouse(ctx, CreateWarehouseInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Code:     "WH-SH",
		Name:     "Shanghai Warehouse",
	})
	if err != nil {
		t.Fatalf("create warehouse: %v", err)
	}

	order, err := svc.CreatePurchaseOrder(ctx, CreatePurchaseOrderInput{
		TenantID:    "tenant-a",
		ActorID:     "actor-a",
		SupplierID:  supplier.ID,
		WarehouseID: warehouse.ID,
		Lines: []CreatePurchaseOrderLine{{
			ProductID: product.ID,
			Quantity:  5,
		}},
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.Status != procurement.PurchaseOrderStatusDraft {
		t.Fatalf("expected draft, got %s", order.Status)
	}

	submitted, approvalRequest, err := svc.SubmitPurchaseOrder(ctx, SubmitPurchaseOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "actor-a",
		PurchaseOrderID: order.ID,
	})
	if err != nil {
		t.Fatalf("submit order: %v", err)
	}

	if submitted.Status != procurement.PurchaseOrderStatusPendingApproval {
		t.Fatalf("expected pending approval, got %s", submitted.Status)
	}
	if approvalRequest.Status != approval.StatusPending {
		t.Fatalf("expected approval pending, got %s", approvalRequest.Status)
	}
	if approvalRequest.ResourceID != submitted.ID {
		t.Fatalf("expected approval resource id %s, got %s", submitted.ID, approvalRequest.ResourceID)
	}
}
```

- [ ] **Step 2: Run the test to verify the package does not exist yet**

Run:

```bash
go test ./internal/application/admin/supplychain -run TestServiceCreatesAndSubmitsPurchaseOrderForApproval -v
```

Expected:

- FAIL with missing package / undefined identifiers for the new service and domain types

- [ ] **Step 3: Implement the minimal domain model and contracts**

Create the domain packages with the minimum rules needed to satisfy the test:

- `masterdata` entities with required identifiers and validation
- `procurement.PurchaseOrder` with `Draft`, `PendingApproval`, `Approved`, `Rejected` statuses
- `approval.Request` with `Pending`, `Approved`, `Rejected` statuses
- repository interfaces keyed by tenant-aware IDs

Minimum shape to preserve:

```go
type PurchaseOrder struct {
	ID           string
	TenantID     string
	SupplierID   string
	WarehouseID  string
	Status       PurchaseOrderStatus
	Lines        []Line
	ApprovalID   string
}

func (po *PurchaseOrder) Submit(approvalID string) error
func (po *PurchaseOrder) MarkApproved() error
func (po *PurchaseOrder) MarkRejected() error
```

- [ ] **Step 4: Re-run the workflow test and keep it green**

Run:

```bash
go test ./internal/application/admin/supplychain -run TestServiceCreatesAndSubmitsPurchaseOrderForApproval -v
```

Expected:

- PASS

- [ ] **Step 5: Add focused rule tests for invalid transitions**

Expand `internal/application/admin/supplychain/service_test.go` or add adjacent tests covering:

- submit fails for empty lines
- submit fails when order is not `Draft`
- approve fails when approval request is not `Pending`
- reject fails when approval request is already terminal

- [ ] **Step 6: Run the package tests**

Run:

```bash
go test ./internal/application/admin/supplychain -v
```

Expected:

- PASS

### Task 2: Add the Application Service and In-Memory Persistence Adapter

**Files:**
- Create: `internal/application/admin/supplychain/commands.go`
- Create: `internal/application/admin/supplychain/service.go`
- Create: `internal/infrastructure/persistence/memory/supplychain.go`
- Modify: `internal/bootstrap/container.go`
- Test: `internal/application/admin/supplychain/service_test.go`

- [ ] **Step 1: Write the next failing service test for approval completion**

Add:

```go
func TestServiceApprovesSubmittedPurchaseOrder(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	order, approvalRequest := createSubmittedOrder(t, ctx, svc)

	approvedOrder, approvedRequest, err := svc.ApproveRequest(ctx, ResolveApprovalInput{
		TenantID:   "tenant-a",
		ActorID:    "manager-a",
		ApprovalID: approvalRequest.ID,
	})
	if err != nil {
		t.Fatalf("approve request: %v", err)
	}

	if approvedRequest.Status != approval.StatusApproved {
		t.Fatalf("expected approved request, got %s", approvedRequest.Status)
	}
	if approvedOrder.Status != procurement.PurchaseOrderStatusApproved {
		t.Fatalf("expected approved order, got %s", approvedOrder.Status)
	}
	if approvedOrder.ID != order.ID {
		t.Fatalf("expected same order id %s, got %s", order.ID, approvedOrder.ID)
	}
}
```

- [ ] **Step 2: Run the targeted test and watch it fail**

Run:

```bash
go test ./internal/application/admin/supplychain -run TestServiceApprovesSubmittedPurchaseOrder -v
```

Expected:

- FAIL because approval resolution behavior is not implemented yet

- [ ] **Step 3: Implement the application orchestration and memory repositories**

Create `service.go` with a tenant-aware orchestration service that:

- validates supplier / product / warehouse references before creating a purchase order
- invokes `shared.Pipeline.Execute` for mutating commands
- persists orders and approval requests through repository interfaces
- exposes read methods used by the future HTTP layer

Create `memory/supplychain.go` with concurrency-safe in-memory repositories:

```go
type SupplyChainStore struct {
	mu sync.Mutex
	suppliers map[string]masterdata.Supplier
	products map[string]masterdata.Product
	warehouses map[string]masterdata.Warehouse
	orders map[string]procurement.PurchaseOrder
	approvals map[string]approval.Request
}
```

Update `bootstrap.Container` to construct the store once and expose:

- `SupplyChainService *supplychain.Service`

- [ ] **Step 4: Re-run the service package**

Run:

```bash
go test ./internal/application/admin/supplychain -v
```

Expected:

- PASS

- [ ] **Step 5: Add list / get behavior needed by the admin API**

Extend the service with:

- `GetPurchaseOrder(ctx, tenantID, orderID string) (procurement.PurchaseOrder, approval.Request, error)`

Cover it with one small unit test asserting that a submitted order returns its linked approval request.

### Task 3: Expose the Workflow Through Admin HTTP Routes

**Files:**
- Modify: `internal/interfaces/http/router/admin.go`
- Modify: `internal/interfaces/http/presenter/response.go`
- Test: `test/integration/admin_supply_chain_test.go`

- [ ] **Step 1: Write the failing integration test for the admin flow**

Create `test/integration/admin_supply_chain_test.go` with a full HTTP flow:

```go
func TestAdminSupplyChainFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	supplierID := postJSON(t, h, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-001",
		"name": "Acme Supply",
	})

	productID := postJSON(t, h, "/api/admin/v1/master-data/products", map[string]any{
		"sku":  "SKU-001",
		"name": "Copper Wire",
		"unit": "roll",
	})

	warehouseID := postJSON(t, h, "/api/admin/v1/master-data/warehouses", map[string]any{
		"code": "WH-SH",
		"name": "Shanghai Warehouse",
	})

	orderResp := postJSONBody(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id": supplierID,
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity": 5,
		}},
	})

	orderID := orderResp["id"].(string)
	submitResp := postJSONBody(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/submit", map[string]any{})
	approvalID := submitResp["approval"]["id"].(string)

	approveResp := postJSONBody(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})
	if approveResp["status"] != "approved" {
		t.Fatalf("expected approved status, got %#v", approveResp["status"])
	}
}
```

- [ ] **Step 2: Run the integration test to verify the routes do not exist**

Run:

```bash
go test ./test/integration -run TestAdminSupplyChainFlow -v
```

Expected:

- FAIL with `404` or missing route behavior

- [ ] **Step 3: Implement minimal admin handlers**

Add route groups in `internal/interfaces/http/router/admin.go`:

- `POST /master-data/suppliers`
- `POST /master-data/products`
- `POST /master-data/warehouses`
- `POST /procurement/purchase-orders`
- `POST /procurement/purchase-orders/:id/submit`
- `GET /procurement/purchase-orders/:id`
- `POST /approvals/:id/approve`
- `POST /approvals/:id/reject`

Handler rules:

- read tenant and actor from middleware-backed context
- bind JSON payloads into small request structs
- call the supply-chain service
- return `400` for validation failures and `404` for missing objects
- use presenter helpers for success and error responses

- [ ] **Step 4: Re-run the targeted integration test**

Run:

```bash
go test ./test/integration -run TestAdminSupplyChainFlow -v
```

Expected:

- PASS

- [ ] **Step 5: Add one negative-path integration test**

Cover:

- creating a purchase order with an unknown supplier returns `404`

### Task 4: Prepare Forward-Looking Phase 2 Storage Schema

**Files:**
- Create: `migrations/000002_init_phase2_wave1_tables.up.sql`
- Create: `migrations/000002_init_phase2_wave1_tables.down.sql`
- Test: `test/integration/compose_contract_test.go`

- [ ] **Step 1: Write a failing migration contract assertion**

Add one assertion to an existing integration test or create a focused migration contract test that checks the new migration contains the expected tables:

- `supplier`
- `product`
- `warehouse`
- `purchase_order`
- `purchase_order_line`
- `approval_request`

- [ ] **Step 2: Run the contract test and confirm it fails**

Run:

```bash
go test ./test/integration -run TestPhase2Wave1MigrationContract -v
```

Expected:

- FAIL because the migration files do not exist yet

- [ ] **Step 3: Add the migration files**

Create the SQL migration with tenant-aware tables and explicit status columns:

```sql
create table if not exists supplier (
    id text primary key,
    tenant_id text not null,
    code text not null,
    name text not null,
    created_at timestamptz not null default now(),
    unique (tenant_id, code)
);
```

Repeat the pattern for product, warehouse, purchase_order, purchase_order_line, and approval_request. The down migration should drop the tables in dependency-safe reverse order.

- [ ] **Step 4: Re-run the migration contract test**

Run:

```bash
go test ./test/integration -run TestPhase2Wave1MigrationContract -v
```

Expected:

- PASS

### Task 5: Run the Full Verification Sweep

**Files:**
- Verify only

- [ ] **Step 1: Run the application tests**

```bash
go test ./internal/application/admin/supplychain -v
```

- [ ] **Step 2: Run the integration tests**

```bash
go test ./test/integration -v
```

- [ ] **Step 3: Run the full repository test suite**

```bash
go test ./... 
```

Expected:

- All tests PASS
- No new route panics
- No race-prone in-memory store behavior under normal tests

- [ ] **Step 4: Update docs if behavior changed materially**

If the implemented API or workflow differs from the plan, update:

- `README.md`
- `docs/phase-2-coverage-status.md`

- [ ] **Step 5: Prepare the next Phase 2 slice**

Record the follow-up backlog for the next implementation cycle:

- receipt / inbound inventory
- inventory ledger and balance queries
- payable bill creation on approved receipt
