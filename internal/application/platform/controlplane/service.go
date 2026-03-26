package controlplane

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/iam"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
	"github.com/nikkofu/erp-claw/internal/platform/tenant"
)

var ErrPolicyRuleStoreUnavailable = errors.New("policy rule store is unavailable")

type ServiceDeps struct {
	TenantCatalog   tenant.Catalog
	IAMDirectory    iam.Directory
	Sessions        platformruntime.SessionRepository
	Tasks           platformruntime.TaskRepository
	AuditReader     audit.Reader
	PolicyRules     policy.RuleStore
	WorkspaceEvents platformruntime.WorkspaceEventSink
	Pipeline        *shared.Pipeline
}

type Service struct {
	tenantCatalog   tenant.Catalog
	iamDirectory    iam.Directory
	sessions        platformruntime.SessionRepository
	tasks           platformruntime.TaskRepository
	auditReader     audit.Reader
	policyRules     policy.RuleStore
	workspaceEvents platformruntime.WorkspaceEventSink
	pipeline        *shared.Pipeline
}

var ids atomic.Uint64

func NewService(deps ServiceDeps) *Service {
	if deps.Pipeline == nil {
		deps.Pipeline = shared.NewPipeline(shared.PipelineDeps{})
	}
	return &Service{
		tenantCatalog:   deps.TenantCatalog,
		iamDirectory:    deps.IAMDirectory,
		sessions:        deps.Sessions,
		tasks:           deps.Tasks,
		auditReader:     deps.AuditReader,
		policyRules:     deps.PolicyRules,
		workspaceEvents: deps.WorkspaceEvents,
		pipeline:        deps.Pipeline,
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

type GetTenantInput struct {
	OperatorTenantID string
	OperatorActorID  string
	Code             string
}

func (s *Service) GetTenant(ctx context.Context, input GetTenantInput) (tenant.Tenant, error) {
	var value tenant.Tenant
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "controlplane.tenants.get",
		TenantID: input.OperatorTenantID,
		ActorID:  input.OperatorActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		current, err := s.tenantCatalog.Get(txCtx, strings.TrimSpace(input.Code))
		if err != nil {
			return err
		}
		value = current
		return nil
	})
	return value, err
}

type ListTenantsInput struct {
	OperatorTenantID string
	OperatorActorID  string
}

func (s *Service) ListTenants(ctx context.Context, input ListTenantsInput) ([]tenant.Tenant, error) {
	var values []tenant.Tenant
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "controlplane.tenants.list",
		TenantID: input.OperatorTenantID,
		ActorID:  input.OperatorActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		listed, err := s.tenantCatalog.List(txCtx)
		if err != nil {
			return err
		}
		values = append([]tenant.Tenant(nil), listed...)
		return nil
	})
	return values, err
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

type GetActorInput struct {
	OperatorTenantID string
	OperatorActorID  string
	TenantID         string
	ActorID          string
}

func (s *Service) GetActor(ctx context.Context, input GetActorInput) (iam.Actor, error) {
	var actor iam.Actor
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "controlplane.actors.get",
		TenantID: input.OperatorTenantID,
		ActorID:  input.OperatorActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		targetTenantID := strings.TrimSpace(input.TenantID)
		if targetTenantID == "" {
			targetTenantID = strings.TrimSpace(input.OperatorTenantID)
		}
		current, err := s.iamDirectory.Get(txCtx, targetTenantID, strings.TrimSpace(input.ActorID))
		if err != nil {
			return err
		}
		actor = current
		return nil
	})
	return actor, err
}

type ListActorsInput struct {
	OperatorTenantID string
	OperatorActorID  string
	TenantID         string
}

