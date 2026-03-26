package command

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var errAssignUserDepartmentHandlerBindingRepositoryRequired = errors.New("assign user department handler requires binding repository")

type AssignUserDepartment struct {
	TenantID     string
	UserID       string
	DepartmentID string
}

type AssignUserDepartmentHandler struct {
	Bindings  controlplane.UserDepartmentBindingRepository
	Authorize func(context.Context, AssignUserDepartment) error
	Audit     func(context.Context, controlplane.UserDepartmentBinding) error
}

func (h AssignUserDepartmentHandler) Handle(ctx context.Context, cmd AssignUserDepartment) (controlplane.UserDepartmentBinding, error) {
	binding, err := controlplane.NewUserDepartmentBinding(cmd.TenantID, cmd.UserID, cmd.DepartmentID)
	if err != nil {
		return controlplane.UserDepartmentBinding{}, err
	}
	if h.Bindings == nil {
		return controlplane.UserDepartmentBinding{}, errAssignUserDepartmentHandlerBindingRepositoryRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, cmd); err != nil {
			return controlplane.UserDepartmentBinding{}, err
		}
	}

	created, err := h.Bindings.AssignUserDepartment(ctx, binding)
	if err != nil {
		return controlplane.UserDepartmentBinding{}, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, created); err != nil {
			return controlplane.UserDepartmentBinding{}, err
		}
	}

	return created, nil
}
