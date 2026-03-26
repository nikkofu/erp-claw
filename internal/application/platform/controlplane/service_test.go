package controlplane

import (
	"context"
	"errors"
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

func TestServiceCloseSessionEmitsWorkspaceEventAndRejectsEnqueueOnClosedSession(t *testing.T) {
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

	closed, err := svc.CloseSession(ctx, CloseSessionInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: session.ID,
	})
	if err != nil {
		t.Fatalf("close session: %v", err)
	}
	if closed.Status != platformruntime.SessionStatusClosed {
		t.Fatalf("expected closed session, got %s", closed.Status)
	}
	if closed.EndedAt.IsZero() {
		t.Fatal("expected closed session ended_at")
	}

	_, err = svc.EnqueueTask(ctx, EnqueueTaskInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: session.ID,
		TaskID:    "task-001",
		TaskType:  "tool.call",
	})
	if !errors.Is(err, platformruntime.ErrInvalidSessionTransition) {
		t.Fatalf("expected ErrInvalidSessionTransition, got %v", err)
	}

	got := sink.Events()
	if len(got) != 2 {
		t.Fatalf("expected 2 workspace events, got %d", len(got))
	}
	assertWorkspaceEvent(t, got[0], "runtime.session.opened", "tenant-a", session.ID, "")
	assertWorkspaceEvent(t, got[1], "runtime.session.closed", "tenant-a", session.ID, "")
}

func TestServiceListSessionsReturnsTenantScopedSessions(t *testing.T) {
	store := memory.NewControlPlaneStore()
	svc := NewService(ServiceDeps{
		TenantCatalog: store.TenantCatalog(),
		IAMDirectory:  store.IAMDirectory(),
		Sessions:      store.SessionRepository(),
		Tasks:         store.TaskRepository(),
		Pipeline: shared.NewPipeline(shared.PipelineDeps{
			Policy: policy.StaticEvaluator(policy.DecisionAllow),
		}),
	})

	ctx := context.Background()
	if _, err := svc.OpenSession(ctx, OpenSessionInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-002",
	}); err != nil {
		t.Fatalf("open tenant-a sess-002: %v", err)
	}
	if _, err := svc.OpenSession(ctx, OpenSessionInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-001",
	}); err != nil {
		t.Fatalf("open tenant-a sess-001: %v", err)
	}
	if _, err := svc.OpenSession(ctx, OpenSessionInput{
		TenantID:  "tenant-b",
		ActorID:   "actor-b",
		SessionID: "sess-003",
	}); err != nil {
		t.Fatalf("open tenant-b sess-003: %v", err)
	}

	got, err := svc.ListSessions(ctx, ListSessionsInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
	})
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 tenant sessions, got %d", len(got))
	}

	wantIDs := map[string]struct{}{
		"sess-001": {},
		"sess-002": {},
	}
	for _, session := range got {
		if session.TenantID != "tenant-a" {
			t.Fatalf("expected tenant-a session, got tenant %s", session.TenantID)
		}
		if _, ok := wantIDs[session.ID]; !ok {
			t.Fatalf("unexpected session id %s", session.ID)
		}
		delete(wantIDs, session.ID)
	}
	if len(wantIDs) != 0 {
		t.Fatalf("missing expected sessions: %#v", wantIDs)
	}
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
