package query

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var errListAgentProfilesHandlerRepositoryRequired = errors.New("list agent profiles handler requires agent profile repository")

type ListAgentProfiles struct {
	TenantID string
}

type ListAgentProfilesHandler struct {
	Profiles  controlplane.AgentProfileRepository
	Authorize func(context.Context, ListAgentProfiles) error
	Audit     func(context.Context, []controlplane.AgentProfile) error
}

func (h ListAgentProfilesHandler) Handle(ctx context.Context, q ListAgentProfiles) ([]controlplane.AgentProfile, error) {
	if h.Profiles == nil {
		return nil, errListAgentProfilesHandlerRepositoryRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, q); err != nil {
			return nil, err
		}
	}

	profiles, err := h.Profiles.ListAgentProfiles(ctx, q.TenantID)
	if err != nil {
		return nil, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, profiles); err != nil {
			return nil, err
		}
	}

	return profiles, nil
}
