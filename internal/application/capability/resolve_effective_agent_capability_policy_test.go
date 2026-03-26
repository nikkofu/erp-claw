package capability

import (
	"context"
	"reflect"
	"testing"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

func TestResolveEffectiveAgentCapabilityPolicyHandlerSeparatesActiveAndStaleBindings(t *testing.T) {
	t.Parallel()

	repo := &fakeAgentCapabilityPolicyRepository{
		profiles: []controlplane.AgentProfile{{TenantID: "tenant-a", ID: "profile-1"}},
		getPolicy: &domaincap.AgentCapabilityPolicy{
			TenantID:             "tenant-a",
			AgentProfileID:       "profile-1",
			AllowedModelEntryIDs: []string{"model-1", "model-2"},
			AllowedToolEntryIDs:  []string{"tool-1", "tool-2"},
		},
		models: []*domaincap.ModelCatalogEntry{
			{TenantID: "tenant-a", EntryID: "model-1", Status: domaincap.CatalogStatusActive},
			{TenantID: "tenant-a", EntryID: "model-2", Status: domaincap.CatalogStatusInactive},
		},
		tools: []*domaincap.ToolCatalogEntry{
			{TenantID: "tenant-a", EntryID: "tool-1", Status: domaincap.CatalogStatusActive},
		},
	}
	handler, err := NewResolveEffectiveAgentCapabilityPolicyHandler(repo, repo, repo, repo)
	if err != nil {
		t.Fatalf("handler init failed: %v", err)
	}

	policy, err := handler.Handle(context.Background(), "tenant-a", "profile-1")
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if !reflect.DeepEqual(policy.EffectiveModelEntryIDs, []string{"model-1"}) {
		t.Fatalf("unexpected effective model ids: %v", policy.EffectiveModelEntryIDs)
	}
	if !reflect.DeepEqual(policy.StaleModelEntryIDs, []string{"model-2"}) {
		t.Fatalf("unexpected stale model ids: %v", policy.StaleModelEntryIDs)
	}
	if !reflect.DeepEqual(policy.EffectiveToolEntryIDs, []string{"tool-1"}) {
		t.Fatalf("unexpected effective tool ids: %v", policy.EffectiveToolEntryIDs)
	}
	if !reflect.DeepEqual(policy.StaleToolEntryIDs, []string{"tool-2"}) {
		t.Fatalf("unexpected stale tool ids: %v", policy.StaleToolEntryIDs)
	}
}

func TestResolveEffectiveAgentCapabilityPolicyHandlerReturnsEmptyEffectivePolicyWhenNoBindingExists(t *testing.T) {
	t.Parallel()

	repo := &fakeAgentCapabilityPolicyRepository{
		profiles: []controlplane.AgentProfile{{TenantID: "tenant-a", ID: "profile-1"}},
		models:   []*domaincap.ModelCatalogEntry{{TenantID: "tenant-a", EntryID: "model-1", Status: domaincap.CatalogStatusActive}},
		tools:    []*domaincap.ToolCatalogEntry{{TenantID: "tenant-a", EntryID: "tool-1", Status: domaincap.CatalogStatusActive}},
	}
	handler, err := NewResolveEffectiveAgentCapabilityPolicyHandler(repo, repo, repo, repo)
	if err != nil {
		t.Fatalf("handler init failed: %v", err)
	}

	policy, err := handler.Handle(context.Background(), "tenant-a", "profile-1")
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if len(policy.EffectiveModelEntryIDs) != 0 || len(policy.EffectiveToolEntryIDs) != 0 {
		t.Fatalf("expected empty effective policy, got %+v", policy)
	}
	if len(policy.StaleModelEntryIDs) != 0 || len(policy.StaleToolEntryIDs) != 0 {
		t.Fatalf("expected no stale bindings, got %+v", policy)
	}
}
