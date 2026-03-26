package command

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var errCreateTenantHandlerTenantRepositoryRequired = errors.New("create tenant handler requires tenant repository")

type CreateTenant struct {
	Code string
	Name string
}

type CreateTenantHandler struct {
	Tenants   controlplane.TenantRepository
	Authorize func(context.Context, CreateTenant) error
	Audit     func(context.Context, controlplane.Tenant) error
}

func (h CreateTenantHandler) Handle(ctx context.Context, cmd CreateTenant) (controlplane.Tenant, error) {
	tenant, err := controlplane.NewTenant(cmd.Code, cmd.Name)
	if err != nil {
		return controlplane.Tenant{}, err
	}
	if h.Tenants == nil {
		return controlplane.Tenant{}, errCreateTenantHandlerTenantRepositoryRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, cmd); err != nil {
			return controlplane.Tenant{}, err
		}
	}

	created, err := h.Tenants.CreateTenant(ctx, tenant)
	if err != nil {
		return controlplane.Tenant{}, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, created); err != nil {
			return controlplane.Tenant{}, err
		}
	}

	return created, nil
}
