# Phase 2 Wave 2 Inbound Inventory Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the next executable Phase 2 business slice: receive approved purchase orders into inventory, persist inbound receipt truth, and expose inventory balance queries through the admin API.

**Architecture:** Extend the existing Phase 2 Wave 1 supply-chain slice instead of branching into a parallel flow. Keep inbound receipt and inventory ledger rules inside a new `internal/domain/inventory` package, let the existing `internal/application/admin/supplychain` service orchestrate purchase-order receipt posting plus ledger writes, and continue using the shared in-memory adapter/runtime container until PostgreSQL repositories are introduced.

**Tech Stack:** Go 1.25, Gin, existing command pipeline/audit/policy seams, in-memory repositories for runtime wiring, SQL migrations for forward-looking PostgreSQL schema, standard-library tests

---

**Spec Reference:** `docs/superpowers/specs/2026-03-25-agentic-ai-native-erp-design.md`

**Coverage Reference:** `docs/phase-2-coverage-status.md`

**Scope Note:** This plan narrows Phase 2 Wave 2 to the inventory truth slice already called out in the coverage doc:

- inbound receipt posting against approved purchase orders
- purchase-order lifecycle progression from `approved` to `received`
- immutable inbound inventory ledger entries
- inventory balance query derived from ledger truth
- admin API routes for receipt posting and balance inspection
- forward-looking inventory schema migration

**Out of Scope For This Plan:**

- outbound issue / reservation / transfer
- payable bill generation
- inventory projection tables beyond direct balance aggregation
- PostgreSQL repository implementation
- workspace or integration APIs for inventory

## File Structure Map

The implementation produced by this plan should create or modify the following structure.

```text
erp-claw/
  docs/
    superpowers/
      plans/
        2026-03-26-phase-2-wave-2-inbound-inventory-implementation-plan.md
  internal/
    application/
      admin/
        supplychain/
          commands.go
          service.go
          service_test.go
    bootstrap/
      container.go
    domain/
      inventory/
        ledger.go
        receipt.go
        repository.go
      procurement/
        purchase_order.go
    infrastructure/
      persistence/
        memory/
          supplychain.go
          supplychain_test.go
    interfaces/
      http/
        router/
          admin.go
  migrations/
    000003_init_phase2_wave2_inventory_tables.up.sql
    000003_init_phase2_wave2_inventory_tables.down.sql
  test/
    integration/
      admin_inventory_test.go
      compose_contract_test.go
```

### Task 1: Model Inbound Receipt and Inventory Ledger Rules

**Files:**
- Create: `internal/domain/inventory/receipt.go`
- Create: `internal/domain/inventory/ledger.go`
- Create: `internal/domain/inventory/repository.go`
- Modify: `internal/domain/procurement/purchase_order.go`
- Test: `internal/application/admin/supplychain/service_test.go`

- [ ] **Step 1: Write the failing service workflow test**

Add this test to `internal/application/admin/supplychain/service_test.go`:

```go
func TestServiceReceivesApprovedPurchaseOrderIntoInventory(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	order, approvalRequest := createSubmittedOrder(t, ctx, svc)

	approvedOrder, _, err := svc.ApproveRequest(ctx, ResolveApprovalInput{
		TenantID:   "tenant-a",
		ActorID:    "manager-a",
		ApprovalID: approvalRequest.ID,
	})
	if err != nil {
		t.Fatalf("approve request: %v", err)
	}

	receipt, ledgerEntries, receivedOrder, err := svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "receiver-a",
		PurchaseOrderID: approvedOrder.ID,
		Lines: []ReceivePurchaseOrderLine{{
			ProductID: approvedOrder.Lines[0].ProductID,
			Quantity:  approvedOrder.Lines[0].Quantity,
		}},
	})
	if err != nil {
		t.Fatalf("receive purchase order: %v", err)
	}

	if receipt.Status != inventory.ReceiptStatusPosted {
		t.Fatalf("expected posted receipt, got %s", receipt.Status)
	}
	if receivedOrder.Status != procurement.PurchaseOrderStatusReceived {
		t.Fatalf("expected received order, got %s", receivedOrder.Status)
	}
	if len(ledgerEntries) != 1 {
		t.Fatalf("expected 1 ledger entry, got %d", len(ledgerEntries))
	}
	if ledgerEntries[0].Quantity != approvedOrder.Lines[0].Quantity {
		t.Fatalf("expected quantity %d, got %d", approvedOrder.Lines[0].Quantity, ledgerEntries[0].Quantity)
	}
}
```

- [ ] **Step 2: Run the targeted test and verify it fails**

Run:

```bash
go test ./internal/application/admin/supplychain -run TestServiceReceivesApprovedPurchaseOrderIntoInventory -v
```

Expected:

- FAIL because `ReceivePurchaseOrderInput`, `ReceivePurchaseOrder`, and the new inventory domain types do not exist yet

- [ ] **Step 3: Implement minimal domain rules**

Create the inventory domain package with these minimum shapes:

