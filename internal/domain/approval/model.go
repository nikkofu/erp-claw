package approval

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrTenantIDRequired             = errors.New("approval tenant id is required")
	ErrDefinitionIDRequired         = errors.New("approval definition id is required")
	ErrDefinitionNameRequired       = errors.New("approval definition name is required")
	ErrDefinitionApproverRequired   = errors.New("approval definition approver id is required")
	ErrDefinitionNotFound           = errors.New("approval definition not found")
	ErrDefinitionInactive           = errors.New("approval definition is inactive")
	ErrInstanceNotFound             = errors.New("approval instance not found")
	ErrInstanceDefinitionRequired   = errors.New("approval instance definition id is required")
	ErrInstanceResourceTypeRequired = errors.New("approval instance resource type is required")
	ErrInstanceResourceIDRequired   = errors.New("approval instance resource id is required")
	ErrInstanceRequesterRequired    = errors.New("approval instance requested by is required")
	ErrTaskNotFound                 = errors.New("approval task not found")
	ErrTaskInstanceRequired         = errors.New("approval task instance id is required")
	ErrTaskApproverRequired         = errors.New("approval task approver id is required")
	ErrDecisionActorMismatch        = errors.New("approval decision actor does not match approver")
	errInvalidInstanceTransition    = errors.New("approval instance status transition is invalid")
	errInvalidTaskTransition        = errors.New("approval task status transition is invalid")
)

type Definition struct {
	TenantID   string
	ID         string
	Name       string
	ApproverID string
	Active     bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type InstanceStatus string

const (
	InstanceStatusPending  InstanceStatus = "pending"
	InstanceStatusApproved InstanceStatus = "approved"
	InstanceStatusRejected InstanceStatus = "rejected"
)

type Instance struct {
	TenantID     string
	ID           string
	DefinitionID string
	ResourceType string
	ResourceID   string
	RequestedBy  string
	Status       InstanceStatus
	CreatedAt    time.Time
	DecidedAt    *time.Time
}

type TaskStatus string

const (
	TaskStatusPending  TaskStatus = "pending"
	TaskStatusApproved TaskStatus = "approved"
	TaskStatusRejected TaskStatus = "rejected"
)

type Task struct {
	TenantID   string
	ID         string
	InstanceID string
	ApproverID string
	Status     TaskStatus
	DecidedBy  string
	Comment    string
	CreatedAt  time.Time
	DecidedAt  *time.Time
}

func NewDefinition(tenantID, id, name, approverID string, active bool) (Definition, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Definition{}, ErrTenantIDRequired
	}
	if strings.TrimSpace(id) == "" {
		return Definition{}, ErrDefinitionIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return Definition{}, ErrDefinitionNameRequired
	}
	if strings.TrimSpace(approverID) == "" {
		return Definition{}, ErrDefinitionApproverRequired
	}

	now := time.Now().UTC()
	return Definition{
		TenantID:   tenantID,
		ID:         id,
		Name:       name,
		ApproverID: approverID,
		Active:     active,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func NewInstance(tenantID, definitionID, resourceType, resourceID, requestedBy string) (Instance, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Instance{}, ErrTenantIDRequired
	}
	if strings.TrimSpace(definitionID) == "" {
		return Instance{}, ErrInstanceDefinitionRequired
	}
	if strings.TrimSpace(resourceType) == "" {
		return Instance{}, ErrInstanceResourceTypeRequired
	}
	if strings.TrimSpace(resourceID) == "" {
		return Instance{}, ErrInstanceResourceIDRequired
	}
	if strings.TrimSpace(requestedBy) == "" {
		return Instance{}, ErrInstanceRequesterRequired
	}

	return Instance{
		TenantID:     tenantID,
		DefinitionID: definitionID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		RequestedBy:  requestedBy,
		Status:       InstanceStatusPending,
		CreatedAt:    time.Now().UTC(),
	}, nil
}

func (i *Instance) TransitionTo(next InstanceStatus) error {
	if i.Status == next {
		return nil
	}
	if i.Status != InstanceStatusPending {
		return errInvalidInstanceTransition
	}
	if next != InstanceStatusApproved && next != InstanceStatusRejected {
		return errInvalidInstanceTransition
	}

	now := time.Now().UTC()
	i.Status = next
	i.DecidedAt = &now
	return nil
}

func NewTask(tenantID, instanceID, approverID string) (Task, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Task{}, ErrTenantIDRequired
	}
	if strings.TrimSpace(instanceID) == "" {
		return Task{}, ErrTaskInstanceRequired
	}
	if strings.TrimSpace(approverID) == "" {
		return Task{}, ErrTaskApproverRequired
	}

	return Task{
		TenantID:   tenantID,
		InstanceID: instanceID,
		ApproverID: approverID,
		Status:     TaskStatusPending,
		CreatedAt:  time.Now().UTC(),
	}, nil
}

func (t *Task) Decide(status TaskStatus, decidedBy, comment string) error {
	if t.Status != TaskStatusPending {
		return errInvalidTaskTransition
	}
	if status != TaskStatusApproved && status != TaskStatusRejected {
		return errInvalidTaskTransition
	}
	if strings.TrimSpace(decidedBy) == "" || strings.TrimSpace(decidedBy) != strings.TrimSpace(t.ApproverID) {
		return ErrDecisionActorMismatch
	}

	now := time.Now().UTC()
	t.Status = status
	t.DecidedBy = strings.TrimSpace(decidedBy)
	t.Comment = comment
	t.DecidedAt = &now
	return nil
}

type DefinitionRepository interface {
	SaveDefinition(ctx context.Context, definition Definition) (Definition, error)
	GetDefinitionByID(ctx context.Context, tenantID, definitionID string) (Definition, error)
}

type InstanceRepository interface {
	CreateInstance(ctx context.Context, instance Instance) (Instance, error)
	GetInstanceByID(ctx context.Context, tenantID, instanceID string) (Instance, error)
	UpdateInstanceStatus(ctx context.Context, tenantID, instanceID string, status InstanceStatus) error
}

type TaskRepository interface {
	CreateTask(ctx context.Context, task Task) (Task, error)
	GetTaskByID(ctx context.Context, tenantID, taskID string) (Task, error)
	UpdateTaskDecision(ctx context.Context, tenantID, taskID string, status TaskStatus, decidedBy, comment string) error
	ListTasksByInstance(ctx context.Context, tenantID, instanceID string) ([]Task, error)
}
