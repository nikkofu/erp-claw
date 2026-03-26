package bootstrap

import (
	appagentruntime "github.com/nikkofu/erp-claw/internal/application/agentruntime"
	domainagentruntime "github.com/nikkofu/erp-claw/internal/domain/agentruntime"
	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
	"github.com/nikkofu/erp-claw/internal/interfaces/ws"
	"github.com/nikkofu/erp-claw/internal/platform/health"
)

type ControlPlaneCatalog interface {
	controlplane.TenantRepository
	controlplane.UserRepository
	controlplane.RoleRepository
	controlplane.DepartmentRepository
	controlplane.UserRoleBindingRepository
	controlplane.UserDepartmentBindingRepository
	controlplane.AgentProfileRepository
}

type AgentRuntimeCatalog interface {
	domainagentruntime.SessionRepository
	domainagentruntime.TaskRepository
	appagentruntime.SessionReader
	appagentruntime.TaskReader
}

type Container struct {
	Config              Config
	Health              *health.Service
	ControlPlaneCatalog ControlPlaneCatalog
	AgentRuntimeCatalog AgentRuntimeCatalog
	WorkspaceGateway    *ws.WorkspaceGateway
}

func NewContainer(cfg Config) *Container {
	return &Container{
		Config:              cfg,
		Health:              health.NewService(),
		ControlPlaneCatalog: newControlPlaneCatalog(cfg),
		AgentRuntimeCatalog: newAgentRuntimeCatalog(cfg),
		WorkspaceGateway:    ws.NewWorkspaceGateway(),
	}
}
