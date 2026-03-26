package command

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var errCreateAgentProfileHandlerRepositoryRequired = errors.New("create agent profile handler requires agent profile repository")

type CreateAgentProfile struct {
	TenantID string
	Name     string
	Model    string
}

type CreateAgentProfileHandler struct {
	Profiles  controlplane.AgentProfileRepository
	Authorize func(context.Context, CreateAgentProfile) error
	Audit     func(context.Context, controlplane.AgentProfile) error
}

func (h CreateAgentProfileHandler) Handle(ctx context.Context, cmd CreateAgentProfile) (controlplane.AgentProfile, error) {
	profile, err := controlplane.NewAgentProfile(cmd.TenantID, cmd.Name, cmd.Model)
	if err != nil {
		return controlplane.AgentProfile{}, err
	}
	if h.Profiles == nil {
		return controlplane.AgentProfile{}, errCreateAgentProfileHandlerRepositoryRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, cmd); err != nil {
			return controlplane.AgentProfile{}, err
		}
	}

	created, err := h.Profiles.CreateAgentProfile(ctx, profile)
	if err != nil {
		return controlplane.AgentProfile{}, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, created); err != nil {
			return controlplane.AgentProfile{}, err
		}
	}

	return created, nil
}
