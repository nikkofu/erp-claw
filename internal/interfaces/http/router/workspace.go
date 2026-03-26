package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	agentruntimeapp "github.com/nikkofu/erp-claw/internal/application/agentruntime"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/presenter"
)

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
