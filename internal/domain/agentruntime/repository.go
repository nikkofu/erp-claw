package agentruntime

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSessionNotFound = errors.New("agent runtime session not found")
	ErrTaskNotFound    = errors.New("agent runtime task not found")
)

type SessionRepository interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
	GetSessionByTenantAndKey(ctx context.Context, tenantID, sessionKey string) (Session, error)
	UpdateSessionStatus(ctx context.Context, tenantID, sessionKey string, status SessionStatus, endedAt *time.Time) error
}

type TaskRepository interface {
	CreateTask(ctx context.Context, task Task) (Task, error)
	GetTaskByID(ctx context.Context, tenantID, taskID string) (Task, error)
	UpdateTaskStatus(ctx context.Context, tenantID, taskID string, status TaskStatus, output map[string]any, completedAt *time.Time) error
}
