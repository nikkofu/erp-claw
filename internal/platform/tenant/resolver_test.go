package tenant

import (
	"context"
	"testing"
)

func TestChainResolverUsesPrimaryWhenTenantExists(t *testing.T) {
	catalog := NewInMemoryCatalog()
	value, err := NewTenant("tenant-001", "tenant-admin", "Admin")
	if err != nil {
		t.Fatalf("new tenant: %v", err)
	}
	if err := catalog.Save(context.Background(), value); err != nil {
		t.Fatalf("save tenant: %v", err)
	}

	resolver := ChainResolver{
		Primary:  CatalogResolver{Catalog: catalog},
		Fallback: SimpleResolver{},
	}
	route, err := resolver.Resolve("tenant-admin")
	if err != nil {
		t.Fatalf("resolve tenant: %v", err)
	}
	if route.Isolation != "logical_cell" {
		t.Fatalf("expected logical_cell from catalog resolver, got %s", route.Isolation)
	}
}

func TestChainResolverFallsBackWhenTenantMissingFromCatalog(t *testing.T) {
	resolver := ChainResolver{
		Primary:  CatalogResolver{Catalog: NewInMemoryCatalog()},
		Fallback: SimpleResolver{},
	}

	route, err := resolver.Resolve("tenant-missing")
	if err != nil {
		t.Fatalf("resolve tenant with fallback: %v", err)
	}
	if route.TenantID != "tenant-missing" {
		t.Fatalf("expected fallback tenant id tenant-missing, got %s", route.TenantID)
	}
	if route.Isolation != "" {
		t.Fatalf("expected empty isolation from simple fallback, got %s", route.Isolation)
	}
}
