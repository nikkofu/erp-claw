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
