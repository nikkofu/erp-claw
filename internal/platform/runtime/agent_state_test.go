package runtime

import (
	"errors"
	"testing"
	"time"
)

func TestSessionLifecycleTransitions(t *testing.T) {
	now := time.Date(2026, 3, 26, 8, 0, 0, 0, time.UTC)
	session, err := NewSession("sess-001", "tenant-admin", "actor-a", map[string]any{"channel": "workspace"}, now)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if session.Status != SessionStatusOpen {
		t.Fatalf("expected session open, got %s", session.Status)
	}

	closedAt := now.Add(5 * time.Minute)
	if err := session.Close(closedAt); err != nil {
		t.Fatalf("close session: %v", err)
	}
	if session.Status != SessionStatusClosed {
		t.Fatalf("expected session closed, got %s", session.Status)
	}
	if session.EndedAt.IsZero() {
		t.Fatal("expected ended_at to be set")
	}
}

func TestTaskLifecycleTransitions(t *testing.T) {
	now := time.Date(2026, 3, 26, 8, 0, 0, 0, time.UTC)
	task, err := NewTask("task-001", "tenant-admin", "sess-001", "tool.call", map[string]any{"tool": "search"}, now)
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	if task.Status != TaskStatusPending {
		t.Fatalf("expected task pending, got %s", task.Status)
	}

	startedAt := now.Add(1 * time.Minute)
	if err := task.Start(startedAt); err != nil {
		t.Fatalf("start task: %v", err)
	}
	if task.Status != TaskStatusRunning {
		t.Fatalf("expected task running, got %s", task.Status)
	}

	doneAt := now.Add(2 * time.Minute)
	if err := task.Complete(map[string]any{"ok": true}, doneAt); err != nil {
		t.Fatalf("complete task: %v", err)
	}
	if task.Status != TaskStatusSucceeded {
		t.Fatalf("expected task succeeded, got %s", task.Status)
	}
	if task.CompletedAt.IsZero() {
		t.Fatal("expected completed_at to be set")
	}
}

func TestTaskCannotCompleteBeforeStart(t *testing.T) {
	now := time.Date(2026, 3, 26, 8, 0, 0, 0, time.UTC)
	task, err := NewTask("task-001", "tenant-admin", "sess-001", "tool.call", nil, now)
	if err != nil {
		t.Fatalf("new task: %v", err)
	}

	err = task.Complete(map[string]any{"ok": true}, now.Add(time.Minute))
	if !errors.Is(err, ErrInvalidTaskTransition) {
		t.Fatalf("expected ErrInvalidTaskTransition, got %v", err)
	}
}
