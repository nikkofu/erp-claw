package integration

import (
	"context"
	"testing"
	"time"

	application "github.com/nikkofu/erp-claw/internal/application/agentruntime"
	domain "github.com/nikkofu/erp-claw/internal/domain/agentruntime"
	"github.com/nikkofu/erp-claw/internal/interfaces/ws"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
)

func TestAgentRuntimeFailureFlowPersistsFailedAndCanceledTasksAndBroadcastsEvents(t *testing.T) {
	sessions := newMemorySessionRepository()
	tasks := newMemoryTaskRepository()
	gateway := ws.NewWorkspaceGateway()

	events, err := gateway.RegisterChannel("workspace-failure-session", 8)
	if err != nil {
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

	if _, err := service.CreateSession(context.Background(), "2001", "workspace-failure-session", map[string]any{"source": "integration-test"}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	failedTask, err := service.CreateTask(context.Background(), "2001", "workspace-failure-session", "plan", map[string]any{"prompt": "fail me"})
	if err != nil {
		t.Fatalf("create failed task: %v", err)
	}
	if _, err := service.StartTask(context.Background(), "2001", failedTask.ID); err != nil {
		t.Fatalf("start failed task: %v", err)
	}
	if _, err := service.FailTask(context.Background(), "2001", failedTask.ID, map[string]any{"error": "tool crashed"}); err != nil {
		t.Fatalf("fail task: %v", err)
	}

	canceledTask, err := service.CreateTask(context.Background(), "2001", "workspace-failure-session", "plan", map[string]any{"prompt": "cancel me"})
	if err != nil {
		t.Fatalf("create canceled task: %v", err)
	}
	if _, err := service.CancelTask(context.Background(), "2001", canceledTask.ID, map[string]any{"reason": "user-request"}); err != nil {
		t.Fatalf("cancel task: %v", err)
	}

	persistedFailedTask, err := tasks.GetTaskByID(context.Background(), "2001", failedTask.ID)
	if err != nil {
		t.Fatalf("get failed task: %v", err)
	}
	if persistedFailedTask.Status != domain.TaskStatusFailed {
		t.Fatalf("expected failed status, got %q", persistedFailedTask.Status)
	}
	if persistedFailedTask.CompletedAt == nil {
		t.Fatal("expected failed task completed_at to be set")
	}
	if persistedFailedTask.Output["error"] != "tool crashed" {
		t.Fatalf("expected failed task output to include error, got %#v", persistedFailedTask.Output)
	}

	persistedCanceledTask, err := tasks.GetTaskByID(context.Background(), "2001", canceledTask.ID)
	if err != nil {
		t.Fatalf("get canceled task: %v", err)
	}
	if persistedCanceledTask.Status != domain.TaskStatusCanceled {
		t.Fatalf("expected canceled status, got %q", persistedCanceledTask.Status)
	}
	if persistedCanceledTask.CompletedAt == nil {
		t.Fatal("expected canceled task completed_at to be set")
	}
	if persistedCanceledTask.Output["reason"] != "user-request" {
		t.Fatalf("expected canceled task output to include reason, got %#v", persistedCanceledTask.Output)
	}

	got := collectWorkspaceEvents(t, events, 3, 250*time.Millisecond)
	if got[0].Type != platformruntime.WorkspaceEventTypeTaskStatusChanged {
		t.Fatalf("unexpected first event type: %s", got[0].Type)
	}
	if got[0].TaskID != failedTask.ID {
		t.Fatalf("unexpected first event task id: %s", got[0].TaskID)
	}
	payload0, ok := got[0].Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", got[0].Payload)
	}
	if payload0["status"] != string(domain.TaskStatusRunning) {
		t.Fatalf("expected running status payload, got %#v", payload0)
	}
	if got[1].Type != platformruntime.WorkspaceEventTypeTaskStatusChanged {
		t.Fatalf("unexpected second event type: %s", got[1].Type)
	}
	if got[1].TaskID != failedTask.ID {
		t.Fatalf("unexpected second event task id: %s", got[1].TaskID)
	}
	payload1, ok := got[1].Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", got[1].Payload)
	}
	if payload1["status"] != string(domain.TaskStatusFailed) {
		t.Fatalf("expected failed status payload, got %#v", payload1)
	}
	if got[2].Type != platformruntime.WorkspaceEventTypeTaskStatusChanged {
		t.Fatalf("unexpected third event type: %s", got[2].Type)
	}
	if got[2].TaskID != canceledTask.ID {
		t.Fatalf("unexpected third event task id: %s", got[2].TaskID)
	}
	payload2, ok := got[2].Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", got[2].Payload)
	}
	if payload2["status"] != string(domain.TaskStatusCanceled) {
		t.Fatalf("expected canceled status payload, got %#v", payload2)
	}
}
