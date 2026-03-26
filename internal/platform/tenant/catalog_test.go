package tenant

import (
	"context"
	"errors"
	"testing"
)

func TestInMemoryCatalogSavesAndLoadsTenantByCode(t *testing.T) {
	catalog := NewInMemoryCatalog()

	created, err := NewTenant("tenant-001", "tenant-admin", "Admin Tenant")
	if err != nil {
		t.Fatalf("new tenant: %v", err)
	}
	if err := catalog.Save(context.Background(), created); err != nil {
		t.Fatalf("save tenant: %v", err)
	}

	got, err := catalog.Get(context.Background(), "tenant-admin")
	if err != nil {
		t.Fatalf("get tenant: %v", err)
	}
	if got.ID != "tenant-001" {
		t.Fatalf("expected tenant id tenant-001, got %s", got.ID)
	}
	if got.Name != "Admin Tenant" {
		t.Fatalf("expected tenant name Admin Tenant, got %s", got.Name)
	}
}

func TestInMemoryCatalogReturnsNotFoundForUnknownCode(t *testing.T) {
	catalog := NewInMemoryCatalog()

	_, err := catalog.Get(context.Background(), "tenant-missing")
	if !errors.Is(err, ErrTenantNotFound) {
		t.Fatalf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestCatalogResolverUsesCatalogRoute(t *testing.T) {
	catalog := NewInMemoryCatalog()
	created, err := NewTenant("tenant-001", "tenant-admin", "Admin Tenant")
	if err != nil {
		t.Fatalf("new tenant: %v", err)
	}
	if err := catalog.Save(context.Background(), created); err != nil {
		t.Fatalf("save tenant: %v", err)
	}

	resolver := CatalogResolver{Catalog: catalog}
	route, err := resolver.Resolve("tenant-admin")
	if err != nil {
		t.Fatalf("resolve tenant: %v", err)
	}
	if route.TenantID != "tenant-admin" {
		t.Fatalf("expected route tenant tenant-admin, got %s", route.TenantID)
	}
	if route.Isolation != "logical_cell" {
		t.Fatalf("expected logical_cell isolation, got %s", route.Isolation)
	}
}
