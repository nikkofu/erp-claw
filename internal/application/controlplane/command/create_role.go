package command

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var errCreateRoleHandlerRoleRepositoryRequired = errors.New("create role handler requires role repository")

type CreateRole struct {
	TenantID    string
	Name        string
	Description string
}

type CreateRoleHandler struct {
	Roles     controlplane.RoleRepository
	Authorize func(context.Context, CreateRole) error
	Audit     func(context.Context, controlplane.Role) error
}

func (h CreateRoleHandler) Handle(ctx context.Context, cmd CreateRole) (controlplane.Role, error) {
	role, err := controlplane.NewRole(cmd.TenantID, cmd.Name, cmd.Description)
	if err != nil {
		return controlplane.Role{}, err
	}
	if h.Roles == nil {
		return controlplane.Role{}, errCreateRoleHandlerRoleRepositoryRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, cmd); err != nil {
			return controlplane.Role{}, err
		}
	}

	created, err := h.Roles.CreateRole(ctx, role)
	if err != nil {
		return controlplane.Role{}, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, created); err != nil {
			return controlplane.Role{}, err
		}
	}

	return created, nil
}
