package agentruntime

import (
	"errors"
	"strings"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSucceeded TaskStatus = "succeeded"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCanceled  TaskStatus = "canceled"
)

var (
	errTaskTenantIDRequired        = errors.New("agent runtime task tenant id is required")
	errTaskTypeRequired            = errors.New("agent runtime task type is required")
	errTaskInvalidStatusTransition = errors.New("agent runtime task status transition is invalid")
)

type Task struct {
	ID          string
	TenantID    string
	SessionID   string
	TaskType    string
	Status      TaskStatus
	Input       map[string]any
	Output      map[string]any
	Attempts    int
	QueuedAt    time.Time
	CompletedAt *time.Time
}

func NewTask(tenantID, sessionID, taskType string, input map[string]any) (Task, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Task{}, errTaskTenantIDRequired
	}
	if strings.TrimSpace(taskType) == "" {
		return Task{}, errTaskTypeRequired
	}

	return Task{
		TenantID:  tenantID,
		SessionID: sessionID,
		TaskType:  taskType,
		Status:    TaskStatusPending,
		Input:     copyAnyMap(input),
		Output:    map[string]any{},
	}, nil
}

func (t *Task) TransitionTo(next TaskStatus, output map[string]any, completedAt time.Time) error {
	if !isValidTaskStatus(next) {
		return errTaskInvalidStatusTransition
	}
	if t.Status == next {
		return nil
	}
	if !canTransitionTaskStatus(t.Status, next) {
		return errTaskInvalidStatusTransition
	}

	t.Status = next
	if output != nil {
		t.Output = copyAnyMap(output)
	}

	switch next {
	case TaskStatusSucceeded, TaskStatusFailed, TaskStatusCanceled:
		if completedAt.IsZero() {
			completedAt = time.Now().UTC()
		}
		t.CompletedAt = &completedAt
	default:
		t.CompletedAt = nil
	}

	return nil
}

func isValidTaskStatus(status TaskStatus) bool {
	switch status {
	case TaskStatusPending, TaskStatusRunning, TaskStatusSucceeded, TaskStatusFailed, TaskStatusCanceled:
		return true
	default:
		return false
	}
}

func canTransitionTaskStatus(current, next TaskStatus) bool {
	switch current {
	case TaskStatusPending:
		return next == TaskStatusRunning || next == TaskStatusFailed || next == TaskStatusCanceled
	case TaskStatusRunning:
		return next == TaskStatusSucceeded || next == TaskStatusFailed || next == TaskStatusCanceled
	default:
		return false
	}
}

func copyAnyMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}

	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
