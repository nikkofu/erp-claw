package ws

import (
	"context"
	"testing"
	"time"

	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
)

func TestGatewayRegistersWorkspaceChannel(t *testing.T) {
	gateway := NewWorkspaceGateway()
	if gateway == nil {
		t.Fatal("expected workspace gateway to be initialized")
	}
}

func TestGatewayAppendsRuntimeEvent(t *testing.T) {
	gateway := NewWorkspaceGateway()
	stream, err := gateway.RegisterChannel("session-a", 1)
	if err != nil {
		t.Fatalf("register channel: %v", err)
	}

	evt := platformruntime.WorkspaceEvent{
		Type:      platformruntime.WorkspaceEventTypeTaskStatusChanged,
		TenantID:  "tenant-a",
		SessionID: "session-a",
		TaskID:    "task-a",
		Payload:   map[string]any{"status": "running"},
	}
	if err := gateway.AppendWorkspaceEvent(context.Background(), evt); err != nil {
		t.Fatalf("append runtime event: %v", err)
	}

	select {
	case got := <-stream:
		if got.TaskID != "task-a" {
			t.Fatalf("expected task-a, got %s", got.TaskID)
		}
	default:
		t.Fatal("expected runtime event to be delivered")
	}
}

func TestGatewayAppendsRuntimeEventWithoutRegisteredChannel(t *testing.T) {
	gateway := NewWorkspaceGateway()

	evt := platformruntime.WorkspaceEvent{
		Type:      platformruntime.WorkspaceEventTypeTaskStatusChanged,
		TenantID:  "tenant-a",
		SessionID: "session-missing",
		TaskID:    "task-a",
		Payload:   map[string]any{"status": "running"},
	}
	if err := gateway.AppendWorkspaceEvent(context.Background(), evt); err != nil {
		t.Fatalf("append runtime event without live channel: %v", err)
	}

	events, err := gateway.ListWorkspaceEvents(context.Background(), "tenant-a", "session-missing")
	if err != nil {
		t.Fatalf("list workspace events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one stored event, got %d", len(events))
	}
	if events[0].TaskID != "task-a" {
		t.Fatalf("expected task-a, got %s", events[0].TaskID)
	}
}

func TestGatewayListsWorkspaceEventsByTenantAndSession(t *testing.T) {
	gateway := NewWorkspaceGateway()
	if _, err := gateway.RegisterChannel("session-a", 4); err != nil {
		t.Fatalf("register session-a: %v", err)
	}
	if _, err := gateway.RegisterChannel("session-b", 4); err != nil {
		t.Fatalf("register session-b: %v", err)
	}

	events := []platformruntime.WorkspaceEvent{
		{
			Type:       platformruntime.WorkspaceEventTypeTaskStatusChanged,
			TenantID:   "tenant-a",
			SessionID:  "session-a",
			TaskID:     "task-a",
			Payload:    map[string]any{"status": "running"},
			OccurredAt: mustTime(t, "2026-03-26T10:00:00Z"),
		},
		{
			Type:       platformruntime.WorkspaceEventTypeTaskStatusChanged,
			TenantID:   "tenant-a",
			SessionID:  "session-a",
			TaskID:     "task-a",
			Payload:    map[string]any{"status": "succeeded"},
			OccurredAt: mustTime(t, "2026-03-26T10:00:01Z"),
		},
		{
			Type:       platformruntime.WorkspaceEventTypeTaskStatusChanged,
			TenantID:   "tenant-b",
			SessionID:  "session-b",
			TaskID:     "task-b",
			Payload:    map[string]any{"status": "running"},
			OccurredAt: mustTime(t, "2026-03-26T10:00:02Z"),
		},
	}
	for _, evt := range events {
		if err := gateway.AppendWorkspaceEvent(context.Background(), evt); err != nil {
			t.Fatalf("append workspace event: %v", err)
		}
	}

	got, err := gateway.ListWorkspaceEvents(context.Background(), "tenant-a", "session-a")
	if err != nil {
		t.Fatalf("list workspace events: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
	if got[0].Payload.(map[string]any)["status"] != "running" {
		t.Fatalf("unexpected first event: %#v", got[0])
	}
	if got[1].Payload.(map[string]any)["status"] != "succeeded" {
		t.Fatalf("unexpected second event: %#v", got[1])
	}
}

func mustTime(t *testing.T, raw string) time.Time {
	t.Helper()

	got, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("parse time %q: %v", raw, err)
	}
	return got
}
