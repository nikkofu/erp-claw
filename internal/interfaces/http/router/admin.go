package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	approvalapp "github.com/nikkofu/erp-claw/internal/application/approval"
	capabilityapp "github.com/nikkofu/erp-claw/internal/application/capability"
	controlcommand "github.com/nikkofu/erp-claw/internal/application/controlplane/command"
	controlquery "github.com/nikkofu/erp-claw/internal/application/controlplane/query"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/presenter"
)

type createTenantRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type createUserRequest struct {
	TenantID    string `json:"tenant_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type createRoleRequest struct {
	TenantID    string `json:"tenant_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type createDepartmentRequest struct {
	TenantID           string `json:"tenant_id"`
	Name               string `json:"name"`
	ParentDepartmentID string `json:"parent_department_id"`
}

type assignUserRoleRequest struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
	RoleID   string `json:"role_id"`
}

type assignUserDepartmentRequest struct {
	TenantID     string `json:"tenant_id"`
	UserID       string `json:"user_id"`
	DepartmentID string `json:"department_id"`
}

type createAgentProfileRequest struct {
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
	Model    string `json:"model"`
}

type createApprovalDefinitionRequest struct {
	TenantID   string `json:"tenant_id"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	ApproverID string `json:"approver_id"`
	Active     bool   `json:"active"`
}

type createApprovalInstanceRequest struct {
	TenantID     string `json:"tenant_id"`
	DefinitionID string `json:"definition_id"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	RequestedBy  string `json:"requested_by"`
}

type decideApprovalTaskRequest struct {
	TenantID string `json:"tenant_id"`
	ActorID  string `json:"actor_id"`
	Comment  string `json:"comment"`
}

type createModelCatalogEntryRequest struct {
	EntryID     string `json:"entry_id"`
	ModelKey    string `json:"model_key"`
	DisplayName string `json:"display_name"`
	Provider    string `json:"provider"`
	Status      string `json:"status"`
}

type createToolCatalogEntryRequest struct {
	EntryID     string `json:"entry_id"`
	ToolKey     string `json:"tool_key"`
	DisplayName string `json:"display_name"`
	RiskLevel   string `json:"risk_level"`
	Status      string `json:"status"`
}