func (s *Service) ListActors(ctx context.Context, input ListActorsInput) ([]iam.Actor, error) {
	var actors []iam.Actor
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "controlplane.actors.list",
		TenantID: input.OperatorTenantID,
		ActorID:  input.OperatorActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		targetTenantID := strings.TrimSpace(input.TenantID)
		if targetTenantID == "" {
			targetTenantID = strings.TrimSpace(input.OperatorTenantID)
		}
		listed, err := s.iamDirectory.List(txCtx, targetTenantID)
		if err != nil {
			return err
		}
		actors = append([]iam.Actor(nil), listed...)
		return nil
	})
	return actors, err
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
		if err := s.emitWorkspaceEvent(platformruntime.WorkspaceEvent{
			Type:      "runtime.session.opened",
			TenantID:  next.TenantID,
			SessionID: next.ID,
			Payload: map[string]any{
				"actor_id": next.ActorID,
				"status":   string(next.Status),
			},
		}); err != nil {
			return err
		}
		session = next
		return nil
	})
	return session, err
}

type CloseSessionInput struct {
	TenantID  string
	ActorID   string
	SessionID string
}

func (s *Service) CloseSession(ctx context.Context, input CloseSessionInput) (platformruntime.Session, error) {
	var session platformruntime.Session
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.sessions.close",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		current, err := s.sessions.Get(txCtx, strings.TrimSpace(input.TenantID), strings.TrimSpace(input.SessionID))
		if err != nil {
			return err
		}
		tasks, err := s.tasks.ListBySession(txCtx, current.TenantID, current.ID)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			if task.Status == platformruntime.TaskStatusPending || task.Status == platformruntime.TaskStatusRunning {
				return platformruntime.ErrInvalidSessionTransition
			}
		}
		if err := current.Close(time.Now().UTC()); err != nil {
			return err
		}
		if err := s.sessions.Save(txCtx, current); err != nil {
			return err
		}
		if err := s.emitWorkspaceEvent(platformruntime.WorkspaceEvent{
			Type:      "runtime.session.closed",
			TenantID:  current.TenantID,
			SessionID: current.ID,
			Payload: map[string]any{
				"actor_id": current.ActorID,
				"status":   string(current.Status),
			},
		}); err != nil {
			return err
		}
		session = current
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

		session, err := s.sessions.Get(txCtx, created.TenantID, created.SessionID)
		if err != nil {
			return err
		}
		if session.Status != platformruntime.SessionStatusOpen {
			return platformruntime.ErrInvalidSessionTransition
		}
		if err := s.tasks.Save(txCtx, created); err != nil {
			return err
		}
		if err := s.emitWorkspaceEvent(platformruntime.WorkspaceEvent{
			Type:      "runtime.task.enqueued",
			TenantID:  created.TenantID,
			SessionID: created.SessionID,
			TaskID:    created.ID,
			Payload: map[string]any{
				"task_type": created.Type,
				"status":    string(created.Status),
			},
		}); err != nil {
			return err
		}
		task = created
		return nil
	})
	return task, err
}

type ListSessionsInput struct {
	TenantID string
	ActorID  string
	Status   platformruntime.SessionStatus
	Offset   int
	Limit    int
}

func (s *Service) ListSessions(ctx context.Context, input ListSessionsInput) ([]platformruntime.Session, error) {
	var sessions []platformruntime.Session
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.sessions.list",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		current, err := s.sessions.ListByTenant(txCtx, strings.TrimSpace(input.TenantID))
		if err != nil {
			return err
		}
		targetStatus := input.Status
		filtered := make([]platformruntime.Session, 0, len(current))
		for _, session := range current {
			if targetStatus != "" && session.Status != targetStatus {
				continue
			}
			filtered = append(filtered, session)
		}

		start := input.Offset
		if start < 0 {
			start = 0
		}
		if start >= len(filtered) {
			sessions = []platformruntime.Session{}
			return nil
		}

		end := len(filtered)
		if input.Limit > 0 && start+input.Limit < end {
			end = start + input.Limit
		}
		sessions = append([]platformruntime.Session(nil), filtered[start:end]...)
		return nil
	})
	return sessions, err
}

type AdvanceTaskInput struct {
	TenantID string
	ActorID  string
	TaskID   string
	Output   map[string]any
	Reason   string
}

