package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/nikkofu/erp-claw/internal/domain/masterdata"
	"github.com/nikkofu/erp-claw/internal/platform/iam"
	"github.com/nikkofu/erp-claw/internal/platform/tenant"
)

func TestAtomicTransactionManagerWithinTransactionNilHandlerNoop(t *testing.T) {
	manager := NewAtomicTransactionManager(NewSupplyChainStore(), NewControlPlaneStore())
	if err := manager.WithinTransaction(context.Background(), nil); err != nil {
		t.Fatalf("nil handler should not fail: %v", err)
	}
}

func TestAtomicTransactionManagerWithinTransactionCommitsOnSuccess(t *testing.T) {
	ctx := context.Background()
	supplyChainStore := NewSupplyChainStore()
	controlPlaneStore := NewControlPlaneStore()
	manager := NewAtomicTransactionManager(supplyChainStore, controlPlaneStore)

	supplierRepo := supplyChainStore.MasterDataRepository()
	tenantRepo := controlPlaneStore.TenantCatalog()

	supplier, err := masterdata.NewSupplier("sup-001", "tenant-a", "SUP-001", "Northwind")
	if err != nil {
		t.Fatalf("new supplier: %v", err)
	}
	tenantValue, err := tenant.NewTenant("tenant-a", "tenant-a", "Tenant A")
	if err != nil {
		t.Fatalf("new tenant: %v", err)
	}

	if err := manager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := supplierRepo.SaveSupplier(txCtx, supplier); err != nil {
			return err
		}
		return tenantRepo.Save(txCtx, tenantValue)
	}); err != nil {
		t.Fatalf("commit transaction: %v", err)
	}

	if _, err := supplierRepo.GetSupplier(ctx, supplier.TenantID, supplier.ID); err != nil {
		t.Fatalf("supplier should be persisted: %v", err)
	}
	if _, err := tenantRepo.Get(ctx, tenantValue.Code); err != nil {
		t.Fatalf("tenant should be persisted: %v", err)
	}
}

func TestAtomicTransactionManagerWithinTransactionRollsBackSupplyChainOnError(t *testing.T) {
	ctx := context.Background()
	supplyChainStore := NewSupplyChainStore()
	controlPlaneStore := NewControlPlaneStore()
	manager := NewAtomicTransactionManager(supplyChainStore, controlPlaneStore)

	supplierRepo := supplyChainStore.MasterDataRepository()
	seed, err := masterdata.NewSupplier("sup-001", "tenant-a", "SUP-001", "Original")
	if err != nil {
		t.Fatalf("new seed supplier: %v", err)
	}
	if err := supplierRepo.SaveSupplier(ctx, seed); err != nil {
		t.Fatalf("save seed supplier: %v", err)
	}

	updated, err := masterdata.NewSupplier("sup-001", "tenant-a", "SUP-001", "Updated")
	if err != nil {
		t.Fatalf("new updated supplier: %v", err)
	}
	transient, err := masterdata.NewSupplier("sup-002", "tenant-a", "SUP-002", "Transient")
	if err != nil {
		t.Fatalf("new transient supplier: %v", err)
	}

	expectedErr := errors.New("force rollback")
	err = manager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := supplierRepo.SaveSupplier(txCtx, updated); err != nil {
			return err
		}
		if err := supplierRepo.SaveSupplier(txCtx, transient); err != nil {
			return err
		}
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected rollback error %v, got %v", expectedErr, err)
	}

	reloaded, err := supplierRepo.GetSupplier(ctx, seed.TenantID, seed.ID)
	if err != nil {
		t.Fatalf("reload seed supplier: %v", err)
	}
	if reloaded.Name != "Original" {
		t.Fatalf("expected seed supplier name Original, got %s", reloaded.Name)
	}
	if _, err := supplierRepo.GetSupplier(ctx, transient.TenantID, transient.ID); !errors.Is(err, masterdata.ErrSupplierNotFound) {
		t.Fatalf("transient supplier should be rolled back, got %v", err)
	}
}

func TestAtomicTransactionManagerWithinTransactionRollsBackControlPlaneOnError(t *testing.T) {
	ctx := context.Background()
	controlPlaneStore := NewControlPlaneStore()
	manager := NewAtomicTransactionManager(nil, controlPlaneStore)

	tenantRepo := controlPlaneStore.TenantCatalog()
	iamRepo := controlPlaneStore.IAMDirectory()

	expectedErr := errors.New("force rollback")
	err := manager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := tenantRepo.Save(txCtx, tenant.Tenant{
			ID:     "tenant-a",
			Code:   "tenant-a",
			Name:   "Tenant A",
			Status: tenant.TenantStatusActive,
		}); err != nil {
			return err
		}
		if err := iamRepo.Save(txCtx, "tenant-a", iam.Actor{
			ID:    "actor-a",
			Roles: []string{"platform_admin"},
		}); err != nil {
			return err
		}
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected rollback error %v, got %v", expectedErr, err)
	}

	if _, err := tenantRepo.Get(ctx, "tenant-a"); !errors.Is(err, tenant.ErrTenantNotFound) {
		t.Fatalf("tenant should be rolled back, got %v", err)
	}
	if _, err := iamRepo.Get(ctx, "tenant-a", "actor-a"); !errors.Is(err, iam.ErrActorNotFound) {
		t.Fatalf("actor should be rolled back, got %v", err)
	}
}
