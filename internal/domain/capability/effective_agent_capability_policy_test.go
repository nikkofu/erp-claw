package capability

import (
	"reflect"
	"testing"
)

func TestNewEffectiveAgentCapabilityPolicyNormalizesLists(t *testing.T) {
	t.Parallel()

	policy, err := NewEffectiveAgentCapabilityPolicy(
		"tenant-a",
		"profile-1",
		[]string{"model-2", "model-1", "model-2"},
		[]string{"tool-2", "tool-1", "tool-2"},
		[]string{"model-stale", "model-stale"},
		[]string{"tool-stale", "tool-stale"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(policy.EffectiveModelEntryIDs, []string{"model-1", "model-2"}) {
		t.Fatalf("unexpected effective model ids: %v", policy.EffectiveModelEntryIDs)
	}
	if !reflect.DeepEqual(policy.EffectiveToolEntryIDs, []string{"tool-1", "tool-2"}) {
		t.Fatalf("unexpected effective tool ids: %v", policy.EffectiveToolEntryIDs)
	}
	if !reflect.DeepEqual(policy.StaleModelEntryIDs, []string{"model-stale"}) {
		t.Fatalf("unexpected stale model ids: %v", policy.StaleModelEntryIDs)
	}
	if !reflect.DeepEqual(policy.StaleToolEntryIDs, []string{"tool-stale"}) {
		t.Fatalf("unexpected stale tool ids: %v", policy.StaleToolEntryIDs)
	}
}
