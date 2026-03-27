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
		page, err := parsePositiveSalesOrderQueryInt(c.Query("page"), 1)
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		pageSize, err := parsePositiveSalesOrderQueryInt(c.Query("page_size"), 20)
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}

		orders, err := container.SupplyChain.ListSalesOrders(c.Request.Context(), supplychain.ListSalesOrdersInput{
			TenantID: tenantIDFromContext(c),
			Status:   c.Query("status"),
			Sort:     c.DefaultQuery("sort", "id_desc"),
			Page:     page,
			PageSize: pageSize,
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
		page, err := parsePositivePayableQueryInt(c.Query("page"), 1)
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		pageSize, err := parsePositivePayableQueryInt(c.Query("page_size"), 20)
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}

		bills, err := container.SupplyChain.ListPayableBills(c.Request.Context(), supplychain.ListPayableBillsInput{
			TenantID: tenantIDFromContext(c),
			Status:   c.Query("status"),
			Sort:     c.DefaultQuery("sort", "id_desc"),
			Page:     page,
			PageSize: pageSize,
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
		page, err := parsePositiveReceivableQueryInt(c.Query("page"), 1)
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		pageSize, err := parsePositiveReceivableQueryInt(c.Query("page_size"), 20)
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}

		bills, err := container.SupplyChain.ListReceivableBills(c.Request.Context(), supplychain.ListReceivableBillsInput{
			TenantID: tenantIDFromContext(c),
			Status:   c.Query("status"),
			Sort:     c.DefaultQuery("sort", "id_desc"),
			Page:     page,
			PageSize: pageSize,
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
