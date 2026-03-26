package capability

import (
	"context"
	"reflect"
	"testing"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

func TestSaveAgentCapabilityPolicyHandlerNormalizesAndSaves(t *testing.T) {
	t.Parallel()

	repo := &fakeAgentCapabilityPolicyRepository{
		profiles: []controlplane.AgentProfile{{TenantID: "tenant-a", ID: "profile-1"}},
		models: []*domaincap.ModelCatalogEntry{
			{TenantID: "tenant-a", EntryID: "model-1"},
			{TenantID: "tenant-a", EntryID: "model-2"},
		},
		tools: []*domaincap.ToolCatalogEntry{
			{TenantID: "tenant-a", EntryID: "tool-1"},
			{TenantID: "tenant-a", EntryID: "tool-2"},
		},
	}
	handler, err := NewSaveAgentCapabilityPolicyHandler(repo, repo, repo, repo)
	if err != nil {
		t.Fatalf("handler init failed: %v", err)
	}

	policy, err := handler.Handle(context.Background(), "tenant-a", SaveAgentCapabilityPolicyPayload{
		AgentProfileID:      "profile-1",
		AllowedModelEntryIDs: []string{"model-2", "model-1", "model-2"},
		AllowedToolEntryIDs:  []string{"tool-2", "tool-1", "tool-2"},
	})
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if len(repo.saved) != 1 {
		t.Fatalf("expected repository save called once, got %d", len(repo.saved))
	}

	wantModels := []string{"model-1", "model-2"}
	if !reflect.DeepEqual(policy.AllowedModelEntryIDs, wantModels) {
		t.Fatalf("unexpected model ids: got %v want %v", policy.AllowedModelEntryIDs, wantModels)
	}
	if !reflect.DeepEqual(repo.saved[0].AllowedModelEntryIDs, wantModels) {
		t.Fatalf("repository received unexpected model ids: %v", repo.saved[0].AllowedModelEntryIDs)
	}
}

func TestGetAgentCapabilityPolicyHandlerDelegates(t *testing.T) {
	t.Parallel()

	repo := &fakeAgentCapabilityPolicyRepository{
		profiles: []controlplane.AgentProfile{{TenantID: "tenant-a", ID: "profile-1"}},
		getPolicy: &domaincap.AgentCapabilityPolicy{
			TenantID:             "tenant-a",
			AgentProfileID:       "profile-1",
			AllowedModelEntryIDs: []string{"model-1"},
			AllowedToolEntryIDs:  []string{"tool-1"},
		},
	}
	handler, err := NewGetAgentCapabilityPolicyHandler(repo, repo)
	if err != nil {
		t.Fatalf("handler init failed: %v", err)
	}

	policy, err := handler.Handle(context.Background(), "tenant-a", "profile-1")
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	if repo.lastTenant != "tenant-a" || repo.lastProfile != "profile-1" {
		t.Fatalf("unexpected repository lookup: tenant=%s profile=%s", repo.lastTenant, repo.lastProfile)
	}
	if policy.AgentProfileID != "profile-1" {
		t.Fatalf("unexpected policy: %+v", policy)
	}
}

func TestSaveAgentCapabilityPolicyHandlerRejectsUnknownTenantLocalReferences(t *testing.T) {
	t.Parallel()

	repo := &fakeAgentCapabilityPolicyRepository{
		profiles: []controlplane.AgentProfile{{TenantID: "tenant-a", ID: "profile-1"}},
		models:   []*domaincap.ModelCatalogEntry{{TenantID: "tenant-b", EntryID: "model-foreign"}},
		tools:    []*domaincap.ToolCatalogEntry{{TenantID: "tenant-b", EntryID: "tool-foreign"}},
	}
	handler, err := NewSaveAgentCapabilityPolicyHandler(repo, repo, repo, repo)
	if err != nil {
		t.Fatalf("handler init failed: %v", err)
	}

	_, err = handler.Handle(context.Background(), "tenant-a", SaveAgentCapabilityPolicyPayload{
		AgentProfileID:       "profile-1",
		AllowedModelEntryIDs: []string{"model-foreign"},
		AllowedToolEntryIDs:  []string{"tool-foreign"},
	})
	if err == nil {
		t.Fatal("expected unknown tenant-local reference error")
	}
}

type fakeAgentCapabilityPolicyRepository struct {
	saved       []*domaincap.AgentCapabilityPolicy
	getPolicy   *domaincap.AgentCapabilityPolicy
	lastTenant  string
	lastProfile string
	profiles    []controlplane.AgentProfile
	models      []*domaincap.ModelCatalogEntry
	tools       []*domaincap.ToolCatalogEntry
}

func (f *fakeAgentCapabilityPolicyRepository) SaveAgentCapabilityPolicy(_ context.Context, policy *domaincap.AgentCapabilityPolicy) error {
	f.saved = append(f.saved, policy)
	return nil
}

func (f *fakeAgentCapabilityPolicyRepository) GetAgentCapabilityPolicy(_ context.Context, tenantID, agentProfileID string) (*domaincap.AgentCapabilityPolicy, error) {
	f.lastTenant = tenantID
	f.lastProfile = agentProfileID
	return f.getPolicy, nil
}

func (f *fakeAgentCapabilityPolicyRepository) ListAgentProfiles(_ context.Context, tenantID string) ([]controlplane.AgentProfile, error) {
	out := make([]controlplane.AgentProfile, 0, len(f.profiles))
	for _, profile := range f.profiles {
		if profile.TenantID != tenantID {
			continue
		}
		out = append(out, profile)
	}
	return out, nil
}

func (f *fakeAgentCapabilityPolicyRepository) ListByTenant(_ context.Context, tenantID string) ([]*domaincap.ModelCatalogEntry, error) {
	out := make([]*domaincap.ModelCatalogEntry, 0, len(f.models))
	for _, entry := range f.models {
		if entry.TenantID != tenantID {
			continue
		}
		out = append(out, entry)
	}
	return out, nil
}

func (f *fakeAgentCapabilityPolicyRepository) ListToolsByTenant(_ context.Context, tenantID string) ([]*domaincap.ToolCatalogEntry, error) {
	out := make([]*domaincap.ToolCatalogEntry, 0, len(f.tools))
	for _, entry := range f.tools {
		if entry.TenantID != tenantID {
			continue
		}
		out = append(out, entry)
	}
	return out, nil
}
