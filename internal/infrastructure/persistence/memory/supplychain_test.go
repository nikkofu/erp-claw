package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/nikkofu/erp-claw/internal/domain/inventory"
	"github.com/nikkofu/erp-claw/internal/domain/payable"
	"github.com/nikkofu/erp-claw/internal/domain/procurement"
)

func TestPurchaseOrderRepositoryGetReturnsDetachedCopy(t *testing.T) {
	ctx := context.Background()
	repo := NewSupplyChainStore().PurchaseOrderRepository()

	order := procurement.PurchaseOrder{
		ID:          "po-001",
		TenantID:    "tenant-a",
		SupplierID:  "sup-001",
		WarehouseID: "wh-001",
		Status:      procurement.PurchaseOrderStatusDraft,
		Lines: []procurement.Line{{
			ProductID: "prd-001",
			Quantity:  5,
		}},
	}
	if err := repo.Save(ctx, order); err != nil {
		t.Fatalf("save order: %v", err)
	}

	got, err := repo.Get(ctx, order.TenantID, order.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	got.Lines[0].Quantity = 99

	reloaded, err := repo.Get(ctx, order.TenantID, order.ID)
	if err != nil {
		t.Fatalf("reload order: %v", err)
	}
	if reloaded.Lines[0].Quantity != 5 {
		t.Fatalf("expected stored quantity 5, got %d", reloaded.Lines[0].Quantity)
	}
}

func TestPurchaseOrderRepositorySaveDetachesCallerSlice(t *testing.T) {
	ctx := context.Background()
	repo := NewSupplyChainStore().PurchaseOrderRepository()

	lines := []procurement.Line{{
		ProductID: "prd-001",
		Quantity:  5,
	}}
	order := procurement.PurchaseOrder{
		ID:          "po-002",
		TenantID:    "tenant-a",
		SupplierID:  "sup-001",
		WarehouseID: "wh-001",
		Status:      procurement.PurchaseOrderStatusDraft,
		Lines:       lines,
	}
	if err := repo.Save(ctx, order); err != nil {
		t.Fatalf("save order: %v", err)
	}

	lines[0].Quantity = 77

	reloaded, err := repo.Get(ctx, order.TenantID, order.ID)
	if err != nil {
		t.Fatalf("reload order: %v", err)
	}
	if reloaded.Lines[0].Quantity != 5 {
		t.Fatalf("expected stored quantity 5, got %d", reloaded.Lines[0].Quantity)
	}
}

func TestInventoryRepositoryListLedgerEntriesReturnsDetachedCopy(t *testing.T) {
	ctx := context.Background()
	repo := NewSupplyChainStore().InventoryRepository()

	entry, err := inventory.NewInboundLedgerEntry("led-001", "tenant-a", "prd-001", "wh-001", "receipt", "rcv-001", 5)
	if err != nil {
		t.Fatalf("new ledger entry: %v", err)
	}
	if err := repo.AppendLedgerEntries(ctx, []inventory.LedgerEntry{entry}); err != nil {
		t.Fatalf("append ledger entries: %v", err)
	}

	got, err := repo.ListLedgerEntries(ctx, "tenant-a", "prd-001", "wh-001")
	if err != nil {
		t.Fatalf("list ledger entries: %v", err)
	}
	got[0].QuantityDelta = 99

	reloaded, err := repo.ListLedgerEntries(ctx, "tenant-a", "prd-001", "wh-001")
	if err != nil {
		t.Fatalf("reload ledger entries: %v", err)
	}
	if reloaded[0].QuantityDelta != 5 {
		t.Fatalf("expected stored quantity delta 5, got %d", reloaded[0].QuantityDelta)
	}
}

func TestPayableRepositoryGetReturnsDetachedCopy(t *testing.T) {
	ctx := context.Background()
	repo := NewSupplyChainStore().PayableRepository()

	bill, err := payable.NewBill("pab-001", "tenant-a", "po-001", "finance-a")
	if err != nil {
		t.Fatalf("new bill: %v", err)
	}
	if err := repo.Save(ctx, bill); err != nil {
		t.Fatalf("save bill: %v", err)
	}

	got, err := repo.Get(ctx, bill.TenantID, bill.ID)
	if err != nil {
		t.Fatalf("get bill: %v", err)
	}
	got.CreatedBy = "tampered"

	reloaded, err := repo.Get(ctx, bill.TenantID, bill.ID)
	if err != nil {
		t.Fatalf("reload bill: %v", err)
	}
	if reloaded.CreatedBy != "finance-a" {
		t.Fatalf("expected stored created_by finance-a, got %s", reloaded.CreatedBy)
	}
}

func TestPayableRepositoryRejectsDuplicateBillPerPurchaseOrder(t *testing.T) {
	ctx := context.Background()
	repo := NewSupplyChainStore().PayableRepository()

	first, err := payable.NewBill("pab-001", "tenant-a", "po-001", "finance-a")
	if err != nil {
		t.Fatalf("new first bill: %v", err)
	}
	if err := repo.Save(ctx, first); err != nil {
		t.Fatalf("save first bill: %v", err)
	}

	second, err := payable.NewBill("pab-002", "tenant-a", "po-001", "finance-a")
	if err != nil {
		t.Fatalf("new second bill: %v", err)
	}
	err = repo.Save(ctx, second)
	if !errors.Is(err, payable.ErrBillAlreadyExists) {
		t.Fatalf("expected bill already exists, got %v", err)
	}
}

func TestPayableRepositoryListPaymentPlansByBillReturnsDetachedCopy(t *testing.T) {
	ctx := context.Background()
	repo := NewSupplyChainStore().PayableRepository()

	bill, err := payable.NewBill("pab-001", "tenant-a", "po-001", "finance-a")
	if err != nil {
		t.Fatalf("new bill: %v", err)
	}
	if err := repo.Save(ctx, bill); err != nil {
		t.Fatalf("save bill: %v", err)
	}

	plan, err := payable.NewPaymentPlan("ppm-001", "tenant-a", bill.ID, "finance-a", "2026-04-01")
	if err != nil {
		t.Fatalf("new payment plan: %v", err)
	}
	if err := repo.SavePaymentPlan(ctx, plan); err != nil {
		t.Fatalf("save payment plan: %v", err)
	}

	got, err := repo.ListPaymentPlansByBill(ctx, "tenant-a", bill.ID)
	if err != nil {
		t.Fatalf("list payment plans: %v", err)
	}
	got[0].DueDateISO8601 = "2099-12-31"

	reloaded, err := repo.ListPaymentPlansByBill(ctx, "tenant-a", bill.ID)
	if err != nil {
		t.Fatalf("reload payment plans: %v", err)
	}
	if reloaded[0].DueDateISO8601 != "2026-04-01" {
		t.Fatalf("expected stored due date 2026-04-01, got %s", reloaded[0].DueDateISO8601)
	}
}

func TestPayableRepositorySavePaymentPlanFailsWhenBillNotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewSupplyChainStore().PayableRepository()

	plan, err := payable.NewPaymentPlan("ppm-001", "tenant-a", "pab-missing", "finance-a", "2026-04-01")
	if err != nil {
		t.Fatalf("new payment plan: %v", err)
	}
	err = repo.SavePaymentPlan(ctx, plan)
	if !errors.Is(err, payable.ErrBillNotFound) {
		t.Fatalf("expected bill not found, got %v", err)
	}
}

func TestPayableRepositoryListByTenantScopesResults(t *testing.T) {
	ctx := context.Background()
	repo := NewSupplyChainStore().PayableRepository()

	billA, err := payable.NewBill("pab-001", "tenant-a", "po-001", "finance-a")
	if err != nil {
		t.Fatalf("new tenant-a bill: %v", err)
	}
	if err := repo.Save(ctx, billA); err != nil {
		t.Fatalf("save tenant-a bill: %v", err)
	}

	billB, err := payable.NewBill("pab-002", "tenant-b", "po-002", "finance-b")
	if err != nil {
		t.Fatalf("new tenant-b bill: %v", err)
	}
	if err := repo.Save(ctx, billB); err != nil {
		t.Fatalf("save tenant-b bill: %v", err)
	}

	got, err := repo.ListByTenant(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("list tenant-a bills: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 tenant-a bill, got %d", len(got))
	}
	if got[0].ID != billA.ID {
		t.Fatalf("expected tenant-a bill id %s, got %s", billA.ID, got[0].ID)
	}
}
