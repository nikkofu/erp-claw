package controlplane

import (
	"context"
	"errors"
	"testing"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/memory"
	"github.com/nikkofu/erp-claw/internal/platform/iam"
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

func TestServiceEmitsCanceledEventOnTaskCancel(t *testing.T) {
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

	canceled, err := svc.CancelTask(ctx, AdvanceTaskInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		TaskID:   task.ID,
		Reason:   "manual cancel",
	})
	if err != nil {
		t.Fatalf("cancel task: %v", err)
	}
	if canceled.Status != platformruntime.TaskStatusCanceled {
		t.Fatalf("expected canceled task, got %s", canceled.Status)
	}
	if canceled.FailureReason != "manual cancel" {
		t.Fatalf("expected cancel reason manual cancel, got %q", canceled.FailureReason)
	}

	got := sink.Events()
	if len(got) != 3 {
		t.Fatalf("expected 3 workspace events, got %d", len(got))
	}
	assertWorkspaceEvent(t, got[2], "runtime.task.canceled", "tenant-a", session.ID, task.ID)
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

func TestServiceCloseSessionRejectsWhenSessionHasActiveTasks(t *testing.T) {
	cases := []struct {
		name          string
		prepareActive func(t *testing.T, svc *Service, taskID string)
	}{
		{
			name: "pending task",
			prepareActive: func(t *testing.T, svc *Service, taskID string) {
				t.Helper()
			},
		},
		{
			name: "running task",
			prepareActive: func(t *testing.T, svc *Service, taskID string) {
				t.Helper()
				if _, err := svc.StartTask(context.Background(), AdvanceTaskInput{
					TenantID: "tenant-a",
					ActorID:  "actor-a",
					TaskID:   taskID,
				}); err != nil {
					t.Fatalf("start task: %v", err)
				}
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
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
			tc.prepareActive(t, svc, task.ID)

			_, err = svc.CloseSession(ctx, CloseSessionInput{
				TenantID:  "tenant-a",
				ActorID:   "actor-a",
				SessionID: session.ID,
			})
			if !errors.Is(err, platformruntime.ErrInvalidSessionTransition) {
				t.Fatalf("expected ErrInvalidSessionTransition, got %v", err)
			}

			current, err := svc.GetSession(ctx, GetSessionInput{
				TenantID:  "tenant-a",
				ActorID:   "actor-a",
				SessionID: session.ID,
			})
			if err != nil {
				t.Fatalf("get session: %v", err)
			}
			if current.Status != platformruntime.SessionStatusOpen {
				t.Fatalf("expected session to remain open, got %s", current.Status)
			}

			for _, evt := range sink.Events() {
				if evt.Type == "runtime.session.closed" {
					t.Fatalf("unexpected session closed event in %s case", tc.name)
				}
			}
		})
	}
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

func TestServiceListTasksReturnsTenantScopedAndFilteredTasks(t *testing.T) {
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
	mustOpen := func(tenantID, actorID, sessionID string) {
		t.Helper()
		if _, err := svc.OpenSession(ctx, OpenSessionInput{
			TenantID:  tenantID,
			ActorID:   actorID,
			SessionID: sessionID,
		}); err != nil {
			t.Fatalf("open session %s: %v", sessionID, err)
		}
	}
	mustOpen("tenant-a", "actor-a", "sess-001")
	mustOpen("tenant-a", "actor-a", "sess-002")
	mustOpen("tenant-b", "actor-b", "sess-003")

	taskPending, err := svc.EnqueueTask(ctx, EnqueueTaskInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-001",
		TaskID:    "task-pending-001",
		TaskType:  "tool.call",
	})
	if err != nil {
		t.Fatalf("enqueue task-pending-001: %v", err)
	}
	taskRunning, err := svc.EnqueueTask(ctx, EnqueueTaskInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-002",
		TaskID:    "task-running-001",
		TaskType:  "tool.call",
	})
	if err != nil {
		t.Fatalf("enqueue task-running-001: %v", err)
	}
	if _, err := svc.StartTask(ctx, AdvanceTaskInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		TaskID:   taskRunning.ID,
	}); err != nil {
		t.Fatalf("start task-running-001: %v", err)
	}
	if _, err := svc.EnqueueTask(ctx, EnqueueTaskInput{
		TenantID:  "tenant-b",
		ActorID:   "actor-b",
		SessionID: "sess-003",
		TaskID:    "task-other-tenant",
		TaskType:  "tool.call",
	}); err != nil {
		t.Fatalf("enqueue task-other-tenant: %v", err)
	}

	allTenantTasks, err := svc.ListTasks(ctx, ListTasksInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
	})
	if err != nil {
		t.Fatalf("list tenant tasks: %v", err)
	}
	if len(allTenantTasks) != 2 {
		t.Fatalf("expected 2 tenant-a tasks, got %d", len(allTenantTasks))
	}

	sessionFiltered, err := svc.ListTasks(ctx, ListTasksInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-001",
	})
	if err != nil {
		t.Fatalf("list session tasks: %v", err)
	}
	if len(sessionFiltered) != 1 || sessionFiltered[0].ID != taskPending.ID {
		t.Fatalf("expected only %s, got %#v", taskPending.ID, sessionFiltered)
	}

	statusFiltered, err := svc.ListTasks(ctx, ListTasksInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Status:   platformruntime.TaskStatusRunning,
	})
	if err != nil {
		t.Fatalf("list running tasks: %v", err)
	}
	if len(statusFiltered) != 1 || statusFiltered[0].ID != taskRunning.ID {
		t.Fatalf("expected only %s, got %#v", taskRunning.ID, statusFiltered)
	}
}

