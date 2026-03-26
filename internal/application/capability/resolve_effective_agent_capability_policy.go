package capability

import (
	"context"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

type ResolveEffectiveAgentCapabilityPolicyHandler struct {
	policies domaincap.AgentCapabilityPolicyRepository
	profiles agentCapabilityProfileReader
	models   modelCatalogReader
	tools    toolCatalogReader
}

func NewResolveEffectiveAgentCapabilityPolicyHandler(
	policies domaincap.AgentCapabilityPolicyRepository,
	profiles agentCapabilityProfileReader,
	models modelCatalogReader,
	tools toolCatalogReader,
) (*ResolveEffectiveAgentCapabilityPolicyHandler, error) {
	if policies == nil {
		return nil, ErrAgentCapabilityPolicyRepositoryRequired
	}
	if profiles == nil {
		return nil, errAgentCapabilityPolicyProfileRepositoryRequired
	}
	if models == nil {
		return nil, ErrRepositoryRequired
	}
	if tools == nil {
		return nil, ErrToolRepositoryRequired
	}
	return &ResolveEffectiveAgentCapabilityPolicyHandler{
		policies: policies,
		profiles: profiles,
		models:   models,
		tools:    tools,
	}, nil
}

func (h *ResolveEffectiveAgentCapabilityPolicyHandler) Handle(ctx context.Context, tenantID, agentProfileID string) (*domaincap.EffectiveAgentCapabilityPolicy, error) {
	if err := ensureAgentProfile(ctx, h.profiles, tenantID, agentProfileID); err != nil {
		return nil, err
	}

	policy, err := h.policies.GetAgentCapabilityPolicy(ctx, tenantID, agentProfileID)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		policy, err = domaincap.NewAgentCapabilityPolicy(tenantID, agentProfileID, nil, nil)
		if err != nil {
			return nil, err
		}
	}

	models, err := h.models.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	tools, err := h.tools.ListToolsByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	activeModels := make(map[string]bool, len(models))
	for _, entry := range models {
		activeModels[entry.EntryID] = entry.IsActive()
	}

	activeTools := make(map[string]bool, len(tools))
	for _, entry := range tools {
		activeTools[entry.EntryID] = entry.IsActive()
	}

	effectiveModelIDs := make([]string, 0, len(policy.AllowedModelEntryIDs))
	staleModelIDs := make([]string, 0)
	for _, entryID := range policy.AllowedModelEntryIDs {
		if activeModels[entryID] {
			effectiveModelIDs = append(effectiveModelIDs, entryID)
			continue
		}
		staleModelIDs = append(staleModelIDs, entryID)
	}

	effectiveToolIDs := make([]string, 0, len(policy.AllowedToolEntryIDs))
	staleToolIDs := make([]string, 0)
	for _, entryID := range policy.AllowedToolEntryIDs {
		if activeTools[entryID] {
			effectiveToolIDs = append(effectiveToolIDs, entryID)
			continue
		}
		staleToolIDs = append(staleToolIDs, entryID)
	}

	return domaincap.NewEffectiveAgentCapabilityPolicy(
		tenantID,
		agentProfileID,
		effectiveModelIDs,
		effectiveToolIDs,
		staleModelIDs,
		staleToolIDs,
	)
}
