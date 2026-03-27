package supplychain

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/domain/approval"
	"github.com/nikkofu/erp-claw/internal/domain/inventory"
	"github.com/nikkofu/erp-claw/internal/domain/masterdata"
	"github.com/nikkofu/erp-claw/internal/domain/payable"
	"github.com/nikkofu/erp-claw/internal/domain/procurement"
	"github.com/nikkofu/erp-claw/internal/domain/receivable"
	"github.com/nikkofu/erp-claw/internal/domain/sales"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/memory"
)

func TestServiceCreatesAndSubmitsPurchaseOrderForApproval(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

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
		t.Fatalf("expected draft status, got %s", order.Status)
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
		t.Fatalf("expected pending approval status, got %s", submitted.Status)
	}

	if approvalRequest.Status != approval.StatusPending {
		t.Fatalf("expected pending approval request, got %s", approvalRequest.Status)
	}

	if approvalRequest.ResourceID != submitted.ID {
		t.Fatalf("expected approval resource %s, got %s", submitted.ID, approvalRequest.ResourceID)
	}
}

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
		t.Fatalf("expected approved order %s, got %s", order.ID, approvedOrder.ID)
	}
}

func TestServiceRejectsSubmittedPurchaseOrder(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	order, approvalRequest := createSubmittedOrder(t, ctx, svc)

	rejectedOrder, rejectedRequest, err := svc.RejectRequest(ctx, ResolveApprovalInput{
		TenantID:   "tenant-a",
		ActorID:    "manager-a",
		ApprovalID: approvalRequest.ID,
	})
	if err != nil {
		t.Fatalf("reject request: %v", err)
	}

	if rejectedRequest.Status != approval.StatusRejected {
		t.Fatalf("expected rejected request, got %s", rejectedRequest.Status)
	}

	if rejectedOrder.Status != procurement.PurchaseOrderStatusRejected {
		t.Fatalf("expected rejected order, got %s", rejectedOrder.Status)
	}

	if rejectedOrder.ID != order.ID {
		t.Fatalf("expected rejected order %s, got %s", order.ID, rejectedOrder.ID)
	}
}

func TestServiceGetPurchaseOrderReturnsLinkedApproval(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	order, approvalRequest := createSubmittedOrder(t, ctx, svc)

	gotOrder, gotApproval, err := svc.GetPurchaseOrder(ctx, "tenant-a", order.ID)
	if err != nil {
		t.Fatalf("get purchase order: %v", err)
	}

	if gotOrder.ID != order.ID {
		t.Fatalf("expected order %s, got %s", order.ID, gotOrder.ID)
	}

	if gotApproval.ID != approvalRequest.ID {
		t.Fatalf("expected approval %s, got %s", approvalRequest.ID, gotApproval.ID)
	}
}

func TestServiceListApprovalRequestsSupportsStatusFilter(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	_, pendingRequest := createSubmittedOrder(t, ctx, svc)
	_, approvedRequest := createSubmittedOrder(t, ctx, svc)

	if _, _, err := svc.ApproveRequest(ctx, ResolveApprovalInput{
		TenantID:   approvedRequest.TenantID,
		ActorID:    "manager-a",
		ApprovalID: approvedRequest.ID,
	}); err != nil {
		t.Fatalf("approve request: %v", err)
	}

	pending, err := svc.ListApprovalRequests(ctx, ListApprovalRequestsInput{
		TenantID: "tenant-a",
		Status:   "pending",
	})
	if err != nil {
		t.Fatalf("list pending approvals: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending approval, got %d", len(pending))
	}
	if pending[0].ID != pendingRequest.ID {
		t.Fatalf("expected pending approval id %s, got %s", pendingRequest.ID, pending[0].ID)
	}

	approved, err := svc.ListApprovalRequests(ctx, ListApprovalRequestsInput{
		TenantID: "tenant-a",
		Status:   "approved",
	})
	if err != nil {
		t.Fatalf("list approved approvals: %v", err)
	}
	if len(approved) != 1 {
		t.Fatalf("expected 1 approved approval, got %d", len(approved))
	}
	if approved[0].ID != approvedRequest.ID {
		t.Fatalf("expected approved approval id %s, got %s", approvedRequest.ID, approved[0].ID)
	}
}

func TestServiceListApprovalRequestsFailsForInvalidStatus(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	_, err := svc.ListApprovalRequests(ctx, ListApprovalRequestsInput{
		TenantID: "tenant-a",
		Status:   "invalid",
	})
	if !errors.Is(err, approval.ErrInvalidRequestQuery) {
		t.Fatalf("expected invalid approval request query, got %v", err)
	}
}

func TestServiceListApprovalRequestsSupportsSortAndPagination(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	_, reqA := createSubmittedOrder(t, ctx, svc)
	_, reqB := createSubmittedOrder(t, ctx, svc)
	_, reqC := createSubmittedOrder(t, ctx, svc)

	ids := []string{reqA.ID, reqB.ID, reqC.ID}
	sort.Strings(ids)

	page1, err := svc.ListApprovalRequests(ctx, ListApprovalRequestsInput{
		TenantID: "tenant-a",
		Sort:     "id_asc",
		Page:     1,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("list approvals page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 approvals in page1, got %d", len(page1))
	}
	if page1[0].ID != ids[0] || page1[1].ID != ids[1] {
		t.Fatalf("expected page1 ids [%s,%s], got [%s,%s]", ids[0], ids[1], page1[0].ID, page1[1].ID)
	}

	page2, err := svc.ListApprovalRequests(ctx, ListApprovalRequestsInput{
		TenantID: "tenant-a",
		Sort:     "id_asc",
		Page:     2,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("list approvals page2: %v", err)
	}
	if len(page2) != 1 {
		t.Fatalf("expected 1 approval in page2, got %d", len(page2))
	}
	if page2[0].ID != ids[2] {
		t.Fatalf("expected page2 id %s, got %s", ids[2], page2[0].ID)
	}
}

func TestServiceListApprovalRequestsFailsForInvalidSortAndPagination(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	_, err := svc.ListApprovalRequests(ctx, ListApprovalRequestsInput{
		TenantID: "tenant-a",
		Sort:     "unknown",
	})
	if !errors.Is(err, approval.ErrInvalidRequestQuery) {
		t.Fatalf("expected invalid approval request query for invalid sort, got %v", err)
	}

	_, err = svc.ListApprovalRequests(ctx, ListApprovalRequestsInput{
		TenantID: "tenant-a",
		Page:     -1,
		PageSize: 10,
	})
	if !errors.Is(err, approval.ErrInvalidRequestQuery) {
		t.Fatalf("expected invalid approval request query for invalid page, got %v", err)
	}
}

func TestServiceSubmitPurchaseOrderFailsForEmptyLines(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	supplier, product, warehouse := createMasterData(t, ctx, svc)
	_ = product

	order := procurement.PurchaseOrder{
		ID:          "po-empty-lines",
		TenantID:    "tenant-a",
		SupplierID:  supplier.ID,
		WarehouseID: warehouse.ID,
		Status:      procurement.PurchaseOrderStatusDraft,
	}
	if err := svc.purchaseOrders.Save(ctx, order); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	_, _, err := svc.SubmitPurchaseOrder(ctx, SubmitPurchaseOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "actor-a",
		PurchaseOrderID: order.ID,
	})
	if !errors.Is(err, procurement.ErrInvalidPurchaseOrder) {
		t.Fatalf("expected invalid purchase order, got %v", err)
	}
}

func TestServiceSubmitPurchaseOrderFailsWhenOrderIsNotDraft(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	order, _ := createSubmittedOrder(t, ctx, svc)

	_, _, err := svc.SubmitPurchaseOrder(ctx, SubmitPurchaseOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "actor-a",
		PurchaseOrderID: order.ID,
	})
	if !errors.Is(err, procurement.ErrPurchaseOrderAlreadySubmitted) {
		t.Fatalf("expected non-draft error, got %v", err)
	}
}

