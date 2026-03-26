package agentruntime

import (
	"context"
	"testing"
	"time"

	domain "github.com/nikkofu/erp-claw/internal/domain/agentruntime"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
)

func TestNewServiceRejectsNilTaskRepository(t *testing.T) {
	_, err := NewService(ServiceDeps{
		Sessions: &stubSessionRepository{},
	})
	if err == nil {
		t.Fatal("expected nil task repository to fail")
	}
}

func TestCreateSessionPersistsMetadata(t *testing.T) {
	sessions := &stubSessionRepository{}
	tasks := &stubTaskRepository{}
	svc, err := NewService(ServiceDeps{
		Sessions: sessions,
		Tasks:    tasks,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	session, err := svc.CreateSession(context.Background(), "tenant-a", "session-a", map[string]any{"channel": "workspace"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if session.Status != domain.SessionStatusOpen {
		t.Fatalf("expected open status, got %q", session.Status)
	}
	if sessions.createCalls != 1 {
		t.Fatalf("expected one create call, got %d", sessions.createCalls)
	}
}

func TestStartTaskTransitionsPendingToRunningAndAppendsEvent(t *testing.T) {
	task, err := domain.NewTask("tenant-a", "session-a", "plan", map[string]any{"prompt": "hello"})
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	task.ID = "task-a"

	tasks := &stubTaskRepository{tasks: map[string]domain.Task{"task-a": task}}
	events := &stubWorkspaceEventAppender{}
	svc, err := NewService(ServiceDeps{
		Sessions: &stubSessionRepository{},
		Tasks:    tasks,
		Events:   events,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	updated, err := svc.StartTask(context.Background(), "tenant-a", "task-a")
	if err != nil {
		t.Fatalf("start task: %v", err)
	}

	if updated.Status != domain.TaskStatusRunning {
		t.Fatalf("expected running status, got %q", updated.Status)
	}
	if len(events.events) != 1 {
		t.Fatalf("expected one appended event, got %d", len(events.events))
	}
	if events.events[0].Type != platformruntime.WorkspaceEventTypeTaskStatusChanged {
		t.Fatalf("unexpected event type: %s", events.events[0].Type)
	}
}

func TestCompleteTaskRejectsPendingTask(t *testing.T) {
	task, err := domain.NewTask("tenant-a", "session-a", "plan", map[string]any{"prompt": "hello"})
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	task.ID = "task-a"

	tasks := &stubTaskRepository{tasks: map[string]domain.Task{"task-a": task}}
	events := &stubWorkspaceEventAppender{}
	svc, err := NewService(ServiceDeps{
		Sessions: &stubSessionRepository{},
		Tasks:    tasks,
		Events:   events,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = svc.CompleteTask(context.Background(), "tenant-a", "task-a", map[string]any{"result": "ok"})
	if err == nil {
		t.Fatal("expected complete task to reject pending status")
	}
	if len(events.events) != 0 {
		t.Fatalf("expected no event append, got %d", len(events.events))
	}
}

func TestFailTaskTransitionsRunningToFailedAndAppendsEvent(t *testing.T) {
	task, err := domain.NewTask("tenant-a", "session-a", "plan", map[string]any{"prompt": "hello"})
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	task.ID = "task-a"
	if err := task.TransitionTo(domain.TaskStatusRunning, nil, time.Time{}); err != nil {
		t.Fatalf("pending -> running: %v", err)
	}

	tasks := &stubTaskRepository{tasks: map[string]domain.Task{"task-a": task}}
	events := &stubWorkspaceEventAppender{}
	svc, err := NewService(ServiceDeps{
		Sessions: &stubSessionRepository{},
		Tasks:    tasks,
		Events:   events,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	updated, err := svc.FailTask(context.Background(), "tenant-a", "task-a", map[string]any{"error": "timeout"})
	if err != nil {
		t.Fatalf("fail task: %v", err)
	}

	if updated.Status != domain.TaskStatusFailed {
		t.Fatalf("expected failed status, got %q", updated.Status)
	}
	if updated.CompletedAt == nil {
		t.Fatal("expected completed_at to be set")
	}
	if updated.Output["error"] != "timeout" {
		t.Fatalf("expected error output to be persisted, got %#v", updated.Output)
	}
	if len(events.events) != 1 {
		t.Fatalf("expected one appended event, got %d", len(events.events))
	}
	if events.events[0].Type != platformruntime.WorkspaceEventTypeTaskStatusChanged {
		t.Fatalf("unexpected event type: %s", events.events[0].Type)
	}
	payload, ok := events.events[0].Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", events.events[0].Payload)
	}
	if payload["status"] != string(domain.TaskStatusFailed) {
		t.Fatalf("expected failed status payload, got %#v", payload)
	}
}

func TestFailTaskRejectsSucceededTask(t *testing.T) {
	task, err := domain.NewTask("tenant-a", "session-a", "plan", map[string]any{"prompt": "hello"})
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	task.ID = "task-a"
	if err := task.TransitionTo(domain.TaskStatusRunning, nil, time.Time{}); err != nil {
		t.Fatalf("pending -> running: %v", err)
	}
	if err := task.TransitionTo(domain.TaskStatusSucceeded, map[string]any{"result": "ok"}, time.Now().UTC()); err != nil {
		t.Fatalf("running -> succeeded: %v", err)
	}

	tasks := &stubTaskRepository{tasks: map[string]domain.Task{"task-a": task}}
	events := &stubWorkspaceEventAppender{}
	svc, err := NewService(ServiceDeps{
		Sessions: &stubSessionRepository{},
		Tasks:    tasks,
		Events:   events,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = svc.FailTask(context.Background(), "tenant-a", "task-a", map[string]any{"error": "timeout"})
	if err == nil {
		t.Fatal("expected fail task to reject succeeded status")
	}
	if len(events.events) != 0 {
		t.Fatalf("expected no event append, got %d", len(events.events))
	}
}

func TestCancelTaskTransitionsPendingToCanceledAndAppendsEvent(t *testing.T) {
	task, err := domain.NewTask("tenant-a", "session-a", "plan", map[string]any{"prompt": "hello"})
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	task.ID = "task-a"

	tasks := &stubTaskRepository{tasks: map[string]domain.Task{"task-a": task}}
	events := &stubWorkspaceEventAppender{}
	svc, err := NewService(ServiceDeps{
		Sessions: &stubSessionRepository{},
		Tasks:    tasks,
		Events:   events,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	updated, err := svc.CancelTask(context.Background(), "tenant-a", "task-a", map[string]any{"reason": "user requested"})
	if err != nil {
		t.Fatalf("cancel task: %v", err)
	}

	if updated.Status != domain.TaskStatusCanceled {
		t.Fatalf("expected canceled status, got %q", updated.Status)
	}
	if updated.CompletedAt == nil {
		t.Fatal("expected completed_at to be set")
	}
	if updated.Output["reason"] != "user requested" {
		t.Fatalf("expected cancel reason output to be persisted, got %#v", updated.Output)
	}
	if len(events.events) != 1 {
		t.Fatalf("expected one appended event, got %d", len(events.events))
	}
	if events.events[0].Type != platformruntime.WorkspaceEventTypeTaskStatusChanged {
		t.Fatalf("unexpected event type: %s", events.events[0].Type)
	}
	payload, ok := events.events[0].Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", events.events[0].Payload)
	}
	if payload["status"] != string(domain.TaskStatusCanceled) {
		t.Fatalf("expected canceled status payload, got %#v", payload)
	}
}

func TestCancelTaskRejectsSucceededTask(t *testing.T) {
	task, err := domain.NewTask("tenant-a", "session-a", "plan", map[string]any{"prompt": "hello"})
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	task.ID = "task-a"
	if err := task.TransitionTo(domain.TaskStatusRunning, nil, time.Time{}); err != nil {
		t.Fatalf("pending -> running: %v", err)
	}
	if err := task.TransitionTo(domain.TaskStatusSucceeded, map[string]any{"result": "ok"}, time.Now().UTC()); err != nil {
		t.Fatalf("running -> succeeded: %v", err)
	}

	tasks := &stubTaskRepository{tasks: map[string]domain.Task{"task-a": task}}
	events := &stubWorkspaceEventAppender{}
	svc, err := NewService(ServiceDeps{
		Sessions: &stubSessionRepository{},
		Tasks:    tasks,
		Events:   events,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = svc.CancelTask(context.Background(), "tenant-a", "task-a", map[string]any{"reason": "too late"})
	if err == nil {
		t.Fatal("expected cancel task to reject succeeded status")
	}
	if len(events.events) != 0 {
		t.Fatalf("expected no event append, got %d", len(events.events))
	}
}

func TestCloseSessionTransitionsOpenToClosedAndAppendsEvent(t *testing.T) {
	sessions := &stubSessionRepository{}
	tasks := &stubTaskRepository{}
	events := &stubWorkspaceEventAppender{}
	svc, err := NewService(ServiceDeps{
		Sessions: sessions,
		Tasks:    tasks,
		Events:   events,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = svc.CreateSession(context.Background(), "tenant-a", "session-a", map[string]any{"channel": "workspace"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	closed, err := svc.CloseSession(context.Background(), "tenant-a", "session-a")
	if err != nil {
		t.Fatalf("close session: %v", err)
	}

	if closed.Status != domain.SessionStatusClosed {
		t.Fatalf("expected closed status, got %q", closed.Status)
	}
	if closed.EndedAt == nil {
		t.Fatal("expected ended_at to be set")
	}
	if len(events.events) != 1 {
		t.Fatalf("expected one appended event, got %d", len(events.events))
	}
	if events.events[0].Type != platformruntime.WorkspaceEventTypeSessionStatusChanged {
		t.Fatalf("unexpected event type: %s", events.events[0].Type)
	}
}

type stubSessionRepository struct {
	createCalls int
	sessions    map[string]domain.Session
}

func (s *stubSessionRepository) CreateSession(_ context.Context, session domain.Session) (domain.Session, error) {
	s.createCalls++
	if session.ID == "" {
		session.ID = "session-row-1"
	}
	if s.sessions == nil {
		s.sessions = map[string]domain.Session{}
	}
	s.sessions[session.SessionKey] = session
	return session, nil
}

func (s *stubSessionRepository) GetSessionByTenantAndKey(_ context.Context, tenantID, sessionKey string) (domain.Session, error) {
	session, ok := s.sessions[sessionKey]
	if !ok || session.TenantID != tenantID {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	return session, nil
}

func (s *stubSessionRepository) UpdateSessionStatus(_ context.Context, tenantID, sessionKey string, status domain.SessionStatus, endedAt *time.Time) error {
	session, ok := s.sessions[sessionKey]
	if !ok || session.TenantID != tenantID {
		return domain.ErrSessionNotFound
	}
	session.Status = status
	session.EndedAt = endedAt
	s.sessions[sessionKey] = session
	return nil
}

type stubTaskRepository struct {
	tasks map[string]domain.Task
}

func (s *stubTaskRepository) CreateTask(_ context.Context, task domain.Task) (domain.Task, error) {
	if s.tasks == nil {
		s.tasks = map[string]domain.Task{}
	}
	if task.ID == "" {
		task.ID = "task-row-1"
	}
	s.tasks[task.ID] = task
	return task, nil
}

func (s *stubTaskRepository) GetTaskByID(_ context.Context, tenantID, taskID string) (domain.Task, error) {
	task, ok := s.tasks[taskID]
	if !ok || task.TenantID != tenantID {
		return domain.Task{}, domain.ErrTaskNotFound
	}
	return task, nil
}

func (s *stubTaskRepository) UpdateTaskStatus(_ context.Context, tenantID, taskID string, status domain.TaskStatus, output map[string]any, completedAt *time.Time) error {
	task, ok := s.tasks[taskID]
	if !ok || task.TenantID != tenantID {
		return domain.ErrTaskNotFound
	}
	task.Status = status
	if output != nil {
		task.Output = output
	}
	task.CompletedAt = completedAt
	s.tasks[taskID] = task
	return nil
}

type stubWorkspaceEventAppender struct {
	events []platformruntime.WorkspaceEvent
}

func (s *stubWorkspaceEventAppender) AppendWorkspaceEvent(_ context.Context, evt platformruntime.WorkspaceEvent) error {
	s.events = append(s.events, evt)
	return nil
}
