package query

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var errListDepartmentsHandlerDepartmentRepositoryRequired = errors.New("list departments handler requires department repository")

type ListDepartments struct {
	TenantID string
}

type ListDepartmentsHandler struct {
	Departments controlplane.DepartmentRepository
	Authorize   func(context.Context, ListDepartments) error
	Audit       func(context.Context, []controlplane.Department) error
}

func (h ListDepartmentsHandler) Handle(ctx context.Context, q ListDepartments) ([]controlplane.Department, error) {
	if h.Departments == nil {
		return nil, errListDepartmentsHandlerDepartmentRepositoryRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, q); err != nil {
			return nil, err
		}
	}

	departments, err := h.Departments.ListDepartments(ctx, q.TenantID)
	if err != nil {
		return nil, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, departments); err != nil {
			return nil, err
		}
	}

	return departments, nil
}
