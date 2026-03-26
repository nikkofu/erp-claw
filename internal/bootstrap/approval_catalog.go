package bootstrap

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	domainapproval "github.com/nikkofu/erp-claw/internal/domain/approval"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/postgres"
)

func newApprovalCatalog(cfg Config) ApprovalCatalog {
	if shouldUseInMemoryCatalogFallback(cfg) {
		return newInMemoryApprovalCatalog()
	}

	catalog, err := newPostgresApprovalCatalog(cfg.Database)
	if err == nil {
		return catalog
	}

	panic(fmt.Errorf("bootstrap: approval catalog init failed: %w", err))
}

func newPostgresApprovalCatalog(cfg DatabaseConfig) (ApprovalCatalog, error) {
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

	repo, err := postgres.NewApprovalRepository(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return repo, nil
}

type inMemoryApprovalCatalog struct {
	mu               sync.RWMutex
	nextDefinitionID int64
	nextInstanceID   int64
	nextTaskID       int64
	definitions      map[string]domainapproval.Definition
	instances        map[string]domainapproval.Instance
	tasks            map[string]domainapproval.Task
}

func newInMemoryApprovalCatalog() *inMemoryApprovalCatalog {
	return &inMemoryApprovalCatalog{
		definitions: make(map[string]domainapproval.Definition),
		instances:   make(map[string]domainapproval.Instance),
		tasks:       make(map[string]domainapproval.Task),
	}
}

func (r *inMemoryApprovalCatalog) SaveDefinition(_ context.Context, definition domainapproval.Definition) (domainapproval.Definition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	if strings.TrimSpace(definition.ID) == "" {
		r.nextDefinitionID++
		definition.ID = "approval-definition-" + strconv.FormatInt(r.nextDefinitionID, 10)
	}
	key := r.definitionKey(definition.TenantID, definition.ID)
	if existing, ok := r.definitions[key]; ok {
		definition.CreatedAt = existing.CreatedAt
	} else if definition.CreatedAt.IsZero() {
		definition.CreatedAt = now
	}
	definition.UpdatedAt = now
	r.definitions[key] = definition
	return definition, nil
}

func (r *inMemoryApprovalCatalog) ListDefinitions(_ context.Context, tenantID string) ([]domainapproval.Definition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	definitions := make([]domainapproval.Definition, 0)
	for _, definition := range r.definitions {
		if definition.TenantID == tenantID {
			definitions = append(definitions, definition)
		}
	}
	sort.Slice(definitions, func(i, j int) bool {
		if definitions[i].CreatedAt.Equal(definitions[j].CreatedAt) {
			return definitions[i].ID < definitions[j].ID
		}
		return definitions[i].CreatedAt.Before(definitions[j].CreatedAt)
	})
	return definitions, nil
}

func (r *inMemoryApprovalCatalog) GetDefinitionByID(_ context.Context, tenantID, definitionID string) (domainapproval.Definition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	definition, ok := r.definitions[r.definitionKey(tenantID, definitionID)]
	if !ok {
		return domainapproval.Definition{}, domainapproval.ErrDefinitionNotFound
	}
	return definition, nil
}

func (r *inMemoryApprovalCatalog) CreateInstance(_ context.Context, instance domainapproval.Instance) (domainapproval.Instance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(instance.ID) == "" {
		r.nextInstanceID++
		instance.ID = "approval-instance-" + strconv.FormatInt(r.nextInstanceID, 10)
	}
	if instance.CreatedAt.IsZero() {
		instance.CreatedAt = time.Now().UTC()
	}
	r.instances[r.instanceKey(instance.TenantID, instance.ID)] = instance
	return instance, nil
}

func (r *inMemoryApprovalCatalog) ListInstances(_ context.Context, tenantID string) ([]domainapproval.Instance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instances := make([]domainapproval.Instance, 0)
	for _, instance := range r.instances {
		if instance.TenantID == tenantID {
			instances = append(instances, instance)
		}
	}
	sort.Slice(instances, func(i, j int) bool {
		if instances[i].CreatedAt.Equal(instances[j].CreatedAt) {
			return instances[i].ID < instances[j].ID
		}
		return instances[i].CreatedAt.Before(instances[j].CreatedAt)
	})
	return instances, nil
}

func (r *inMemoryApprovalCatalog) GetInstanceByID(_ context.Context, tenantID, instanceID string) (domainapproval.Instance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instance, ok := r.instances[r.instanceKey(tenantID, instanceID)]
	if !ok {
		return domainapproval.Instance{}, domainapproval.ErrInstanceNotFound
	}
	return instance, nil
}

func (r *inMemoryApprovalCatalog) UpdateInstanceStatus(_ context.Context, tenantID, instanceID string, status domainapproval.InstanceStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.instanceKey(tenantID, instanceID)
	instance, ok := r.instances[key]
	if !ok {
		return domainapproval.ErrInstanceNotFound
	}
	now := time.Now().UTC()
	instance.Status = status
	instance.DecidedAt = &now
	r.instances[key] = instance
	return nil
}

func (r *inMemoryApprovalCatalog) DeleteInstance(_ context.Context, tenantID, instanceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.instanceKey(tenantID, instanceID)
	if _, ok := r.instances[key]; !ok {
		return domainapproval.ErrInstanceNotFound
	}
	delete(r.instances, key)
	return nil
}

func (r *inMemoryApprovalCatalog) CreateTask(_ context.Context, task domainapproval.Task) (domainapproval.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(task.ID) == "" {
		r.nextTaskID++
		task.ID = "approval-task-" + strconv.FormatInt(r.nextTaskID, 10)
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now().UTC()
	}
	r.tasks[r.taskKey(task.TenantID, task.ID)] = task
	return task, nil
}

func (r *inMemoryApprovalCatalog) ListTasks(_ context.Context, tenantID string) ([]domainapproval.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]domainapproval.Task, 0)
	for _, task := range r.tasks {
		if task.TenantID == tenantID {
			tasks = append(tasks, task)
		}
	}
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].CreatedAt.Equal(tasks[j].CreatedAt) {
			return tasks[i].ID < tasks[j].ID
		}
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})
	return tasks, nil
}

