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

	salesGroup := rg.Group("/sales-orders")
	salesGroup.GET("", func(c *gin.Context) {
		orders, err := container.SupplyChain.ListSalesOrders(c.Request.Context(), supplychain.ListSalesOrdersInput{
			TenantID: tenantIDFromContext(c),
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, salesOrdersResponse(orders))
	})
	salesGroup.GET("/:id", func(c *gin.Context) {
		order, err := container.SupplyChain.GetSalesOrder(c.Request.Context(), supplychain.GetSalesOrderInput{
			TenantID:     tenantIDFromContext(c),
			SalesOrderID: c.Param("id"),
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, salesOrderResponse(order))
	})

	payableGroup := rg.Group("/payables")
	payableGroup.GET("", func(c *gin.Context) {
		bills, err := container.SupplyChain.ListPayableBills(c.Request.Context(), supplychain.ListPayableBillsInput{
			TenantID: tenantIDFromContext(c),
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, payableBillsResponse(bills))
	})
	payableGroup.GET("/:id", func(c *gin.Context) {
		bill, err := container.SupplyChain.GetPayableBill(c.Request.Context(), supplychain.GetPayableBillInput{
			TenantID: tenantIDFromContext(c),
			BillID:   c.Param("id"),
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}

		plans, err := container.SupplyChain.ListPayablePaymentPlans(c.Request.Context(), supplychain.ListPayablePaymentPlansInput{
			TenantID:      tenantIDFromContext(c),
			PayableBillID: bill.ID,
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, payableBillDetailResponse(bill, plans))
	})

	receivableGroup := rg.Group("/receivables")
	receivableGroup.GET("", func(c *gin.Context) {
		bills, err := container.SupplyChain.ListReceivableBills(c.Request.Context(), supplychain.ListReceivableBillsInput{
			TenantID: tenantIDFromContext(c),
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, receivableBillsResponse(bills))
	})
	receivableGroup.GET("/:id", func(c *gin.Context) {
		bill, err := container.SupplyChain.GetReceivableBill(c.Request.Context(), supplychain.GetReceivableBillInput{
			TenantID: tenantIDFromContext(c),
			BillID:   c.Param("id"),
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, receivableBillResponse(bill))
	})
}
