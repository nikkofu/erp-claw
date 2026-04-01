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
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
	"github.com/nikkofu/erp-claw/internal/platform/tenant"
)

type ServiceDeps struct {
	TenantCatalog   tenant.Catalog
	IAMDirectory    iam.Directory
	Sessions        platformruntime.SessionRepository
	Tasks           platformruntime.TaskRepository
	Deliveries      platformruntime.DeliveryRepository
	AuditReader     audit.Reader
	WorkspaceEvents platformruntime.WorkspaceEventSink
	Pipeline        *shared.Pipeline
}

type Service struct {
	tenantCatalog   tenant.Catalog
	iamDirectory    iam.Directory
	sessions        platformruntime.SessionRepository
	tasks           platformruntime.TaskRepository
	deliveries      platformruntime.DeliveryRepository
	auditReader     audit.Reader
	workspaceEvents platformruntime.WorkspaceEventSink
	pipeline        *shared.Pipeline
}

var ids atomic.Uint64

var ErrGovernanceCommandNotImplemented = errors.New("governance command not implemented")

func NewService(deps ServiceDeps) *Service {
	if deps.Pipeline == nil {
		deps.Pipeline = shared.NewPipeline(shared.PipelineDeps{})
	}
	return &Service{
		tenantCatalog:   deps.TenantCatalog,
		iamDirectory:    deps.IAMDirectory,
		sessions:        deps.Sessions,
		tasks:           deps.Tasks,
		deliveries:      deps.Deliveries,
		auditReader:     deps.AuditReader,
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

type AdvanceTaskInput struct {
	TenantID string
	ActorID  string
	TaskID   string
	Output   map[string]any
	Reason   string
}

func (s *Service) StartTask(ctx context.Context, input AdvanceTaskInput) (platformruntime.Task, error) {
	input, _ = normalizeAdvanceTaskInput(ctx, input)
	return s.mutateTask(ctx, "runtime.tasks.start", "runtime.task.running", input, func(task *platformruntime.Task) error {
		return task.Start(time.Now().UTC())
	})
}

func (s *Service) CompleteTask(ctx context.Context, input AdvanceTaskInput) (platformruntime.Task, error) {
	input, _ = normalizeAdvanceTaskInput(ctx, input)
	return s.mutateTask(ctx, "runtime.tasks.complete", "runtime.task.succeeded", input, func(task *platformruntime.Task) error {
		return task.Complete(input.Output, time.Now().UTC())
	})
}

func (s *Service) FailTask(ctx context.Context, input AdvanceTaskInput) (platformruntime.Task, error) {
	input, _ = normalizeAdvanceTaskInput(ctx, input)
	return s.mutateTask(ctx, "runtime.tasks.fail", "runtime.task.failed", input, func(task *platformruntime.Task) error {
		return task.Fail(input.Reason, time.Now().UTC())
	})
}

func (s *Service) PauseTask(ctx context.Context, input AdvanceTaskInput) (platformruntime.Task, error) {
	input, actorProvided := normalizeAdvanceTaskInput(ctx, input)
	if !actorProvided {
		input.ActorID = ""
	}
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.tasks.pause",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  governanceAuditPayload(ctx, s.tasks, "runtime.tasks.pause", input),
	}, func(context.Context, shared.Command) error {
		return ErrGovernanceCommandNotImplemented
	})
	return platformruntime.Task{}, err
}

func (s *Service) ResumeTask(ctx context.Context, input AdvanceTaskInput) (platformruntime.Task, error) {
	input, actorProvided := normalizeAdvanceTaskInput(ctx, input)
	if !actorProvided {
		input.ActorID = ""
	}
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.tasks.resume",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  governanceAuditPayload(ctx, s.tasks, "runtime.tasks.resume", input),
	}, func(context.Context, shared.Command) error {
		return ErrGovernanceCommandNotImplemented
	})
	return platformruntime.Task{}, err
}

