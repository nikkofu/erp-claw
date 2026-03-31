package controlplane

import (
	"context"
	"testing"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/memory"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
)

func TestServiceTaskAndSessionQueriesEnforceTenantIsolationAndPaging(t *testing.T) {
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

	for _, sessionID := range []string{"sess-a-001", "sess-a-002"} {
		if _, err := svc.OpenSession(ctx, OpenSessionInput{
			TenantID:  "tenant-a",
			ActorID:   "actor-a",
			SessionID: sessionID,
		}); err != nil {
			t.Fatalf("open session %s: %v", sessionID, err)
		}
	}
	if _, err := svc.OpenSession(ctx, OpenSessionInput{
		TenantID:  "tenant-b",
		ActorID:   "actor-b",
		SessionID: "sess-b-001",
	}); err != nil {
		t.Fatalf("open session tenant-b: %v", err)
	}

	for _, taskID := range []string{"task-a-001", "task-a-002", "task-a-003"} {
		if _, err := svc.EnqueueTask(ctx, EnqueueTaskInput{
			TenantID:  "tenant-a",
			ActorID:   "actor-a",
			SessionID: "sess-a-001",
			TaskID:    taskID,
			TaskType:  "tool.call",
		}); err != nil {
			t.Fatalf("enqueue task %s: %v", taskID, err)
		}
		if _, err := svc.StartTask(ctx, AdvanceTaskInput{
			TenantID: "tenant-a",
			ActorID:  "actor-a",
			TaskID:   taskID,
		}); err != nil {
			t.Fatalf("start task %s: %v", taskID, err)
		}
	}

	if _, err := svc.EnqueueTask(ctx, EnqueueTaskInput{
		TenantID:  "tenant-b",
		ActorID:   "actor-b",
		SessionID: "sess-b-001",
		TaskID:    "task-b-001",
		TaskType:  "tool.call",
	}); err != nil {
		t.Fatalf("enqueue tenant-b task: %v", err)
	}
	if _, err := svc.StartTask(ctx, AdvanceTaskInput{
		TenantID: "tenant-b",
		ActorID:  "actor-b",
		TaskID:   "task-b-001",
	}); err != nil {
		t.Fatalf("start tenant-b task: %v", err)
	}

	taskPage1, err := svc.ListTasks(ctx, ListTasksInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-a-001",
		Status:    platformruntime.TaskStatusRunning,
		Limit:     2,
	})
	if err != nil {
		t.Fatalf("list tasks page1: %v", err)
	}
	if len(taskPage1.Items) != 2 {
		t.Fatalf("expected 2 tasks on first page, got %d", len(taskPage1.Items))
	}
	if taskPage1.NextCursor == "" {
		t.Fatal("expected next cursor on first page")
	}
	if taskPage1.AsOf.IsZero() {
		t.Fatal("expected non-zero as_of on first page")
	}
	for _, item := range taskPage1.Items {
		if item.TenantID != "tenant-a" {
			t.Fatalf("expected tenant-a only, got %s", item.TenantID)
		}
		if item.SessionID != "sess-a-001" {
			t.Fatalf("expected session sess-a-001 only, got %s", item.SessionID)
		}
		if item.Status != platformruntime.TaskStatusRunning {
			t.Fatalf("expected running status only, got %s", item.Status)
		}
	}

	taskPage2, err := svc.ListTasks(ctx, ListTasksInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-a-001",
		Status:    platformruntime.TaskStatusRunning,
		Limit:     2,
		Cursor:    taskPage1.NextCursor,
	})
	if err != nil {
		t.Fatalf("list tasks page2: %v", err)
	}
	if len(taskPage2.Items) != 1 {
		t.Fatalf("expected 1 task on second page, got %d", len(taskPage2.Items))
	}
	if taskPage2.NextCursor != "" {
		t.Fatalf("expected empty next cursor on final page, got %s", taskPage2.NextCursor)
	}
	if taskPage2.AsOf.IsZero() {
		t.Fatal("expected non-zero as_of on second page")
	}

	sessionPage, err := svc.ListSessions(ctx, ListSessionsInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Status:   platformruntime.SessionStatusOpen,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessionPage.Items) != 2 {
		t.Fatalf("expected 2 open sessions for tenant-a, got %d", len(sessionPage.Items))
	}
	if sessionPage.AsOf.IsZero() {
		t.Fatal("expected non-zero sessions as_of")
	}
	for _, session := range sessionPage.Items {
		if session.TenantID != "tenant-a" {
			t.Fatalf("expected tenant-a session only, got %s", session.TenantID)
		}
		if session.Status != platformruntime.SessionStatusOpen {
			t.Fatalf("expected open session only, got %s", session.Status)
		}
	}

	if _, err := svc.GetSession(ctx, GetSessionInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-b-001",
	}); err == nil {
		t.Fatal("expected tenant isolation error on get session")
	}

	if _, err := svc.GetTask(ctx, GetTaskInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		TaskID:   "task-b-001",
	}); err == nil {
		t.Fatal("expected tenant isolation error on get task")
	}

	if _, err := svc.GetSession(ctx, GetSessionInput{
		TenantID:  "tenant-a",
		ActorID:   "actor-a",
		SessionID: "sess-a-001",
	}); err != nil {
		t.Fatalf("get own session: %v", err)
	}

	if _, err := svc.GetTask(ctx, GetTaskInput{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		TaskID:   "task-a-001",
	}); err != nil {
		t.Fatalf("get own task: %v", err)
	}
}
