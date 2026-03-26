package bootstrap

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/application/admin/supplychain"
	"github.com/nikkofu/erp-claw/internal/application/platform/controlplane"
	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/memory"
	"github.com/nikkofu/erp-claw/internal/interfaces/ws"
	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/health"
	"github.com/nikkofu/erp-claw/internal/platform/iam"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
	"github.com/nikkofu/erp-claw/internal/platform/tenant"
)

type Container struct {
	Config           Config
	Health           *health.Service
	SupplyChain      *supplychain.Service
	ControlPlane     *controlplane.Service
	TenantCatalog    tenant.Catalog
	WorkspaceGateway *ws.WorkspaceGateway
}

func NewContainer(cfg Config) *Container {
	supplyChainStore := memory.NewSupplyChainStore()
	controlPlaneStore := memory.NewControlPlaneStore()
	auditRecorder := audit.NewInMemoryRecorder()
	workspaceGateway := ws.NewWorkspaceGateway()

	lookupRoles := func(ctx context.Context, tenantID, actorID string) ([]string, error) {
		if actorID == iam.SystemActor.ID {
			return append([]string(nil), iam.SystemActor.Roles...), nil
		}
		actor, err := controlPlaneStore.IAMDirectory().Get(ctx, tenantID, actorID)
		if err != nil {
			if errors.Is(err, iam.ErrActorNotFound) {
				return nil, nil
			}
			return nil, err
		}
		return append([]string(nil), actor.Roles...), nil
	}

	evaluator := policy.NewRoleEvaluator(
		lookupRoles,
		[]policy.Rule{
			{CommandPrefix: "masterdata.", AnyOfRoles: []string{"platform_admin", "supplychain_operator"}},
			{CommandPrefix: "procurement.", AnyOfRoles: []string{"platform_admin", "supplychain_operator"}},
			{CommandPrefix: "approval.", AnyOfRoles: []string{"platform_admin", "supplychain_operator", "approver"}},
			{CommandPrefix: "controlplane.", AnyOfRoles: []string{"platform_admin"}},
			{CommandPrefix: "runtime.", AnyOfRoles: []string{"platform_admin", "workspace_operator"}},
			{CommandPrefix: "platform.audit.", AnyOfRoles: []string{"platform_admin"}},
		},
	)

	pipeline := shared.NewPipeline(shared.PipelineDeps{
		Policy: evaluator,
		Audit:  auditRecorder,
	})

	tenantCatalog := controlPlaneStore.TenantCatalog()
	return &Container{
		Config: cfg,
		Health: health.NewService(),
		SupplyChain: supplychain.NewService(supplychain.ServiceDeps{
			MasterData:     supplyChainStore.MasterDataRepository(),
			PurchaseOrders: supplyChainStore.PurchaseOrderRepository(),
			Approvals:      supplyChainStore.ApprovalRepository(),
			Inventory:      supplyChainStore.InventoryRepository(),
			Payables:       supplyChainStore.PayableRepository(),
			Pipeline:       pipeline,
		}),
		ControlPlane: controlplane.NewService(controlplane.ServiceDeps{
			TenantCatalog:   tenantCatalog,
			IAMDirectory:    controlPlaneStore.IAMDirectory(),
			Sessions:        controlPlaneStore.SessionRepository(),
			Tasks:           controlPlaneStore.TaskRepository(),
			AuditReader:     auditRecorder,
			WorkspaceEvents: workspaceGateway,
			Pipeline:        pipeline,
		}),
		TenantCatalog:    tenantCatalog,
		WorkspaceGateway: workspaceGateway,
	}
}
