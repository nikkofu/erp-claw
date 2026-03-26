package bootstrap

import (
	appagentruntime "github.com/nikkofu/erp-claw/internal/application/agentruntime"
	sharedoutbox "github.com/nikkofu/erp-claw/internal/application/shared/outbox"
	domainagentruntime "github.com/nikkofu/erp-claw/internal/domain/agentruntime"
	domainapproval "github.com/nikkofu/erp-claw/internal/domain/approval"
	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
	"github.com/nikkofu/erp-claw/internal/interfaces/ws"
	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/health"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
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

type CapabilityCatalog interface {
	domaincap.ModelCatalogRepository
	domaincap.ToolCatalogRepository
}

type GovernanceCatalog interface {
	policy.RuleRepository
	audit.EventStore
}

type OutboxCatalog interface {
	sharedoutbox.RecoveryRepository
	sharedoutbox.MessageReader
}

type Container struct {
	Config              Config
	Health              *health.Service
	ControlPlaneCatalog ControlPlaneCatalog
	AgentRuntimeCatalog AgentRuntimeCatalog
	ApprovalCatalog     ApprovalCatalog
	CapabilityCatalog   CapabilityCatalog
	GovernanceCatalog   GovernanceCatalog
	OutboxCatalog       OutboxCatalog
	WorkspaceGateway    *ws.WorkspaceGateway
}

func NewContainer(cfg Config) *Container {
	return &Container{
		Config:              cfg,
		Health:              health.NewService(),
		ControlPlaneCatalog: newControlPlaneCatalog(cfg),
		AgentRuntimeCatalog: newAgentRuntimeCatalog(cfg),
		ApprovalCatalog:     newApprovalCatalog(cfg),
		CapabilityCatalog:   newCapabilityCatalog(cfg),
		GovernanceCatalog:   newGovernanceCatalog(cfg),
		OutboxCatalog:       newOutboxCatalog(cfg),
		WorkspaceGateway:    ws.NewWorkspaceGateway(),
	}
}

func NewTestContainer() *Container {
	cfg := DefaultConfig()
	cfg.Env = "test"
	cfg.Database.DSN = ""
	return NewContainer(cfg)
}
