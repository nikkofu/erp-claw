package memory

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nikkofu/erp-claw/internal/platform/iam"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
	"github.com/nikkofu/erp-claw/internal/platform/tenant"
)

type ControlPlaneStore struct {
	mu sync.RWMutex

	tenants     map[string]tenant.Tenant
	actors      map[string]iam.Actor
	sessions    map[string]platformruntime.Session
	tasks       map[string]platformruntime.Task
	sessionTask map[string][]string
}

func NewControlPlaneStore() *ControlPlaneStore {
	return &ControlPlaneStore{
		tenants:     make(map[string]tenant.Tenant),
		actors:      make(map[string]iam.Actor),
		sessions:    make(map[string]platformruntime.Session),
		tasks:       make(map[string]platformruntime.Task),
		sessionTask: make(map[string][]string),
	}
}

func (s *ControlPlaneStore) TenantCatalog() tenant.Catalog {
	return tenantCatalogRepository{s}
}

func (s *ControlPlaneStore) IAMDirectory() iam.Directory {
	return iamDirectoryRepository{s}
}

func (s *ControlPlaneStore) SessionRepository() platformruntime.SessionRepository {
	return sessionRepository{s}
}

func (s *ControlPlaneStore) TaskRepository() platformruntime.TaskRepository {
	return taskRepository{s}
}

type tenantCatalogRepository struct {
	store *ControlPlaneStore
}

func (r tenantCatalogRepository) Save(_ context.Context, value tenant.Tenant) error {
	if value.Status == "" {
		value.Status = tenant.TenantStatusActive
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.tenants[value.Code] = value
	return nil
}

func (r tenantCatalogRepository) Get(_ context.Context, code string) (tenant.Tenant, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	value, ok := r.store.tenants[code]
	if !ok {
		return tenant.Tenant{}, tenant.ErrTenantNotFound
	}
	return value, nil
}

type iamDirectoryRepository struct {
	store *ControlPlaneStore
}

func (r iamDirectoryRepository) Save(_ context.Context, tenantID string, actor iam.Actor) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.actors[key(tenantID, actor.ID)] = cloneActor(actor)
	return nil
}

func (r iamDirectoryRepository) Get(_ context.Context, tenantID, actorID string) (iam.Actor, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	actor, ok := r.store.actors[key(tenantID, actorID)]
	if !ok {
		return iam.Actor{}, iam.ErrActorNotFound
	}
	return cloneActor(actor), nil
}

type sessionRepository struct {
	store *ControlPlaneStore
}

func (r sessionRepository) Save(_ context.Context, session platformruntime.Session) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.sessions[key(session.TenantID, session.ID)] = cloneSession(session)
	return nil
}

func (r sessionRepository) Get(_ context.Context, tenantID, sessionID string) (platformruntime.Session, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	session, ok := r.store.sessions[key(tenantID, sessionID)]
	if !ok {
		return platformruntime.Session{}, platformruntime.ErrSessionNotFound
	}
	return cloneSession(session), nil
}

func (r sessionRepository) List(_ context.Context, query platformruntime.SessionListQuery) (platformruntime.SessionListPage, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	items := make([]platformruntime.Session, 0)
	for _, session := range r.store.sessions {
		if query.TenantID != "" && session.TenantID != query.TenantID {
			continue
		}
		if query.ActorID != "" && session.ActorID != query.ActorID {
			continue
		}
		if query.Status != "" && session.Status != query.Status {
			continue
		}
		items = append(items, cloneSession(session))
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].StartedAt.Equal(items[j].StartedAt) {
			return items[i].ID < items[j].ID
		}
		return items[i].StartedAt.Before(items[j].StartedAt)
	})

	limit := query.Limit
	if limit <= 0 || limit > len(items) {
		limit = len(items)
	}
	return platformruntime.SessionListPage{
		Items: items[:limit],
		AsOf:  time.Now().UTC(),
	}, nil
}

type taskRepository struct {
	store *ControlPlaneStore
}

func (r taskRepository) Save(_ context.Context, task platformruntime.Task) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.tasks[key(task.TenantID, task.ID)] = cloneTask(task)

	sessionKey := key(task.TenantID, task.SessionID)
	ids := r.store.sessionTask[sessionKey]
	exists := false
	for _, existing := range ids {
		if existing == task.ID {
			exists = true
			break
		}
	}
	if !exists {
		r.store.sessionTask[sessionKey] = append(ids, task.ID)
	}
	return nil
}

func (r taskRepository) Get(_ context.Context, tenantID, taskID string) (platformruntime.Task, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	task, ok := r.store.tasks[key(tenantID, taskID)]
	if !ok {
		return platformruntime.Task{}, platformruntime.ErrTaskNotFound
	}
	return cloneTask(task), nil
}

func (r taskRepository) ListBySession(_ context.Context, tenantID, sessionID string) ([]platformruntime.Task, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	sessionKey := key(tenantID, sessionID)
	ids := r.store.sessionTask[sessionKey]
	out := make([]platformruntime.Task, 0, len(ids))
	for _, taskID := range ids {
		task, ok := r.store.tasks[key(tenantID, taskID)]
		if !ok {
			continue
		}
		out = append(out, cloneTask(task))
	}
	return out, nil
}

func (r taskRepository) List(_ context.Context, query platformruntime.TaskListQuery) (platformruntime.TaskListPage, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	items := make([]platformruntime.Task, 0)
	for _, task := range r.store.tasks {
		if query.TenantID != "" && task.TenantID != query.TenantID {
			continue
		}
		if query.SessionID != "" && task.SessionID != query.SessionID {
			continue
		}
		if query.Status != "" && task.Status != query.Status {
			continue
		}
		if query.ActorID != "" {
			session, ok := r.store.sessions[key(task.TenantID, task.SessionID)]
			if !ok {
				continue
			}
			if session.ActorID != query.ActorID {
				continue
			}
		}
		items = append(items, cloneTask(task))
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].QueuedAt.Equal(items[j].QueuedAt) {
			return items[i].ID < items[j].ID
		}
		return items[i].QueuedAt.Before(items[j].QueuedAt)
	})

	start := 0
	if query.Cursor != "" {
		parts := strings.Split(query.Cursor, ":")
		if len(parts) == 2 {
			if parsed, err := strconv.Atoi(parts[1]); err == nil && parsed > 0 {
				start = parsed
			}
		}
	}
	if start > len(items) {
		start = len(items)
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}

	nextCursor := ""
	if end < len(items) {
		nextCursor = "idx:" + strconv.Itoa(end)
	}

	return platformruntime.TaskListPage{
		Items:      items[start:end],
		NextCursor: nextCursor,
		AsOf:       time.Now().UTC(),
	}, nil
}

func cloneActor(actor iam.Actor) iam.Actor {
	actor.Roles = append([]string(nil), actor.Roles...)
	return actor
}

func cloneSession(session platformruntime.Session) platformruntime.Session {
	session.Metadata = cloneMap(session.Metadata)
	return session
}

func cloneTask(task platformruntime.Task) platformruntime.Task {
	task.Input = cloneMap(task.Input)
	task.Output = cloneMap(task.Output)
	return task
}

func cloneMap(source map[string]any) map[string]any {
	if source == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}