func TestServiceApproveRequestFailsWhenAlreadyTerminal(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	_, approvalRequest := createSubmittedOrder(t, ctx, svc)

	if _, _, err := svc.ApproveRequest(ctx, ResolveApprovalInput{
		TenantID:   "tenant-a",
		ActorID:    "manager-a",
		ApprovalID: approvalRequest.ID,
	}); err != nil {
		t.Fatalf("first approve request: %v", err)
	}

	_, _, err := svc.ApproveRequest(ctx, ResolveApprovalInput{
		TenantID:   "tenant-a",
		ActorID:    "manager-b",
		ApprovalID: approvalRequest.ID,
	})
	if !errors.Is(err, approval.ErrApprovalNotPending) {
		t.Fatalf("expected approval not pending, got %v", err)
	}
}

func TestServiceRejectRequestFailsWhenAlreadyTerminal(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	_, approvalRequest := createSubmittedOrder(t, ctx, svc)

	if _, _, err := svc.RejectRequest(ctx, ResolveApprovalInput{
		TenantID:   "tenant-a",
		ActorID:    "manager-a",
		ApprovalID: approvalRequest.ID,
	}); err != nil {
		t.Fatalf("first reject request: %v", err)
	}

	_, _, err := svc.RejectRequest(ctx, ResolveApprovalInput{
		TenantID:   "tenant-a",
		ActorID:    "manager-b",
		ApprovalID: approvalRequest.ID,
	})
	if !errors.Is(err, approval.ErrApprovalNotPending) {
		t.Fatalf("expected approval not pending, got %v", err)
	}
}

func TestServiceCreatePurchaseOrderFailsForUnknownSupplier(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	_, product, warehouse := createMasterData(t, ctx, svc)

	_, err := svc.CreatePurchaseOrder(ctx, CreatePurchaseOrderInput{
		TenantID:    "tenant-a",
		ActorID:     "actor-a",
		SupplierID:  "sup-missing",
		WarehouseID: warehouse.ID,
		Lines: []CreatePurchaseOrderLine{{
			ProductID: product.ID,
			Quantity:  5,
		}},
	})
	if !errors.Is(err, masterdata.ErrSupplierNotFound) {
		t.Fatalf("expected supplier not found, got %v", err)
	}
}

func TestServiceReceivesApprovedPurchaseOrderIntoInventory(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	approvedOrder := createApprovedOrder(t, ctx, svc)

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

	if ledgerEntries[0].QuantityDelta != approvedOrder.Lines[0].Quantity {
		t.Fatalf("expected quantity delta %d, got %d", approvedOrder.Lines[0].Quantity, ledgerEntries[0].QuantityDelta)
	}
}

func TestServiceReceivesApprovedPurchaseOrderWithDuplicateProductLines(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	supplier, product, warehouse := createMasterData(t, ctx, svc)

	order, err := svc.CreatePurchaseOrder(ctx, CreatePurchaseOrderInput{
		TenantID:    "tenant-a",
		ActorID:     "actor-a",
		SupplierID:  supplier.ID,
		WarehouseID: warehouse.ID,
		Lines: []CreatePurchaseOrderLine{
			{
				ProductID: product.ID,
				Quantity:  2,
			},
			{
				ProductID: product.ID,
				Quantity:  3,
			},
		},
	})
	if err != nil {
		t.Fatalf("create purchase order: %v", err)
	}

	submittedOrder, approvalRequest, err := svc.SubmitPurchaseOrder(ctx, SubmitPurchaseOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "actor-a",
		PurchaseOrderID: order.ID,
	})
	if err != nil {
		t.Fatalf("submit purchase order: %v", err)
	}
	if submittedOrder.Status != procurement.PurchaseOrderStatusPendingApproval {
		t.Fatalf("expected pending_approval order, got %s", submittedOrder.Status)
	}

	approvedOrder, _, err := svc.ApproveRequest(ctx, ResolveApprovalInput{
		TenantID:   "tenant-a",
		ActorID:    "manager-a",
		ApprovalID: approvalRequest.ID,
	})
	if err != nil {
		t.Fatalf("approve purchase order: %v", err)
	}
	if approvedOrder.Status != procurement.PurchaseOrderStatusApproved {
		t.Fatalf("expected approved order, got %s", approvedOrder.Status)
	}

	receipt, ledgerEntries, receivedOrder, err := svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "receiver-a",
		PurchaseOrderID: order.ID,
		Lines: []ReceivePurchaseOrderLine{{
			ProductID: product.ID,
			Quantity:  5,
		}},
	})
	if err != nil {
		t.Fatalf("receive purchase order: %v", err)
	}
	if receivedOrder.Status != procurement.PurchaseOrderStatusReceived {
		t.Fatalf("expected received order, got %s", receivedOrder.Status)
	}
	if len(receipt.Lines) != 1 {
		t.Fatalf("expected 1 receipt line, got %d", len(receipt.Lines))
	}
	if len(ledgerEntries) != 1 {
		t.Fatalf("expected 1 ledger entry, got %d", len(ledgerEntries))
	}
	if ledgerEntries[0].QuantityDelta != 5 {
		t.Fatalf("expected quantity delta 5, got %d", ledgerEntries[0].QuantityDelta)
	}
}

func TestServiceReceivePurchaseOrderKeepsOrderApprovedWhenSaveReceiptFails(t *testing.T) {
	ctx := context.Background()
	saveReceiptErr := errors.New("save receipt failed")
	svc := newTestServiceWithInventoryRepository(failingInventoryRepository{
		delegate:       memory.NewInventoryRepository(),
		saveReceiptErr: saveReceiptErr,
	})
	approvedOrder := createApprovedOrder(t, ctx, svc)

	_, _, _, err := svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "receiver-a",
		PurchaseOrderID: approvedOrder.ID,
		Lines: []ReceivePurchaseOrderLine{{
			ProductID: approvedOrder.Lines[0].ProductID,
			Quantity:  approvedOrder.Lines[0].Quantity,
		}},
	})
	if !errors.Is(err, saveReceiptErr) {
		t.Fatalf("expected save receipt error, got %v", err)
	}

	storedOrder, err := svc.purchaseOrders.Get(ctx, "tenant-a", approvedOrder.ID)
	if err != nil {
		t.Fatalf("get purchase order: %v", err)
	}
	if storedOrder.Status != procurement.PurchaseOrderStatusApproved {
		t.Fatalf("expected purchase order to remain approved, got %s", storedOrder.Status)
	}
}

func TestServiceReceivePurchaseOrderKeepsOrderApprovedWhenAppendLedgerFails(t *testing.T) {
	ctx := context.Background()
	appendLedgerErr := errors.New("append ledger failed")
	svc := newTestServiceWithInventoryRepository(failingInventoryRepository{
		delegate:        memory.NewInventoryRepository(),
		appendLedgerErr: appendLedgerErr,
	})
	approvedOrder := createApprovedOrder(t, ctx, svc)

	_, _, _, err := svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "receiver-a",
		PurchaseOrderID: approvedOrder.ID,
		Lines: []ReceivePurchaseOrderLine{{
			ProductID: approvedOrder.Lines[0].ProductID,
			Quantity:  approvedOrder.Lines[0].Quantity,
		}},
	})
	if !errors.Is(err, appendLedgerErr) {
		t.Fatalf("expected append ledger error, got %v", err)
	}

	storedOrder, err := svc.purchaseOrders.Get(ctx, "tenant-a", approvedOrder.ID)
	if err != nil {
		t.Fatalf("get purchase order: %v", err)
	}
	if storedOrder.Status != procurement.PurchaseOrderStatusApproved {
		t.Fatalf("expected purchase order to remain approved, got %s", storedOrder.Status)
	}
}