```go
type ReceiptStatus string

const (
	ReceiptStatusPosted ReceiptStatus = "posted"
)

type Receipt struct {
	ID              string
	TenantID        string
	PurchaseOrderID string
	WarehouseID     string
	Status          ReceiptStatus
	Lines           []ReceiptLine
}

type LedgerEntry struct {
	ID              string
	TenantID        string
	ProductID       string
	WarehouseID     string
	MovementType    MovementType
	QuantityDelta   int
	ReferenceType   string
	ReferenceID     string
}
```

Update `internal/domain/procurement/purchase_order.go` so a purchase order can progress from:

- `draft` -> `pending_approval`
- `pending_approval` -> `approved` / `rejected`
- `approved` -> `received`

Preserve the invariant that receiving is allowed only from `approved`.

- [ ] **Step 4: Add focused rule tests**

Add adjacent tests covering:

- receive fails when order is still `draft`
- receive fails when receipt lines are empty
- receive fails when receipt quantity is non-positive
- receive fails when the order is already `received`

- [ ] **Step 5: Re-run the service package**

Run:

```bash
go test ./internal/application/admin/supplychain -v
```

Expected:

- PASS

### Task 2: Extend the Supply-Chain Service and Memory Persistence

**Files:**
- Modify: `internal/application/admin/supplychain/commands.go`
- Modify: `internal/application/admin/supplychain/service.go`
- Modify: `internal/application/admin/supplychain/service_test.go`
- Modify: `internal/bootstrap/container.go`
- Modify: `internal/infrastructure/persistence/memory/supplychain.go`
- Modify: `internal/infrastructure/persistence/memory/supplychain_test.go`

- [ ] **Step 1: Write the next failing query test**

Add this test to `internal/application/admin/supplychain/service_test.go`:

```go
func TestServiceReturnsInventoryBalanceFromPostedReceipts(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	order, approvalRequest := createSubmittedOrder(t, ctx, svc)

	approvedOrder, _, err := svc.ApproveRequest(ctx, ResolveApprovalInput{
		TenantID:   "tenant-a",
		ActorID:    "manager-a",
		ApprovalID: approvalRequest.ID,
	})
	if err != nil {
		t.Fatalf("approve request: %v", err)
	}

	_, _, _, err = svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "receiver-a",
		PurchaseOrderID: approvedOrder.ID,
		Lines: []ReceivePurchaseOrderLine{{
			ProductID: approvedOrder.Lines[0].ProductID,
			Quantity:  approvedOrder.Lines[0].Quantity,
		}},
	})
	if err != nil {
		t.Fatalf("receive order: %v", err)
	}

	balance, err := svc.GetInventoryBalance(ctx, GetInventoryBalanceInput{
		TenantID:    "tenant-a",
		ProductID:   approvedOrder.Lines[0].ProductID,
		WarehouseID: approvedOrder.WarehouseID,
	})
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}

	if balance.OnHand != approvedOrder.Lines[0].Quantity {
		t.Fatalf("expected on hand %d, got %d", approvedOrder.Lines[0].Quantity, balance.OnHand)
	}
}
```

- [ ] **Step 2: Run the targeted test and verify it fails**

Run:

```bash
go test ./internal/application/admin/supplychain -run TestServiceReturnsInventoryBalanceFromPostedReceipts -v
```

Expected:

- FAIL because inventory repository behavior and query methods do not exist yet

- [ ] **Step 3: Implement the application orchestration and memory repositories**

Extend `commands.go` and `service.go` with:

```go
type ReceivePurchaseOrderInput struct {
	TenantID        string
	ActorID         string
	PurchaseOrderID string
	Lines           []ReceivePurchaseOrderLine
}

type GetInventoryBalanceInput struct {
	TenantID    string
	ProductID   string
	WarehouseID string
}
```

Add service methods:

- `ReceivePurchaseOrder(ctx, input) (inventory.Receipt, []inventory.LedgerEntry, procurement.PurchaseOrder, error)`
- `GetInventoryBalance(ctx, input) (inventory.Balance, error)`

The service should:

- fetch the approved purchase order
- validate receipt lines against the order lines
- create a posted receipt
- append immutable inbound ledger entries
- mark the order `received`
- derive balance by aggregating the ledger for `(tenant, warehouse, product)`

Extend the shared in-memory supply-chain store with receipt and ledger collections. Keep repository returns detached from internal slices/maps so reads cannot mutate stored state.

- [ ] **Step 4: Add repository-level regression tests**

Extend `internal/infrastructure/persistence/memory/supplychain_test.go` with one test that proves:

- appending to returned ledger entry slices does not mutate stored state

- [ ] **Step 5: Re-run the relevant packages**

Run:

```bash
go test ./internal/application/admin/supplychain ./internal/infrastructure/persistence/memory -v
```

Expected:

- PASS

### Task 3: Expose Receipt Posting and Inventory Balance Through Admin HTTP Routes

**Files:**
- Modify: `internal/interfaces/http/router/admin.go`
- Test: `test/integration/admin_inventory_test.go`

- [ ] **Step 1: Write the failing integration test**

Create `test/integration/admin_inventory_test.go` with:

