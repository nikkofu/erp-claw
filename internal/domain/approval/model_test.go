package approval

import "testing"

func TestNewDefinitionRejectsEmptyApprover(t *testing.T) {
	_, err := NewDefinition("tenant-a", "def-a", "purchase approval", "", true)
	if err == nil {
		t.Fatal("expected empty approver to fail")
	}
}

func TestNewInstanceStartsPending(t *testing.T) {
	instance, err := NewInstance("tenant-a", "def-a", "purchase_order", "po-1", "user-a")
	if err != nil {
		t.Fatalf("new instance: %v", err)
	}

	if instance.Status != InstanceStatusPending {
		t.Fatalf("expected pending status, got %q", instance.Status)
	}
}

func TestTaskDecideRejectsSecondDecision(t *testing.T) {
	task, err := NewTask("tenant-a", "inst-a", "approver-a")
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	if err := task.Decide(TaskStatusApproved, "approver-a", "ok"); err != nil {
		t.Fatalf("first decision: %v", err)
	}

	if err := task.Decide(TaskStatusRejected, "approver-a", "no"); err == nil {
		t.Fatal("expected second decision to fail")
	}
}

func TestInstanceTransitionRejectsReturningToPending(t *testing.T) {
	instance, err := NewInstance("tenant-a", "def-a", "purchase_order", "po-1", "user-a")
	if err != nil {
		t.Fatalf("new instance: %v", err)
	}
	if err := instance.TransitionTo(InstanceStatusApproved); err != nil {
		t.Fatalf("pending -> approved: %v", err)
	}

	if err := instance.TransitionTo(InstanceStatusPending); err == nil {
		t.Fatal("expected approved -> pending to fail")
	}
}
