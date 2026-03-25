package bootstrap

import "testing"

func TestRuntimeRoleString(t *testing.T) {
	if APIServerRole.String() != "api-server" {
		t.Fatalf("expected api-server, got %q", APIServerRole.String())
	}
}
