package router

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/application/platform/controlplane"
	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/presenter"
	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/iam"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
	"github.com/nikkofu/erp-claw/internal/platform/tenant"
)

func registerControlPlaneRoutes(rg *gin.RouterGroup, container *bootstrap.Container) {
	if container == nil || container.ControlPlane == nil {
		panic("router: platform container must provide control-plane service")
	}

	controlGroup := rg.Group("/control")
	controlGroup.POST("/tenants", func(c *gin.Context) {
		var req createTenantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		created, err := container.ControlPlane.RegisterTenant(c.Request.Context(), controlplane.RegisterTenantInput{
			OperatorTenantID: tenantIDFromContext(c),
			ActorID:          actorIDFromContext(c),
			Code:             req.Code,
			Name:             req.Name,
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, tenantResponse(created))
	})

	controlGroup.POST("/actors", func(c *gin.Context) {
		var req upsertActorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		actor, err := container.ControlPlane.UpsertActor(c.Request.Context(), controlplane.UpsertActorInput{
			OperatorTenantID: tenantIDFromContext(c),
			OperatorActorID:  actorIDFromContext(c),
			TenantID:         req.TenantID,
			ActorID:          req.ActorID,
			Roles:            req.Roles,
			DepartmentID:     req.DepartmentID,
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, actorResponse(actor))
	})

	agentGroup := rg.Group("/agent")
	agentGroup.POST("/sessions", func(c *gin.Context) {
		var req openSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		session, err := container.ControlPlane.OpenSession(c.Request.Context(), controlplane.OpenSessionInput{
			TenantID:  tenantIDFromContext(c),
			ActorID:   actorIDFromContext(c),
			SessionID: req.SessionID,
			Metadata:  req.Metadata,
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, sessionResponse(session))
	})

	agentGroup.GET("/sessions", func(c *gin.Context) {
		limit, err := parseLimit(c.Query("limit"), 20)
		if err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		status := platformruntime.SessionStatus(strings.TrimSpace(c.Query("status")))

		page, err := container.ControlPlane.ListSessions(c.Request.Context(), controlplane.ListSessionsInput{
			TenantID: tenantIDFromContext(c),
			ActorID:  actorIDFromContext(c),
			Status:   status,
			Limit:    limit,
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, sessionListPageResponse(page))
	})

	agentGroup.GET("/sessions/:id", func(c *gin.Context) {
		session, err := container.ControlPlane.GetSession(c.Request.Context(), controlplane.GetSessionInput{
			TenantID:  tenantIDFromContext(c),
			ActorID:   actorIDFromContext(c),
			SessionID: c.Param("id"),
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, sessionResponse(session))
	})

	agentGroup.POST("/sessions/:id/tasks", func(c *gin.Context) {
		var req enqueueTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		task, err := container.ControlPlane.EnqueueTask(c.Request.Context(), controlplane.EnqueueTaskInput{
			TenantID:  tenantIDFromContext(c),
			ActorID:   actorIDFromContext(c),
			SessionID: c.Param("id"),
			TaskID:    req.TaskID,
			TaskType:  req.TaskType,
			Input:     req.Input,
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, taskResponse(task))
	})

	agentGroup.GET("/sessions/:id/tasks", func(c *gin.Context) {
		tasks, err := container.ControlPlane.ListSessionTasks(c.Request.Context(), controlplane.ListSessionTasksInput{
			TenantID:  tenantIDFromContext(c),
			ActorID:   actorIDFromContext(c),
			SessionID: c.Param("id"),
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, gin.H{"tasks": taskListResponse(tasks)})
	})

	agentGroup.GET("/tasks", func(c *gin.Context) {
		limit, err := parseLimit(c.Query("limit"), 20)
		if err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		status := platformruntime.TaskStatus(strings.TrimSpace(c.Query("status")))

		page, err := container.ControlPlane.ListTasks(c.Request.Context(), controlplane.ListTasksInput{
			TenantID:  tenantIDFromContext(c),
			ActorID:   actorIDFromContext(c),
			SessionID: strings.TrimSpace(c.Query("session_id")),
			Status:    status,
			Limit:     limit,
			Cursor:    c.Query("cursor"),
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, taskListPageResponse(page))
	})

	agentGroup.GET("/tasks/:id", func(c *gin.Context) {
		task, err := container.ControlPlane.GetTask(c.Request.Context(), controlplane.GetTaskInput{
			TenantID: tenantIDFromContext(c),
			ActorID:  actorIDFromContext(c),
			TaskID:   c.Param("id"),
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, taskResponse(task))
	})

	agentGroup.POST("/tasks/:id/start", func(c *gin.Context) {
		task, err := container.ControlPlane.StartTask(c.Request.Context(), controlplane.AdvanceTaskInput{
			TenantID: tenantIDFromContext(c),
			ActorID:  actorIDFromContext(c),
			TaskID:   c.Param("id"),
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, taskResponse(task))
	})

	agentGroup.POST("/tasks/:id/complete", func(c *gin.Context) {
		var req completeTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		task, err := container.ControlPlane.CompleteTask(c.Request.Context(), controlplane.AdvanceTaskInput{
			TenantID: tenantIDFromContext(c),
			ActorID:  actorIDFromContext(c),
			TaskID:   c.Param("id"),
			Output:   req.Output,
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, taskResponse(task))
	})

	agentGroup.POST("/tasks/:id/fail", func(c *gin.Context) {
		var req failTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		task, err := container.ControlPlane.FailTask(c.Request.Context(), controlplane.AdvanceTaskInput{
			TenantID: tenantIDFromContext(c),
			ActorID:  actorIDFromContext(c),
			TaskID:   c.Param("id"),
			Reason:   req.Reason,
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, taskResponse(task))
	})

	agentGroup.POST("/tasks/:id/pause", func(c *gin.Context) {
		task, err := container.ControlPlane.PauseTask(c.Request.Context(), controlplane.AdvanceTaskInput{
			TenantID: tenantIDFromContext(c),
			ActorID:  actorIDFromContext(c),
			TaskID:   c.Param("id"),
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, taskResponse(task))
	})

	agentGroup.POST("/tasks/:id/resume", func(c *gin.Context) {
		task, err := container.ControlPlane.ResumeTask(c.Request.Context(), controlplane.AdvanceTaskInput{
			TenantID: tenantIDFromContext(c),
			ActorID:  actorIDFromContext(c),
			TaskID:   c.Param("id"),
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, taskResponse(task))
	})

	agentGroup.POST("/tasks/:id/handoff", func(c *gin.Context) {
		task, err := container.ControlPlane.HandoffTask(c.Request.Context(), controlplane.AdvanceTaskInput{
			TenantID: tenantIDFromContext(c),
			ActorID:  actorIDFromContext(c),
			TaskID:   c.Param("id"),
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, taskResponse(task))
	})

	auditGroup := rg.Group("/audit")
	auditGroup.GET("/records", func(c *gin.Context) {
		limit, err := parseLimit(c.Query("limit"), 20)
		if err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		records, err := container.ControlPlane.ListAudit(c.Request.Context(), controlplane.ListAuditInput{
			TenantID:    tenantIDFromContext(c),
			ActorID:     actorIDFromContext(c),
			CommandName: c.Query("command"),
			Limit:       limit,
		})
		if err != nil {
			renderControlPlaneError(c, err)
			return
		}
		presenter.OK(c, gin.H{"records": auditRecordsResponse(records)})
	})
}

type createTenantRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type upsertActorRequest struct {
	TenantID     string   `json:"tenant_id"`
	ActorID      string   `json:"actor_id"`
	Roles        []string `json:"roles"`
	DepartmentID string   `json:"department_id"`
}

type openSessionRequest struct {
	SessionID string         `json:"session_id"`
	Metadata  map[string]any `json:"metadata"`
}

type enqueueTaskRequest struct {
	TaskID   string         `json:"task_id"`
	TaskType string         `json:"task_type"`
	Input    map[string]any `json:"input"`
}

type completeTaskRequest struct {
	Output map[string]any `json:"output"`
}

type failTaskRequest struct {
	Reason string `json:"reason"`
}

func renderControlPlaneError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, shared.ErrPolicyDenied):
		presenter.Error(c, http.StatusForbidden, err.Error())
	case errors.Is(err, shared.ErrApprovalRequired),
		errors.Is(err, controlplane.ErrGovernanceCommandNotImplemented):
		presenter.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, tenant.ErrTenantNotFound),
		errors.Is(err, iam.ErrActorNotFound),
		errors.Is(err, platformruntime.ErrSessionNotFound),
		errors.Is(err, platformruntime.ErrTaskNotFound):
		presenter.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, tenant.ErrInvalidTenant),
		errors.Is(err, iam.ErrInvalidActor),
		errors.Is(err, platformruntime.ErrInvalidSession),
		errors.Is(err, platformruntime.ErrInvalidTask),
		errors.Is(err, platformruntime.ErrInvalidSessionTransition),
		errors.Is(err, platformruntime.ErrInvalidTaskTransition):
		presenter.Error(c, http.StatusBadRequest, err.Error())
	default:
		presenter.Error(c, http.StatusInternalServerError, err.Error())
	}
}

func tenantResponse(value tenant.Tenant) gin.H {
	return gin.H{
		"id":     value.ID,
		"code":   value.Code,
		"name":   value.Name,
		"status": value.Status,
	}
}

func actorResponse(value iam.Actor) gin.H {
	return gin.H{
		"id":            value.ID,
		"roles":         append([]string(nil), value.Roles...),
		"department_id": value.DepartmentID,
	}
}

func sessionResponse(value platformruntime.Session) gin.H {
	return gin.H{
		"id":         value.ID,
		"tenant_id":  value.TenantID,
		"actor_id":   value.ActorID,
		"status":     string(value.Status),
		"metadata":   value.Metadata,
		"started_at": formatTime(value.StartedAt),
		"ended_at":   formatTime(value.EndedAt),
	}
}

func taskResponse(value platformruntime.Task) gin.H {
	return gin.H{
		"id":             value.ID,
		"tenant_id":      value.TenantID,
		"session_id":     value.SessionID,
		"task_type":      value.Type,
		"status":         string(value.Status),
		"input":          value.Input,
		"output":         value.Output,
		"failure_reason": value.FailureReason,
		"attempts":       value.Attempts,
		"queued_at":      formatTime(value.QueuedAt),
		"started_at":     formatTime(value.StartedAt),
		"completed_at":   formatTime(value.CompletedAt),
	}
}

func auditRecordsResponse(records []audit.Record) []gin.H {
	out := make([]gin.H, 0, len(records))
	for _, record := range records {
		out = append(out, gin.H{
			"command_name": record.CommandName,
			"tenant_id":    record.TenantID,
			"actor_id":     record.ActorID,
			"decision":     string(record.Decision),
			"outcome":      record.Outcome,
			"error":        record.Error,
			"occurred_at":  formatTime(record.OccurredAt),
		})
	}
	return out
}

func taskListResponse(tasks []platformruntime.Task) []gin.H {
	out := make([]gin.H, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, taskResponse(task))
	}
	return out
}

func taskListPageResponse(page platformruntime.TaskListPage) gin.H {
	return gin.H{
		"items":       taskListResponse(page.Items),
		"next_cursor": page.NextCursor,
		"as_of":       formatTime(page.AsOf),
	}
}

func sessionListResponse(sessions []platformruntime.Session) []gin.H {
	out := make([]gin.H, 0, len(sessions))
	for _, session := range sessions {
		out = append(out, sessionResponse(session))
	}
	return out
}

func sessionListPageResponse(page platformruntime.SessionListPage) gin.H {
	return gin.H{
		"items": sessionListResponse(page.Items),
		"as_of": formatTime(page.AsOf),
	}
}

func formatTime(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.UTC().Format(time.RFC3339)
}

func parseLimit(raw string, fallback int) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if limit <= 0 {
		return 0, errors.New("limit must be greater than 0")
	}
	return limit, nil
}
