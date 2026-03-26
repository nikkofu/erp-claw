package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	agentruntimeapp "github.com/nikkofu/erp-claw/internal/application/agentruntime"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/presenter"
)

type createWorkspaceSessionRequest struct {
	TenantID   string         `json:"tenant_id"`
	SessionKey string         `json:"session_key"`
	Metadata   map[string]any `json:"metadata"`
}

type createWorkspaceTaskRequest struct {
	TenantID  string         `json:"tenant_id"`
	SessionID string         `json:"session_id"`
	TaskType  string         `json:"task_type"`
	Input     map[string]any `json:"input"`
}

type workspaceTaskActionRequest struct {
	TenantID string         `json:"tenant_id"`
	Output   map[string]any `json:"output"`
}

type workspaceSessionActionRequest struct {
	TenantID string `json:"tenant_id"`
}

func registerWorkspaceRoutes(rg *gin.RouterGroup, container *bootstrap.Container) {
	if container == nil {
		panic("router: container must not be nil")
	}
	if container.AgentRuntimeCatalog == nil {
		panic("router: agent runtime catalog must not be nil")
	}
	if container.WorkspaceGateway == nil {
		panic("router: workspace gateway must not be nil")
	}

	listSessionsHandler := agentruntimeapp.ListSessionsHandler{Sessions: container.AgentRuntimeCatalog}
	listTasksHandler := agentruntimeapp.ListTasksHandler{Tasks: container.AgentRuntimeCatalog}
	replayEventsHandler := agentruntimeapp.ReplayWorkspaceEventsHandler{Events: container.WorkspaceGateway}
	runtimeService, err := agentruntimeapp.NewService(agentruntimeapp.ServiceDeps{
		Sessions: container.AgentRuntimeCatalog,
		Tasks:    container.AgentRuntimeCatalog,
		Events:   container.WorkspaceGateway,
	})
	if err != nil {
		panic("router: agent runtime service init failed: " + err.Error())
	}

	rg.POST("/sessions", func(c *gin.Context) {
		var req createWorkspaceSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		created, err := runtimeService.CreateSession(
			c.Request.Context(),
			tenantIDFromValueOrHeader(req.TenantID, c),
			strings.TrimSpace(req.SessionKey),
			req.Metadata,
		)
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, created)
	})

	rg.GET("/sessions", func(c *gin.Context) {
		sessions, err := listSessionsHandler.Handle(c.Request.Context(), agentruntimeapp.ListSessions{
			TenantID: tenantIDFromQueryOrHeader(c),
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		presenter.OK(c, sessions)
	})

	rg.POST("/sessions/:session_key/close", func(c *gin.Context) {
		var req workspaceSessionActionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		updated, err := runtimeService.CloseSession(
			c.Request.Context(),
			tenantIDFromValueOrHeader(req.TenantID, c),
			c.Param("session_key"),
		)
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, updated)
	})

	rg.POST("/tasks", func(c *gin.Context) {
		var req createWorkspaceTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		created, err := runtimeService.CreateTask(
			c.Request.Context(),
			tenantIDFromValueOrHeader(req.TenantID, c),
			strings.TrimSpace(req.SessionID),
			strings.TrimSpace(req.TaskType),
			req.Input,
		)
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, created)
	})

	rg.GET("/tasks", func(c *gin.Context) {
		tasks, err := listTasksHandler.Handle(c.Request.Context(), agentruntimeapp.ListTasks{
			TenantID:  tenantIDFromQueryOrHeader(c),
			SessionID: c.Query("session_id"),
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		presenter.OK(c, tasks)
	})

	rg.POST("/tasks/:task_id/start", func(c *gin.Context) {
		var req workspaceTaskActionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		updated, err := runtimeService.StartTask(
			c.Request.Context(),
			tenantIDFromValueOrHeader(req.TenantID, c),
			c.Param("task_id"),
		)
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, updated)
	})

	rg.POST("/tasks/:task_id/complete", func(c *gin.Context) {
		var req workspaceTaskActionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		updated, err := runtimeService.CompleteTask(
			c.Request.Context(),
			tenantIDFromValueOrHeader(req.TenantID, c),
			c.Param("task_id"),
			req.Output,
		)
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, updated)
	})

	rg.POST("/tasks/:task_id/fail", func(c *gin.Context) {
		var req workspaceTaskActionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		updated, err := runtimeService.FailTask(
			c.Request.Context(),
			tenantIDFromValueOrHeader(req.TenantID, c),
			c.Param("task_id"),
			req.Output,
		)
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, updated)
	})

	rg.POST("/tasks/:task_id/cancel", func(c *gin.Context) {
		var req workspaceTaskActionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		updated, err := runtimeService.CancelTask(
			c.Request.Context(),
			tenantIDFromValueOrHeader(req.TenantID, c),
			c.Param("task_id"),
			req.Output,
		)
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, updated)
	})

	rg.GET("/events", func(c *gin.Context) {
		events, err := replayEventsHandler.Handle(c.Request.Context(), agentruntimeapp.ReplayWorkspaceEvents{
			TenantID:  tenantIDFromQueryOrHeader(c),
			SessionID: c.Query("session_id"),
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		presenter.OK(c, events)
	})
}
