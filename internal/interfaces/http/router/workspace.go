package router

import (
	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/application/admin/supplychain"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/presenter"
)

func registerWorkspaceRoutes(rg *gin.RouterGroup, container *bootstrap.Container) {
	if container == nil || container.SupplyChain == nil {
		panic("router: workspace container must provide supply-chain service")
	}

	inventoryGroup := rg.Group("/inventory")
	inventoryGroup.GET("/balances", func(c *gin.Context) {
		balance, err := container.SupplyChain.GetInventoryBalance(c.Request.Context(), supplychain.GetInventoryBalanceInput{
			TenantID:    tenantIDFromContext(c),
			ProductID:   c.Query("product_id"),
			WarehouseID: c.Query("warehouse_id"),
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, inventoryBalanceResponse(balance))
	})
	inventoryGroup.GET("/ledger", func(c *gin.Context) {
		entries, err := container.SupplyChain.ListInventoryLedger(c.Request.Context(), supplychain.ListInventoryLedgerInput{
			TenantID:    tenantIDFromContext(c),
			ProductID:   c.Query("product_id"),
			WarehouseID: c.Query("warehouse_id"),
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, ledgerEntriesResponse(entries))
	})
}
