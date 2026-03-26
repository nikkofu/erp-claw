package query

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var errListTenantsHandlerTenantRepositoryRequired = errors.New("list tenants handler requires tenant repository")

type ListTenants struct{}

type ListTenantsHandler struct {
	Tenants   controlplane.TenantRepository
	Authorize func(context.Context, ListTenants) error
	Audit     func(context.Context, []controlplane.Tenant) error
}

func (h ListTenantsHandler) Handle(ctx context.Context, q ListTenants) ([]controlplane.Tenant, error) {
	if h.Tenants == nil {
		return nil, errListTenantsHandlerTenantRepositoryRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, q); err != nil {
			return nil, err
		}
	}

	tenants, err := h.Tenants.ListTenants(ctx)
	if err != nil {
		return nil, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, tenants); err != nil {
			return nil, err
		}
	}

	return tenants, nil
}