func TestServiceReturnsInventoryBalanceFromPostedReceipts(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	approvedOrder := createApprovedOrder(t, ctx, svc)

	_, _, _, err := svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
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

	balance, err := svc.GetInventoryBalance(ctx, GetInventoryBalanceInput{
		TenantID:    "tenant-a",
		ProductID:   approvedOrder.Lines[0].ProductID,
		WarehouseID: approvedOrder.WarehouseID,
	})
	if err != nil {
		t.Fatalf("get inventory balance: %v", err)
	}

	if balance.OnHand != approvedOrder.Lines[0].Quantity {
		t.Fatalf("expected on hand %d, got %d", approvedOrder.Lines[0].Quantity, balance.OnHand)
	}
	if balance.Reserved != 0 {
		t.Fatalf("expected reserved 0, got %d", balance.Reserved)
	}
	if balance.Available != approvedOrder.Lines[0].Quantity {
		t.Fatalf("expected available %d, got %d", approvedOrder.Lines[0].Quantity, balance.Available)
	}
}

func TestServiceReservesInventoryAndUpdatesAvailableBalance(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	approvedOrder := createApprovedOrder(t, ctx, svc)

	_, _, _, err := svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
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

	reservation, err := svc.ReserveInventory(ctx, ReserveInventoryInput{
		TenantID:      "tenant-a",
		ActorID:       "planner-a",
		ProductID:     approvedOrder.Lines[0].ProductID,
		WarehouseID:   approvedOrder.WarehouseID,
		Quantity:      2,
		ReferenceType: "sales_order",
		ReferenceID:   "so-001",
	})
	if err != nil {
		t.Fatalf("reserve inventory: %v", err)
	}
	if reservation.Status != inventory.ReservationStatusActive {
		t.Fatalf("expected active reservation, got %s", reservation.Status)
	}

	balance, err := svc.GetInventoryBalance(ctx, GetInventoryBalanceInput{
		TenantID:    "tenant-a",
		ProductID:   approvedOrder.Lines[0].ProductID,
		WarehouseID: approvedOrder.WarehouseID,
	})
	if err != nil {
		t.Fatalf("get inventory balance: %v", err)
	}
	if balance.OnHand != approvedOrder.Lines[0].Quantity {
		t.Fatalf("expected on hand %d, got %d", approvedOrder.Lines[0].Quantity, balance.OnHand)
	}
	if balance.Reserved != 2 {
		t.Fatalf("expected reserved 2, got %d", balance.Reserved)
	}
	if balance.Available != approvedOrder.Lines[0].Quantity-2 {
		t.Fatalf("expected available %d, got %d", approvedOrder.Lines[0].Quantity-2, balance.Available)
	}
}

func TestServiceReserveInventoryFailsWhenQuantityExceedsAvailable(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	approvedOrder := createApprovedOrder(t, ctx, svc)

	_, _, _, err := svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
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

	_, err = svc.ReserveInventory(ctx, ReserveInventoryInput{
		TenantID:      "tenant-a",
		ActorID:       "planner-a",
		ProductID:     approvedOrder.Lines[0].ProductID,
		WarehouseID:   approvedOrder.WarehouseID,
		Quantity:      approvedOrder.Lines[0].Quantity + 1,
		ReferenceType: "sales_order",
		ReferenceID:   "so-002",
	})
	if !errors.Is(err, inventory.ErrInsufficientAvailableInventory) {
		t.Fatalf("expected insufficient available inventory, got %v", err)
	}
}

func TestServiceIssuesInventoryAndReducesAvailableBalance(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	approvedOrder := createApprovedOrder(t, ctx, svc)

	_, _, _, err := svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
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

	outbound, err := svc.IssueInventory(ctx, IssueInventoryInput{
		TenantID:      "tenant-a",
		ActorID:       "warehouse-a",
		ProductID:     approvedOrder.Lines[0].ProductID,
		WarehouseID:   approvedOrder.WarehouseID,
		Quantity:      2,
		ReferenceType: "shipment",
		ReferenceID:   "shp-001",
	})
	if err != nil {
		t.Fatalf("issue inventory: %v", err)
	}
	if outbound.MovementType != inventory.MovementTypeOutbound {
		t.Fatalf("expected outbound movement, got %s", outbound.MovementType)
	}
	if outbound.QuantityDelta != -2 {
		t.Fatalf("expected quantity delta -2, got %d", outbound.QuantityDelta)
	}

	balance, err := svc.GetInventoryBalance(ctx, GetInventoryBalanceInput{
		TenantID:    "tenant-a",
		ProductID:   approvedOrder.Lines[0].ProductID,
		WarehouseID: approvedOrder.WarehouseID,
	})
	if err != nil {
		t.Fatalf("get inventory balance: %v", err)
	}
	if balance.OnHand != approvedOrder.Lines[0].Quantity-2 {
		t.Fatalf("expected on hand %d, got %d", approvedOrder.Lines[0].Quantity-2, balance.OnHand)
	}
	if balance.Reserved != 0 {
		t.Fatalf("expected reserved 0, got %d", balance.Reserved)
	}
	if balance.Available != approvedOrder.Lines[0].Quantity-2 {
		t.Fatalf("expected available %d, got %d", approvedOrder.Lines[0].Quantity-2, balance.Available)
	}
}

func TestServiceIssueInventoryFailsWhenQuantityExceedsAvailable(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	approvedOrder := createApprovedOrder(t, ctx, svc)

	_, _, _, err := svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
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

	_, err = svc.ReserveInventory(ctx, ReserveInventoryInput{
		TenantID:      "tenant-a",
		ActorID:       "planner-a",
		ProductID:     approvedOrder.Lines[0].ProductID,
		WarehouseID:   approvedOrder.WarehouseID,
		Quantity:      approvedOrder.Lines[0].Quantity - 1,
		ReferenceType: "sales_order",
		ReferenceID:   "so-003",
	})
	if err != nil {
		t.Fatalf("reserve inventory: %v", err)
	}

	_, err = svc.IssueInventory(ctx, IssueInventoryInput{
		TenantID:      "tenant-a",
		ActorID:       "warehouse-a",
		ProductID:     approvedOrder.Lines[0].ProductID,
		WarehouseID:   approvedOrder.WarehouseID,
		Quantity:      2,
		ReferenceType: "shipment",
		ReferenceID:   "shp-002",
	})
	if !errors.Is(err, inventory.ErrInsufficientAvailableInventory) {
		t.Fatalf("expected insufficient available inventory, got %v", err)
	}
}

func TestServiceTransfersInventoryBetweenWarehousesAndUpdatesBalances(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	approvedOrder := createApprovedOrder(t, ctx, svc)

	targetWarehouse, err := svc.CreateWarehouse(ctx, CreateWarehouseInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Code:     "WH-BJ",
		Name:     "Beijing Warehouse",
	})
	if err != nil {
		t.Fatalf("create target warehouse: %v", err)
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
		t.Fatalf("receive purchase order: %v", err)
	}

	entries, err := svc.TransferInventory(ctx, TransferInventoryInput{
		TenantID:        "tenant-a",
		ActorID:         "warehouse-a",
		ProductID:       approvedOrder.Lines[0].ProductID,
		FromWarehouseID: approvedOrder.WarehouseID,
		ToWarehouseID:   targetWarehouse.ID,
		Quantity:        2,
		ReferenceType:   "transfer_order",
		ReferenceID:     "trf-001",
	})
	if err != nil {
		t.Fatalf("transfer inventory: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 transfer ledger entries, got %d", len(entries))
	}
	if entries[0].MovementType != inventory.MovementTypeOutbound || entries[0].QuantityDelta != -2 {
		t.Fatalf("expected first entry outbound -2, got %s %d", entries[0].MovementType, entries[0].QuantityDelta)
	}
	if entries[1].MovementType != inventory.MovementTypeInbound || entries[1].QuantityDelta != 2 {
		t.Fatalf("expected second entry inbound 2, got %s %d", entries[1].MovementType, entries[1].QuantityDelta)
	}

	sourceBalance, err := svc.GetInventoryBalance(ctx, GetInventoryBalanceInput{
		TenantID:    "tenant-a",
		ProductID:   approvedOrder.Lines[0].ProductID,
		WarehouseID: approvedOrder.WarehouseID,
	})
	if err != nil {
		t.Fatalf("get source inventory balance: %v", err)
	}
	if sourceBalance.OnHand != approvedOrder.Lines[0].Quantity-2 {
		t.Fatalf("expected source on hand %d, got %d", approvedOrder.Lines[0].Quantity-2, sourceBalance.OnHand)
	}
	if sourceBalance.Available != approvedOrder.Lines[0].Quantity-2 {
		t.Fatalf("expected source available %d, got %d", approvedOrder.Lines[0].Quantity-2, sourceBalance.Available)
	}

	targetBalance, err := svc.GetInventoryBalance(ctx, GetInventoryBalanceInput{
		TenantID:    "tenant-a",
		ProductID:   approvedOrder.Lines[0].ProductID,
		WarehouseID: targetWarehouse.ID,
	})
	if err != nil {
		t.Fatalf("get target inventory balance: %v", err)
	}
	if targetBalance.OnHand != 2 {
		t.Fatalf("expected target on hand 2, got %d", targetBalance.OnHand)
	}
	if targetBalance.Available != 2 {
		t.Fatalf("expected target available 2, got %d", targetBalance.Available)
	}
}

