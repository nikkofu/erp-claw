package tenant

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
)

var (
	ErrTenantNotFound = errors.New("tenant not found")
	ErrInvalidTenant  = errors.New("invalid tenant")
)

const TenantStatusActive = "active"

// Tenant is the minimal control-plane tenant aggregate for Phase 1.
type Tenant struct {
	ID     string
	Code   string
	Name   string
	Status string
}

func NewTenant(id, code, name string) (Tenant, error) {
	tenant := Tenant{
		ID:     strings.TrimSpace(id),
		Code:   strings.TrimSpace(code),
		Name:   strings.TrimSpace(name),
		Status: TenantStatusActive,
	}
	if tenant.ID == "" || tenant.Code == "" || tenant.Name == "" {
		return Tenant{}, ErrInvalidTenant
	}
	return tenant, nil
}

// Catalog stores and resolves tenant metadata.
type Catalog interface {
	Save(ctx context.Context, tenant Tenant) error
	Get(ctx context.Context, code string) (Tenant, error)
	List(ctx context.Context) ([]Tenant, error)
}

type InMemoryCatalog struct {
	mu     sync.RWMutex
	byCode map[string]Tenant
}

func NewInMemoryCatalog() *InMemoryCatalog {
	return &InMemoryCatalog{
		byCode: make(map[string]Tenant),
	}
}

func (c *InMemoryCatalog) Save(_ context.Context, tenant Tenant) error {
	tenant.ID = strings.TrimSpace(tenant.ID)
	tenant.Code = strings.TrimSpace(tenant.Code)
	tenant.Name = strings.TrimSpace(tenant.Name)
	if tenant.ID == "" || tenant.Code == "" || tenant.Name == "" {
		return ErrInvalidTenant
	}
	if tenant.Status == "" {
		tenant.Status = TenantStatusActive
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.byCode[tenant.Code] = tenant
	return nil
}

func (c *InMemoryCatalog) Get(_ context.Context, code string) (Tenant, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return Tenant{}, ErrTenantNotFound
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	tenant, ok := c.byCode[code]
	if !ok {
		return Tenant{}, ErrTenantNotFound
	}
	return tenant, nil
}

func (c *InMemoryCatalog) List(_ context.Context) ([]Tenant, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]Tenant, 0, len(c.byCode))
	for _, value := range c.byCode {
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Code < out[j].Code
	})
	return out, nil
}

// CatalogResolver resolves tenant routing information from a tenant catalog.
type CatalogResolver struct {
	Catalog Catalog
}

func (r CatalogResolver) Resolve(tenantCode string) (CellRoute, error) {
	tenantCode = strings.TrimSpace(tenantCode)
	if tenantCode == "" {
		return CellRoute{}, errUnknownTenant
	}
	if r.Catalog == nil {
		return SimpleResolver{}.Resolve(tenantCode)
	}

	tenant, err := r.Catalog.Get(context.Background(), tenantCode)
	if err != nil {
		return CellRoute{}, errUnknownTenant
	}

	return CellRoute{
		TenantID:      tenant.Code,
		Isolation:     "logical_cell",
		CachePrefix:   tenant.Code + ":",
		StoragePrefix: tenant.Code + "/",
	}, nil
}
