package approval

import (
	"context"
	"testing"

	domainapproval "github.com/nikkofu/erp-claw/internal/domain/approval"
)

func TestSaveDefinitionHandlerCreatesDefinition(t *testing.T) {
	repo := newFakeRepository()
	handler := SaveDefinitionHandler{Definitions: repo}

	definition, err := handler.Handle(context.Background(), SaveDefinition{
		TenantID:   "tenant-a",
		ID:         "def-a",
		Name:       "purchase approval",
		ApproverID: "manager-a",
		Active:     true,
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if definition.ID != "def-a" {
		t.Fatalf("expected definition id def-a, got %s", definition.ID)
	}
	if got := repo.definitions["tenant-a|def-a"]; got.Name != "purchase approval" {
		t.Fatalf("definition not stored correctly: %#v", got)
	}
}

func TestListDefinitionsHandlerReturnsTenantScopedDefinitions(t *testing.T) {
	repo := newFakeRepository()
	repo.definitions["tenant-a|def-a"] = domainapproval.Definition{TenantID: "tenant-a", ID: "def-a", Name: "purchase approval", ApproverID: "manager-a", Active: true}
	repo.definitions["tenant-b|def-b"] = domainapproval.Definition{TenantID: "tenant-b", ID: "def-b", Name: "expense approval", ApproverID: "manager-b", Active: true}

	handler := ListDefinitionsHandler{Definitions: repo}
	definitions, err := handler.Handle(context.Background(), ListDefinitions{TenantID: "tenant-a"})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(definitions) != 1 {
		t.Fatalf("expected one definition, got %d", len(definitions))
	}
	if definitions[0].ID != "def-a" {
		t.Fatalf("expected def-a, got %s", definitions[0].ID)
	}
}

func TestListInstancesHandlerReturnsTenantScopedInstances(t *testing.T) {
	repo := newFakeRepository()
	repo.instances["tenant-a|inst-a"] = domainapproval.Instance{TenantID: "tenant-a", ID: "inst-a", DefinitionID: "def-a", Status: domainapproval.InstanceStatusPending}
	repo.instances["tenant-b|inst-b"] = domainapproval.Instance{TenantID: "tenant-b", ID: "inst-b", DefinitionID: "def-b", Status: domainapproval.InstanceStatusApproved}

	handler := ListInstancesHandler{Instances: repo}
	instances, err := handler.Handle(context.Background(), ListInstances{TenantID: "tenant-a"})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected one instance, got %d", len(instances))
	}
	if instances[0].ID != "inst-a" {
		t.Fatalf("expected inst-a, got %s", instances[0].ID)
	}
}

func TestListTasksHandlerReturnsTenantScopedTasks(t *testing.T) {
	repo := newFakeRepository()
	repo.tasks["tenant-a|task-a"] = domainapproval.Task{TenantID: "tenant-a", ID: "task-a", InstanceID: "inst-a", ApproverID: "manager-a", Status: domainapproval.TaskStatusPending}
	repo.tasks["tenant-b|task-b"] = domainapproval.Task{TenantID: "tenant-b", ID: "task-b", InstanceID: "inst-b", ApproverID: "manager-b", Status: domainapproval.TaskStatusApproved}

	handler := ListTasksHandler{Tasks: repo}
	tasks, err := handler.Handle(context.Background(), ListTasks{TenantID: "tenant-a"})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected one task, got %d", len(tasks))
	}
	if tasks[0].ID != "task-a" {
		t.Fatalf("expected task-a, got %s", tasks[0].ID)
	}
}