func (r *inMemoryApprovalCatalog) GetTaskByID(_ context.Context, tenantID, taskID string) (domainapproval.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[r.taskKey(tenantID, taskID)]
	if !ok {
		return domainapproval.Task{}, domainapproval.ErrTaskNotFound
	}
	return task, nil
}

func (r *inMemoryApprovalCatalog) UpdateTaskDecision(_ context.Context, tenantID, taskID string, status domainapproval.TaskStatus, decidedBy, comment string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.taskKey(tenantID, taskID)
	task, ok := r.tasks[key]
	if !ok {
		return domainapproval.ErrTaskNotFound
	}
	now := time.Now().UTC()
	task.Status = status
	task.DecidedBy = decidedBy
	task.Comment = comment
	task.DecidedAt = &now
	r.tasks[key] = task
	return nil
}

func (r *inMemoryApprovalCatalog) ListTasksByInstance(_ context.Context, tenantID, instanceID string) ([]domainapproval.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]domainapproval.Task, 0)
	for _, task := range r.tasks {
		if task.TenantID == tenantID && task.InstanceID == instanceID {
			tasks = append(tasks, task)
		}
	}
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].CreatedAt.Equal(tasks[j].CreatedAt) {
			return tasks[i].ID < tasks[j].ID
		}
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})
	return tasks, nil
}

func (r *inMemoryApprovalCatalog) definitionKey(tenantID, definitionID string) string {
	return tenantID + "|" + definitionID
}

func (r *inMemoryApprovalCatalog) instanceKey(tenantID, instanceID string) string {
	return tenantID + "|" + instanceID
}

func (r *inMemoryApprovalCatalog) taskKey(tenantID, taskID string) string {
	return tenantID + "|" + taskID
}
