package runtime

import (
	"context"
	"errors"
)

var (
	ErrSessionNotFound = errors.New("agent session not found")
	ErrTaskNotFound    = errors.New("agent task not found")
)

type SessionRepository interface {
	Save(ctx context.Context, session Session) error
	Get(ctx context.Context, tenantID, sessionID string) (Session, error)
	List(ctx context.Context, query SessionListQuery) (SessionListPage, error)
}

type TaskRepository interface {
	Save(ctx context.Context, task Task) error
	Get(ctx context.Context, tenantID, taskID string) (Task, error)
	ListBySession(ctx context.Context, tenantID, sessionID string) ([]Task, error)
	List(ctx context.Context, query TaskListQuery) (TaskListPage, error)
}
