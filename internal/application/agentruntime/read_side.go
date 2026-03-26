package agentruntime

import (
	"context"
	"errors"

	domain "github.com/nikkofu/erp-claw/internal/domain/agentruntime"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
)

var (
	errListSessionsHandlerSessionReaderRequired            = errors.New("list sessions handler requires session reader")
	errListTasksHandlerTaskReaderRequired                  = errors.New("list tasks handler requires task reader")
	errReplayWorkspaceEventsHandlerWorkspaceReaderRequired = errors.New("replay workspace events handler requires workspace event reader")
)

type SessionReader interface {
	ListSessions(ctx context.Context, tenantID string) ([]domain.Session, error)
}

type TaskReader interface {
	ListTasks(ctx context.Context, tenantID, sessionID string) ([]domain.Task, error)
}

type WorkspaceEventReader interface {
	ListWorkspaceEvents(ctx context.Context, tenantID, sessionID string) ([]platformruntime.WorkspaceEvent, error)
}

type ListSessions struct {
	TenantID string
}

type ListSessionsHandler struct {
	Sessions  SessionReader
	Authorize func(context.Context, ListSessions) error
}

func (h ListSessionsHandler) Handle(ctx context.Context, q ListSessions) ([]domain.Session, error) {
	if h.Sessions == nil {
		return nil, errListSessionsHandlerSessionReaderRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, q); err != nil {
			return nil, err
		}
	}

	return h.Sessions.ListSessions(ctx, q.TenantID)
}

type ListTasks struct {
	TenantID  string
	SessionID string
}

type ListTasksHandler struct {
	Tasks     TaskReader
	Authorize func(context.Context, ListTasks) error
}

func (h ListTasksHandler) Handle(ctx context.Context, q ListTasks) ([]domain.Task, error) {
	if h.Tasks == nil {
		return nil, errListTasksHandlerTaskReaderRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, q); err != nil {
			return nil, err
		}
	}

	return h.Tasks.ListTasks(ctx, q.TenantID, q.SessionID)
}

type ReplayWorkspaceEvents struct {
	TenantID  string
	SessionID string
}

type ReplayWorkspaceEventsHandler struct {
	Events    WorkspaceEventReader
	Authorize func(context.Context, ReplayWorkspaceEvents) error
}

func (h ReplayWorkspaceEventsHandler) Handle(ctx context.Context, q ReplayWorkspaceEvents) ([]platformruntime.WorkspaceEvent, error) {
	if h.Events == nil {
		return nil, errReplayWorkspaceEventsHandlerWorkspaceReaderRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, q); err != nil {
			return nil, err
		}
	}

	return h.Events.ListWorkspaceEvents(ctx, q.TenantID, q.SessionID)
}
