package controlplane

import (
	"context"
	"testing"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/memory"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
)

func TestServiceEmitsWorkspaceEventsForSessionAndTaskLifecycle(t *testing.T) {
	store := memory.NewControlPlaneStore()
	sink := &recordingWorkspaceEventSink{}
	svc := NewService(ServiceDeps{
		TenantCatalog:   store.TenantCatalog(),
		IAMDirectory:    store.IAMDirectory(),
		Sessions:        store.SessionRepository(),
		Tasks:           store.TaskRepository(),
		WorkspaceEvents: sink,
		Pipeline: shared.NewPipeline(shared.PipelineDeps{
			Policy: policy.StaticEvaluator(policy.DecisionAllow),
		}),
	})

	ctx := context.Background()
	session, err := svc.OpenSession(ctx, OpenSessionInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-001",
	})
	if err != nil {
		t.Fatalf("open session: %v", err)
	}

	task, err := svc.EnqueueTask(ctx, EnqueueTaskInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: session.ID,
		TaskID:    "task-001",
		TaskType:  "tool.call",
	})
	if err != nil {
		t.Fatalf("enqueue task: %v", err)
	}

	if _, err := svc.StartTask(ctx, AdvanceTaskInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		TaskID:   task.ID,
	}); err != nil {
		t.Fatalf("start task: %v", err)
	}
	if _, err := svc.CompleteTask(ctx, AdvanceTaskInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		TaskID:   task.ID,
		Output:   map[string]any{"ok": true},
	}); err != nil {
		t.Fatalf("complete task: %v", err)
	}

	got := sink.Events()
	if len(got) != 4 {
		t.Fatalf("expected 4 workspace events, got %d", len(got))
	}

	assertWorkspaceEvent(t, got[0], "runtime.session.opened", "tenant-a", session.ID, "")
	assertWorkspaceEvent(t, got[1], "runtime.task.enqueued", "tenant-a", session.ID, task.ID)
	assertWorkspaceEvent(t, got[2], "runtime.task.running", "tenant-a", session.ID, task.ID)
	assertWorkspaceEvent(t, got[3], "runtime.task.succeeded", "tenant-a", session.ID, task.ID)
}

func TestServiceEmitsFailureEventOnTaskFail(t *testing.T) {
	store := memory.NewControlPlaneStore()
	sink := &recordingWorkspaceEventSink{}
	svc := NewService(ServiceDeps{
		TenantCatalog:   store.TenantCatalog(),
		IAMDirectory:    store.IAMDirectory(),
		Sessions:        store.SessionRepository(),
		Tasks:           store.TaskRepository(),
		WorkspaceEvents: sink,
		Pipeline: shared.NewPipeline(shared.PipelineDeps{
			Policy: policy.StaticEvaluator(policy.DecisionAllow),
		}),
	})

	ctx := context.Background()
	session, err := svc.OpenSession(ctx, OpenSessionInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-001",
	})
	if err != nil {
		t.Fatalf("open session: %v", err)
	}
	task, err := svc.EnqueueTask(ctx, EnqueueTaskInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: session.ID,
		TaskID:    "task-001",
		TaskType:  "tool.call",
	})
	if err != nil {
		t.Fatalf("enqueue task: %v", err)
	}
	if _, err := svc.StartTask(ctx, AdvanceTaskInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		TaskID:   task.ID,
	}); err != nil {
		t.Fatalf("start task: %v", err)
	}
	if _, err := svc.FailTask(ctx, AdvanceTaskInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		TaskID:   task.ID,
		Reason:   "tool timeout",
	}); err != nil {
		t.Fatalf("fail task: %v", err)
	}

	got := sink.Events()
	if len(got) != 4 {
		t.Fatalf("expected 4 workspace events, got %d", len(got))
	}
	assertWorkspaceEvent(t, got[3], "runtime.task.failed", "tenant-a", session.ID, task.ID)
}

type recordingWorkspaceEventSink struct {
	events []platformruntime.WorkspaceEvent
}

func (r *recordingWorkspaceEventSink) Broadcast(evt platformruntime.WorkspaceEvent) error {
	r.events = append(r.events, evt)
	return nil
}

func (r *recordingWorkspaceEventSink) Events() []platformruntime.WorkspaceEvent {
	out := make([]platformruntime.WorkspaceEvent, len(r.events))
	copy(out, r.events)
	return out
}

func assertWorkspaceEvent(
	t *testing.T,
	evt platformruntime.WorkspaceEvent,
	eventType string,
	tenantID string,
	sessionID string,
	taskID string,
) {
	t.Helper()

	if evt.Type != eventType {
		t.Fatalf("expected event type %s, got %s", eventType, evt.Type)
	}
	if evt.TenantID != tenantID {
		t.Fatalf("expected event tenant %s, got %s", tenantID, evt.TenantID)
	}
	if evt.SessionID != sessionID {
		t.Fatalf("expected event session %s, got %s", sessionID, evt.SessionID)
	}
	if evt.TaskID != taskID {
		t.Fatalf("expected event task %s, got %s", taskID, evt.TaskID)
	}
}