func (s *Service) HandoffTask(ctx context.Context, input AdvanceTaskInput) (platformruntime.Task, error) {
	input, actorProvided := normalizeAdvanceTaskInput(ctx, input)
	if !actorProvided {
		input.ActorID = ""
	}
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.tasks.handoff",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  governanceAuditPayload(ctx, s.tasks, "runtime.tasks.handoff", input),
	}, func(context.Context, shared.Command) error {
		return ErrGovernanceCommandNotImplemented
	})
	return platformruntime.Task{}, err
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
		if input.ActorID != "" && current.ActorID != input.ActorID {
			return platformruntime.ErrSessionNotFound
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
		if input.ActorID != "" {
			session, err := s.sessions.Get(txCtx, current.TenantID, current.SessionID)
			if err != nil {
				return err
			}
			if session.ActorID != input.ActorID {
				return platformruntime.ErrTaskNotFound
			}
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
		session, err := s.sessions.Get(txCtx, input.TenantID, input.SessionID)
		if err != nil {
			return err
		}
		if input.ActorID != "" && session.ActorID != input.ActorID {
			return platformruntime.ErrSessionNotFound
		}
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
	Limit     int
	Cursor    string
}

func (s *Service) ListTasks(ctx context.Context, input ListTasksInput) (platformruntime.TaskListPage, error) {
	var page platformruntime.TaskListPage
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.tasks.list",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		listed, err := s.tasks.List(txCtx, platformruntime.TaskListQuery{
			TenantID:  strings.TrimSpace(input.TenantID),
			ActorID:   strings.TrimSpace(input.ActorID),
			SessionID: strings.TrimSpace(input.SessionID),
			Status:    input.Status,
			Limit:     input.Limit,
			Cursor:    strings.TrimSpace(input.Cursor),
		})
		if err != nil {
			return err
		}
		page = listed
		return nil
	})
	return page, err
}

type ListSessionsInput struct {
	TenantID string
	ActorID  string
	Status   platformruntime.SessionStatus
	Limit    int
}

func (s *Service) ListSessions(ctx context.Context, input ListSessionsInput) (platformruntime.SessionListPage, error) {
	var page platformruntime.SessionListPage
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.sessions.list",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		listed, err := s.sessions.List(txCtx, platformruntime.SessionListQuery{
			TenantID: strings.TrimSpace(input.TenantID),
			ActorID:  strings.TrimSpace(input.ActorID),
			Status:   input.Status,
			Limit:    input.Limit,
		})
		if err != nil {
			return err
		}
		page = listed
		return nil
	})
	return page, err
}

type ListTimelineInput struct {
	TenantID  string
	ActorID   string
	SessionID string
	TaskID    string
	Limit     int
	Cursor    string
}

func (s *Service) ListTimeline(ctx context.Context, input ListTimelineInput) (platformruntime.ReadSnapshot[platformruntime.TimelineEntry], error) {
	var page platformruntime.ReadSnapshot[platformruntime.TimelineEntry]
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.timeline.list",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		listed, err := s.tasks.ListTimeline(
			txCtx,
			strings.TrimSpace(input.TenantID),
			strings.TrimSpace(input.SessionID),
			strings.TrimSpace(input.TaskID),
			input.Limit,
			strings.TrimSpace(input.Cursor),
		)
		if err != nil {
			return err
		}
		page = listed
		return nil
	})
	return page, err
}

type ListEvidenceInput struct {
	TenantID  string
	ActorID   string
	TaskID    string
	RequestID string
	Limit     int
	Cursor    string
}

func (s *Service) ListEvidence(ctx context.Context, input ListEvidenceInput) (platformruntime.ReadSnapshot[platformruntime.EvidenceEntry], error) {
	var page platformruntime.ReadSnapshot[platformruntime.EvidenceEntry]
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.evidence.list",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		listed, err := s.tasks.ListEvidence(
			txCtx,
			strings.TrimSpace(input.TenantID),
			strings.TrimSpace(input.TaskID),
			strings.TrimSpace(input.RequestID),
			input.Limit,
			strings.TrimSpace(input.Cursor),
		)
		if err != nil {
			return err
		}
		page = listed
		return nil
	})
	return page, err
}

type ListDeliveriesInput struct {
	TenantID  string
	ActorID   string
	Status    platformruntime.DeliveryStatus
	SessionID string
	TaskID    string
	Limit     int
}

