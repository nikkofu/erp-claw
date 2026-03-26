package integration

import (
	"context"
	"errors"
	"testing"

	appapproval "github.com/nikkofu/erp-claw/internal/application/approval"
	"github.com/nikkofu/erp-claw/internal/application/shared"
	domainapproval "github.com/nikkofu/erp-claw/internal/domain/approval"
	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestPipelineRequireApprovalStartsApprovalInstanceAndTask(t *testing.T) {
	repo := newIntegrationApprovalRepo(t)
	definition, err := domainapproval.NewDefinition("tenant-a", "def-a", "purchase approval", "manager-a", true)
	if err != nil {
		t.Fatalf("new definition: %v", err)
	}
	repo.definitions["tenant-a|def-a"] = definition

	auditStore := audit.NewInMemoryStore()
	auditService, err := audit.NewService(auditStore)
	if err != nil {
		t.Fatalf("new audit service: %v", err)
	}

	startHandler := appapproval.StartApprovalHandler{
		Definitions: repo,
		Instances:   repo,
		Tasks:       repo,
	}
	pipeline := shared.NewPipeline(shared.PipelineDeps{
		Policy:    policy.StaticEvaluator(policy.DecisionRequireApproval),
		Audit:     auditService,
		Approvals: appapproval.SharedCommandApprovalStarter{Handler: startHandler},
	})

	err = pipeline.Execute(context.Background(), shared.Command{
		Name:     "purchase.submit",
		TenantID: "tenant-a",
		ActorID:  "user-a",
		Payload: map[string]any{
			"approval_definition_id": "def-a",
			"resource_type":          "purchase_order",
			"resource_id":            "po-1",
		},
	})
	if !errors.Is(err, shared.ErrApprovalRequired) {
		t.Fatalf("expected approval required error, got %v", err)
	}

	if len(repo.instances) != 1 {
		t.Fatalf("expected one approval instance, got %d", len(repo.instances))
	}
	if len(repo.tasks) != 1 {
		t.Fatalf("expected one approval task, got %d", len(repo.tasks))
	}

	events, err := auditService.List(context.Background(), audit.Query{
		TenantID:    "tenant-a",
		CommandName: "purchase.submit",
		ActorID:     "user-a",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(events))
	}
	if events[0].Outcome != "pending_approval" {
		t.Fatalf("expected pending_approval outcome, got %s", events[0].Outcome)
	}
}

type integrationApprovalRepo struct {
	definitions map[string]domainapproval.Definition
	instances   map[string]domainapproval.Instance
	tasks       map[string]domainapproval.Task
}

func newIntegrationApprovalRepo(t *testing.T) *integrationApprovalRepo {
	t.Helper()
	return &integrationApprovalRepo{
		definitions: make(map[string]domainapproval.Definition),
		instances:   make(map[string]domainapproval.Instance),
		tasks:       make(map[string]domainapproval.Task),
	}
}

func (r *integrationApprovalRepo) SaveDefinition(_ context.Context, definition domainapproval.Definition) (domainapproval.Definition, error) {
	r.definitions[definition.TenantID+"|"+definition.ID] = definition
	return definition, nil
}

func (r *integrationApprovalRepo) GetDefinitionByID(_ context.Context, tenantID, definitionID string) (domainapproval.Definition, error) {
	definition, ok := r.definitions[tenantID+"|"+definitionID]
	if !ok {
		return domainapproval.Definition{}, domainapproval.ErrDefinitionNotFound
	}
	return definition, nil
}

func (r *integrationApprovalRepo) CreateInstance(_ context.Context, instance domainapproval.Instance) (domainapproval.Instance, error) {
	if instance.ID == "" {
		instance.ID = "inst-a"
	}
	r.instances[instance.TenantID+"|"+instance.ID] = instance
	return instance, nil
}

func (r *integrationApprovalRepo) GetInstanceByID(_ context.Context, tenantID, instanceID string) (domainapproval.Instance, error) {
	instance, ok := r.instances[tenantID+"|"+instanceID]
	if !ok {
		return domainapproval.Instance{}, domainapproval.ErrInstanceNotFound
	}
	return instance, nil
}

func (r *integrationApprovalRepo) UpdateInstanceStatus(_ context.Context, tenantID, instanceID string, status domainapproval.InstanceStatus) error {
	instance, ok := r.instances[tenantID+"|"+instanceID]
	if !ok {
		return domainapproval.ErrInstanceNotFound
	}
	instance.Status = status
	r.instances[tenantID+"|"+instanceID] = instance
	return nil
}

func (r *integrationApprovalRepo) DeleteInstance(_ context.Context, tenantID, instanceID string) error {
	delete(r.instances, tenantID+"|"+instanceID)
	return nil
}

func (r *integrationApprovalRepo) CreateTask(_ context.Context, task domainapproval.Task) (domainapproval.Task, error) {
	if task.ID == "" {
		task.ID = "task-a"
	}
	r.tasks[task.TenantID+"|"+task.ID] = task
	return task, nil
}

func (r *integrationApprovalRepo) GetTaskByID(_ context.Context, tenantID, taskID string) (domainapproval.Task, error) {
	task, ok := r.tasks[tenantID+"|"+taskID]
	if !ok {
		return domainapproval.Task{}, domainapproval.ErrTaskNotFound
	}
	return task, nil
}

func (r *integrationApprovalRepo) UpdateTaskDecision(_ context.Context, tenantID, taskID string, status domainapproval.TaskStatus, decidedBy, comment string) error {
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

func (r *integrationApprovalRepo) ListTasksByInstance(_ context.Context, tenantID, instanceID string) ([]domainapproval.Task, error) {
	tasks := make([]domainapproval.Task, 0)
	for _, task := range r.tasks {
		if task.TenantID == tenantID && task.InstanceID == instanceID {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}