func TestServiceTransferInventoryFailsWhenQuantityExceedsAvailable(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	approvedOrder := createApprovedOrder(t, ctx, svc)

	targetWarehouse, err := svc.CreateWarehouse(ctx, CreateWarehouseInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Code:     "WH-BJ",
		Name:     "Beijing Warehouse",
	})
	if err != nil {
		t.Fatalf("create target warehouse: %v", err)
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
		t.Fatalf("receive purchase order: %v", err)
	}

	_, err = svc.TransferInventory(ctx, TransferInventoryInput{
		TenantID:        "tenant-a",
		ActorID:         "warehouse-a",
		ProductID:       approvedOrder.Lines[0].ProductID,
		FromWarehouseID: approvedOrder.WarehouseID,
		ToWarehouseID:   targetWarehouse.ID,
		Quantity:        approvedOrder.Lines[0].Quantity + 1,
		ReferenceType:   "transfer_order",
		ReferenceID:     "trf-002",
	})
	if !errors.Is(err, inventory.ErrInsufficientAvailableInventory) {
		t.Fatalf("expected insufficient available inventory, got %v", err)
	}
}

func TestServiceCreatesAndExecutesTransferOrder(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	receivedOrder := createReceivedOrder(t, ctx, svc)

	targetWarehouse, err := svc.CreateWarehouse(ctx, CreateWarehouseInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Code:     "WH-BJ-ORDER",
		Name:     "Beijing Warehouse Transfer",
	})
	if err != nil {
		t.Fatalf("create target warehouse: %v", err)
	}

	order, err := svc.CreateTransferOrder(ctx, CreateTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "planner-a",
		ProductID:       receivedOrder.Lines[0].ProductID,
		FromWarehouseID: receivedOrder.WarehouseID,
		ToWarehouseID:   targetWarehouse.ID,
		Quantity:        2,
	})
	if err != nil {
		t.Fatalf("create transfer order: %v", err)
	}
	if order.Status != inventory.TransferOrderStatusPlanned {
		t.Fatalf("expected planned transfer order, got %s", order.Status)
	}

	executed, entries, err := svc.ExecuteTransferOrder(ctx, ExecuteTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "warehouse-a",
		TransferOrderID: order.ID,
	})
	if err != nil {
		t.Fatalf("execute transfer order: %v", err)
	}
	if executed.Status != inventory.TransferOrderStatusExecuted {
		t.Fatalf("expected executed transfer order, got %s", executed.Status)
	}
	if executed.ExecutedBy != "warehouse-a" {
		t.Fatalf("expected executed_by warehouse-a, got %s", executed.ExecutedBy)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 transfer order ledger entries, got %d", len(entries))
	}
	if entries[0].ReferenceType != "transfer_order" || entries[0].ReferenceID != order.ID {
		t.Fatalf("expected outbound reference transfer_order/%s, got %s/%s", order.ID, entries[0].ReferenceType, entries[0].ReferenceID)
	}
	if entries[1].ReferenceType != "transfer_order" || entries[1].ReferenceID != order.ID {
		t.Fatalf("expected inbound reference transfer_order/%s, got %s/%s", order.ID, entries[1].ReferenceType, entries[1].ReferenceID)
	}

	storedOrder, err := svc.GetTransferOrder(ctx, GetTransferOrderInput{
		TenantID:        "tenant-a",
		TransferOrderID: order.ID,
	})
	if err != nil {
		t.Fatalf("get transfer order: %v", err)
	}
	if storedOrder.Status != inventory.TransferOrderStatusExecuted {
		t.Fatalf("expected stored transfer order status executed, got %s", storedOrder.Status)
	}

	sourceBalance, err := svc.GetInventoryBalance(ctx, GetInventoryBalanceInput{
		TenantID:    "tenant-a",
		ProductID:   receivedOrder.Lines[0].ProductID,
		WarehouseID: receivedOrder.WarehouseID,
	})
	if err != nil {
		t.Fatalf("get source inventory balance: %v", err)
	}
	if sourceBalance.OnHand != receivedOrder.Lines[0].Quantity-2 {
		t.Fatalf("expected source on hand %d, got %d", receivedOrder.Lines[0].Quantity-2, sourceBalance.OnHand)
	}

	targetBalance, err := svc.GetInventoryBalance(ctx, GetInventoryBalanceInput{
		TenantID:    "tenant-a",
		ProductID:   receivedOrder.Lines[0].ProductID,
		WarehouseID: targetWarehouse.ID,
	})
	if err != nil {
		t.Fatalf("get target inventory balance: %v", err)
	}
	if targetBalance.OnHand != 2 {
		t.Fatalf("expected target on hand 2, got %d", targetBalance.OnHand)
	}
}

func TestServiceExecuteTransferOrderFailsWhenAlreadyExecuted(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	receivedOrder := createReceivedOrder(t, ctx, svc)

	targetWarehouse, err := svc.CreateWarehouse(ctx, CreateWarehouseInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Code:     "WH-BJ-REPLAY",
		Name:     "Beijing Warehouse Replay",
	})
	if err != nil {
		t.Fatalf("create target warehouse: %v", err)
	}

	order, err := svc.CreateTransferOrder(ctx, CreateTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "planner-a",
		ProductID:       receivedOrder.Lines[0].ProductID,
		FromWarehouseID: receivedOrder.WarehouseID,
		ToWarehouseID:   targetWarehouse.ID,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("create transfer order: %v", err)
	}
	if _, _, err := svc.ExecuteTransferOrder(ctx, ExecuteTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "warehouse-a",
		TransferOrderID: order.ID,
	}); err != nil {
		t.Fatalf("first execute transfer order: %v", err)
	}

	_, _, err = svc.ExecuteTransferOrder(ctx, ExecuteTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "warehouse-a",
		TransferOrderID: order.ID,
	})
	if !errors.Is(err, inventory.ErrTransferOrderNotExecutable) {
		t.Fatalf("expected transfer order not executable, got %v", err)
	}
}

