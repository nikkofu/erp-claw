package bootstrap

import (
	appagentruntime "github.com/nikkofu/erp-claw/internal/application/agentruntime"
	domainagentruntime "github.com/nikkofu/erp-claw/internal/domain/agentruntime"
	domainapproval "github.com/nikkofu/erp-claw/internal/domain/approval"
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

type ApprovalCatalog interface {
	domainapproval.DefinitionRepository
	domainapproval.InstanceRepository
	domainapproval.TaskRepository
}

type Container struct {
	Config              Config
	Health              *health.Service
	ControlPlaneCatalog ControlPlaneCatalog
	AgentRuntimeCatalog AgentRuntimeCatalog
	ApprovalCatalog     ApprovalCatalog
	WorkspaceGateway    *ws.WorkspaceGateway
}

func NewContainer(cfg Config) *Container {
	return &Container{
		Config:              cfg,
		Health:              health.NewService(),
		ControlPlaneCatalog: newControlPlaneCatalog(cfg),
		AgentRuntimeCatalog: newAgentRuntimeCatalog(cfg),
		ApprovalCatalog:     newApprovalCatalog(cfg),
		WorkspaceGateway:    ws.NewWorkspaceGateway(),
	}
}

func NewTestContainer() *Container {
	cfg := DefaultConfig()
	cfg.Env = "test"
	cfg.Database.DSN = ""
	return NewContainer(cfg)
}
