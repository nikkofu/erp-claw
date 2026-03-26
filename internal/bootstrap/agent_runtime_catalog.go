package bootstrap

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	agentruntimeapp "github.com/nikkofu/erp-claw/internal/application/agentruntime"
	domain "github.com/nikkofu/erp-claw/internal/domain/agentruntime"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/postgres"
)

func newAgentRuntimeCatalog(cfg Config) AgentRuntimeCatalog {
	if shouldUseInMemoryCatalogFallback(cfg) {
		return NewInMemoryAgentRuntimeCatalogForTest()
	}

	catalog, err := newPostgresAgentRuntimeCatalog(cfg.Database)
	if err == nil {
		return catalog
	}

	panic(fmt.Errorf("bootstrap: agent runtime catalog init failed: %w", err))
}

func newPostgresAgentRuntimeCatalog(cfg DatabaseConfig) (AgentRuntimeCatalog, error) {
	db, err := postgres.New(postgres.Config{
		DSN:          cfg.DSN,
		MaxOpenConns: cfg.MaxOpenConns,
		MaxIdleConns: cfg.MaxIdleConns,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	repo, err := postgres.NewAgentRuntimeRepository(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return repo, nil
}

type inMemoryAgentRuntimeCatalog struct {
	mu            sync.RWMutex
	nextSessionID int64
	nextTaskID    int64
	sessions      map[string]domain.Session
	tasks         map[string]domain.Task
}

var (
	_ AgentRuntimeCatalog           = (*inMemoryAgentRuntimeCatalog)(nil)
	_ agentruntimeapp.SessionReader = (*inMemoryAgentRuntimeCatalog)(nil)
	_ agentruntimeapp.TaskReader    = (*inMemoryAgentRuntimeCatalog)(nil)
)

func NewInMemoryAgentRuntimeCatalogForTest() *inMemoryAgentRuntimeCatalog {
	return &inMemoryAgentRuntimeCatalog{
		sessions: make(map[string]domain.Session),
		tasks:    make(map[string]domain.Task),
	}
}

func (r *inMemoryAgentRuntimeCatalog) CreateSession(_ context.Context, session domain.Session) (domain.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(session.ID) == "" {
		r.nextSessionID++
		session.ID = strconv.FormatInt(r.nextSessionID, 10)
	}

	r.sessions[r.sessionKey(session.TenantID, session.SessionKey)] = session
	return session, nil
}

func (r *inMemoryAgentRuntimeCatalog) GetSessionByTenantAndKey(_ context.Context, tenantID, sessionKey string) (domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[r.sessionKey(tenantID, sessionKey)]
	if !ok {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	return session, nil
}

func (r *inMemoryAgentRuntimeCatalog) UpdateSessionStatus(_ context.Context, tenantID, sessionKey string, status domain.SessionStatus, endedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.sessionKey(tenantID, sessionKey)
	session, ok := r.sessions[key]
	if !ok {
		return domain.ErrSessionNotFound
	}
	session.Status = status
	session.EndedAt = endedAt
	r.sessions[key] = session
	return nil
}

func (r *inMemoryAgentRuntimeCatalog) ListSessions(_ context.Context, tenantID string) ([]domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]domain.Session, 0, len(r.sessions))
	for _, session := range r.sessions {
		if session.TenantID != tenantID {
			continue
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (r *inMemoryAgentRuntimeCatalog) CreateTask(_ context.Context, task domain.Task) (domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(task.ID) == "" {
		r.nextTaskID++
		task.ID = strconv.FormatInt(r.nextTaskID, 10)
	}

	r.tasks[r.taskKey(task.TenantID, task.ID)] = task
	return task, nil
}

func (r *inMemoryAgentRuntimeCatalog) GetTaskByID(_ context.Context, tenantID, taskID string) (domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[r.taskKey(tenantID, taskID)]
	if !ok {
		return domain.Task{}, domain.ErrTaskNotFound
	}
	return task, nil
}

func (r *inMemoryAgentRuntimeCatalog) UpdateTaskStatus(_ context.Context, tenantID, taskID string, status domain.TaskStatus, output map[string]any, completedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.taskKey(tenantID, taskID)
	task, ok := r.tasks[key]
	if !ok {
		return domain.ErrTaskNotFound
	}
	task.Status = status
	task.Output = output
	task.CompletedAt = completedAt
	r.tasks[key] = task
	return nil
}

func (r *inMemoryAgentRuntimeCatalog) ListTasks(_ context.Context, tenantID, sessionID string) ([]domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]domain.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
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

func (r *inMemoryAgentRuntimeCatalog) sessionKey(tenantID, sessionKey string) string {
	return tenantID + "|" + sessionKey
}

func (r *inMemoryAgentRuntimeCatalog) taskKey(tenantID, taskID string) string {
	return tenantID + "|" + taskID
}