func TestServiceListTransferOrdersSupportsStatusSortAndPagination(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	receivedOrder := createReceivedOrder(t, ctx, svc)

	warehouseA, err := svc.CreateWarehouse(ctx, CreateWarehouseInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Code:     "WH-LIST-A",
		Name:     "Warehouse List A",
	})
	if err != nil {
		t.Fatalf("create warehouse A: %v", err)
	}
	warehouseB, err := svc.CreateWarehouse(ctx, CreateWarehouseInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Code:     "WH-LIST-B",
		Name:     "Warehouse List B",
	})
	if err != nil {
		t.Fatalf("create warehouse B: %v", err)
	}
	warehouseC, err := svc.CreateWarehouse(ctx, CreateWarehouseInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Code:     "WH-LIST-C",
		Name:     "Warehouse List C",
	})
	if err != nil {
		t.Fatalf("create warehouse C: %v", err)
	}

	orderA, err := svc.CreateTransferOrder(ctx, CreateTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "planner-a",
		ProductID:       receivedOrder.Lines[0].ProductID,
		FromWarehouseID: receivedOrder.WarehouseID,
		ToWarehouseID:   warehouseA.ID,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("create transfer order A: %v", err)
	}
	orderB, err := svc.CreateTransferOrder(ctx, CreateTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "planner-a",
		ProductID:       receivedOrder.Lines[0].ProductID,
		FromWarehouseID: receivedOrder.WarehouseID,
		ToWarehouseID:   warehouseB.ID,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("create transfer order B: %v", err)
	}
	orderC, err := svc.CreateTransferOrder(ctx, CreateTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "planner-a",
		ProductID:       receivedOrder.Lines[0].ProductID,
		FromWarehouseID: receivedOrder.WarehouseID,
		ToWarehouseID:   warehouseC.ID,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("create transfer order C: %v", err)
	}

	if _, _, err := svc.ExecuteTransferOrder(ctx, ExecuteTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "warehouse-a",
		TransferOrderID: orderB.ID,
	}); err != nil {
		t.Fatalf("execute transfer order B: %v", err)
	}

	page1, err := svc.ListTransferOrders(ctx, ListTransferOrdersInput{
		TenantID: "tenant-a",
		Status:   string(inventory.TransferOrderStatusPlanned),
		Sort:     "id_asc",
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("list transfer orders page 1: %v", err)
	}
	if len(page1) != 1 {
		t.Fatalf("expected 1 planned order on page1, got %d", len(page1))
	}
	if page1[0].ID != orderA.ID {
		t.Fatalf("expected first planned order %s, got %s", orderA.ID, page1[0].ID)
	}

	page2, err := svc.ListTransferOrders(ctx, ListTransferOrdersInput{
		TenantID: "tenant-a",
		Status:   string(inventory.TransferOrderStatusPlanned),
		Sort:     "id_asc",
		Page:     2,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("list transfer orders page 2: %v", err)
	}
	if len(page2) != 1 {
		t.Fatalf("expected 1 planned order on page2, got %d", len(page2))
	}
	if page2[0].ID != orderC.ID {
		t.Fatalf("expected second planned order %s, got %s", orderC.ID, page2[0].ID)
	}

	executed, err := svc.ListTransferOrders(ctx, ListTransferOrdersInput{
		TenantID: "tenant-a",
		Status:   string(inventory.TransferOrderStatusExecuted),
		Sort:     "id_desc",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("list executed transfer orders: %v", err)
	}
	if len(executed) != 1 {
		t.Fatalf("expected 1 executed transfer order, got %d", len(executed))
	}
	if executed[0].ID != orderB.ID {
		t.Fatalf("expected executed transfer order %s, got %s", orderB.ID, executed[0].ID)
	}
}

func TestServiceListTransferOrdersFailsForInvalidQuery(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	_, err := svc.ListTransferOrders(ctx, ListTransferOrdersInput{
		TenantID: "tenant-a",
		Sort:     "unknown",
	})
	if !errors.Is(err, inventory.ErrInvalidTransferOrderQuery) {
		t.Fatalf("expected invalid transfer order query, got %v", err)
	}
}

func TestServiceCancelsTransferOrder(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	receivedOrder := createReceivedOrder(t, ctx, svc)

	targetWarehouse, err := svc.CreateWarehouse(ctx, CreateWarehouseInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Code:     "WH-CANCEL-1",
		Name:     "Warehouse Cancel 1",
	})
	if err != nil {
		t.Fatalf("create target warehouse: %v", err)
	}

	order, err := svc.CreateTransferOrder(ctx, CreateTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "planner-a",
		ProductID:       receivedOrder.Lines[0].ProductID,
		FromWarehouseID: receivedOrder.WarehouseID,
		ToWarehouseID:   targetWarehouse.ID,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("create transfer order: %v", err)
	}

	canceled, err := svc.CancelTransferOrder(ctx, CancelTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "planner-a",
		TransferOrderID: order.ID,
	})
	if err != nil {
		t.Fatalf("cancel transfer order: %v", err)
	}
	if canceled.Status != inventory.TransferOrderStatusCancelled {
		t.Fatalf("expected canceled status, got %s", canceled.Status)
	}
	if canceled.CanceledBy != "planner-a" {
		t.Fatalf("expected canceled_by planner-a, got %s", canceled.CanceledBy)
	}

	_, _, err = svc.ExecuteTransferOrder(ctx, ExecuteTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "warehouse-a",
		TransferOrderID: order.ID,
	})
	if !errors.Is(err, inventory.ErrTransferOrderNotExecutable) {
		t.Fatalf("expected transfer order not executable after cancellation, got %v", err)
	}
}

func TestServiceCancelTransferOrderFailsWhenAlreadyExecuted(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	receivedOrder := createReceivedOrder(t, ctx, svc)

	targetWarehouse, err := svc.CreateWarehouse(ctx, CreateWarehouseInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Code:     "WH-CANCEL-2",
		Name:     "Warehouse Cancel 2",
	})
	if err != nil {
		t.Fatalf("create target warehouse: %v", err)
	}

	order, err := svc.CreateTransferOrder(ctx, CreateTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "planner-a",
		ProductID:       receivedOrder.Lines[0].ProductID,
		FromWarehouseID: receivedOrder.WarehouseID,
		ToWarehouseID:   targetWarehouse.ID,
		Quantity:        1,
	})
	if err != nil {
		t.Fatalf("create transfer order: %v", err)
	}

	if _, _, err := svc.ExecuteTransferOrder(ctx, ExecuteTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "warehouse-a",
		TransferOrderID: order.ID,
	}); err != nil {
		t.Fatalf("execute transfer order: %v", err)
	}

	_, err = svc.CancelTransferOrder(ctx, CancelTransferOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "planner-a",
		TransferOrderID: order.ID,
	})
	if !errors.Is(err, inventory.ErrTransferOrderNotCancelable) {
		t.Fatalf("expected transfer order not cancelable, got %v", err)
	}
}

func TestServiceListsInventoryLedgerEntriesForWarehouseProduct(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	approvedOrder := createApprovedOrder(t, ctx, svc)

	_, _, _, err := svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
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
	_, err = svc.IssueInventory(ctx, IssueInventoryInput{
		TenantID:      "tenant-a",
		ActorID:       "warehouse-a",
		ProductID:     approvedOrder.Lines[0].ProductID,
		WarehouseID:   approvedOrder.WarehouseID,
		Quantity:      2,
		ReferenceType: "shipment",
		ReferenceID:   "shp-ledger-001",
	})
	if err != nil {
		t.Fatalf("issue inventory: %v", err)
	}

	entries, err := svc.ListInventoryLedger(ctx, ListInventoryLedgerInput{
		TenantID:    "tenant-a",
		ProductID:   approvedOrder.Lines[0].ProductID,
		WarehouseID: approvedOrder.WarehouseID,
	})
	if err != nil {
		t.Fatalf("list inventory ledger: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 ledger entries, got %d", len(entries))
	}
	if entries[0].MovementType != inventory.MovementTypeInbound || entries[0].QuantityDelta != approvedOrder.Lines[0].Quantity {
		t.Fatalf("expected first ledger entry inbound %d, got %s %d", approvedOrder.Lines[0].Quantity, entries[0].MovementType, entries[0].QuantityDelta)
	}
	if entries[1].MovementType != inventory.MovementTypeOutbound || entries[1].QuantityDelta != -2 {
		t.Fatalf("expected second ledger entry outbound -2, got %s %d", entries[1].MovementType, entries[1].QuantityDelta)
	}
}

