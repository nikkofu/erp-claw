package capability

import (
	"context"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

type GetAgentCapabilityPolicyHandler struct {
	repo     domaincap.AgentCapabilityPolicyRepository
	profiles agentCapabilityProfileReader
}

func NewGetAgentCapabilityPolicyHandler(
	repo domaincap.AgentCapabilityPolicyRepository,
	profiles agentCapabilityProfileReader,
) (*GetAgentCapabilityPolicyHandler, error) {
	if repo == nil {
		return nil, ErrAgentCapabilityPolicyRepositoryRequired
	}
	if profiles == nil {
		return nil, errAgentCapabilityPolicyProfileRepositoryRequired
	}
	return &GetAgentCapabilityPolicyHandler{repo: repo, profiles: profiles}, nil
}

func (h *GetAgentCapabilityPolicyHandler) Handle(ctx context.Context, tenantID, agentProfileID string) (*domaincap.AgentCapabilityPolicy, error) {
	if err := ensureAgentProfile(ctx, h.profiles, tenantID, agentProfileID); err != nil {
		return nil, err
	}

	policy, err := h.repo.GetAgentCapabilityPolicy(ctx, tenantID, agentProfileID)
	if err != nil {
		return nil, err
	}
	if policy != nil {
		return policy, nil
	}
	return domaincap.NewAgentCapabilityPolicy(tenantID, agentProfileID, nil, nil)
}
