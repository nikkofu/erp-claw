package memory

import (
	"context"
	"sort"
	"strings"
	"sync"

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

func (r tenantCatalogRepository) List(_ context.Context) ([]tenant.Tenant, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	out := make([]tenant.Tenant, 0, len(r.store.tenants))
	for _, value := range r.store.tenants {
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Code < out[j].Code
	})
	return out, nil
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

func (r iamDirectoryRepository) List(_ context.Context, tenantID string) ([]iam.Actor, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	prefix := tenantID + "/"
	out := make([]iam.Actor, 0)
	for key, actor := range r.store.actors {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		out = append(out, cloneActor(actor))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (r iamDirectoryRepository) Delete(_ context.Context, tenantID, actorID string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	actorKey := key(tenantID, actorID)
	if _, ok := r.store.actors[actorKey]; !ok {
		return iam.ErrActorNotFound
	}
	delete(r.store.actors, actorKey)
	return nil
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

func (r sessionRepository) ListByTenant(_ context.Context, tenantID string) ([]platformruntime.Session, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	prefix := tenantID + "/"
	out := make([]platformruntime.Session, 0)
	for sessionKey, session := range r.store.sessions {
		if !strings.HasPrefix(sessionKey, prefix) {
			continue
		}
		out = append(out, cloneSession(session))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
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

func (r taskRepository) ListByTenant(_ context.Context, tenantID string) ([]platformruntime.Task, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	prefix := tenantID + "/"
	out := make([]platformruntime.Task, 0)
	for taskKey, task := range r.store.tasks {
		if !strings.HasPrefix(taskKey, prefix) {
			continue
		}
		out = append(out, cloneTask(task))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
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
