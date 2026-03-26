package capability

import (
	"context"
	"errors"
	"fmt"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var ErrAgentCapabilityPolicyRepositoryRequired = errors.New("agent capability policy repository is required")
var errAgentCapabilityPolicyProfileRepositoryRequired = errors.New("agent profile repository is required")
var ErrAgentProfileNotFound = errors.New("agent profile not found")

type agentCapabilityProfileReader interface {
	ListAgentProfiles(ctx context.Context, tenantID string) ([]controlplane.AgentProfile, error)
}

type modelCatalogReader interface {
	ListByTenant(ctx context.Context, tenantID string) ([]*domaincap.ModelCatalogEntry, error)
}

type toolCatalogReader interface {
	ListToolsByTenant(ctx context.Context, tenantID string) ([]*domaincap.ToolCatalogEntry, error)
}

type SaveAgentCapabilityPolicyPayload struct {
	AgentProfileID       string
	AllowedModelEntryIDs []string
	AllowedToolEntryIDs  []string
}

type SaveAgentCapabilityPolicyHandler struct {
	repo     domaincap.AgentCapabilityPolicyRepository
	profiles agentCapabilityProfileReader
	models   modelCatalogReader
	tools    toolCatalogReader
}

func NewSaveAgentCapabilityPolicyHandler(
	repo domaincap.AgentCapabilityPolicyRepository,
	profiles agentCapabilityProfileReader,
	models modelCatalogReader,
	tools toolCatalogReader,
) (*SaveAgentCapabilityPolicyHandler, error) {
	if repo == nil {
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
	return &SaveAgentCapabilityPolicyHandler{
		repo:     repo,
		profiles: profiles,
		models:   models,
		tools:    tools,
	}, nil
}

func (h *SaveAgentCapabilityPolicyHandler) Handle(ctx context.Context, tenantID string, payload SaveAgentCapabilityPolicyPayload) (*domaincap.AgentCapabilityPolicy, error) {
	policy, err := domaincap.NewAgentCapabilityPolicy(
		tenantID,
		payload.AgentProfileID,
		payload.AllowedModelEntryIDs,
		payload.AllowedToolEntryIDs,
	)
	if err != nil {
		return nil, err
	}

	if err := ensureAgentProfile(ctx, h.profiles, tenantID, payload.AgentProfileID); err != nil {
		return nil, err
	}
	if err := ensureModelEntryIDs(ctx, h.models, tenantID, policy.AllowedModelEntryIDs); err != nil {
		return nil, err
	}
	if err := ensureToolEntryIDs(ctx, h.tools, tenantID, policy.AllowedToolEntryIDs); err != nil {
		return nil, err
	}

	if err := h.repo.SaveAgentCapabilityPolicy(ctx, policy); err != nil {
		return nil, err
	}

	return policy, nil
}

func ensureAgentProfile(ctx context.Context, profiles agentCapabilityProfileReader, tenantID, agentProfileID string) error {
	available, err := profiles.ListAgentProfiles(ctx, tenantID)
	if err != nil {
		return err
	}
	for _, profile := range available {
		if profile.ID == agentProfileID {
			return nil
		}
	}
	return fmt.Errorf("%w: %q", ErrAgentProfileNotFound, agentProfileID)
}

func ensureModelEntryIDs(ctx context.Context, models modelCatalogReader, tenantID string, entryIDs []string) error {
	available, err := models.ListByTenant(ctx, tenantID)
	if err != nil {
		return err
	}
	index := make(map[string]struct{}, len(available))
	active := make(map[string]bool, len(available))
	for _, entry := range available {
		index[entry.EntryID] = struct{}{}
		active[entry.EntryID] = entry.IsActive()
	}
	for _, entryID := range entryIDs {
		if _, ok := index[entryID]; !ok {
			return fmt.Errorf("model catalog entry %q not found", entryID)
		}
		if !active[entryID] {
			return fmt.Errorf("model catalog entry %q is inactive", entryID)
		}
	}
	return nil
}

func ensureToolEntryIDs(ctx context.Context, tools toolCatalogReader, tenantID string, entryIDs []string) error {
	available, err := tools.ListToolsByTenant(ctx, tenantID)
	if err != nil {
		return err
	}
	index := make(map[string]struct{}, len(available))
	active := make(map[string]bool, len(available))
	for _, entry := range available {
		index[entry.EntryID] = struct{}{}
		active[entry.EntryID] = entry.IsActive()
	}
	for _, entryID := range entryIDs {
		if _, ok := index[entryID]; !ok {
			return fmt.Errorf("tool catalog entry %q not found", entryID)
		}
		if !active[entryID] {
			return fmt.Errorf("tool catalog entry %q is inactive", entryID)
		}
	}
	return nil
}
