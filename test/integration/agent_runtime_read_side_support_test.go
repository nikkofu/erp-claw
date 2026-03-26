package integration

import (
	"context"

	domain "github.com/nikkofu/erp-claw/internal/domain/agentruntime"
)

func (r *memorySessionRepository) ListSessions(_ context.Context, tenantID string) ([]domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]domain.Session, 0, len(r.data))
	for _, session := range r.data {
		if session.TenantID != tenantID {
			continue
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (r *memoryTaskRepository) ListTasks(_ context.Context, tenantID, sessionID string) ([]domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]domain.Task, 0, len(r.data))
	for _, task := range r.data {
		if task.TenantID != tenantID {
			continue
		}
		if sessionID != "" && task.SessionID != sessionID {
			continue
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