func TestServiceCreatesAndShipsSalesOrder(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	receivedOrder := createReceivedOrder(t, ctx, svc)

	order, err := svc.CreateSalesOrder(ctx, CreateSalesOrderInput{
		TenantID:    "tenant-a",
		ActorID:     "sales-a",
		WarehouseID: receivedOrder.WarehouseID,
		ExternalRef: "SO-001",
		Lines: []CreateSalesOrderLine{{
			ProductID: receivedOrder.Lines[0].ProductID,
			Quantity:  2,
		}},
	})
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}
	if order.Status != sales.OrderStatusDraft {
		t.Fatalf("expected draft sales order, got %s", order.Status)
	}

	shippingOrder, entries, err := svc.ShipSalesOrder(ctx, ShipSalesOrderInput{
		TenantID:     "tenant-a",
		ActorID:      "warehouse-a",
		SalesOrderID: order.ID,
	})
	if err != nil {
		t.Fatalf("ship sales order: %v", err)
	}
	if shippingOrder.Status != sales.OrderStatusShipped {
		t.Fatalf("expected shipped sales order, got %s", shippingOrder.Status)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 shipment ledger entry, got %d", len(entries))
	}
	if entries[0].MovementType != inventory.MovementTypeOutbound || entries[0].QuantityDelta != -2 {
		t.Fatalf("expected outbound shipment entry -2, got %s %d", entries[0].MovementType, entries[0].QuantityDelta)
	}

	balance, err := svc.GetInventoryBalance(ctx, GetInventoryBalanceInput{
		TenantID:    "tenant-a",
		ProductID:   receivedOrder.Lines[0].ProductID,
		WarehouseID: receivedOrder.WarehouseID,
	})
	if err != nil {
		t.Fatalf("get inventory balance: %v", err)
	}
	if balance.OnHand != receivedOrder.Lines[0].Quantity-2 {
		t.Fatalf("expected on hand %d, got %d", receivedOrder.Lines[0].Quantity-2, balance.OnHand)
	}
	if balance.Available != receivedOrder.Lines[0].Quantity-2 {
		t.Fatalf("expected available %d, got %d", receivedOrder.Lines[0].Quantity-2, balance.Available)
	}
}

func TestServiceShipSalesOrderFailsWhenQuantityExceedsAvailable(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	receivedOrder := createReceivedOrder(t, ctx, svc)

	order, err := svc.CreateSalesOrder(ctx, CreateSalesOrderInput{
		TenantID:    "tenant-a",
		ActorID:     "sales-a",
		WarehouseID: receivedOrder.WarehouseID,
		ExternalRef: "SO-002",
		Lines: []CreateSalesOrderLine{{
			ProductID: receivedOrder.Lines[0].ProductID,
			Quantity:  receivedOrder.Lines[0].Quantity + 1,
		}},
	})
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}

	_, _, err = svc.ShipSalesOrder(ctx, ShipSalesOrderInput{
		TenantID:     "tenant-a",
		ActorID:      "warehouse-a",
		SalesOrderID: order.ID,
	})
	if !errors.Is(err, inventory.ErrInsufficientAvailableInventory) {
		t.Fatalf("expected insufficient available inventory, got %v", err)
	}
}

func TestServiceListsSalesOrdersByTenant(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	receivedOrder := createReceivedOrder(t, ctx, svc)

	orderA, err := svc.CreateSalesOrder(ctx, CreateSalesOrderInput{
		TenantID:    "tenant-a",
		ActorID:     "sales-a",
		WarehouseID: receivedOrder.WarehouseID,
		ExternalRef: "SO-003",
		Lines: []CreateSalesOrderLine{{
			ProductID: receivedOrder.Lines[0].ProductID,
			Quantity:  1,
		}},
	})
	if err != nil {
		t.Fatalf("create tenant-a sales order: %v", err)
	}
	if _, err := svc.CreateSalesOrder(ctx, CreateSalesOrderInput{
		TenantID:    "tenant-b",
		ActorID:     "sales-b",
		WarehouseID: receivedOrder.WarehouseID,
		ExternalRef: "SO-B-001",
		Lines: []CreateSalesOrderLine{{
			ProductID: receivedOrder.Lines[0].ProductID,
			Quantity:  1,
		}},
	}); err == nil {
		t.Fatalf("expected tenant-b create to fail due to tenant-scoped master data")
	}

	orders, err := svc.ListSalesOrders(ctx, ListSalesOrdersInput{TenantID: "tenant-a"})
	if err != nil {
		t.Fatalf("list sales orders: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 tenant-a sales order, got %d", len(orders))
	}
	if orders[0].ID != orderA.ID {
		t.Fatalf("expected tenant-a sales order id %s, got %s", orderA.ID, orders[0].ID)
	}
}

func TestServiceBuildsBackofficeOverviewReadModel(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	receivedOrder := createReceivedOrder(t, ctx, svc)

	if _, err := svc.CreatePayableBill(ctx, CreatePayableBillInput{
		TenantID:        "tenant-a",
		ActorID:         "finance-a",
		PurchaseOrderID: receivedOrder.ID,
	}); err != nil {
		t.Fatalf("create payable bill: %v", err)
	}
	if _, err := svc.CreateReceivableBill(ctx, CreateReceivableBillInput{
		TenantID:    "tenant-a",
		ActorID:     "finance-a",
		ExternalRef: "SO-OV-001",
	}); err != nil {
		t.Fatalf("create receivable bill: %v", err)
	}

	shipOrder, err := svc.CreateSalesOrder(ctx, CreateSalesOrderInput{
		TenantID:    "tenant-a",
		ActorID:     "sales-a",
		WarehouseID: receivedOrder.WarehouseID,
		ExternalRef: "SO-OV-002",
		Lines: []CreateSalesOrderLine{{
			ProductID: receivedOrder.Lines[0].ProductID,
			Quantity:  2,
		}},
	})
	if err != nil {
		t.Fatalf("create sales order to ship: %v", err)
	}
	if _, _, err := svc.ShipSalesOrder(ctx, ShipSalesOrderInput{
		TenantID:     "tenant-a",
		ActorID:      "warehouse-a",
		SalesOrderID: shipOrder.ID,
	}); err != nil {
		t.Fatalf("ship sales order: %v", err)
	}
	if _, err := svc.CreateSalesOrder(ctx, CreateSalesOrderInput{
		TenantID:    "tenant-a",
		ActorID:     "sales-a",
		WarehouseID: receivedOrder.WarehouseID,
		ExternalRef: "SO-OV-003",
		Lines: []CreateSalesOrderLine{{
			ProductID: receivedOrder.Lines[0].ProductID,
			Quantity:  1,
		}},
	}); err != nil {
		t.Fatalf("create draft sales order: %v", err)
	}

	overview, err := svc.GetBackofficeOverview(ctx, GetBackofficeOverviewInput{
		TenantID: "tenant-a",
	})
	if err != nil {
		t.Fatalf("get backoffice overview: %v", err)
	}
	if overview.Payable.OpenCount != 1 {
		t.Fatalf("expected payable open_count 1, got %d", overview.Payable.OpenCount)
	}
	if overview.Receivable.OpenCount != 1 {
		t.Fatalf("expected receivable open_count 1, got %d", overview.Receivable.OpenCount)
	}
	if overview.Sales.DraftCount != 1 {
		t.Fatalf("expected sales draft_count 1, got %d", overview.Sales.DraftCount)
	}
	if overview.Sales.ShippedCount != 1 {
		t.Fatalf("expected sales shipped_count 1, got %d", overview.Sales.ShippedCount)
	}
	if overview.Sales.TotalCount != 2 {
		t.Fatalf("expected sales total_count 2, got %d", overview.Sales.TotalCount)
	}
}

func TestServiceReceivePurchaseOrderFailsWhenOrderNotApproved(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	order, _ := createSubmittedOrder(t, ctx, svc)

	_, _, _, err := svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "receiver-a",
		PurchaseOrderID: order.ID,
		Lines: []ReceivePurchaseOrderLine{{
			ProductID: order.Lines[0].ProductID,
			Quantity:  order.Lines[0].Quantity,
		}},
	})
	if !errors.Is(err, procurement.ErrPurchaseOrderNotReceivable) {
		t.Fatalf("expected purchase order not receivable, got %v", err)
	}
}