func (s *Service) StartTask(ctx context.Context, input AdvanceTaskInput) (platformruntime.Task, error) {
	return s.mutateTask(ctx, "runtime.tasks.start", "runtime.task.running", input, func(task *platformruntime.Task) error {
		return task.Start(time.Now().UTC())
	})
}

func (s *Service) CompleteTask(ctx context.Context, input AdvanceTaskInput) (platformruntime.Task, error) {
	return s.mutateTask(ctx, "runtime.tasks.complete", "runtime.task.succeeded", input, func(task *platformruntime.Task) error {
		return task.Complete(input.Output, time.Now().UTC())
	})
}

func (s *Service) FailTask(ctx context.Context, input AdvanceTaskInput) (platformruntime.Task, error) {
	return s.mutateTask(ctx, "runtime.tasks.fail", "runtime.task.failed", input, func(task *platformruntime.Task) error {
		return task.Fail(input.Reason, time.Now().UTC())
	})
}

func (s *Service) CancelTask(ctx context.Context, input AdvanceTaskInput) (platformruntime.Task, error) {
	return s.mutateTask(ctx, "runtime.tasks.cancel", "runtime.task.canceled", input, func(task *platformruntime.Task) error {
		return task.Cancel(input.Reason, time.Now().UTC())
	})
}

type GetSessionInput struct {
	TenantID  string
	ActorID   string
	SessionID string
}

func (s *Service) GetSession(ctx context.Context, input GetSessionInput) (platformruntime.Session, error) {
	var session platformruntime.Session
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.sessions.get",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		current, err := s.sessions.Get(txCtx, input.TenantID, input.SessionID)
		if err != nil {
			return err
		}
		session = current
		return nil
	})
	return session, err
}

type GetTaskInput struct {
	TenantID string
	ActorID  string
	TaskID   string
}

func (s *Service) GetTask(ctx context.Context, input GetTaskInput) (platformruntime.Task, error) {
	var task platformruntime.Task
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.tasks.get",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		current, err := s.tasks.Get(txCtx, input.TenantID, input.TaskID)
		if err != nil {
			return err
		}
		task = current
		return nil
	})
	return task, err
}

type ListSessionTasksInput struct {
	TenantID  string
	ActorID   string
	SessionID string
}

func (s *Service) ListSessionTasks(ctx context.Context, input ListSessionTasksInput) ([]platformruntime.Task, error) {
	var tasks []platformruntime.Task
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.tasks.list_by_session",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		current, err := s.tasks.ListBySession(txCtx, input.TenantID, input.SessionID)
		if err != nil {
			return err
		}
		tasks = current
		return nil
	})
	return tasks, err
}

type ListTasksInput struct {
	TenantID  string
	ActorID   string
	SessionID string
	Status    platformruntime.TaskStatus
	Offset    int
	Limit     int
}

func (s *Service) ListTasks(ctx context.Context, input ListTasksInput) ([]platformruntime.Task, error) {
	var tasks []platformruntime.Task
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.tasks.list",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		current, err := s.tasks.ListByTenant(txCtx, strings.TrimSpace(input.TenantID))
		if err != nil {
			return err
		}

		targetSessionID := strings.TrimSpace(input.SessionID)
		targetStatus := input.Status
		filtered := make([]platformruntime.Task, 0, len(current))
		for _, task := range current {
			if targetSessionID != "" && task.SessionID != targetSessionID {
				continue
			}
			if targetStatus != "" && task.Status != targetStatus {
				continue
			}
			filtered = append(filtered, task)
		}

		start := input.Offset
		if start < 0 {
			start = 0
		}
		if start >= len(filtered) {
			tasks = []platformruntime.Task{}
			return nil
		}

		end := len(filtered)
		if input.Limit > 0 && start+input.Limit < end {
			end = start + input.Limit
		}
		tasks = append([]platformruntime.Task(nil), filtered[start:end]...)
		return nil
	})
	return tasks, err
}

type UpsertPolicyRuleInput struct {
	OperatorTenantID string
	OperatorActorID  string
	TenantID         string
	CommandPrefix    string
	AnyOfRoles       []string
}