func registerAdminRoutes(rg *gin.RouterGroup, container *bootstrap.Container) {
	if container == nil {
		panic("router: container must not be nil")
	}
	if container.ControlPlaneCatalog == nil {
		panic("router: control-plane catalog must not be nil")
	}
	if container.ApprovalCatalog == nil {
		panic("router: approval catalog must not be nil")
	}
	if container.CapabilityCatalog == nil {
		panic("router: capability catalog must not be nil")
	}

	catalog := container.ControlPlaneCatalog
	approvalCatalog := container.ApprovalCatalog
	capabilityCatalog := container.CapabilityCatalog
	createTenantHandler := controlcommand.CreateTenantHandler{Tenants: catalog}
	listTenantsHandler := controlquery.ListTenantsHandler{Tenants: catalog}
	createUserHandler := controlcommand.CreateUserHandler{Users: catalog}
	listUsersHandler := controlquery.ListUsersHandler{Users: catalog}
	createRoleHandler := controlcommand.CreateRoleHandler{Roles: catalog}
	listRolesHandler := controlquery.ListRolesHandler{Roles: catalog}
	createDepartmentHandler := controlcommand.CreateDepartmentHandler{Departments: catalog}
	listDepartmentsHandler := controlquery.ListDepartmentsHandler{Departments: catalog}
	assignUserRoleHandler := controlcommand.AssignUserRoleHandler{Bindings: catalog}
	assignUserDepartmentHandler := controlcommand.AssignUserDepartmentHandler{Bindings: catalog}
	createAgentProfileHandler := controlcommand.CreateAgentProfileHandler{Profiles: catalog}
	listAgentProfilesHandler := controlquery.ListAgentProfilesHandler{Profiles: catalog}
	saveApprovalDefinitionHandler := approvalapp.SaveDefinitionHandler{Definitions: approvalCatalog}
	listApprovalDefinitionsHandler := approvalapp.ListDefinitionsHandler{Definitions: approvalCatalog}
	startApprovalHandler := approvalapp.StartApprovalHandler{
		Definitions: approvalCatalog,
		Instances:   approvalCatalog,
		Tasks:       approvalCatalog,
	}
	listApprovalInstancesHandler := approvalapp.ListInstancesHandler{Instances: approvalCatalog}
	listApprovalTasksHandler := approvalapp.ListTasksHandler{Tasks: approvalCatalog}
	approveApprovalTaskHandler := approvalapp.ApproveTaskHandler{
		Instances: approvalCatalog,
		Tasks:     approvalCatalog,
	}
	rejectApprovalTaskHandler := approvalapp.RejectTaskHandler{
		Instances: approvalCatalog,
		Tasks:     approvalCatalog,
	}
	createModelCatalogEntryHandler, err := capabilityapp.NewCreateModelCatalogEntryHandler(capabilityCatalog)
	if err != nil {
		panic("router: model catalog handler init failed: " + err.Error())
	}
	listModelCatalogEntriesHandler, err := capabilityapp.NewListModelCatalogEntriesHandler(capabilityCatalog)
	if err != nil {
		panic("router: model catalog list handler init failed: " + err.Error())
	}
	createToolCatalogEntryHandler, err := capabilityapp.NewCreateToolCatalogEntryHandler(capabilityCatalog)
	if err != nil {
		panic("router: tool catalog handler init failed: " + err.Error())
	}
	listToolCatalogEntriesHandler, err := capabilityapp.NewListToolCatalogEntriesHandler(capabilityCatalog)
	if err != nil {
		panic("router: tool catalog list handler init failed: " + err.Error())
	}

	rg.POST("/tenants", func(c *gin.Context) {
		var req createTenantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		created, err := createTenantHandler.Handle(c.Request.Context(), controlcommand.CreateTenant{
			Code: req.Code,
			Name: req.Name,
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, created)
	})

	rg.GET("/tenants", func(c *gin.Context) {
		tenants, err := listTenantsHandler.Handle(c.Request.Context(), controlquery.ListTenants{})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		presenter.OK(c, tenants)
	})

	rg.POST("/users", func(c *gin.Context) {
		var req createUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		tenantID := strings.TrimSpace(req.TenantID)
		if tenantID == "" {
			tenantID = strings.TrimSpace(c.GetHeader("X-Tenant-ID"))
		}

		created, err := createUserHandler.Handle(c.Request.Context(), controlcommand.CreateUser{
			TenantID:    tenantID,
			Email:       req.Email,
			DisplayName: req.DisplayName,
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, created)
	})

	rg.GET("/users", func(c *gin.Context) {
		users, err := listUsersHandler.Handle(c.Request.Context(), controlquery.ListUsers{
			TenantID: tenantIDFromQueryOrHeader(c),
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		presenter.OK(c, users)
	})

	rg.POST("/roles", func(c *gin.Context) {
		var req createRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		created, err := createRoleHandler.Handle(c.Request.Context(), controlcommand.CreateRole{
			TenantID:    tenantIDFromValueOrHeader(req.TenantID, c),
			Name:        req.Name,
			Description: req.Description,
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, created)
	})

	rg.GET("/roles", func(c *gin.Context) {
		roles, err := listRolesHandler.Handle(c.Request.Context(), controlquery.ListRoles{
			TenantID: tenantIDFromQueryOrHeader(c),
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		presenter.OK(c, roles)
	})

	rg.POST("/departments", func(c *gin.Context) {
		var req createDepartmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		created, err := createDepartmentHandler.Handle(c.Request.Context(), controlcommand.CreateDepartment{
			TenantID:           tenantIDFromValueOrHeader(req.TenantID, c),
			Name:               req.Name,
			ParentDepartmentID: req.ParentDepartmentID,
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, created)
	})

	rg.GET("/departments", func(c *gin.Context) {
		departments, err := listDepartmentsHandler.Handle(c.Request.Context(), controlquery.ListDepartments{
			TenantID: tenantIDFromQueryOrHeader(c),
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		presenter.OK(c, departments)
	})

	rg.POST("/user-role-bindings", func(c *gin.Context) {
		var req assignUserRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		created, err := assignUserRoleHandler.Handle(c.Request.Context(), controlcommand.AssignUserRole{
			TenantID: tenantIDFromValueOrHeader(req.TenantID, c),
			UserID:   req.UserID,
			RoleID:   req.RoleID,
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, created)
	})

	rg.POST("/user-department-bindings", func(c *gin.Context) {
		var req assignUserDepartmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		created, err := assignUserDepartmentHandler.Handle(c.Request.Context(), controlcommand.AssignUserDepartment{
			TenantID:     tenantIDFromValueOrHeader(req.TenantID, c),
			UserID:       req.UserID,
			DepartmentID: req.DepartmentID,
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, created)
	})

	rg.POST("/agent-profiles", func(c *gin.Context) {
		var req createAgentProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		tenantID := strings.TrimSpace(req.TenantID)
		if tenantID == "" {
			tenantID = strings.TrimSpace(c.GetHeader("X-Tenant-ID"))
		}

		created, err := createAgentProfileHandler.Handle(c.Request.Context(), controlcommand.CreateAgentProfile{
			TenantID: tenantID,
			Name:     req.Name,
			Model:    req.Model,
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, created)
	})

	rg.GET("/agent-profiles", func(c *gin.Context) {
		profiles, err := listAgentProfilesHandler.Handle(c.Request.Context(), controlquery.ListAgentProfiles{
			TenantID: tenantIDFromQueryOrHeader(c),
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		presenter.OK(c, profiles)
	})

	rg.POST("/approval-definitions", func(c *gin.Context) {
		var req createApprovalDefinitionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		created, err := saveApprovalDefinitionHandler.Handle(c.Request.Context(), approvalapp.SaveDefinition{
			TenantID:   tenantIDFromValueOrHeader(req.TenantID, c),
			ID:         req.ID,
			Name:       req.Name,
			ApproverID: req.ApproverID,
			Active:     req.Active,
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, created)
	})

	rg.GET("/approval-definitions", func(c *gin.Context) {
		definitions, err := listApprovalDefinitionsHandler.Handle(c.Request.Context(), approvalapp.ListDefinitions{
			TenantID: tenantIDFromQueryOrHeader(c),
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		presenter.OK(c, definitions)
	})

	rg.POST("/approval-instances", func(c *gin.Context) {
		var req createApprovalInstanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		created, err := startApprovalHandler.Handle(c.Request.Context(), approvalapp.StartApproval{
			TenantID:     tenantIDFromValueOrHeader(req.TenantID, c),
			DefinitionID: req.DefinitionID,
			ResourceType: req.ResourceType,
			ResourceID:   req.ResourceID,
			RequestedBy:  req.RequestedBy,
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, created.Instance)
	})

	rg.GET("/approval-instances", func(c *gin.Context) {
		instances, err := listApprovalInstancesHandler.Handle(c.Request.Context(), approvalapp.ListInstances{
			TenantID: tenantIDFromQueryOrHeader(c),
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		presenter.OK(c, instances)
	})

	rg.GET("/approval-tasks", func(c *gin.Context) {
		tasks, err := listApprovalTasksHandler.Handle(c.Request.Context(), approvalapp.ListTasks{
			TenantID: tenantIDFromQueryOrHeader(c),
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		presenter.OK(c, tasks)
	})

	rg.POST("/approval-tasks/:task_id/approve", func(c *gin.Context) {
		var req decideApprovalTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		updated, err := approveApprovalTaskHandler.Handle(c.Request.Context(), approvalapp.ApproveTask{
			TenantID: tenantIDFromValueOrHeader(req.TenantID, c),
			TaskID:   c.Param("task_id"),
			ActorID:  req.ActorID,
			Comment:  req.Comment,
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, updated)
	})

	rg.POST("/approval-tasks/:task_id/reject", func(c *gin.Context) {
		var req decideApprovalTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		updated, err := rejectApprovalTaskHandler.Handle(c.Request.Context(), approvalapp.RejectTask{
			TenantID: tenantIDFromValueOrHeader(req.TenantID, c),
			TaskID:   c.Param("task_id"),
			ActorID:  req.ActorID,
			Comment:  req.Comment,
		})
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, updated)
	})

	rg.POST("/model-catalog-entries", func(c *gin.Context) {
		var req createModelCatalogEntryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		tenantID := tenantIDFromQueryOrHeader(c)
		if err := createModelCatalogEntryHandler.Handle(c.Request.Context(), tenantID, capabilityapp.CreateModelCatalogEntryPayload{
			ID:          req.EntryID,
			ModelKey:    req.ModelKey,
			DisplayName: req.DisplayName,
			Provider:    req.Provider,
			Status:      req.Status,
		}); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, gin.H{
			"TenantID":    tenantID,
			"EntryID":     req.EntryID,
			"ModelKey":    req.ModelKey,
			"DisplayName": req.DisplayName,
			"Provider":    req.Provider,
			"Status":      req.Status,
		})
	})

	rg.GET("/model-catalog-entries", func(c *gin.Context) {
		entries, err := listModelCatalogEntriesHandler.Handle(c.Request.Context(), tenantIDFromQueryOrHeader(c))
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		presenter.OK(c, entries)
	})

	rg.POST("/tool-catalog-entries", func(c *gin.Context) {
		var req createToolCatalogEntryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		tenantID := tenantIDFromQueryOrHeader(c)
		if err := createToolCatalogEntryHandler.Handle(c.Request.Context(), tenantID, capabilityapp.CreateToolCatalogEntryPayload{
			ID:          req.EntryID,
			ToolKey:     req.ToolKey,
			DisplayName: req.DisplayName,
			RiskLevel:   req.RiskLevel,
			Status:      req.Status,
		}); err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		adminCreated(c, gin.H{
			"TenantID":    tenantID,
			"EntryID":     req.EntryID,
			"ToolKey":     req.ToolKey,
			"DisplayName": req.DisplayName,
			"RiskLevel":   req.RiskLevel,
			"Status":      req.Status,
		})
	})

	rg.GET("/tool-catalog-entries", func(c *gin.Context) {
		entries, err := listToolCatalogEntriesHandler.Handle(c.Request.Context(), tenantIDFromQueryOrHeader(c))
		if err != nil {
			adminError(c, http.StatusBadRequest, err)
			return
		}

		presenter.OK(c, entries)
	})
}

func tenantIDFromQueryOrHeader(c *gin.Context) string {
	tenantID := strings.TrimSpace(c.Query("tenant_id"))
	if tenantID != "" {
		return tenantID
	}
	return strings.TrimSpace(c.GetHeader("X-Tenant-ID"))
}

func tenantIDFromValueOrHeader(tenantID string, c *gin.Context) string {
	trimmed := strings.TrimSpace(tenantID)
	if trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(c.GetHeader("X-Tenant-ID"))
}

func adminCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{
		"data": data,
		"meta": gin.H{"request_id": c.GetString("request_id")},
	})
}

func adminError(c *gin.Context, status int, err error) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": err.Error(),
		"meta":  gin.H{"request_id": c.GetString("request_id")},
	})
}