func TestServiceReceivePurchaseOrderFailsForEmptyLines(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	approvedOrder := createApprovedOrder(t, ctx, svc)

	_, _, _, err := svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "receiver-a",
		PurchaseOrderID: approvedOrder.ID,
	})
	if !errors.Is(err, inventory.ErrInvalidReceipt) {
		t.Fatalf("expected invalid receipt, got %v", err)
	}
}

func TestServiceCreatesPayableBillForReceivedPurchaseOrder(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	receivedOrder := createReceivedOrder(t, ctx, svc)

	bill, err := svc.CreatePayableBill(ctx, CreatePayableBillInput{
		TenantID:        "tenant-a",
		ActorID:         "finance-a",
		PurchaseOrderID: receivedOrder.ID,
	})
	if err != nil {
		t.Fatalf("create payable bill: %v", err)
	}

	if bill.Status != payable.BillStatusOpen {
		t.Fatalf("expected open payable bill, got %s", bill.Status)
	}
	if bill.PurchaseOrderID != receivedOrder.ID {
		t.Fatalf("expected purchase order id %s, got %s", receivedOrder.ID, bill.PurchaseOrderID)
	}
}

func TestServiceCreatePayableBillFailsWhenOrderNotReceived(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	approvedOrder := createApprovedOrder(t, ctx, svc)

	_, err := svc.CreatePayableBill(ctx, CreatePayableBillInput{
		TenantID:        "tenant-a",
		ActorID:         "finance-a",
		PurchaseOrderID: approvedOrder.ID,
	})
	if !errors.Is(err, payable.ErrOrderNotBillable) {
		t.Fatalf("expected order not billable, got %v", err)
	}
}

func TestServiceCreatePayableBillFailsWhenBillAlreadyExists(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	receivedOrder := createReceivedOrder(t, ctx, svc)

	_, err := svc.CreatePayableBill(ctx, CreatePayableBillInput{
		TenantID:        "tenant-a",
		ActorID:         "finance-a",
		PurchaseOrderID: receivedOrder.ID,
	})
	if err != nil {
		t.Fatalf("create first payable bill: %v", err)
	}

	_, err = svc.CreatePayableBill(ctx, CreatePayableBillInput{
		TenantID:        "tenant-a",
		ActorID:         "finance-a",
		PurchaseOrderID: receivedOrder.ID,
	})
	if !errors.Is(err, payable.ErrBillAlreadyExists) {
		t.Fatalf("expected bill already exists, got %v", err)
	}
}

func TestServiceCreatesPayablePaymentPlanForBill(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	receivedOrder := createReceivedOrder(t, ctx, svc)

	bill, err := svc.CreatePayableBill(ctx, CreatePayableBillInput{
		TenantID:        "tenant-a",
		ActorID:         "finance-a",
		PurchaseOrderID: receivedOrder.ID,
	})
	if err != nil {
		t.Fatalf("create payable bill: %v", err)
	}

	plan, err := svc.CreatePayablePaymentPlan(ctx, CreatePayablePaymentPlanInput{
		TenantID:       "tenant-a",
		ActorID:        "finance-a",
		PayableBillID:  bill.ID,
		DueDateISO8601: "2026-04-01",
	})
	if err != nil {
		t.Fatalf("create payable payment plan: %v", err)
	}

	if plan.Status != payable.PaymentPlanStatusPlanned {
		t.Fatalf("expected planned payment plan, got %s", plan.Status)
	}
	if plan.PayableBillID != bill.ID {
		t.Fatalf("expected payable bill id %s, got %s", bill.ID, plan.PayableBillID)
	}
	if plan.DueDateISO8601 != "2026-04-01" {
		t.Fatalf("expected due date 2026-04-01, got %s", plan.DueDateISO8601)
	}
}

func TestServiceCreatePayablePaymentPlanFailsWhenBillNotFound(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	_, err := svc.CreatePayablePaymentPlan(ctx, CreatePayablePaymentPlanInput{
		TenantID:       "tenant-a",
		ActorID:        "finance-a",
		PayableBillID:  "pab-missing",
		DueDateISO8601: "2026-04-01",
	})
	if !errors.Is(err, payable.ErrBillNotFound) {
		t.Fatalf("expected bill not found, got %v", err)
	}
}

func TestServiceCreatePayablePaymentPlanFailsForInvalidDueDate(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	receivedOrder := createReceivedOrder(t, ctx, svc)

	bill, err := svc.CreatePayableBill(ctx, CreatePayableBillInput{
		TenantID:        "tenant-a",
		ActorID:         "finance-a",
		PurchaseOrderID: receivedOrder.ID,
	})
	if err != nil {
		t.Fatalf("create payable bill: %v", err)
	}

	_, err = svc.CreatePayablePaymentPlan(ctx, CreatePayablePaymentPlanInput{
		TenantID:       "tenant-a",
		ActorID:        "finance-a",
		PayableBillID:  bill.ID,
		DueDateISO8601: "20260401",
	})
	if !errors.Is(err, payable.ErrInvalidPaymentPlan) {
		t.Fatalf("expected invalid payment plan, got %v", err)
	}
}

func TestServiceListPayableBillsReturnsTenantScopedBills(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	receivedOrderA := createReceivedOrder(t, ctx, svc)
	billA, err := svc.CreatePayableBill(ctx, CreatePayableBillInput{
		TenantID:        "tenant-a",
		ActorID:         "finance-a",
		PurchaseOrderID: receivedOrderA.ID,
	})
	if err != nil {
		t.Fatalf("create tenant-a payable bill: %v", err)
	}

	supplierB, err := svc.CreateSupplier(ctx, CreateSupplierInput{
		TenantID: "tenant-b",
		ActorID:  "actor-b",
		Code:     "SUP-B-001",
		Name:     "Tenant B Supply",
	})
	if err != nil {
		t.Fatalf("create tenant-b supplier: %v", err)
	}
	productB, err := svc.CreateProduct(ctx, CreateProductInput{
		TenantID: "tenant-b",
		ActorID:  "actor-b",
		SKU:      "SKU-B-001",
		Name:     "Tenant B Product",
		Unit:     "pcs",
	})
	if err != nil {
		t.Fatalf("create tenant-b product: %v", err)
	}
	warehouseB, err := svc.CreateWarehouse(ctx, CreateWarehouseInput{
		TenantID: "tenant-b",
		ActorID:  "actor-b",
		Code:     "WH-BJ",
		Name:     "Beijing Warehouse",
	})
	if err != nil {
		t.Fatalf("create tenant-b warehouse: %v", err)
	}
	orderB, err := svc.CreatePurchaseOrder(ctx, CreatePurchaseOrderInput{
		TenantID:    "tenant-b",
		ActorID:     "actor-b",
		SupplierID:  supplierB.ID,
		WarehouseID: warehouseB.ID,
		Lines: []CreatePurchaseOrderLine{{
			ProductID: productB.ID,
			Quantity:  2,
		}},
	})
	if err != nil {
		t.Fatalf("create tenant-b order: %v", err)
	}
	submittedB, approvalB, err := svc.SubmitPurchaseOrder(ctx, SubmitPurchaseOrderInput{
		TenantID:        "tenant-b",
		ActorID:         "actor-b",
		PurchaseOrderID: orderB.ID,
	})
	if err != nil {
		t.Fatalf("submit tenant-b order: %v", err)
	}
	approvedB, _, err := svc.ApproveRequest(ctx, ResolveApprovalInput{
		TenantID:   "tenant-b",
		ActorID:    "manager-b",
		ApprovalID: approvalB.ID,
	})
	if err != nil {
		t.Fatalf("approve tenant-b order: %v", err)
	}
	_, _, _, err = svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
		TenantID:        "tenant-b",
		ActorID:         "receiver-b",
		PurchaseOrderID: submittedB.ID,
		Lines: []ReceivePurchaseOrderLine{{
			ProductID: approvedB.Lines[0].ProductID,
			Quantity:  approvedB.Lines[0].Quantity,
		}},
	})
	if err != nil {
		t.Fatalf("receive tenant-b order: %v", err)
	}
	if _, err := svc.CreatePayableBill(ctx, CreatePayableBillInput{
		TenantID:        "tenant-b",
		ActorID:         "finance-b",
		PurchaseOrderID: approvedB.ID,
	}); err != nil {
		t.Fatalf("create tenant-b payable bill: %v", err)
	}

	bills, err := svc.ListPayableBills(ctx, ListPayableBillsInput{TenantID: "tenant-a"})
	if err != nil {
		t.Fatalf("list tenant-a payable bills: %v", err)
	}
	if len(bills) != 1 {
		t.Fatalf("expected 1 tenant-a payable bill, got %d", len(bills))
	}
	if bills[0].ID != billA.ID {
		t.Fatalf("expected tenant-a payable bill id %s, got %s", billA.ID, bills[0].ID)
	}
}

