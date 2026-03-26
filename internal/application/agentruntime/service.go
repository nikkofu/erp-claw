package agentruntime

import (
	"context"
	"errors"
	"time"

	domain "github.com/nikkofu/erp-claw/internal/domain/agentruntime"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
)

var (
	errAgentRuntimeServiceSessionRepositoryRequired = errors.New("agent runtime service requires session repository")
	errAgentRuntimeServiceTaskRepositoryRequired    = errors.New("agent runtime service requires task repository")
)

type WorkspaceEventAppender interface {
	AppendWorkspaceEvent(ctx context.Context, evt platformruntime.WorkspaceEvent) error
}

type ServiceDeps struct {
	Sessions domain.SessionRepository
	Tasks    domain.TaskRepository
	Events   WorkspaceEventAppender
}

type Service struct {
	sessions domain.SessionRepository
	tasks    domain.TaskRepository
	events   WorkspaceEventAppender
}

func NewService(deps ServiceDeps) (*Service, error) {
	if deps.Sessions == nil {
		return nil, errAgentRuntimeServiceSessionRepositoryRequired
	}
	if deps.Tasks == nil {
		return nil, errAgentRuntimeServiceTaskRepositoryRequired
	}

	return &Service{
		sessions: deps.Sessions,
		tasks:    deps.Tasks,
		events:   deps.Events,
	}, nil
}

func (s *Service) CreateSession(ctx context.Context, tenantID, sessionKey string, metadata map[string]any) (domain.Session, error) {
	session, err := domain.NewSession(tenantID, sessionKey, metadata)
	if err != nil {
		return domain.Session{}, err
	}

	return s.sessions.CreateSession(ctx, session)
}

func (s *Service) CreateTask(ctx context.Context, tenantID, sessionID, taskType string, input map[string]any) (domain.Task, error) {
	task, err := domain.NewTask(tenantID, sessionID, taskType, input)
	if err != nil {
		return domain.Task{}, err
	}

	return s.tasks.CreateTask(ctx, task)
}

func (s *Service) StartTask(ctx context.Context, tenantID, taskID string) (domain.Task, error) {
	task, err := s.tasks.GetTaskByID(ctx, tenantID, taskID)
	if err != nil {
		return domain.Task{}, err
	}

	if err := task.TransitionTo(domain.TaskStatusRunning, nil, time.Time{}); err != nil {
		return domain.Task{}, err
	}
	if err := s.tasks.UpdateTaskStatus(ctx, tenantID, taskID, task.Status, task.Output, task.CompletedAt); err != nil {
		return domain.Task{}, err
	}

	if err := s.appendTaskStatusChanged(ctx, task); err != nil {
		return domain.Task{}, err
	}

	return task, nil
}

func (s *Service) CompleteTask(ctx context.Context, tenantID, taskID string, output map[string]any) (domain.Task, error) {
	task, err := s.tasks.GetTaskByID(ctx, tenantID, taskID)
	if err != nil {
		return domain.Task{}, err
	}

	if err := task.TransitionTo(domain.TaskStatusSucceeded, output, time.Now().UTC()); err != nil {
		return domain.Task{}, err
	}
	if err := s.tasks.UpdateTaskStatus(ctx, tenantID, taskID, task.Status, task.Output, task.CompletedAt); err != nil {
		return domain.Task{}, err
	}

	if err := s.appendTaskStatusChanged(ctx, task); err != nil {
		return domain.Task{}, err
	}

	return task, nil
}

func (s *Service) FailTask(ctx context.Context, tenantID, taskID string, output map[string]any) (domain.Task, error) {
	task, err := s.tasks.GetTaskByID(ctx, tenantID, taskID)
	if err != nil {
		return domain.Task{}, err
	}

	if err := task.TransitionTo(domain.TaskStatusFailed, output, time.Now().UTC()); err != nil {
		return domain.Task{}, err
	}
	if err := s.tasks.UpdateTaskStatus(ctx, tenantID, taskID, task.Status, task.Output, task.CompletedAt); err != nil {
		return domain.Task{}, err
	}

	if err := s.appendTaskStatusChanged(ctx, task); err != nil {
		return domain.Task{}, err
	}

	return task, nil
}

func (s *Service) CancelTask(ctx context.Context, tenantID, taskID string, output map[string]any) (domain.Task, error) {
	task, err := s.tasks.GetTaskByID(ctx, tenantID, taskID)
	if err != nil {
		return domain.Task{}, err
	}

	if err := task.TransitionTo(domain.TaskStatusCanceled, output, time.Now().UTC()); err != nil {
		return domain.Task{}, err
	}
	if err := s.tasks.UpdateTaskStatus(ctx, tenantID, taskID, task.Status, task.Output, task.CompletedAt); err != nil {
		return domain.Task{}, err
	}

	if err := s.appendTaskStatusChanged(ctx, task); err != nil {
		return domain.Task{}, err
	}

	return task, nil
}

func (s *Service) CloseSession(ctx context.Context, tenantID, sessionKey string) (domain.Session, error) {
	session, err := s.sessions.GetSessionByTenantAndKey(ctx, tenantID, sessionKey)
	if err != nil {
		return domain.Session{}, err
	}

	if err := session.TransitionTo(domain.SessionStatusClosed, time.Now().UTC()); err != nil {
		return domain.Session{}, err
	}
	if err := s.sessions.UpdateSessionStatus(ctx, tenantID, sessionKey, session.Status, session.EndedAt); err != nil {
		return domain.Session{}, err
	}

	if err := s.appendSessionStatusChanged(ctx, session); err != nil {
		return domain.Session{}, err
	}

	return session, nil
}

func (s *Service) appendTaskStatusChanged(ctx context.Context, task domain.Task) error {
	if s.events == nil {
		return nil
	}

	return s.events.AppendWorkspaceEvent(ctx, platformruntime.WorkspaceEvent{
		Type:       platformruntime.WorkspaceEventTypeTaskStatusChanged,
		TenantID:   task.TenantID,
		SessionID:  task.SessionID,
		TaskID:     task.ID,
		Payload:    map[string]any{"status": string(task.Status)},
		OccurredAt: time.Now().UTC(),
	})
}

func (s *Service) appendSessionStatusChanged(ctx context.Context, session domain.Session) error {
	if s.events == nil {
		return nil
	}

	return s.events.AppendWorkspaceEvent(ctx, platformruntime.WorkspaceEvent{
		Type:       platformruntime.WorkspaceEventTypeSessionStatusChanged,
		TenantID:   session.TenantID,
		SessionID:  session.SessionKey,
		Payload:    map[string]any{"status": string(session.Status)},
		OccurredAt: time.Now().UTC(),
	})
}
