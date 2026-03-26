package agentruntime

import (
	"context"
	"testing"
	"time"

	domain "github.com/nikkofu/erp-claw/internal/domain/agentruntime"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
)

func TestListSessionsHandlerUsesTenantScope(t *testing.T) {
	repo := &stubSessionLister{
		list: []domain.Session{{TenantID: "tenant-a", SessionKey: "session-a"}},
	}
	handler := ListSessionsHandler{Sessions: repo}

	sessions, err := handler.Handle(context.Background(), ListSessions{TenantID: "tenant-a"})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if repo.lastTenantID != "tenant-a" {
		t.Fatalf("expected tenant scope forwarded, got %q", repo.lastTenantID)
	}
	if len(sessions) != 1 || sessions[0].SessionKey != "session-a" {
		t.Fatalf("unexpected sessions: %+v", sessions)
	}
}

func TestListTasksHandlerUsesTenantAndSessionScope(t *testing.T) {
	repo := &stubTaskLister{
		list: []domain.Task{{TenantID: "tenant-a", SessionID: "session-a", TaskType: "plan"}},
	}
	handler := ListTasksHandler{Tasks: repo}

	tasks, err := handler.Handle(context.Background(), ListTasks{TenantID: "tenant-a", SessionID: "session-a"})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if repo.lastTenantID != "tenant-a" || repo.lastSessionID != "session-a" {
		t.Fatalf("unexpected scopes: tenant=%q session=%q", repo.lastTenantID, repo.lastSessionID)
	}
	if len(tasks) != 1 || tasks[0].TaskType != "plan" {
		t.Fatalf("unexpected tasks: %+v", tasks)
	}
}

func TestReplayWorkspaceEventsHandlerDelegates(t *testing.T) {
	reader := &stubWorkspaceEventReader{
		list: []platformruntime.WorkspaceEvent{
			{Type: platformruntime.WorkspaceEventTypeTaskStatusChanged, SessionID: "session-a", OccurredAt: time.Date(2026, 3, 26, 10, 0, 0, 0, time.UTC)},
		},
	}
	handler := ReplayWorkspaceEventsHandler{Events: reader}

	events, err := handler.Handle(context.Background(), ReplayWorkspaceEvents{TenantID: "tenant-a", SessionID: "session-a"})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if reader.lastTenantID != "tenant-a" || reader.lastSessionID != "session-a" {
		t.Fatalf("unexpected scopes: tenant=%q session=%q", reader.lastTenantID, reader.lastSessionID)
	}
	if len(events) != 1 || events[0].SessionID != "session-a" {
		t.Fatalf("unexpected events: %+v", events)
	}
}

type stubSessionLister struct {
	list         []domain.Session
	lastTenantID string
}

func (s *stubSessionLister) ListSessions(_ context.Context, tenantID string) ([]domain.Session, error) {
	s.lastTenantID = tenantID
	return s.list, nil
}

type stubTaskLister struct {
	list          []domain.Task
	lastTenantID  string
	lastSessionID string
}

func (s *stubTaskLister) ListTasks(_ context.Context, tenantID, sessionID string) ([]domain.Task, error) {
	s.lastTenantID = tenantID
	s.lastSessionID = sessionID
	return s.list, nil
}

type stubWorkspaceEventReader struct {
	list          []platformruntime.WorkspaceEvent
	lastTenantID  string
	lastSessionID string
}

func (s *stubWorkspaceEventReader) ListWorkspaceEvents(_ context.Context, tenantID, sessionID string) ([]platformruntime.WorkspaceEvent, error) {
	s.lastTenantID = tenantID
	s.lastSessionID = sessionID
	return s.list, nil
}
