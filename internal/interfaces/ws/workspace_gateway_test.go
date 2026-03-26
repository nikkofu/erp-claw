package ws

import (
	"testing"

	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
)

func TestGatewayRegistersWorkspaceChannel(t *testing.T) {
	gateway := NewWorkspaceGateway()
	if gateway == nil {
		t.Fatal("expected workspace gateway to be initialized")
	}
}

func TestGatewayBroadcastUnknownSessionDoesNotFail(t *testing.T) {
	gateway := NewWorkspaceGateway()
	err := gateway.Broadcast(platformruntime.WorkspaceEvent{
		Type:      "runtime.session.opened",
		TenantID:  "tenant-a",
		SessionID: "sess-missing",
	})
	if err != nil {
		t.Fatalf("expected unknown session broadcast to be dropped, got %v", err)
	}
}

func TestGatewayBroadcastDeliversToRegisteredSession(t *testing.T) {
	gateway := NewWorkspaceGateway()
	ch, err := gateway.RegisterChannel("sess-001", 1)
	if err != nil {
		t.Fatalf("register channel: %v", err)
	}
	defer gateway.UnregisterChannel("sess-001")

	evt := platformruntime.WorkspaceEvent{
		Type:      "runtime.task.succeeded",
		TenantID:  "tenant-a",
		SessionID: "sess-001",
		TaskID:    "task-001",
	}
	if err := gateway.Broadcast(evt); err != nil {
		t.Fatalf("broadcast event: %v", err)
	}

	got := <-ch
	if got.Type != evt.Type {
		t.Fatalf("expected event type %s, got %s", evt.Type, got.Type)
	}
	if got.TaskID != evt.TaskID {
		t.Fatalf("expected task %s, got %s", evt.TaskID, got.TaskID)
	}
}
