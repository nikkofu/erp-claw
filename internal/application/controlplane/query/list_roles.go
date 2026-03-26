package query

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var errListRolesHandlerRoleRepositoryRequired = errors.New("list roles handler requires role repository")

type ListRoles struct {
	TenantID string
}

type ListRolesHandler struct {
	Roles     controlplane.RoleRepository
	Authorize func(context.Context, ListRoles) error
	Audit     func(context.Context, []controlplane.Role) error
}

func (h ListRolesHandler) Handle(ctx context.Context, q ListRoles) ([]controlplane.Role, error) {
	if h.Roles == nil {
		return nil, errListRolesHandlerRoleRepositoryRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, q); err != nil {
			return nil, err
		}
	}

	roles, err := h.Roles.ListRoles(ctx, q.TenantID)
	if err != nil {
		return nil, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, roles); err != nil {
			return nil, err
		}
	}

	return roles, nil
}
