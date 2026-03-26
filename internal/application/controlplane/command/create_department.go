package command

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var errCreateDepartmentHandlerDepartmentRepositoryRequired = errors.New("create department handler requires department repository")

type CreateDepartment struct {
	TenantID           string
	Name               string
	ParentDepartmentID string
}

type CreateDepartmentHandler struct {
	Departments controlplane.DepartmentRepository
	Authorize   func(context.Context, CreateDepartment) error
	Audit       func(context.Context, controlplane.Department) error
}

func (h CreateDepartmentHandler) Handle(ctx context.Context, cmd CreateDepartment) (controlplane.Department, error) {
	department, err := controlplane.NewDepartment(cmd.TenantID, cmd.Name, cmd.ParentDepartmentID)
	if err != nil {
		return controlplane.Department{}, err
	}
	if h.Departments == nil {
		return controlplane.Department{}, errCreateDepartmentHandlerDepartmentRepositoryRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, cmd); err != nil {
			return controlplane.Department{}, err
		}
	}

	created, err := h.Departments.CreateDepartment(ctx, department)
	if err != nil {
		return controlplane.Department{}, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, created); err != nil {
			return controlplane.Department{}, err
		}
	}

	return created, nil
}
