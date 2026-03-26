package command

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var errAssignUserRoleHandlerBindingRepositoryRequired = errors.New("assign user role handler requires binding repository")

type AssignUserRole struct {
	TenantID string
	UserID   string
	RoleID   string
}

type AssignUserRoleHandler struct {
	Bindings  controlplane.UserRoleBindingRepository
	Authorize func(context.Context, AssignUserRole) error
	Audit     func(context.Context, controlplane.UserRoleBinding) error
}

func (h AssignUserRoleHandler) Handle(ctx context.Context, cmd AssignUserRole) (controlplane.UserRoleBinding, error) {
	binding, err := controlplane.NewUserRoleBinding(cmd.TenantID, cmd.UserID, cmd.RoleID)
	if err != nil {
		return controlplane.UserRoleBinding{}, err
	}
	if h.Bindings == nil {
		return controlplane.UserRoleBinding{}, errAssignUserRoleHandlerBindingRepositoryRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, cmd); err != nil {
			return controlplane.UserRoleBinding{}, err
		}
	}

	created, err := h.Bindings.AssignUserRole(ctx, binding)
	if err != nil {
		return controlplane.UserRoleBinding{}, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, created); err != nil {
			return controlplane.UserRoleBinding{}, err
		}
	}

	return created, nil
}
