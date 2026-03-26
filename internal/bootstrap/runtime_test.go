package bootstrap

import "testing"

func TestRuntimeRoleString(t *testing.T) {
	if APIServerRole.String() != "api-server" {
		t.Fatalf("expected api-server, got %q", APIServerRole.String())
	}
}

func TestNewContainerProvidesSupplyChainService(t *testing.T) {
	container := NewContainer(DefaultConfig())
	if container.SupplyChain == nil {
		t.Fatal("expected supply-chain service to be wired")
	}
}
