package approval

import (
	"context"
	"errors"

	domainapproval "github.com/nikkofu/erp-claw/internal/domain/approval"
)

var (
	errSaveDefinitionHandlerDefinitionRepositoryRequired  = errors.New("save definition handler requires definition repository")
	errListDefinitionsHandlerDefinitionRepositoryRequired = errors.New("list definitions handler requires definition repository")
	errListInstancesHandlerInstanceRepositoryRequired     = errors.New("list instances handler requires instance repository")
	errListTasksHandlerTaskRepositoryRequired             = errors.New("list tasks handler requires task repository")
)

type SaveDefinition struct {
	TenantID   string
	ID         string
	Name       string
	ApproverID string
	Active     bool
}

type SaveDefinitionHandler struct {
	Definitions domainapproval.DefinitionRepository
}

func (h SaveDefinitionHandler) Handle(ctx context.Context, cmd SaveDefinition) (domainapproval.Definition, error) {
	if h.Definitions == nil {
		return domainapproval.Definition{}, errSaveDefinitionHandlerDefinitionRepositoryRequired
	}

	definition, err := domainapproval.NewDefinition(cmd.TenantID, cmd.ID, cmd.Name, cmd.ApproverID, cmd.Active)
	if err != nil {
		return domainapproval.Definition{}, err
	}

	return h.Definitions.SaveDefinition(ctx, definition)
}

type ListDefinitions struct {
	TenantID string
}

type ListDefinitionsHandler struct {
	Definitions domainapproval.DefinitionRepository
}

func (h ListDefinitionsHandler) Handle(ctx context.Context, q ListDefinitions) ([]domainapproval.Definition, error) {
	if h.Definitions == nil {
		return nil, errListDefinitionsHandlerDefinitionRepositoryRequired
	}

	return h.Definitions.ListDefinitions(ctx, q.TenantID)
}

type ListInstances struct {
	TenantID string
}

type ListInstancesHandler struct {
	Instances domainapproval.InstanceRepository
}

func (h ListInstancesHandler) Handle(ctx context.Context, q ListInstances) ([]domainapproval.Instance, error) {
	if h.Instances == nil {
		return nil, errListInstancesHandlerInstanceRepositoryRequired
	}

	return h.Instances.ListInstances(ctx, q.TenantID)
}

type ListTasks struct {
	TenantID string
}

type ListTasksHandler struct {
	Tasks domainapproval.TaskRepository
}

func (h ListTasksHandler) Handle(ctx context.Context, q ListTasks) ([]domainapproval.Task, error) {
	if h.Tasks == nil {
		return nil, errListTasksHandlerTaskRepositoryRequired
	}

	return h.Tasks.ListTasks(ctx, q.TenantID)
}
