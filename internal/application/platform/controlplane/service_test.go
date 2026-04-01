package controlplane

import (
	"context"
	"errors"
	"testing"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/memory"
	"github.com/nikkofu/erp-claw/internal/platform/audit"
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

func TestAuditRecordsContainApprovalLifecycleEvidence(t *testing.T) {
	store := memory.NewControlPlaneStore()
	auditRecorder := audit.NewInMemoryRecorder()
	svc := NewService(ServiceDeps{
		TenantCatalog:   store.TenantCatalog(),
		IAMDirectory:    store.IAMDirectory(),
		Sessions:        store.SessionRepository(),
		Tasks:           store.TaskRepository(),
		WorkspaceEvents: &recordingWorkspaceEventSink{},
		Pipeline: shared.NewPipeline(shared.PipelineDeps{
			Policy: policy.StaticEvaluator(policy.DecisionAllow),
			Audit:  auditRecorder,
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
	if _, err := svc.EnqueueTask(ctx, EnqueueTaskInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: session.ID,
		TaskID:    "task-001",
		TaskType:  "tool.call",
	}); err != nil {
		t.Fatalf("enqueue task: %v", err)
	}

	for _, action := range []struct {
		name string
		run  func() error
	}{
		{
			name: "runtime.tasks.pause",
			run: func() error {
				_, err := svc.PauseTask(ctx, AdvanceTaskInput{TenantID: "tenant-a", ActorID: "actor-a", TaskID: "task-001"})
				return err
			},
		},
		{
			name: "runtime.tasks.resume",
			run: func() error {
				_, err := svc.ResumeTask(ctx, AdvanceTaskInput{TenantID: "tenant-a", ActorID: "actor-a", TaskID: "task-001"})
				return err
			},
		},
		{
			name: "runtime.tasks.handoff",
			run: func() error {
				_, err := svc.HandoffTask(ctx, AdvanceTaskInput{TenantID: "tenant-a", ActorID: "actor-a", TaskID: "task-001"})
				return err
			},
		},
	} {
		if err := action.run(); err == nil {
			t.Fatalf("expected %s to be rejected", action.name)
		}
	}

	records, err := auditRecorder.List(ctx, audit.Query{TenantID: "tenant-a", Limit: 20})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}

	required := map[string]bool{
		"runtime.tasks.pause":   false,
		"runtime.tasks.resume":  false,
		"runtime.tasks.handoff": false,
	}
	for _, record := range records {
		if _, ok := required[record.CommandName]; !ok {
			continue
		}
		required[record.CommandName] = true
		if record.Decision != policy.DecisionAllow {
			t.Fatalf("expected decision allow for %s, got %s", record.CommandName, record.Decision)
		}
		if record.Outcome != "failed" {
			t.Fatalf("expected outcome failed for %s, got %s", record.CommandName, record.Outcome)
		}
		if record.CorrelationID == "" {
			t.Fatalf("expected correlation id for %s", record.CommandName)
		}
		if record.ResourceType == "" || record.ResourceID == "" {
			t.Fatalf("expected resource evidence for %s", record.CommandName)
		}
	}

	for command, seen := range required {
		if !seen {
			t.Fatalf("expected audit record for %s", command)
		}
	}
}

func TestEventPublishFirstSuccessMarksDelivered(t *testing.T) {
	store := memory.NewControlPlaneStore()
	sink := &recordingWorkspaceEventSink{}
	svc := NewService(ServiceDeps{
		TenantCatalog:   store.TenantCatalog(),
		IAMDirectory:    store.IAMDirectory(),
		Sessions:        store.SessionRepository(),
		Tasks:           store.TaskRepository(),
		Deliveries:      store.DeliveryRepository(),
		WorkspaceEvents: sink,
		Pipeline: shared.NewPipeline(shared.PipelineDeps{
			Policy: policy.StaticEvaluator(policy.DecisionAllow),
		}),
	})

	event := platformruntime.WorkspaceEvent{
		Type:      "runtime.task.running",
		TenantID:  "tenant-e2",
		SessionID: "sess-e2",
		TaskID:    "task-e2",
	}

	if err := svc.emitWorkspaceEvent(event); err != nil {
		t.Fatalf("first publish: %v", err)
	}

	delivered, err := svc.ListDeliveries(context.Background(), ListDeliveriesInput{
		TenantID: "tenant-e2",
		Status:   platformruntime.DeliveryStatusDelivered,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("list delivered deliveries: %v", err)
	}
	if len(delivered.Items) != 1 {
		t.Fatalf("expected 1 delivered delivery after first success, got %d", len(delivered.Items))
	}

	pending, err := svc.ListDeliveries(context.Background(), ListDeliveriesInput{
		TenantID: "tenant-e2",
		Status:   platformruntime.DeliveryStatusPending,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("list pending deliveries: %v", err)
	}
	if len(pending.Items) != 0 {
		t.Fatalf("expected 0 pending deliveries after first success, got %d", len(pending.Items))
	}
}

func TestEventPublishFailureTransitionsToRecoveredOnRetry(t *testing.T) {
	store := memory.NewControlPlaneStore()
	sink := &recordingWorkspaceEventSink{fail: true}
	svc := NewService(ServiceDeps{
		TenantCatalog:   store.TenantCatalog(),
		IAMDirectory:    store.IAMDirectory(),
		Sessions:        store.SessionRepository(),
		Tasks:           store.TaskRepository(),
		Deliveries:      store.DeliveryRepository(),
		WorkspaceEvents: sink,
		Pipeline: shared.NewPipeline(shared.PipelineDeps{
			Policy: policy.StaticEvaluator(policy.DecisionAllow),
		}),
	})

	event := platformruntime.WorkspaceEvent{
		Type:      "runtime.task.running",
		TenantID:  "tenant-e2",
		SessionID: "sess-e2",
		TaskID:    "task-e2",
	}

	if err := svc.emitWorkspaceEvent(event); err == nil {
		t.Fatal("expected first publish to fail")
	}

	sink.fail = false
	if err := svc.emitWorkspaceEvent(event); err != nil {
		t.Fatalf("retry publish: %v", err)
	}

	recovered, err := svc.ListDeliveries(context.Background(), ListDeliveriesInput{
		TenantID: "tenant-e2",
		Status:   platformruntime.DeliveryStatusRecovered,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("list recovered deliveries: %v", err)
	}
	if len(recovered.Items) != 1 {
		t.Fatalf("expected 1 recovered delivery, got %d", len(recovered.Items))
	}
	if recovered.Items[0].AttemptCount != 2 {
		t.Fatalf("expected attempt_count 2 after retry, got %d", recovered.Items[0].AttemptCount)
	}
}

func TestGovernanceTaskCommandsRejectWithoutSuccessNoop(t *testing.T) {
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

	if _, err := svc.PauseTask(ctx, AdvanceTaskInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		TaskID:   task.ID,
	}); err == nil {
		t.Fatal("expected pause task to be rejected")
	}
	if _, err := svc.ResumeTask(ctx, AdvanceTaskInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		TaskID:   task.ID,
	}); err == nil {
		t.Fatal("expected resume task to be rejected")
	}
	if _, err := svc.HandoffTask(ctx, AdvanceTaskInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		TaskID:   task.ID,
	}); err == nil {
		t.Fatal("expected handoff task to be rejected")
	}

	current, err := svc.GetTask(ctx, GetTaskInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		TaskID:   task.ID,
	})
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if current.Status != platformruntime.TaskStatusPending {
		t.Fatalf("expected task to stay pending, got %s", current.Status)
	}

	got := sink.Events()
	if len(got) != 2 {
		t.Fatalf("expected no extra events from governance commands, got %d events", len(got))
	}
	assertWorkspaceEvent(t, got[0], "runtime.session.opened", "tenant-a", session.ID, "")
	assertWorkspaceEvent(t, got[1], "runtime.task.enqueued", "tenant-a", session.ID, task.ID)
}

type recordingWorkspaceEventSink struct {
	events []platformruntime.WorkspaceEvent
	fail   bool
}

func (r *recordingWorkspaceEventSink) Broadcast(evt platformruntime.WorkspaceEvent) error {
	if r.fail {
		return errors.New("workspace gateway unavailable")
	}
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
