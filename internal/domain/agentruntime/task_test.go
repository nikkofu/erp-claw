package agentruntime

import (
	"testing"
	"time"
)

func TestNewTaskRequiresTaskType(t *testing.T) {
	_, err := NewTask("1001", "2001", "", map[string]any{"prompt": "hello"})
	if err == nil {
		t.Fatal("expected empty task type to fail")
	}
}

func TestTaskTransitionPendingToRunningToSucceeded(t *testing.T) {
	task, err := NewTask("1001", "2001", "plan", map[string]any{"prompt": "hello"})
	if err != nil {
		t.Fatalf("new task: %v", err)
	}

	if err := task.TransitionTo(TaskStatusRunning, nil, time.Time{}); err != nil {
		t.Fatalf("pending -> running: %v", err)
	}
	if task.Status != TaskStatusRunning {
		t.Fatalf("expected running status, got %q", task.Status)
	}

	completedAt := time.Now().UTC()
	output := map[string]any{"result": "ok"}
	if err := task.TransitionTo(TaskStatusSucceeded, output, completedAt); err != nil {
		t.Fatalf("running -> succeeded: %v", err)
	}
	if task.Status != TaskStatusSucceeded {
		t.Fatalf("expected succeeded status, got %q", task.Status)
	}
	if task.CompletedAt == nil || !task.CompletedAt.Equal(completedAt) {
		t.Fatalf("expected completed_at to be set to %s", completedAt)
	}
}

func TestTaskTransitionFromSucceededToRunningRejected(t *testing.T) {
	task, err := NewTask("1001", "2001", "plan", map[string]any{"prompt": "hello"})
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	if err := task.TransitionTo(TaskStatusRunning, nil, time.Time{}); err != nil {
		t.Fatalf("pending -> running: %v", err)
	}
	if err := task.TransitionTo(TaskStatusSucceeded, map[string]any{"result": "ok"}, time.Now().UTC()); err != nil {
		t.Fatalf("running -> succeeded: %v", err)
	}

	err = task.TransitionTo(TaskStatusRunning, nil, time.Time{})
	if err == nil {
		t.Fatal("expected succeeded -> running transition to fail")
	}
}

func TestTaskTransitionRunningToFailedSetsCompletion(t *testing.T) {
	task, err := NewTask("1001", "2001", "plan", map[string]any{"prompt": "hello"})
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	if err := task.TransitionTo(TaskStatusRunning, nil, time.Time{}); err != nil {
		t.Fatalf("pending -> running: %v", err)
	}

	completedAt := time.Now().UTC()
	output := map[string]any{"error": "model timeout"}
	if err := task.TransitionTo(TaskStatusFailed, output, completedAt); err != nil {
		t.Fatalf("running -> failed: %v", err)
	}

	if task.Status != TaskStatusFailed {
		t.Fatalf("expected failed status, got %q", task.Status)
	}
	if task.CompletedAt == nil || !task.CompletedAt.Equal(completedAt) {
		t.Fatalf("expected completed_at to be set to %s", completedAt)
	}
	if task.Output["error"] != "model timeout" {
		t.Fatalf("expected error output to be persisted, got %#v", task.Output)
	}
}

func TestTaskTransitionPendingToCanceledSetsCompletion(t *testing.T) {
	task, err := NewTask("1001", "2001", "plan", map[string]any{"prompt": "hello"})
	if err != nil {
		t.Fatalf("new task: %v", err)
	}

	completedAt := time.Now().UTC()
	output := map[string]any{"reason": "user-requested"}
	if err := task.TransitionTo(TaskStatusCanceled, output, completedAt); err != nil {
		t.Fatalf("pending -> canceled: %v", err)
	}

	if task.Status != TaskStatusCanceled {
		t.Fatalf("expected canceled status, got %q", task.Status)
	}
	if task.CompletedAt == nil || !task.CompletedAt.Equal(completedAt) {
		t.Fatalf("expected completed_at to be set to %s", completedAt)
	}
	if task.Output["reason"] != "user-requested" {
		t.Fatalf("expected cancel reason output to be persisted, got %#v", task.Output)
	}
}

func TestTaskTransitionPendingToSucceededRejected(t *testing.T) {
	task, err := NewTask("1001", "2001", "plan", map[string]any{"prompt": "hello"})
	if err != nil {
		t.Fatalf("new task: %v", err)
	}

	err = task.TransitionTo(TaskStatusSucceeded, map[string]any{"result": "ok"}, time.Now().UTC())
	if err == nil {
		t.Fatal("expected pending -> succeeded transition to fail")
	}
}

func TestTaskTransitionFromCanceledToFailedRejected(t *testing.T) {
	task, err := NewTask("1001", "2001", "plan", map[string]any{"prompt": "hello"})
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	if err := task.TransitionTo(TaskStatusCanceled, map[string]any{"reason": "user-requested"}, time.Now().UTC()); err != nil {
		t.Fatalf("pending -> canceled: %v", err)
	}

	err = task.TransitionTo(TaskStatusFailed, map[string]any{"error": "late failure"}, time.Now().UTC())
	if err == nil {
		t.Fatal("expected canceled -> failed transition to fail")
	}
}