func TestServiceListSessionsSupportsStatusFilter(t *testing.T) {
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
		SessionID: "sess-open",
	}); err != nil {
		t.Fatalf("open sess-open: %v", err)
	}
	if _, err := svc.OpenSession(ctx, OpenSessionInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-closed",
	}); err != nil {
		t.Fatalf("open sess-closed: %v", err)
	}
	if _, err := svc.CloseSession(ctx, CloseSessionInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-closed",
	}); err != nil {
		t.Fatalf("close sess-closed: %v", err)
	}

	closedSessions, err := svc.ListSessions(ctx, ListSessionsInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Status:   platformruntime.SessionStatusClosed,
	})
	if err != nil {
		t.Fatalf("list closed sessions: %v", err)
	}
	if len(closedSessions) != 1 {
		t.Fatalf("expected 1 closed session, got %d", len(closedSessions))
	}
	if closedSessions[0].ID != "sess-closed" {
		t.Fatalf("expected sess-closed, got %s", closedSessions[0].ID)
	}
}

func TestServiceListTasksSupportsOffsetAndLimit(t *testing.T) {
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
		SessionID: "sess-page-001",
	}); err != nil {
		t.Fatalf("open session: %v", err)
	}

	for _, taskID := range []string{"task-page-001", "task-page-002", "task-page-003"} {
		if _, err := svc.EnqueueTask(ctx, EnqueueTaskInput{
			TenantID:  "tenant-a",
			ActorID:   "actor-a",
			SessionID: "sess-page-001",
			TaskID:    taskID,
			TaskType:  "tool.call",
		}); err != nil {
			t.Fatalf("enqueue %s: %v", taskID, err)
		}
	}

	paged, err := svc.ListTasks(ctx, ListTasksInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Offset:   1,
		Limit:    1,
	})
	if err != nil {
		t.Fatalf("list paged tasks: %v", err)
	}
	if len(paged) != 1 {
		t.Fatalf("expected 1 paged task, got %d", len(paged))
	}
	if paged[0].ID != "task-page-002" {
		t.Fatalf("expected task-page-002, got %s", paged[0].ID)
	}
}

func TestServiceListSessionsSupportsOffsetAndLimit(t *testing.T) {
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
	for _, sessionID := range []string{"sess-page-001", "sess-page-002", "sess-page-003"} {
		if _, err := svc.OpenSession(ctx, OpenSessionInput{
			TenantID:  "tenant-a",
			ActorID:   "actor-a",
			SessionID: sessionID,
		}); err != nil {
			t.Fatalf("open %s: %v", sessionID, err)
		}
	}

	paged, err := svc.ListSessions(ctx, ListSessionsInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Offset:   1,
		Limit:    1,
	})
	if err != nil {
		t.Fatalf("list paged sessions: %v", err)
	}
	if len(paged) != 1 {
		t.Fatalf("expected 1 paged session, got %d", len(paged))
	}
	if paged[0].ID != "sess-page-002" {
		t.Fatalf("expected sess-page-002, got %s", paged[0].ID)
	}
}

func TestServiceListSessionTasksSupportsStatusAndPagination(t *testing.T) {
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
		SessionID: "sess-filter-001",
	}); err != nil {
		t.Fatalf("open session: %v", err)
	}
	for _, taskID := range []string{"task-001", "task-002", "task-003"} {
		if _, err := svc.EnqueueTask(ctx, EnqueueTaskInput{
			TenantID:  "tenant-a",
			ActorID:   "actor-a",
			SessionID: "sess-filter-001",
			TaskID:    taskID,
			TaskType:  "tool.call",
		}); err != nil {
			t.Fatalf("enqueue %s: %v", taskID, err)
		}
	}
	if _, err := svc.StartTask(ctx, AdvanceTaskInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		TaskID:   "task-002",
	}); err != nil {
		t.Fatalf("start task-002: %v", err)
	}

	running, err := svc.ListSessionTasks(ctx, ListSessionTasksInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-filter-001",
		Status:    platformruntime.TaskStatusRunning,
	})
	if err != nil {
		t.Fatalf("list running session tasks: %v", err)
	}
	if len(running) != 1 || running[0].ID != "task-002" {
		t.Fatalf("expected only task-002, got %#v", running)
	}

	paged, err := svc.ListSessionTasks(ctx, ListSessionTasksInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-filter-001",
		Offset:    1,
		Limit:     1,
	})
	if err != nil {
		t.Fatalf("list paged session tasks: %v", err)
	}
	if len(paged) != 1 || paged[0].ID != "task-002" {
		t.Fatalf("expected only task-002 in paged result, got %#v", paged)
	}
}

func TestServiceDeleteActorRemovesActorFromDirectory(t *testing.T) {
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
	if _, err := svc.UpsertActor(ctx, UpsertActorInput{
		OperatorTenantID: "tenant-admin",
		OperatorActorID:  "actor-admin",
		TenantID:         "tenant-a",
		ActorID:          "actor-delete-me",
		Roles:            []string{"viewer"},
	}); err != nil {
		t.Fatalf("upsert actor: %v", err)
	}

	if err := svc.DeleteActor(ctx, DeleteActorInput{
		OperatorTenantID: "tenant-admin",
		OperatorActorID:  "actor-admin",
		TenantID:         "tenant-a",
		ActorID:          "actor-delete-me",
	}); err != nil {
		t.Fatalf("delete actor: %v", err)
	}

	_, err := svc.GetActor(ctx, GetActorInput{
		OperatorTenantID: "tenant-admin",
		OperatorActorID:  "actor-admin",
		TenantID:         "tenant-a",
		ActorID:          "actor-delete-me",
	})
	if !errors.Is(err, iam.ErrActorNotFound) {
		t.Fatalf("expected ErrActorNotFound, got %v", err)
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
