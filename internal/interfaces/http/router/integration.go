package router

import (
	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/application/admin/supplychain"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/presenter"
)

func registerIntegrationRoutes(rg *gin.RouterGroup, container *bootstrap.Container) {
	if container == nil || container.SupplyChain == nil {
		panic("router: integration container must provide supply-chain service")
	}

	readModelGroup := rg.Group("/read-models")
	readModelGroup.GET("/overview", func(c *gin.Context) {
		overview, err := container.SupplyChain.GetBackofficeOverview(c.Request.Context(), supplychain.GetBackofficeOverviewInput{
			TenantID: tenantIDFromContext(c),
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, backofficeOverviewResponse(overview))
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
}
