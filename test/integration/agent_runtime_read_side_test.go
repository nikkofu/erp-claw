package integration

import (
	"context"
	"testing"
	"time"

	application "github.com/nikkofu/erp-claw/internal/application/agentruntime"
	"github.com/nikkofu/erp-claw/internal/interfaces/ws"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
)

func TestAgentRuntimeReadSideListsSessionsTasksAndReplaysEvents(t *testing.T) {
	sessions := newMemorySessionRepository()
	tasks := newMemoryTaskRepository()
	gateway := ws.NewWorkspaceGateway()
	if _, err := gateway.RegisterChannel("workspace-read-side", 8); err != nil {
		t.Fatalf("register workspace channel: %v", err)
	}

	service, err := application.NewService(application.ServiceDeps{
		Sessions: sessions,
		Tasks:    tasks,
		Events:   gateway,
	})
	if err != nil {
		t.Fatalf("new runtime service: %v", err)
	}

	session, err := service.CreateSession(context.Background(), "3001", "workspace-read-side", map[string]any{"source": "integration-test"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	task, err := service.CreateTask(context.Background(), "3001", "workspace-read-side", "plan", map[string]any{"prompt": "list me"})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if _, err := service.StartTask(context.Background(), "3001", task.ID); err != nil {
		t.Fatalf("start task: %v", err)
	}
	if _, err := service.CompleteTask(context.Background(), "3001", task.ID, map[string]any{"result": "ok"}); err != nil {
		t.Fatalf("complete task: %v", err)
	}

	listSessions := application.ListSessionsHandler{Sessions: sessions}
	gotSessions, err := listSessions.Handle(context.Background(), application.ListSessions{TenantID: "3001"})
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(gotSessions) != 1 || gotSessions[0].SessionKey != session.SessionKey {
		t.Fatalf("unexpected sessions: %+v", gotSessions)
	}

	listTasks := application.ListTasksHandler{Tasks: tasks}
	gotTasks, err := listTasks.Handle(context.Background(), application.ListTasks{TenantID: "3001", SessionID: "workspace-read-side"})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(gotTasks) != 1 || gotTasks[0].ID != task.ID {
		t.Fatalf("unexpected tasks: %+v", gotTasks)
	}

	replay := application.ReplayWorkspaceEventsHandler{Events: gateway}
	gotEvents, err := replay.Handle(context.Background(), application.ReplayWorkspaceEvents{TenantID: "3001", SessionID: "workspace-read-side"})
	if err != nil {
		t.Fatalf("replay events: %v", err)
	}
	if len(gotEvents) != 2 {
		t.Fatalf("expected 2 replayed events, got %d", len(gotEvents))
	}
	if gotEvents[0].Type != platformruntime.WorkspaceEventTypeTaskStatusChanged {
		t.Fatalf("unexpected first event type: %s", gotEvents[0].Type)
	}
	if gotEvents[1].Payload.(map[string]any)["status"] != "succeeded" {
		t.Fatalf("unexpected final event payload: %#v", gotEvents[1].Payload)
	}
	if gotEvents[0].OccurredAt.After(gotEvents[1].OccurredAt.Add(time.Nanosecond)) {
		t.Fatalf("expected replay order to be stable: %+v", gotEvents)
	}
}
