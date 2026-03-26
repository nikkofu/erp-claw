package capability

import (
	"context"
	"errors"
	"testing"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

func TestSharedCommandCapabilityAuthorizerAllowsEffectiveModelAndTools(t *testing.T) {
	t.Parallel()

	authorizer := newTestSharedCommandCapabilityAuthorizer(t, &fakeAgentCapabilityPolicyRepository{
		profiles: []controlplane.AgentProfile{{TenantID: "tenant-a", ID: "profile-1"}},
		getPolicy: &domaincap.AgentCapabilityPolicy{
			TenantID:             "tenant-a",
			AgentProfileID:       "profile-1",
			AllowedModelEntryIDs: []string{"model-1"},
			AllowedToolEntryIDs:  []string{"tool-1", "tool-2"},
		},
		models: []*domaincap.ModelCatalogEntry{
			{TenantID: "tenant-a", EntryID: "model-1", Status: domaincap.CatalogStatusActive},
		},
		tools: []*domaincap.ToolCatalogEntry{
			{TenantID: "tenant-a", EntryID: "tool-1", Status: domaincap.CatalogStatusActive},
			{TenantID: "tenant-a", EntryID: "tool-2", Status: domaincap.CatalogStatusActive},
		},
	})

	err := authorizer.AuthorizeCommandCapabilities(context.Background(), shared.Command{
		Name:     "agent.run",
		TenantID: "tenant-a",
		ActorID:  "user-a",
		Payload: map[string]any{
			"agent_profile_id": "profile-1",
			"model_entry_id":   "model-1",
			"tool_entry_ids":   []string{"tool-1", "tool-2"},
		},
	})
	if err != nil {
		t.Fatalf("expected authorization success, got %v", err)
	}
}

func TestSharedCommandCapabilityAuthorizerRejectsStaleOrUnboundRequests(t *testing.T) {
	t.Parallel()

	t.Run("stale model", func(t *testing.T) {
		t.Parallel()

		authorizer := newTestSharedCommandCapabilityAuthorizer(t, &fakeAgentCapabilityPolicyRepository{
			profiles: []controlplane.AgentProfile{{TenantID: "tenant-a", ID: "profile-1"}},
			getPolicy: &domaincap.AgentCapabilityPolicy{
				TenantID:             "tenant-a",
				AgentProfileID:       "profile-1",
				AllowedModelEntryIDs: []string{"model-2"},
				AllowedToolEntryIDs:  []string{"tool-1"},
			},
			models: []*domaincap.ModelCatalogEntry{
				{TenantID: "tenant-a", EntryID: "model-2", Status: domaincap.CatalogStatusInactive},
			},
			tools: []*domaincap.ToolCatalogEntry{
				{TenantID: "tenant-a", EntryID: "tool-1", Status: domaincap.CatalogStatusActive},
			},
		})

		err := authorizer.AuthorizeCommandCapabilities(context.Background(), shared.Command{
			Name:     "agent.run",
			TenantID: "tenant-a",
			ActorID:  "user-a",
			Payload: map[string]any{
				"agent_profile_id": "profile-1",
				"model_entry_id":   "model-2",
			},
		})
		if !errors.Is(err, shared.ErrCapabilityDenied) {
			t.Fatalf("expected capability denied error, got %v", err)
		}
	})

	t.Run("unbound tool", func(t *testing.T) {
		t.Parallel()

		authorizer := newTestSharedCommandCapabilityAuthorizer(t, &fakeAgentCapabilityPolicyRepository{
			profiles: []controlplane.AgentProfile{{TenantID: "tenant-a", ID: "profile-1"}},
			getPolicy: &domaincap.AgentCapabilityPolicy{
				TenantID:             "tenant-a",
				AgentProfileID:       "profile-1",
				AllowedModelEntryIDs: []string{"model-1"},
				AllowedToolEntryIDs:  []string{"tool-1"},
			},
			models: []*domaincap.ModelCatalogEntry{
				{TenantID: "tenant-a", EntryID: "model-1", Status: domaincap.CatalogStatusActive},
			},
			tools: []*domaincap.ToolCatalogEntry{
				{TenantID: "tenant-a", EntryID: "tool-1", Status: domaincap.CatalogStatusActive},
				{TenantID: "tenant-a", EntryID: "tool-2", Status: domaincap.CatalogStatusActive},
			},
		})

		err := authorizer.AuthorizeCommandCapabilities(context.Background(), shared.Command{
			Name:     "agent.run",
			TenantID: "tenant-a",
			ActorID:  "user-a",
			Payload: map[string]any{
				"agent_profile_id": "profile-1",
				"tool_entry_ids":   []string{"tool-2"},
			},
		})
		if !errors.Is(err, shared.ErrCapabilityDenied) {
			t.Fatalf("expected capability denied error, got %v", err)
		}
	})
}

func TestSharedCommandCapabilityAuthorizerSkipsCommandsWithoutExplicitCapabilityKeys(t *testing.T) {
	t.Parallel()

	authorizer := newTestSharedCommandCapabilityAuthorizer(t, &fakeAgentCapabilityPolicyRepository{})

	err := authorizer.AuthorizeCommandCapabilities(context.Background(), shared.Command{
		Name:     "agent.run",
		TenantID: "tenant-a",
		ActorID:  "user-a",
		Payload: map[string]any{
			"resource_id": "session-1",
		},
	})
	if err != nil {
		t.Fatalf("expected no-op authorization, got %v", err)
	}
}

func TestSharedCommandCapabilityAuthorizerRejectsMissingAgentProfileID(t *testing.T) {
	t.Parallel()

	authorizer := newTestSharedCommandCapabilityAuthorizer(t, &fakeAgentCapabilityPolicyRepository{})

	err := authorizer.AuthorizeCommandCapabilities(context.Background(), shared.Command{
		Name:     "agent.run",
		TenantID: "tenant-a",
		ActorID:  "user-a",
		Payload: map[string]any{
			"model_entry_id": "model-1",
		},
	})
	if !errors.Is(err, shared.ErrCapabilityDenied) {
		t.Fatalf("expected capability denied error, got %v", err)
	}
}

func newTestSharedCommandCapabilityAuthorizer(t *testing.T, repo *fakeAgentCapabilityPolicyRepository) SharedCommandCapabilityAuthorizer {
	t.Helper()

	resolver, err := NewResolveEffectiveAgentCapabilityPolicyHandler(repo, repo, repo, repo)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}
	return SharedCommandCapabilityAuthorizer{Resolver: resolver}
}
