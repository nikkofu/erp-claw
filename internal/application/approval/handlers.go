package approval

import (
	"context"
	"errors"

	domainapproval "github.com/nikkofu/erp-claw/internal/domain/approval"
)

var (
	errStartApprovalDefinitionsRequired = errors.New("start approval handler requires definition repository")
	errStartApprovalInstancesRequired   = errors.New("start approval handler requires instance repository")
	errStartApprovalTasksRequired       = errors.New("start approval handler requires task repository")
	errApproveTaskInstancesRequired     = errors.New("approve task handler requires instance repository")
	errApproveTaskTasksRequired         = errors.New("approve task handler requires task repository")
	errRejectTaskInstancesRequired      = errors.New("reject task handler requires instance repository")
	errRejectTaskTasksRequired          = errors.New("reject task handler requires task repository")
)

type StartApproval struct {
	TenantID     string
	DefinitionID string
	ResourceType string
	ResourceID   string
	RequestedBy  string
}

type StartApprovalResult struct {
	Instance domainapproval.Instance
	Task     domainapproval.Task
}

type StartApprovalHandler struct {
	Definitions domainapproval.DefinitionRepository
	Instances   domainapproval.InstanceRepository
	Tasks       domainapproval.TaskRepository
}

func (h StartApprovalHandler) Handle(ctx context.Context, cmd StartApproval) (StartApprovalResult, error) {
	if h.Definitions == nil {
		return StartApprovalResult{}, errStartApprovalDefinitionsRequired
	}
	if h.Instances == nil {
		return StartApprovalResult{}, errStartApprovalInstancesRequired
	}
	if h.Tasks == nil {
		return StartApprovalResult{}, errStartApprovalTasksRequired
	}

	definition, err := h.Definitions.GetDefinitionByID(ctx, cmd.TenantID, cmd.DefinitionID)
	if err != nil {
		return StartApprovalResult{}, err
	}
	if !definition.Active {
		return StartApprovalResult{}, domainapproval.ErrDefinitionInactive
	}

	instance, err := domainapproval.NewInstance(cmd.TenantID, cmd.DefinitionID, cmd.ResourceType, cmd.ResourceID, cmd.RequestedBy)
	if err != nil {
		return StartApprovalResult{}, err
	}
	instance, err = h.Instances.CreateInstance(ctx, instance)
	if err != nil {
		return StartApprovalResult{}, err
	}

	task, err := domainapproval.NewTask(cmd.TenantID, instance.ID, definition.ApproverID)
	if err != nil {
		return StartApprovalResult{}, err
	}
	task, err = h.Tasks.CreateTask(ctx, task)
	if err != nil {
		return StartApprovalResult{}, err
	}

	return StartApprovalResult{Instance: instance, Task: task}, nil
}

type ApproveTask struct {
	TenantID string
	TaskID   string
	ActorID  string
	Comment  string
}

type ApproveTaskResult struct {
	Instance domainapproval.Instance
	Task     domainapproval.Task
}

type ApproveTaskHandler struct {
	Instances domainapproval.InstanceRepository
	Tasks     domainapproval.TaskRepository
}

func (h ApproveTaskHandler) Handle(ctx context.Context, cmd ApproveTask) (ApproveTaskResult, error) {
	if h.Instances == nil {
		return ApproveTaskResult{}, errApproveTaskInstancesRequired
	}
	if h.Tasks == nil {
		return ApproveTaskResult{}, errApproveTaskTasksRequired
	}

	task, err := h.Tasks.GetTaskByID(ctx, cmd.TenantID, cmd.TaskID)
	if err != nil {
		return ApproveTaskResult{}, err
	}
	if err := task.Decide(domainapproval.TaskStatusApproved, cmd.ActorID, cmd.Comment); err != nil {
		return ApproveTaskResult{}, err
	}
	if err := h.Tasks.UpdateTaskDecision(ctx, cmd.TenantID, cmd.TaskID, task.Status, task.DecidedBy, task.Comment); err != nil {
		return ApproveTaskResult{}, err
	}

	instance, err := h.Instances.GetInstanceByID(ctx, cmd.TenantID, task.InstanceID)
	if err != nil {
		return ApproveTaskResult{}, err
	}
	tasks, err := h.Tasks.ListTasksByInstance(ctx, cmd.TenantID, task.InstanceID)
	if err != nil {
		return ApproveTaskResult{}, err
	}

	allApproved := true
	for _, current := range tasks {
		if current.ID == task.ID {
			current = task
		}
		if current.Status != domainapproval.TaskStatusApproved {
			allApproved = false
			break
		}
	}
	if allApproved {
		if err := instance.TransitionTo(domainapproval.InstanceStatusApproved); err != nil {
			return ApproveTaskResult{}, err
		}
		if err := h.Instances.UpdateInstanceStatus(ctx, cmd.TenantID, instance.ID, instance.Status); err != nil {
			return ApproveTaskResult{}, err
		}
	}

	return ApproveTaskResult{Instance: instance, Task: task}, nil
}

type RejectTask struct {
	TenantID string
	TaskID   string
	ActorID  string
	Comment  string
}

type RejectTaskResult struct {
	Instance domainapproval.Instance
	Task     domainapproval.Task
}

type RejectTaskHandler struct {
	Instances domainapproval.InstanceRepository
	Tasks     domainapproval.TaskRepository
}

func (h RejectTaskHandler) Handle(ctx context.Context, cmd RejectTask) (RejectTaskResult, error) {
	if h.Instances == nil {
		return RejectTaskResult{}, errRejectTaskInstancesRequired
	}
	if h.Tasks == nil {
		return RejectTaskResult{}, errRejectTaskTasksRequired
	}

	task, err := h.Tasks.GetTaskByID(ctx, cmd.TenantID, cmd.TaskID)
	if err != nil {
		return RejectTaskResult{}, err
	}
	if err := task.Decide(domainapproval.TaskStatusRejected, cmd.ActorID, cmd.Comment); err != nil {
		return RejectTaskResult{}, err
	}
	if err := h.Tasks.UpdateTaskDecision(ctx, cmd.TenantID, cmd.TaskID, task.Status, task.DecidedBy, task.Comment); err != nil {
		return RejectTaskResult{}, err
	}

	instance, err := h.Instances.GetInstanceByID(ctx, cmd.TenantID, task.InstanceID)
	if err != nil {
		return RejectTaskResult{}, err
	}
	if err := instance.TransitionTo(domainapproval.InstanceStatusRejected); err != nil {
		return RejectTaskResult{}, err
	}
	if err := h.Instances.UpdateInstanceStatus(ctx, cmd.TenantID, instance.ID, instance.Status); err != nil {
		return RejectTaskResult{}, err
	}

	return RejectTaskResult{Instance: instance, Task: task}, nil
}
