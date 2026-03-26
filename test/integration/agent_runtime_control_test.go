package integration

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	application "github.com/nikkofu/erp-claw/internal/application/agentruntime"
	domain "github.com/nikkofu/erp-claw/internal/domain/agentruntime"
	"github.com/nikkofu/erp-claw/internal/interfaces/ws"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
)

func TestAgentRuntimeControlMigrationContainsStatusGuardrails(t *testing.T) {
	upSQL, err := os.ReadFile("../../migrations/000004_phase1_agent_runtime_control.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	sql := strings.ToLower(string(upSQL))

	required := []string{
		"agent_session_status_check",
		"check (status in ('open', 'closed'))",
		"agent_task_status_check",
		"check (status in ('pending', 'running', 'succeeded', 'failed', 'canceled'))",
		"create index if not exists idx_agent_task_tenant_session_queued_at",
		"create index if not exists idx_agent_task_tenant_status_queued_at",
	}
	for _, needle := range required {
		if !strings.Contains(sql, needle) {
			t.Fatalf("expected migration to contain %q", needle)
		}
	}
}

func TestAgentRuntimeControlMigrationDownDropsArtifacts(t *testing.T) {
	downSQL, err := os.ReadFile("../../migrations/000004_phase1_agent_runtime_control.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	sql := strings.ToLower(string(downSQL))

	required := []string{
		"drop index if exists idx_agent_task_tenant_status_queued_at",
		"drop index if exists idx_agent_task_tenant_session_queued_at",
		"drop constraint if exists agent_task_status_check",
		"drop constraint if exists agent_session_status_check",
	}
	for _, needle := range required {
		if !strings.Contains(sql, needle) {
			t.Fatalf("expected down migration to contain %q", needle)
		}
	}
}

func TestAgentRuntimeControlServiceFlowPersistsStateAndBroadcastsEvents(t *testing.T) {
	sessions := newMemorySessionRepository()
	tasks := newMemoryTaskRepository()
	gateway := ws.NewWorkspaceGateway()

	events, err := gateway.RegisterChannel("workspace-session", 8)
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

	session, err := service.CreateSession(context.Background(), "1001", "workspace-session", map[string]any{"source": "integration-test"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if session.Status != domain.SessionStatusOpen {
		t.Fatalf("expected open session status, got %q", session.Status)
	}

	task, err := service.CreateTask(context.Background(), "1001", "workspace-session", "plan", map[string]any{"prompt": "close month end"})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if _, err := service.StartTask(context.Background(), "1001", task.ID); err != nil {
		t.Fatalf("start task: %v", err)
	}
	if _, err := service.CompleteTask(context.Background(), "1001", task.ID, map[string]any{"result": "done"}); err != nil {
		t.Fatalf("complete task: %v", err)
	}
	if _, err := service.CloseSession(context.Background(), "1001", "workspace-session"); err != nil {
		t.Fatalf("close session: %v", err)
	}

	persistedTask, err := tasks.GetTaskByID(context.Background(), "1001", task.ID)
	if err != nil {
		t.Fatalf("get persisted task: %v", err)
	}
	if persistedTask.Status != domain.TaskStatusSucceeded {
		t.Fatalf("expected succeeded task status, got %q", persistedTask.Status)
	}
	if persistedTask.CompletedAt == nil {
		t.Fatal("expected task completed_at to be set")
	}

	persistedSession, err := sessions.GetSessionByTenantAndKey(context.Background(), "1001", "workspace-session")
	if err != nil {
		t.Fatalf("get persisted session: %v", err)
	}
	if persistedSession.Status != domain.SessionStatusClosed {
		t.Fatalf("expected closed session status, got %q", persistedSession.Status)
	}
	if persistedSession.EndedAt == nil {
		t.Fatal("expected session ended_at to be set")
	}

	got := collectWorkspaceEvents(t, events, 3, 250*time.Millisecond)
	if got[0].Type != platformruntime.WorkspaceEventTypeTaskStatusChanged {
		t.Fatalf("unexpected first event type: %s", got[0].Type)
	}
	if got[1].Type != platformruntime.WorkspaceEventTypeTaskStatusChanged {
		t.Fatalf("unexpected second event type: %s", got[1].Type)
	}
	if got[2].Type != platformruntime.WorkspaceEventTypeSessionStatusChanged {
		t.Fatalf("unexpected third event type: %s", got[2].Type)
	}
}

func collectWorkspaceEvents(t *testing.T, stream <-chan platformruntime.WorkspaceEvent, want int, timeout time.Duration) []platformruntime.WorkspaceEvent {
	t.Helper()

	events := make([]platformruntime.WorkspaceEvent, 0, want)
	for i := 0; i < want; i++ {
		select {
		case evt := <-stream:
			events = append(events, evt)
		case <-time.After(timeout):
			t.Fatalf("timed out waiting for event %d/%d", i+1, want)
		}
	}
	return events
}

type memorySessionRepository struct {
	mu     sync.RWMutex
	nextID int
	data   map[string]domain.Session
}

func newMemorySessionRepository() *memorySessionRepository {
	return &memorySessionRepository{
		data: map[string]domain.Session{},
	}
}

func (r *memorySessionRepository) CreateSession(_ context.Context, session domain.Session) (domain.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	session.ID = strconv.Itoa(r.nextID)
	r.data[r.sessionKey(session.TenantID, session.SessionKey)] = session
	return session, nil
}

func (r *memorySessionRepository) GetSessionByTenantAndKey(_ context.Context, tenantID, sessionKey string) (domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.data[r.sessionKey(tenantID, sessionKey)]
	if !ok {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	return session, nil
}

func (r *memorySessionRepository) UpdateSessionStatus(_ context.Context, tenantID, sessionKey string, status domain.SessionStatus, endedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.sessionKey(tenantID, sessionKey)
	session, ok := r.data[key]
	if !ok {
		return domain.ErrSessionNotFound
	}
	session.Status = status
	session.EndedAt = endedAt
	r.data[key] = session
	return nil
}

func (r *memorySessionRepository) sessionKey(tenantID, sessionKey string) string {
	return tenantID + "|" + sessionKey
}

type memoryTaskRepository struct {
	mu     sync.RWMutex
	nextID int
	data   map[string]domain.Task
}

func newMemoryTaskRepository() *memoryTaskRepository {
	return &memoryTaskRepository{
		data: map[string]domain.Task{},
	}
}

func (r *memoryTaskRepository) CreateTask(_ context.Context, task domain.Task) (domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	task.ID = strconv.Itoa(r.nextID)
	r.data[r.taskKey(task.TenantID, task.ID)] = task
	return task, nil
}

func (r *memoryTaskRepository) GetTaskByID(_ context.Context, tenantID, taskID string) (domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.data[r.taskKey(tenantID, taskID)]
	if !ok {
		return domain.Task{}, domain.ErrTaskNotFound
	}
	return task, nil
}

func (r *memoryTaskRepository) UpdateTaskStatus(_ context.Context, tenantID, taskID string, status domain.TaskStatus, output map[string]any, completedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.taskKey(tenantID, taskID)
	task, ok := r.data[key]
	if !ok {
		return domain.ErrTaskNotFound
	}
	task.Status = status
	task.Output = output
	task.CompletedAt = completedAt
	r.data[key] = task
	return nil
}

func (r *memoryTaskRepository) taskKey(tenantID, taskID string) string {
	return tenantID + "|" + taskID
}
