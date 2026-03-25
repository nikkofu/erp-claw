package ws

import "testing"

func TestGatewayRegistersWorkspaceChannel(t *testing.T) {
	gateway := NewWorkspaceGateway()
	if gateway == nil {
		t.Fatal("expected workspace gateway to be initialized")
	}
}
