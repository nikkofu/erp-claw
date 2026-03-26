package controlplane

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/iam"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
	"github.com/nikkofu/erp-claw/internal/platform/tenant"
)

type ServiceDeps struct {
	TenantCatalog tenant.Catalog
	IAMDirectory  iam.Directory
	Sessions      platformruntime.SessionRepository
	Tasks         platformruntime.TaskRepository
	AuditReader   audit.Reader
	Pipeline      *shared.Pipeline
}

type Service struct {
	tenantCatalog tenant.Catalog
	iamDirectory  iam.Directory
	sessions      platformruntime.SessionRepository
	tasks         platformruntime.TaskRepository
	auditReader   audit.Reader
	pipeline      *shared.Pipeline
}

var ids atomic.Uint64

func NewService(deps ServiceDeps) *Service {
	if deps.Pipeline == nil {
		deps.Pipeline = shared.NewPipeline(shared.PipelineDeps{})
	}
	return &Service{
		tenantCatalog: deps.TenantCatalog,
		iamDirectory:  deps.IAMDirectory,
		sessions:      deps.Sessions,
		tasks:         deps.Tasks,
		auditReader:   deps.AuditReader,
		pipeline:      deps.Pipeline,
	}
}

type RegisterTenantInput struct {
	OperatorTenantID string
	ActorID          string
	Code             string
	Name             string
}

func (s *Service) RegisterTenant(ctx context.Context, input RegisterTenantInput) (tenant.Tenant, error) {
	var created tenant.Tenant
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "controlplane.tenants.register",
		TenantID: input.OperatorTenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		next, err := tenant.NewTenant(nextID("ten"), strings.TrimSpace(input.Code), strings.TrimSpace(input.Name))
		if err != nil {
			return err
		}
		if err := s.tenantCatalog.Save(txCtx, next); err != nil {
			return err
		}
		created = next
		return nil
	})
	return created, err
}

type UpsertActorInput struct {
	OperatorTenantID string
	OperatorActorID  string
	TenantID         string
	ActorID          string
	Roles            []string
	DepartmentID     string
}

func (s *Service) UpsertActor(ctx context.Context, input UpsertActorInput) (iam.Actor, error) {
	var actor iam.Actor
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "controlplane.actors.upsert",
		TenantID: input.OperatorTenantID,
		ActorID:  input.OperatorActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		targetTenantID := strings.TrimSpace(input.TenantID)
		if targetTenantID == "" {
			targetTenantID = strings.TrimSpace(input.OperatorTenantID)
		}

		actor = iam.Actor{
			ID:           strings.TrimSpace(input.ActorID),
			Roles:        normalizeRoles(input.Roles),
			DepartmentID: strings.TrimSpace(input.DepartmentID),
		}
		if err := s.iamDirectory.Save(txCtx, targetTenantID, actor); err != nil {
			return err
		}
		return nil
	})
	return actor, err
}

type OpenSessionInput struct {
	TenantID  string
	ActorID   string
	SessionID string
	Metadata  map[string]any
}

func (s *Service) OpenSession(ctx context.Context, input OpenSessionInput) (platformruntime.Session, error) {
	var session platformruntime.Session
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.sessions.open",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		sessionID := strings.TrimSpace(input.SessionID)
		if sessionID == "" {
			sessionID = nextID("sess")
		}
		next, err := platformruntime.NewSession(
			sessionID,
			strings.TrimSpace(input.TenantID),
			strings.TrimSpace(input.ActorID),
			input.Metadata,
			time.Now().UTC(),
		)
		if err != nil {
			return err
		}
		if err := s.sessions.Save(txCtx, next); err != nil {
			return err
		}
		session = next
		return nil
	})
	return session, err
}

type EnqueueTaskInput struct {
	TenantID  string
	ActorID   string
	SessionID string
	TaskID    string
	TaskType  string
	Input     map[string]any
}

func (s *Service) EnqueueTask(ctx context.Context, input EnqueueTaskInput) (platformruntime.Task, error) {
	var task platformruntime.Task
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.tasks.enqueue",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		taskID := strings.TrimSpace(input.TaskID)
		if taskID == "" {
			taskID = nextID("task")
		}
		created, err := platformruntime.NewTask(
			taskID,
			strings.TrimSpace(input.TenantID),
			strings.TrimSpace(input.SessionID),
			strings.TrimSpace(input.TaskType),
			input.Input,
			time.Now().UTC(),
		)
		if err != nil {
			return err
		}

		if _, err := s.sessions.Get(txCtx, created.TenantID, created.SessionID); err != nil {
			return err
		}
		if err := s.tasks.Save(txCtx, created); err != nil {
			return err
		}
		task = created
		return nil
	})
	return task, err
}

type AdvanceTaskInput struct {
	TenantID string
	ActorID  string
	TaskID   string
	Output   map[string]any
	Reason   string
}

func (s *Service) StartTask(ctx context.Context, input AdvanceTaskInput) (platformruntime.Task, error) {
	return s.mutateTask(ctx, "runtime.tasks.start", input, func(task *platformruntime.Task) error {
		return task.Start(time.Now().UTC())
	})
}

func (s *Service) CompleteTask(ctx context.Context, input AdvanceTaskInput) (platformruntime.Task, error) {
	return s.mutateTask(ctx, "runtime.tasks.complete", input, func(task *platformruntime.Task) error {
		return task.Complete(input.Output, time.Now().UTC())
	})
}

func (s *Service) FailTask(ctx context.Context, input AdvanceTaskInput) (platformruntime.Task, error) {
	return s.mutateTask(ctx, "runtime.tasks.fail", input, func(task *platformruntime.Task) error {
		return task.Fail(input.Reason, time.Now().UTC())
	})
}

type ListAuditInput struct {
	TenantID    string
	ActorID     string
	CommandName string
	Limit       int
}

func (s *Service) ListAudit(ctx context.Context, input ListAuditInput) ([]audit.Record, error) {
	if s.auditReader == nil {
		return nil, nil
	}

	var records []audit.Record
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "platform.audit.list",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		listed, err := s.auditReader.List(txCtx, audit.Query{
			TenantID:    strings.TrimSpace(input.TenantID),
			CommandName: strings.TrimSpace(input.CommandName),
			Limit:       input.Limit,
		})
		if err != nil {
			return err
		}
		records = listed
		return nil
	})
	return records, err
}

func (s *Service) mutateTask(
	ctx context.Context,
	commandName string,
	input AdvanceTaskInput,
	mutate func(task *platformruntime.Task) error,
) (platformruntime.Task, error) {
	var task platformruntime.Task
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     commandName,
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		current, err := s.tasks.Get(txCtx, input.TenantID, input.TaskID)
		if err != nil {
			return err
		}
		if err := mutate(&current); err != nil {
			return err
		}
		if err := s.tasks.Save(txCtx, current); err != nil {
			return err
		}
		task = current
		return nil
	})
	return task, err
}

func normalizeRoles(roles []string) []string {
	out := make([]string, 0, len(roles))
	seen := map[string]struct{}{}
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		out = append(out, role)
	}
	return out
}

func nextID(prefix string) string {
	return fmt.Sprintf("%s-%06d", prefix, ids.Add(1))
}
