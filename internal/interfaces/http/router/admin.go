package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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

func registerAdminRoutes(rg *gin.RouterGroup, container *bootstrap.Container) {
	if container == nil {
		panic("router: container must not be nil")
	}
	if container.ControlPlaneCatalog == nil {
		panic("router: control-plane catalog must not be nil")
	}

	catalog := container.ControlPlaneCatalog
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
