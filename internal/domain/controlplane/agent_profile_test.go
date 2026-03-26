package controlplane

import "testing"

func TestNewAgentProfileRequiresModel(t *testing.T) {
	_, err := NewAgentProfile("tenant-a", "planner", "")
	if err == nil {
		t.Fatal("expected empty model to fail")
	}
}