func (s *Service) ListDeliveries(ctx context.Context, input ListDeliveriesInput) (platformruntime.DeliveryListPage, error) {
	var page platformruntime.DeliveryListPage
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "runtime.deliveries.list",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		if s.deliveries == nil {
			page = platformruntime.DeliveryListPage{Items: []platformruntime.DeliveryRecord{}, AsOf: time.Now().UTC()}
			return nil
		}
		listed, err := s.deliveries.List(txCtx, platformruntime.DeliveryListQuery{
			TenantID:  strings.TrimSpace(input.TenantID),
			ActorID:   strings.TrimSpace(input.ActorID),
			Status:    input.Status,
			SessionID: strings.TrimSpace(input.SessionID),
			TaskID:    strings.TrimSpace(input.TaskID),
			Limit:     input.Limit,
		})
		if err != nil {
			return err
		}
		page = listed
		return nil
	})
	return page, err
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
	now := time.Now().UTC()
	record := platformruntime.DeliveryRecord{
		EventType:    strings.TrimSpace(evt.Type),
		TenantID:     strings.TrimSpace(evt.TenantID),
		SessionID:    strings.TrimSpace(evt.SessionID),
		TaskID:       strings.TrimSpace(evt.TaskID),
		AttemptCount: 1,
		Status:       platformruntime.DeliveryStatusPending,
		UpdatedAt:    now,
	}
	if s.deliveries != nil {
		if existing, ok, err := s.deliveries.Get(context.Background(), record.TenantID, record.EventType, record.SessionID, record.TaskID); err != nil {
			return err
		} else if ok {
			record = existing
			record.AttemptCount++
			record.UpdatedAt = now
		}
		if err := s.deliveries.Save(context.Background(), record); err != nil {
			return err
		}
	}

	err := s.workspaceEvents.Broadcast(evt)
	if err != nil {
		if s.deliveries != nil {
			record.Status = platformruntime.DeliveryStatusFailed
			record.LastError = err.Error()
			record.UpdatedAt = time.Now().UTC()
			if saveErr := s.deliveries.Save(context.Background(), record); saveErr != nil {
				return saveErr
			}
		}
		return err
	}

	if s.deliveries != nil {
		if record.Status == platformruntime.DeliveryStatusFailed {
			record.Status = platformruntime.DeliveryStatusRecovered
		} else {
			record.Status = platformruntime.DeliveryStatusDelivered
		}
		record.LastError = ""
		record.UpdatedAt = time.Now().UTC()
		if err := s.deliveries.Save(context.Background(), record); err != nil {
			return err
		}
	}
	return nil
}

func normalizeAdvanceTaskInput(ctx context.Context, input AdvanceTaskInput) (AdvanceTaskInput, bool) {
	input.TenantID = strings.TrimSpace(input.TenantID)
	input.ActorID = strings.TrimSpace(input.ActorID)
	input.TaskID = strings.TrimSpace(input.TaskID)
	input.Reason = strings.TrimSpace(input.Reason)

	actorProvided := input.ActorID != ""
	if rc, ok := platformruntime.RequestContextFromContext(ctx); ok && rc != nil {
		if tenantID := strings.TrimSpace(rc.TenantID); tenantID != "" {
			input.TenantID = tenantID
		}
		if actorID := strings.TrimSpace(rc.ActorID); actorID != "" {
			input.ActorID = actorID
		}
		actorProvided = rc.ActorProvided
	}
	return input, actorProvided
}

func governanceAuditPayload(
	ctx context.Context,
	tasks platformruntime.TaskRepository,
	commandName string,
	input AdvanceTaskInput,
) map[string]any {
	sessionID := ""
	if tasks != nil && input.TenantID != "" && input.TaskID != "" {
		if task, err := tasks.Get(ctx, input.TenantID, input.TaskID); err == nil {
			sessionID = task.SessionID
		}
	}
	return map[string]any{
		"action":         strings.TrimPrefix(commandName, "runtime.tasks."),
		"task_id":        input.TaskID,
		"session_id":     sessionID,
		"correlation_id": commandName + ":" + input.TenantID + ":" + input.TaskID,
		"resource_type":  "task",
		"resource_id":    input.TaskID,
	}
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