```go
func TestAdminInventoryReceiptFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	supplierID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-001",
		"name": "Acme Supply",
	}), "id")
	productID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/products", map[string]any{
		"sku": "SKU-001",
		"name": "Copper Wire",
		"unit": "roll",
	}), "id")
	warehouseID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/warehouses", map[string]any{
		"code": "WH-SH",
		"name": "Shanghai Warehouse",
	}), "id")

	orderID := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id": supplierID,
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{"product_id": productID, "quantity": 5}},
	}), "id")

	submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/submit", map[string]any{})
	approvalID := stringField(t, nestedMap(t, submitResp, "approval"), "id")
	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})

	receiptResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/receive", map[string]any{
		"lines": []map[string]any{{"product_id": productID, "quantity": 5}},
	})
	if stringField(t, nestedMap(t, receiptResp, "order"), "status") != "received" {
		t.Fatalf("expected received order status")
	}

	balanceResp := getJSONData(t, h, "/api/admin/v1/inventory/balances?product_id="+productID+"&warehouse_id="+warehouseID)
	if intField(t, balanceResp, "on_hand") != 5 {
		t.Fatalf("expected on_hand 5")
	}
}
```

- [ ] **Step 2: Run the targeted integration test and verify it fails**

Run:

```bash
go test ./test/integration -run TestAdminInventoryReceiptFlow -v
```

Expected:

- FAIL with missing route behavior (`404`) or undefined helpers

- [ ] **Step 3: Implement minimal admin routes**

Add these routes to `internal/interfaces/http/router/admin.go`:

- `POST /procurement/purchase-orders/:id/receive`
- `GET /inventory/balances`

Handler rules:

- bind receipt lines from JSON
- reuse middleware-backed tenant/actor context
- return `400` for invalid transitions or payload validation
- return `404` for missing purchase orders or master-data references
- return `200` with balance fields `tenant_id`, `product_id`, `warehouse_id`, `on_hand`

- [ ] **Step 4: Add one negative-path integration test**

Cover:

- trying to receive a purchase order before approval returns `400`

- [ ] **Step 5: Re-run the targeted integration suite**

Run:

```bash
go test ./test/integration -run 'TestAdminInventoryReceiptFlow|TestAdminInventoryReceiptRequiresApprovedOrder' -v
```

Expected:

- PASS

### Task 4: Prepare Forward-Looking Phase 2 Wave 2 Storage Schema

**Files:**
- Create: `migrations/000003_init_phase2_wave2_inventory_tables.up.sql`
- Create: `migrations/000003_init_phase2_wave2_inventory_tables.down.sql`
- Modify: `test/integration/compose_contract_test.go`

- [ ] **Step 1: Write the failing migration contract assertion**

Extend `test/integration/compose_contract_test.go` with a new contract test that checks the migration contains these tables:

- `receipt`
- `receipt_line`
- `inventory_ledger`

Also assert the migration keeps tenant-aware constraints similar to Wave 1.

- [ ] **Step 2: Run the migration contract test and confirm it fails**

Run:

```bash
go test ./test/integration -run TestPhase2Wave2MigrationContract -v
```

Expected:

- FAIL because the `000003` migration files do not exist yet

- [ ] **Step 3: Add the migration files**

Create the SQL migration with minimum shapes like:

```sql
create table if not exists receipt (
    id text primary key,
    tenant_id text not null,
    purchase_order_id text not null,
    warehouse_id text not null,
    status text not null default 'posted',
    created_by text not null,
    created_at timestamptz not null default now(),
    unique (tenant_id, id),
    foreign key (tenant_id, purchase_order_id) references purchase_order(tenant_id, id),
    foreign key (tenant_id, warehouse_id) references warehouse(tenant_id, id)
);
```

Repeat the tenant-aware pattern for `receipt_line` and `inventory_ledger`. The down migration should drop the tables in dependency-safe reverse order.

- [ ] **Step 4: Re-run the migration contract test**

Run:

```bash
go test ./test/integration -run TestPhase2Wave2MigrationContract -v
```

Expected:

- PASS

### Task 5: Run the Full Verification Sweep and Update Phase 2 Status

**Files:**
- Modify: `README.md`
- Modify: `docs/phase-2-coverage-status.md`
- Verify only

- [ ] **Step 1: Run the application and integration suites**

```bash
go test ./internal/application/admin/supplychain ./internal/infrastructure/persistence/memory -v
go test ./test/integration -v
```

- [ ] **Step 2: Run the full repository test suite**

```bash
go test ./...
```

- [ ] **Step 3: Run the focused race check for new runtime slices**

```bash
go test -race ./internal/application/admin/supplychain ./internal/infrastructure/persistence/memory ./test/integration
```

- [ ] **Step 4: Update docs to reflect Wave 2 progress**

Update:

- `README.md` with the new admin receipt and inventory balance endpoints
- `docs/phase-2-coverage-status.md` so Wave 2 no longer reads as untouched

- [ ] **Step 5: Prepare the next Phase 2 slice**

Record the follow-up backlog for the next implementation cycle:

- payable bill creation on approved and received procurement objects
- inventory projection / list views if direct ledger aggregation becomes too expensive
- PostgreSQL repositories for Phase 2 objects
