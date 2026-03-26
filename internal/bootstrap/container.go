package bootstrap

import (
	"github.com/nikkofu/erp-claw/internal/application/admin/supplychain"
	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/memory"
	"github.com/nikkofu/erp-claw/internal/platform/health"
)

type Container struct {
	Config      Config
	Health      *health.Service
	SupplyChain *supplychain.Service
}

func NewContainer(cfg Config) *Container {
	store := memory.NewSupplyChainStore()
	return &Container{
		Config: cfg,
		Health: health.NewService(),
		SupplyChain: supplychain.NewService(supplychain.ServiceDeps{
			MasterData:     store.MasterDataRepository(),
			PurchaseOrders: store.PurchaseOrderRepository(),
			Approvals:      store.ApprovalRepository(),
			Inventory:      store.InventoryRepository(),
			Pipeline:       shared.NewPipeline(shared.PipelineDeps{}),
		}),
	}
}
