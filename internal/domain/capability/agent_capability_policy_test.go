package capability

import (
	"reflect"
	"testing"
)

func TestNewAgentCapabilityPolicyValidatesTenantAndProfile(t *testing.T) {
	t.Parallel()

	_, err := NewAgentCapabilityPolicy("", "profile-1", nil, nil)
	if err != ErrTenantIDRequired {
		t.Fatalf("expected tenant ID error, got %v", err)
	}

	_, err = NewAgentCapabilityPolicy("tenant-a", "", nil, nil)
	if err != ErrAgentProfileIDRequired {
		t.Fatalf("expected agent profile ID error, got %v", err)
	}
}

func TestNewAgentCapabilityPolicyNormalizesEntryIDs(t *testing.T) {
	t.Parallel()

	policy, err := NewAgentCapabilityPolicy(
		"tenant-a",
		"profile-1",
		[]string{" model-2 ", "model-1", "model-2", ""},
		[]string{"tool-2", "tool-1", " tool-2 ", ""},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if policy.TenantID != "tenant-a" || policy.AgentProfileID != "profile-1" {
		t.Fatalf("unexpected policy identity: %+v", policy)
	}

	wantModels := []string{"model-1", "model-2"}
	if !reflect.DeepEqual(policy.AllowedModelEntryIDs, wantModels) {
		t.Fatalf("unexpected model entry ids: got %v want %v", policy.AllowedModelEntryIDs, wantModels)
	}

	wantTools := []string{"tool-1", "tool-2"}
	if !reflect.DeepEqual(policy.AllowedToolEntryIDs, wantTools) {
		t.Fatalf("unexpected tool entry ids: got %v want %v", policy.AllowedToolEntryIDs, wantTools)
	}
}
