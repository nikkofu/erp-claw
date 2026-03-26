package runtime

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidSession           = errors.New("invalid agent session")
	ErrInvalidSessionTransition = errors.New("invalid session transition")
	ErrInvalidTask              = errors.New("invalid agent task")
	ErrInvalidTaskTransition    = errors.New("invalid task transition")
)

type SessionStatus string

const (
	SessionStatusOpen   SessionStatus = "open"
	SessionStatusClosed SessionStatus = "closed"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSucceeded TaskStatus = "succeeded"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCanceled  TaskStatus = "canceled"
)

// AgentSession describes a workspace session for command/task streaming.
type Session struct {
	ID        string
	TenantID  string
	ActorID   string
	Status    SessionStatus
	Metadata  map[string]any
	StartedAt time.Time
	EndedAt   time.Time
}

func NewSession(id, tenantID, actorID string, metadata map[string]any, now time.Time) (Session, error) {
	id = strings.TrimSpace(id)
	tenantID = strings.TrimSpace(tenantID)
	actorID = strings.TrimSpace(actorID)
	if id == "" || tenantID == "" || actorID == "" {
		return Session{}, ErrInvalidSession
	}

	return Session{
		ID:        id,
		TenantID:  tenantID,
		ActorID:   actorID,
		Status:    SessionStatusOpen,
		Metadata:  cloneMap(metadata),
		StartedAt: normalizeNow(now),
	}, nil
}

func (s *Session) Close(at time.Time) error {
	if s.Status != SessionStatusOpen {
		return ErrInvalidSessionTransition
	}
	s.Status = SessionStatusClosed
	s.EndedAt = normalizeNow(at)
	return nil
}

// Task describes one executable unit associated with an agent session.
type Task struct {
	ID            string
	TenantID      string
	SessionID     string
	Type          string
	Status        TaskStatus
	Input         map[string]any
	Output        map[string]any
	FailureReason string
	Attempts      int
	QueuedAt      time.Time
	StartedAt     time.Time
	CompletedAt   time.Time
}

func NewTask(id, tenantID, sessionID, taskType string, input map[string]any, now time.Time) (Task, error) {
	id = strings.TrimSpace(id)
	tenantID = strings.TrimSpace(tenantID)
	sessionID = strings.TrimSpace(sessionID)
	taskType = strings.TrimSpace(taskType)
	if id == "" || tenantID == "" || sessionID == "" || taskType == "" {
		return Task{}, ErrInvalidTask
	}

	return Task{
		ID:        id,
		TenantID:  tenantID,
		SessionID: sessionID,
		Type:      taskType,
		Status:    TaskStatusPending,
		Input:     cloneMap(input),
		QueuedAt:  normalizeNow(now),
	}, nil
}

func (t *Task) Start(at time.Time) error {
	if t.Status != TaskStatusPending {
		return ErrInvalidTaskTransition
	}
	t.Status = TaskStatusRunning
	t.Attempts++
	t.StartedAt = normalizeNow(at)
	return nil
}

func (t *Task) Complete(output map[string]any, at time.Time) error {
	if t.Status != TaskStatusRunning {
		return ErrInvalidTaskTransition
	}
	t.Status = TaskStatusSucceeded
	t.Output = cloneMap(output)
	t.FailureReason = ""
	t.CompletedAt = normalizeNow(at)
	return nil
}

func (t *Task) Fail(reason string, at time.Time) error {
	if t.Status != TaskStatusRunning {
		return ErrInvalidTaskTransition
	}
	t.Status = TaskStatusFailed
	t.FailureReason = strings.TrimSpace(reason)
	t.CompletedAt = normalizeNow(at)
	return nil
}

func (t *Task) Cancel(reason string, at time.Time) error {
	if t.Status != TaskStatusPending && t.Status != TaskStatusRunning {
		return ErrInvalidTaskTransition
	}
	t.Status = TaskStatusCanceled
	t.FailureReason = strings.TrimSpace(reason)
	t.CompletedAt = normalizeNow(at)
	return nil
}

func (t *Task) Retry(at time.Time) error {
	if t.Status != TaskStatusFailed && t.Status != TaskStatusCanceled {
		return ErrInvalidTaskTransition
	}
	t.Status = TaskStatusPending
	t.Output = map[string]any{}
	t.FailureReason = ""
	t.QueuedAt = normalizeNow(at)
	t.StartedAt = time.Time{}
	t.CompletedAt = time.Time{}
	return nil
}

func normalizeNow(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now().UTC()
	}
	return ts.UTC()
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
