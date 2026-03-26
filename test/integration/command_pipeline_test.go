package integration

import (
	"context"
	"errors"
	"sort"
	"testing"

	appcap "github.com/nikkofu/erp-claw/internal/application/capability"
	"github.com/nikkofu/erp-claw/internal/application/shared"
	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestCommandPipelineRejectsDeniedPolicy(t *testing.T) {
	p := shared.NewPipeline(shared.PipelineDeps{
		Policy: policy.StaticEvaluator(policy.DecisionDeny),
	})

	err := p.Execute(context.Background(), shared.Command{
		Name:     "customers.create",
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Payload:  map[string]any{"name": "Acme"},
	})

	if err == nil {
		t.Fatalf("expected denied policy to reject command")
	}
}

func TestCommandPipelineAllowsEffectiveCapabilityRequest(t *testing.T) {
	repo := newIntegrationCapabilityRepo()
	repo.profiles = []controlplane.AgentProfile{{TenantID: "tenant-a", ID: "profile-1"}}
	repo.getPolicy = &domaincap.AgentCapabilityPolicy{
		TenantID:             "tenant-a",
		AgentProfileID:       "profile-1",
		AllowedModelEntryIDs: []string{"model-1"},
		AllowedToolEntryIDs:  []string{"tool-1"},
	}
	repo.models = []*domaincap.ModelCatalogEntry{{TenantID: "tenant-a", EntryID: "model-1", Status: domaincap.CatalogStatusActive}}
	repo.tools = []*domaincap.ToolCatalogEntry{{TenantID: "tenant-a", EntryID: "tool-1", Status: domaincap.CatalogStatusActive}}

	resolver, err := appcap.NewResolveEffectiveAgentCapabilityPolicyHandler(repo, repo, repo, repo)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	p := shared.NewPipeline(shared.PipelineDeps{
		Policy:       policy.StaticEvaluator(policy.DecisionAllow),
		Capabilities: appcap.SharedCommandCapabilityAuthorizer{Resolver: resolver},
	})

	handlerCalls := 0
	err = p.Execute(context.Background(), shared.Command{
		Name:     "agent.run",
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Payload: map[string]any{
			"agent_profile_id": "profile-1",
			"model_entry_id":   "model-1",
			"tool_entry_ids":   []string{"tool-1"},
		},
	}, func(context.Context, shared.Command) error {
		handlerCalls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected allowed capability request, got %v", err)
	}
	if handlerCalls != 1 {
		t.Fatalf("expected handler to run once, got %d", handlerCalls)
	}
}

func TestCommandPipelineBlocksStaleCapabilityRequest(t *testing.T) {
	repo := newIntegrationCapabilityRepo()
	repo.profiles = []controlplane.AgentProfile{{TenantID: "tenant-a", ID: "profile-1"}}
	repo.getPolicy = &domaincap.AgentCapabilityPolicy{
		TenantID:             "tenant-a",
		AgentProfileID:       "profile-1",
		AllowedModelEntryIDs: []string{"model-2"},
	}
	repo.models = []*domaincap.ModelCatalogEntry{{TenantID: "tenant-a", EntryID: "model-2", Status: domaincap.CatalogStatusInactive}}

	resolver, err := appcap.NewResolveEffectiveAgentCapabilityPolicyHandler(repo, repo, repo, repo)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	p := shared.NewPipeline(shared.PipelineDeps{
		Policy:       policy.StaticEvaluator(policy.DecisionAllow),
		Capabilities: appcap.SharedCommandCapabilityAuthorizer{Resolver: resolver},
	})

	handlerCalls := 0
	err = p.Execute(context.Background(), shared.Command{
		Name:     "agent.run",
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Payload: map[string]any{
			"agent_profile_id": "profile-1",
			"model_entry_id":   "model-2",
		},
	}, func(context.Context, shared.Command) error {
		handlerCalls++
		return nil
	})
	if !errors.Is(err, shared.ErrCapabilityDenied) {
		t.Fatalf("expected capability denied error, got %v", err)
	}
	if handlerCalls != 0 {
		t.Fatalf("expected handler not to run, got %d calls", handlerCalls)
	}
}

type integrationCapabilityRepo struct {
	profiles    []controlplane.AgentProfile
	getPolicy   *domaincap.AgentCapabilityPolicy
	models      []*domaincap.ModelCatalogEntry
	tools       []*domaincap.ToolCatalogEntry
	lastTenant  string
	lastProfile string
}

func newIntegrationCapabilityRepo() *integrationCapabilityRepo {
	return &integrationCapabilityRepo{}
}

func (r *integrationCapabilityRepo) GetAgentCapabilityPolicy(_ context.Context, tenantID, agentProfileID string) (*domaincap.AgentCapabilityPolicy, error) {
	r.lastTenant = tenantID
	r.lastProfile = agentProfileID
	return r.getPolicy, nil
}

func (r *integrationCapabilityRepo) SaveAgentCapabilityPolicy(_ context.Context, _ *domaincap.AgentCapabilityPolicy) error {
	return nil
}

func (r *integrationCapabilityRepo) ListAgentProfiles(_ context.Context, tenantID string) ([]controlplane.AgentProfile, error) {
	out := make([]controlplane.AgentProfile, 0, len(r.profiles))
	for _, profile := range r.profiles {
		if profile.TenantID == tenantID {
			out = append(out, profile)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (r *integrationCapabilityRepo) ListByTenant(_ context.Context, tenantID string) ([]*domaincap.ModelCatalogEntry, error) {
	out := make([]*domaincap.ModelCatalogEntry, 0, len(r.models))
	for _, entry := range r.models {
		if entry.TenantID == tenantID {
			out = append(out, entry)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].EntryID < out[j].EntryID
	})
	return out, nil
}

func (r *integrationCapabilityRepo) ListToolsByTenant(_ context.Context, tenantID string) ([]*domaincap.ToolCatalogEntry, error) {
	out := make([]*domaincap.ToolCatalogEntry, 0, len(r.tools))
	for _, entry := range r.tools {
		if entry.TenantID == tenantID {
			out = append(out, entry)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].EntryID < out[j].EntryID
	})
	return out, nil
}
