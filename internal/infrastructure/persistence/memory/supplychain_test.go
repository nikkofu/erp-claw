package memory

import (
	"context"
	"testing"

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
