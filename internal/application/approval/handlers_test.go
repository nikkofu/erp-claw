package approval

import (
	"context"
	"errors"
	"testing"

	domainapproval "github.com/nikkofu/erp-claw/internal/domain/approval"
)

func TestStartApprovalHandlerCreatesPendingInstanceAndTask(t *testing.T) {
	repo := newFakeRepository()
	repo.definitions["tenant-a|def-a"] = domainapproval.Definition{
		TenantID:   "tenant-a",
		ID:         "def-a",
		Name:       "purchase approval",
		ApproverID: "manager-a",
		Active:     true,
	}

	handler := StartApprovalHandler{
		Definitions: repo,
		Instances:   repo,
		Tasks:       repo,
	}
	result, err := handler.Handle(context.Background(), StartApproval{
		TenantID:     "tenant-a",
		DefinitionID: "def-a",
		ResourceType: "purchase_order",
		ResourceID:   "po-1",
		RequestedBy:  "user-a",
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if result.Instance.Status != domainapproval.InstanceStatusPending {
		t.Fatalf("expected pending instance, got %q", result.Instance.Status)
	}
	if result.Task.ApproverID != "manager-a" {
		t.Fatalf("unexpected approver: %s", result.Task.ApproverID)
	}
}

func TestApproveTaskHandlerApprovesTaskAndInstance(t *testing.T) {
	repo := seededApprovalRepo(t)
	handler := ApproveTaskHandler{
		Instances: repo,
		Tasks:     repo,
	}

	result, err := handler.Handle(context.Background(), ApproveTask{
		TenantID: "tenant-a",
		TaskID:   "task-a",
		ActorID:  "manager-a",
		Comment:  "approved",
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if result.Task.Status != domainapproval.TaskStatusApproved {
		t.Fatalf("expected approved task, got %q", result.Task.Status)
	}
	if result.Instance.Status != domainapproval.InstanceStatusApproved {
		t.Fatalf("expected approved instance, got %q", result.Instance.Status)
	}
}

func TestStartApprovalHandlerDeletesInstanceWhenTaskCreationFails(t *testing.T) {
	repo := newFakeRepository()
	repo.definitions["tenant-a|def-a"] = domainapproval.Definition{
		TenantID:   "tenant-a",
		ID:         "def-a",
		Name:       "purchase approval",
		ApproverID: "manager-a",
		Active:     true,
	}
	repo.createTaskErr = errors.New("task store unavailable")

	handler := StartApprovalHandler{
		Definitions: repo,
		Instances:   repo,
		Tasks:       repo,
	}
	_, err := handler.Handle(context.Background(), StartApproval{
		TenantID:     "tenant-a",
		DefinitionID: "def-a",
		ResourceType: "purchase_order",
		ResourceID:   "po-1",
		RequestedBy:  "user-a",
	})
	if err == nil {
		t.Fatal("expected task creation failure")
	}
	if len(repo.instances) != 0 {
		t.Fatalf("expected no orphaned approval instances, got %d", len(repo.instances))
	}
}

func TestRejectTaskHandlerRejectsTaskAndInstance(t *testing.T) {
	repo := seededApprovalRepo(t)
	handler := RejectTaskHandler{
		Instances: repo,
		Tasks:     repo,
	}

	result, err := handler.Handle(context.Background(), RejectTask{
		TenantID: "tenant-a",
		TaskID:   "task-a",
		ActorID:  "manager-a",
		Comment:  "rejected",
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if result.Task.Status != domainapproval.TaskStatusRejected {
		t.Fatalf("expected rejected task, got %q", result.Task.Status)
	}
	if result.Instance.Status != domainapproval.InstanceStatusRejected {
		t.Fatalf("expected rejected instance, got %q", result.Instance.Status)
	}
}

type fakeRepository struct {
	nextID      int
	definitions map[string]domainapproval.Definition
	instances   map[string]domainapproval.Instance
	tasks       map[string]domainapproval.Task
	createTaskErr error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		definitions: make(map[string]domainapproval.Definition),
		instances:   make(map[string]domainapproval.Instance),
		tasks:       make(map[string]domainapproval.Task),
	}
}

func seededApprovalRepo(t *testing.T) *fakeRepository {
	t.Helper()

	repo := newFakeRepository()
	instance, err := domainapproval.NewInstance("tenant-a", "def-a", "purchase_order", "po-1", "user-a")
	if err != nil {
		t.Fatalf("new instance: %v", err)
	}
	instance.ID = "inst-a"
	repo.instances["tenant-a|inst-a"] = instance

	task, err := domainapproval.NewTask("tenant-a", "inst-a", "manager-a")
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	task.ID = "task-a"
	repo.tasks["tenant-a|task-a"] = task
	return repo
}

func (r *fakeRepository) SaveDefinition(_ context.Context, definition domainapproval.Definition) (domainapproval.Definition, error) {
	r.definitions[definition.TenantID+"|"+definition.ID] = definition
	return definition, nil
}

func (r *fakeRepository) GetDefinitionByID(_ context.Context, tenantID, definitionID string) (domainapproval.Definition, error) {
	definition, ok := r.definitions[tenantID+"|"+definitionID]
	if !ok {
		return domainapproval.Definition{}, domainapproval.ErrDefinitionNotFound
	}
	return definition, nil
}

func (r *fakeRepository) CreateInstance(_ context.Context, instance domainapproval.Instance) (domainapproval.Instance, error) {
	r.nextID++
	if instance.ID == "" {
		instance.ID = "inst-created"
	}
	r.instances[instance.TenantID+"|"+instance.ID] = instance
	return instance, nil
}

func (r *fakeRepository) GetInstanceByID(_ context.Context, tenantID, instanceID string) (domainapproval.Instance, error) {
	instance, ok := r.instances[tenantID+"|"+instanceID]
	if !ok {
		return domainapproval.Instance{}, domainapproval.ErrInstanceNotFound
	}
	return instance, nil
}

func (r *fakeRepository) UpdateInstanceStatus(_ context.Context, tenantID, instanceID string, status domainapproval.InstanceStatus) error {
	instance, ok := r.instances[tenantID+"|"+instanceID]
	if !ok {
		return domainapproval.ErrInstanceNotFound
	}
	instance.Status = status
	r.instances[tenantID+"|"+instanceID] = instance
	return nil
}

func (r *fakeRepository) CreateTask(_ context.Context, task domainapproval.Task) (domainapproval.Task, error) {
	if r.createTaskErr != nil {
		return domainapproval.Task{}, r.createTaskErr
	}
	if task.ID == "" {
		task.ID = "task-created"
	}
	r.tasks[task.TenantID+"|"+task.ID] = task
	return task, nil
}

func (r *fakeRepository) DeleteInstance(_ context.Context, tenantID, instanceID string) error {
	delete(r.instances, tenantID+"|"+instanceID)
	return nil
}

func (r *fakeRepository) GetTaskByID(_ context.Context, tenantID, taskID string) (domainapproval.Task, error) {
	task, ok := r.tasks[tenantID+"|"+taskID]
	if !ok {
		return domainapproval.Task{}, domainapproval.ErrTaskNotFound
	}
	return task, nil
}

func (r *fakeRepository) UpdateTaskDecision(_ context.Context, tenantID, taskID string, status domainapproval.TaskStatus, decidedBy, comment string) error {
	task, ok := r.tasks[tenantID+"|"+taskID]
	if !ok {
		return domainapproval.ErrTaskNotFound
	}
	task.Status = status
	task.DecidedBy = decidedBy
	task.Comment = comment
	r.tasks[tenantID+"|"+taskID] = task
	return nil
}

func (r *fakeRepository) ListTasksByInstance(_ context.Context, tenantID, instanceID string) ([]domainapproval.Task, error) {
	tasks := make([]domainapproval.Task, 0)
	for _, task := range r.tasks {
		if task.TenantID == tenantID && task.InstanceID == instanceID {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}
