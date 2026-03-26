package query

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var errListUsersHandlerUserRepositoryRequired = errors.New("list users handler requires user repository")

type ListUsers struct {
	TenantID string
}

type ListUsersHandler struct {
	Users     controlplane.UserRepository
	Authorize func(context.Context, ListUsers) error
	Audit     func(context.Context, []controlplane.User) error
}

func (h ListUsersHandler) Handle(ctx context.Context, q ListUsers) ([]controlplane.User, error) {
	if h.Users == nil {
		return nil, errListUsersHandlerUserRepositoryRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, q); err != nil {
			return nil, err
		}
	}

	users, err := h.Users.ListUsers(ctx, q.TenantID)
	if err != nil {
		return nil, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, users); err != nil {
			return nil, err
		}
	}

	return users, nil
}
