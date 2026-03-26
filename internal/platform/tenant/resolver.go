package tenant

import "errors"

var errUnknownTenant = errors.New("unknown tenant")

// Resolver resolves tenant IDs to cell routes.
type Resolver interface {
	Resolve(tenantID string) (CellRoute, error)
}

// SimpleResolver is a placeholder resolver used while the platform lacks a catalog.
type SimpleResolver struct{}

// Resolve returns a minimal cell route when the tenant ID is non-empty.
func (SimpleResolver) Resolve(tenantID string) (CellRoute, error) {
	if tenantID == "" {
		return CellRoute{}, errUnknownTenant
	}
	return CellRoute{TenantID: tenantID}, nil
}

// ChainResolver resolves with Primary first and falls back when Primary fails.
type ChainResolver struct {
	Primary  Resolver
	Fallback Resolver
}

func (r ChainResolver) Resolve(tenantID string) (CellRoute, error) {
	if r.Primary != nil {
		route, err := r.Primary.Resolve(tenantID)
		if err == nil {
			return route, nil
		}
	}
	if r.Fallback != nil {
		return r.Fallback.Resolve(tenantID)
	}
	return CellRoute{}, errUnknownTenant
}
