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

func TestTaskCanCancelFromPending(t *testing.T) {
	now := time.Date(2026, 3, 26, 8, 0, 0, 0, time.UTC)
	task, err := NewTask("task-001", "tenant-admin", "sess-001", "tool.call", nil, now)
	if err != nil {
		t.Fatalf("new task: %v", err)
	}

	canceledAt := now.Add(time.Minute)
	if err := task.Cancel("manual cancel", canceledAt); err != nil {
		t.Fatalf("cancel task: %v", err)
	}
	if task.Status != TaskStatusCanceled {
		t.Fatalf("expected task canceled, got %s", task.Status)
	}
	if task.FailureReason != "manual cancel" {
		t.Fatalf("expected cancel reason, got %q", task.FailureReason)
	}
	if task.CompletedAt.IsZero() {
		t.Fatal("expected completed_at to be set for canceled task")
	}
}

func TestTaskCannotCancelAfterCompletion(t *testing.T) {
	now := time.Date(2026, 3, 26, 8, 0, 0, 0, time.UTC)
	task, err := NewTask("task-001", "tenant-admin", "sess-001", "tool.call", nil, now)
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	if err := task.Start(now.Add(time.Minute)); err != nil {
		t.Fatalf("start task: %v", err)
	}
	if err := task.Complete(map[string]any{"ok": true}, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("complete task: %v", err)
	}

	err = task.Cancel("too late", now.Add(3*time.Minute))
	if !errors.Is(err, ErrInvalidTaskTransition) {
		t.Fatalf("expected ErrInvalidTaskTransition, got %v", err)
	}
}

func TestTaskCanRetryFromFailed(t *testing.T) {
	now := time.Date(2026, 3, 26, 8, 0, 0, 0, time.UTC)
	task, err := NewTask("task-001", "tenant-admin", "sess-001", "tool.call", nil, now)
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	if err := task.Start(now.Add(time.Minute)); err != nil {
		t.Fatalf("start task: %v", err)
	}
	if err := task.Fail("tool timeout", now.Add(2*time.Minute)); err != nil {
		t.Fatalf("fail task: %v", err)
	}

	requeuedAt := now.Add(3 * time.Minute)
	if err := task.Retry(requeuedAt); err != nil {
		t.Fatalf("retry task: %v", err)
	}
	if task.Status != TaskStatusPending {
		t.Fatalf("expected pending after retry, got %s", task.Status)
	}
	if task.FailureReason != "" {
		t.Fatalf("expected empty failure reason after retry, got %q", task.FailureReason)
	}
	if !task.QueuedAt.Equal(requeuedAt) {
		t.Fatalf("expected queued_at to be reset to retry time, got %s", task.QueuedAt.Format(time.RFC3339))
	}
	if !task.StartedAt.IsZero() {
		t.Fatalf("expected started_at reset on retry, got %s", task.StartedAt.Format(time.RFC3339))
	}
	if !task.CompletedAt.IsZero() {
		t.Fatalf("expected completed_at reset on retry, got %s", task.CompletedAt.Format(time.RFC3339))
	}
}

func TestTaskRetryRejectsWhenAttemptsExhausted(t *testing.T) {
	now := time.Date(2026, 3, 26, 8, 0, 0, 0, time.UTC)
	task, err := NewTask("task-001", "tenant-admin", "sess-001", "tool.call", nil, now)
	if err != nil {
		t.Fatalf("new task: %v", err)
	}

	for i := 0; i < MaxTaskAttempts; i++ {
		if err := task.Start(now.Add(time.Duration(i+1) * time.Minute)); err != nil {
			t.Fatalf("start attempt %d: %v", i+1, err)
		}
		if err := task.Fail("tool timeout", now.Add(time.Duration(i+1)*time.Minute+30*time.Second)); err != nil {
			t.Fatalf("fail attempt %d: %v", i+1, err)
		}
		if i < MaxTaskAttempts-1 {
			if err := task.Retry(now.Add(time.Duration(i+1)*time.Minute + 45*time.Second)); err != nil {
				t.Fatalf("retry attempt %d: %v", i+1, err)
			}
		}
	}

	err = task.Retry(now.Add(10 * time.Minute))
	if !errors.Is(err, ErrTaskRetryLimitExceeded) {
		t.Fatalf("expected ErrTaskRetryLimitExceeded, got %v", err)
	}
}