func (s *Service) UpsertPolicyRule(ctx context.Context, input UpsertPolicyRuleInput) (policy.Rule, error) {
	if s.policyRules == nil {
		return policy.Rule{}, ErrPolicyRuleStoreUnavailable
	}

	var rule policy.Rule
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "controlplane.policy_rules.upsert",
		TenantID: input.OperatorTenantID,
		ActorID:  input.OperatorActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		targetTenantID := strings.TrimSpace(input.TenantID)
		if targetTenantID == "" {
			targetTenantID = strings.TrimSpace(input.OperatorTenantID)
		}

		rule = policy.Rule{
			CommandPrefix: strings.TrimSpace(input.CommandPrefix),
			AnyOfRoles:    append([]string(nil), input.AnyOfRoles...),
		}
		if err := s.policyRules.Upsert(txCtx, targetTenantID, rule); err != nil {
			return err
		}

		listed, err := s.policyRules.List(txCtx, targetTenantID)
		if err != nil {
			return err
		}
		for _, candidate := range listed {
			if candidate.CommandPrefix == rule.CommandPrefix {
				rule = candidate
				break
			}
		}
		return nil
	})
	return rule, err
}

type ListPolicyRulesInput struct {
	OperatorTenantID string
	OperatorActorID  string
	TenantID         string
}

func (s *Service) ListPolicyRules(ctx context.Context, input ListPolicyRulesInput) ([]policy.Rule, error) {
	if s.policyRules == nil {
		return nil, ErrPolicyRuleStoreUnavailable
	}

	var rules []policy.Rule
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "controlplane.policy_rules.list",
		TenantID: input.OperatorTenantID,
		ActorID:  input.OperatorActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		targetTenantID := strings.TrimSpace(input.TenantID)
		if targetTenantID == "" {
			targetTenantID = strings.TrimSpace(input.OperatorTenantID)
		}

		listed, err := s.policyRules.List(txCtx, targetTenantID)
		if err != nil {
			return err
		}
		rules = append([]policy.Rule(nil), listed...)
		return nil
	})
	return rules, err
}

type DeletePolicyRuleInput struct {
	OperatorTenantID string
	OperatorActorID  string
	TenantID         string
	CommandPrefix    string
}

func (s *Service) DeletePolicyRule(ctx context.Context, input DeletePolicyRuleInput) error {
	if s.policyRules == nil {
		return ErrPolicyRuleStoreUnavailable
	}

	return s.pipeline.Execute(ctx, shared.Command{
		Name:     "controlplane.policy_rules.delete",
		TenantID: input.OperatorTenantID,
		ActorID:  input.OperatorActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		targetTenantID := strings.TrimSpace(input.TenantID)
		if targetTenantID == "" {
			targetTenantID = strings.TrimSpace(input.OperatorTenantID)
		}
		return s.policyRules.Delete(txCtx, targetTenantID, strings.TrimSpace(input.CommandPrefix))
	})
}

type ListAuditInput struct {
	TenantID      string
	ActorID       string
	CommandName   string
	QueryActorID  string
	QueryDecision policy.Decision
	QueryOutcome  string
	Offset        int
	Limit         int
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
			ActorID:     strings.TrimSpace(input.QueryActorID),
			Decision:    input.QueryDecision,
			Outcome:     strings.TrimSpace(input.QueryOutcome),
			CommandName: strings.TrimSpace(input.CommandName),
			Offset:      input.Offset,
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
	eventType string,
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
		if err := s.emitWorkspaceEvent(platformruntime.WorkspaceEvent{
			Type:      eventType,
			TenantID:  current.TenantID,
			SessionID: current.SessionID,
			TaskID:    current.ID,
			Payload: map[string]any{
				"status":   string(current.Status),
				"attempts": current.Attempts,
			},
		}); err != nil {
			return err
		}
		task = current
		return nil
	})
	return task, err
}

func (s *Service) emitWorkspaceEvent(evt platformruntime.WorkspaceEvent) error {
	if s.workspaceEvents == nil {
		return nil
	}
	return s.workspaceEvents.Broadcast(evt)
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
