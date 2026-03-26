package command

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var errCreateUserHandlerUserRepositoryRequired = errors.New("create user handler requires user repository")

type CreateUser struct {
	TenantID    string
	Email       string
	DisplayName string
}

type CreateUserHandler struct {
	Users     controlplane.UserRepository
	Authorize func(context.Context, CreateUser) error
	Audit     func(context.Context, controlplane.User) error
}

func (h CreateUserHandler) Handle(ctx context.Context, cmd CreateUser) (controlplane.User, error) {
	user, err := controlplane.NewUser(cmd.TenantID, cmd.Email, cmd.DisplayName)
	if err != nil {
		return controlplane.User{}, err
	}
	if h.Users == nil {
		return controlplane.User{}, errCreateUserHandlerUserRepositoryRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, cmd); err != nil {
			return controlplane.User{}, err
		}
	}

	created, err := h.Users.CreateUser(ctx, user)
	if err != nil {
		return controlplane.User{}, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, created); err != nil {
			return controlplane.User{}, err
		}
	}

	return created, nil
}