func TestServiceCreatesReceivableBill(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	bill, err := svc.CreateReceivableBill(ctx, CreateReceivableBillInput{
		TenantID:    "tenant-a",
		ActorID:     "finance-a",
		ExternalRef: "SO-001",
	})
	if err != nil {
		t.Fatalf("create receivable bill: %v", err)
	}
	if bill.Status != receivable.BillStatusOpen {
		t.Fatalf("expected open receivable bill, got %s", bill.Status)
	}
	if bill.ExternalRef != "SO-001" {
		t.Fatalf("expected external_ref SO-001, got %s", bill.ExternalRef)
	}
}

func TestServiceCreateReceivableBillFailsForInvalidInput(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	_, err := svc.CreateReceivableBill(ctx, CreateReceivableBillInput{
		TenantID:    "tenant-a",
		ActorID:     "finance-a",
		ExternalRef: "   ",
	})
	if !errors.Is(err, receivable.ErrInvalidBill) {
		t.Fatalf("expected invalid receivable bill, got %v", err)
	}
}

func TestServiceListReceivableBillsReturnsTenantScopedBills(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	billA, err := svc.CreateReceivableBill(ctx, CreateReceivableBillInput{
		TenantID:    "tenant-a",
		ActorID:     "finance-a",
		ExternalRef: "SO-001",
	})
	if err != nil {
		t.Fatalf("create tenant-a receivable bill: %v", err)
	}
	if _, err := svc.CreateReceivableBill(ctx, CreateReceivableBillInput{
		TenantID:    "tenant-b",
		ActorID:     "finance-b",
		ExternalRef: "SO-B-001",
	}); err != nil {
		t.Fatalf("create tenant-b receivable bill: %v", err)
	}

	bills, err := svc.ListReceivableBills(ctx, ListReceivableBillsInput{
		TenantID: "tenant-a",
	})
	if err != nil {
		t.Fatalf("list tenant-a receivable bills: %v", err)
	}
	if len(bills) != 1 {
		t.Fatalf("expected 1 tenant-a receivable bill, got %d", len(bills))
	}
	if bills[0].ID != billA.ID {
		t.Fatalf("expected tenant-a receivable bill id %s, got %s", billA.ID, bills[0].ID)
	}
}

func newTestService() *Service {
	return newTestServiceWithInventoryRepository(memory.NewInventoryRepository())
}

func newTestServiceWithInventoryRepository(inventoryRepo inventory.Repository) *Service {
	if inventoryRepo == nil {
		inventoryRepo = memory.NewInventoryRepository()
	}
	return NewService(ServiceDeps{
		MasterData:     memory.NewMasterDataRepository(),
		PurchaseOrders: memory.NewPurchaseOrderRepository(),
		Approvals:      memory.NewApprovalRepository(),
		Inventory:      inventoryRepo,
		Payables:       memory.NewPayableRepository(),
		Receivables:    memory.NewReceivableRepository(),
		SalesOrders:    memory.NewSalesOrderRepository(),
		Pipeline:       shared.NewPipeline(shared.PipelineDeps{}),
	})
}

type failingInventoryRepository struct {
	delegate        inventory.Repository
	saveReceiptErr  error
	appendLedgerErr error
}

func (r failingInventoryRepository) SaveReceipt(ctx context.Context, receipt inventory.Receipt) error {
	if r.saveReceiptErr != nil {
		return r.saveReceiptErr
	}
	return r.delegate.SaveReceipt(ctx, receipt)
}

func (r failingInventoryRepository) AppendLedgerEntries(ctx context.Context, entries []inventory.LedgerEntry) error {
	if r.appendLedgerErr != nil {
		return r.appendLedgerErr
	}
	return r.delegate.AppendLedgerEntries(ctx, entries)
}

func (r failingInventoryRepository) ListLedgerEntries(ctx context.Context, tenantID, productID, warehouseID string) ([]inventory.LedgerEntry, error) {
	return r.delegate.ListLedgerEntries(ctx, tenantID, productID, warehouseID)
}

func (r failingInventoryRepository) SaveReservation(ctx context.Context, reservation inventory.Reservation) error {
	return r.delegate.SaveReservation(ctx, reservation)
}

func (r failingInventoryRepository) ListReservations(ctx context.Context, tenantID, productID, warehouseID string) ([]inventory.Reservation, error) {
	return r.delegate.ListReservations(ctx, tenantID, productID, warehouseID)
}

func (r failingInventoryRepository) SaveTransferOrder(ctx context.Context, order inventory.TransferOrder) error {
	return r.delegate.SaveTransferOrder(ctx, order)
}

func (r failingInventoryRepository) GetTransferOrder(ctx context.Context, tenantID, transferOrderID string) (inventory.TransferOrder, error) {
	return r.delegate.GetTransferOrder(ctx, tenantID, transferOrderID)
}

func (r failingInventoryRepository) ListTransferOrdersByTenant(ctx context.Context, tenantID string) ([]inventory.TransferOrder, error) {
	return r.delegate.ListTransferOrdersByTenant(ctx, tenantID)
}

func createSubmittedOrder(t *testing.T, ctx context.Context, svc *Service) (procurement.PurchaseOrder, approval.Request) {
	t.Helper()

	supplier, product, warehouse := createMasterData(t, ctx, svc)

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

	submittedOrder, approvalRequest, err := svc.SubmitPurchaseOrder(ctx, SubmitPurchaseOrderInput{
		TenantID:        "tenant-a",
		ActorID:         "actor-a",
		PurchaseOrderID: order.ID,
	})
	if err != nil {
		t.Fatalf("submit order: %v", err)
	}

	return submittedOrder, approvalRequest
}

func createApprovedOrder(t *testing.T, ctx context.Context, svc *Service) procurement.PurchaseOrder {
	t.Helper()

	_, approvalRequest := createSubmittedOrder(t, ctx, svc)
	approvedOrder, _, err := svc.ApproveRequest(ctx, ResolveApprovalInput{
		TenantID:   "tenant-a",
		ActorID:    "manager-a",
		ApprovalID: approvalRequest.ID,
	})
	if err != nil {
		t.Fatalf("approve request: %v", err)
	}
	return approvedOrder
}

func createReceivedOrder(t *testing.T, ctx context.Context, svc *Service) procurement.PurchaseOrder {
	t.Helper()

	approvedOrder := createApprovedOrder(t, ctx, svc)
	_, _, receivedOrder, err := svc.ReceivePurchaseOrder(ctx, ReceivePurchaseOrderInput{
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
	return receivedOrder
}

func createMasterData(t *testing.T, ctx context.Context, svc *Service) (masterdata.Supplier, masterdata.Product, masterdata.Warehouse) {
	t.Helper()

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

	return supplier, product, warehouse
}
