package supplychain

import (
	"context"
	"errors"
	"testing"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/domain/approval"
	"github.com/nikkofu/erp-claw/internal/domain/masterdata"
	"github.com/nikkofu/erp-claw/internal/domain/procurement"
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

func newTestService() *Service {
	return NewService(ServiceDeps{
		MasterData:     memory.NewMasterDataRepository(),
		PurchaseOrders: memory.NewPurchaseOrderRepository(),
		Approvals:      memory.NewApprovalRepository(),
		Pipeline:       shared.NewPipeline(shared.PipelineDeps{}),
	})
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
