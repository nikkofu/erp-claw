package router

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/application/admin/supplychain"
	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/domain/approval"
	"github.com/nikkofu/erp-claw/internal/domain/masterdata"
	"github.com/nikkofu/erp-claw/internal/domain/procurement"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/presenter"
)

func registerAdminRoutes(rg *gin.RouterGroup, container *bootstrap.Container) {
	if container == nil || container.SupplyChain == nil {
		panic("router: admin container must provide supply-chain service")
	}

	masterDataGroup := rg.Group("/master-data")
	masterDataGroup.POST("/suppliers", func(c *gin.Context) {
		var req createSupplierRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		supplier, err := container.SupplyChain.CreateSupplier(c.Request.Context(), supplychain.CreateSupplierInput{
			TenantID: tenantIDFromContext(c),
			ActorID:  actorIDFromContext(c),
			Code:     req.Code,
			Name:     req.Name,
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, supplierResponse(supplier))
	})

	masterDataGroup.POST("/products", func(c *gin.Context) {
		var req createProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		product, err := container.SupplyChain.CreateProduct(c.Request.Context(), supplychain.CreateProductInput{
			TenantID: tenantIDFromContext(c),
			ActorID:  actorIDFromContext(c),
			SKU:      req.SKU,
			Name:     req.Name,
			Unit:     req.Unit,
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, productResponse(product))
	})

	masterDataGroup.POST("/warehouses", func(c *gin.Context) {
		var req createWarehouseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		warehouse, err := container.SupplyChain.CreateWarehouse(c.Request.Context(), supplychain.CreateWarehouseInput{
			TenantID: tenantIDFromContext(c),
			ActorID:  actorIDFromContext(c),
			Code:     req.Code,
			Name:     req.Name,
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, warehouseResponse(warehouse))
	})

	procurementGroup := rg.Group("/procurement/purchase-orders")
	procurementGroup.POST("", func(c *gin.Context) {
		var req createPurchaseOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		lines := make([]supplychain.CreatePurchaseOrderLine, 0, len(req.Lines))
		for _, line := range req.Lines {
			lines = append(lines, supplychain.CreatePurchaseOrderLine{
				ProductID: line.ProductID,
				Quantity:  line.Quantity,
			})
		}

		order, err := container.SupplyChain.CreatePurchaseOrder(c.Request.Context(), supplychain.CreatePurchaseOrderInput{
			TenantID:    tenantIDFromContext(c),
			ActorID:     actorIDFromContext(c),
			SupplierID:  req.SupplierID,
			WarehouseID: req.WarehouseID,
			Lines:       lines,
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, purchaseOrderResponse(order))
	})

	procurementGroup.POST("/:id/submit", func(c *gin.Context) {
		order, request, err := container.SupplyChain.SubmitPurchaseOrder(c.Request.Context(), supplychain.SubmitPurchaseOrderInput{
			TenantID:        tenantIDFromContext(c),
			ActorID:         actorIDFromContext(c),
			PurchaseOrderID: c.Param("id"),
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, purchaseOrderDetailResponse(order, request))
	})

	procurementGroup.GET("/:id", func(c *gin.Context) {
		order, request, err := container.SupplyChain.GetPurchaseOrder(c.Request.Context(), tenantIDFromContext(c), c.Param("id"))
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, purchaseOrderDetailResponse(order, request))
	})

	approvalGroup := rg.Group("/approvals")
	approvalGroup.POST("/:id/approve", func(c *gin.Context) {
		order, request, err := container.SupplyChain.ApproveRequest(c.Request.Context(), supplychain.ResolveApprovalInput{
			TenantID:   tenantIDFromContext(c),
			ActorID:    actorIDFromContext(c),
			ApprovalID: c.Param("id"),
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, purchaseOrderDetailResponse(order, request))
	})

	approvalGroup.POST("/:id/reject", func(c *gin.Context) {
		order, request, err := container.SupplyChain.RejectRequest(c.Request.Context(), supplychain.ResolveApprovalInput{
			TenantID:   tenantIDFromContext(c),
			ActorID:    actorIDFromContext(c),
			ApprovalID: c.Param("id"),
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, purchaseOrderDetailResponse(order, request))
	})
}

type createSupplierRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type createProductRequest struct {
	SKU  string `json:"sku"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type createWarehouseRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type createPurchaseOrderRequest struct {
	SupplierID  string                           `json:"supplier_id"`
	WarehouseID string                           `json:"warehouse_id"`
	Lines       []createPurchaseOrderLineRequest `json:"lines"`
}

type createPurchaseOrderLineRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func tenantIDFromContext(c *gin.Context) string {
	return c.GetString("tenant_id")
}

func actorIDFromContext(c *gin.Context) string {
	return c.GetString("actor_id")
}

func renderSupplyChainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, masterdata.ErrSupplierNotFound),
		errors.Is(err, masterdata.ErrProductNotFound),
		errors.Is(err, masterdata.ErrWarehouseNotFound),
		errors.Is(err, procurement.ErrPurchaseOrderNotFound),
		errors.Is(err, approval.ErrRequestNotFound):
		presenter.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, masterdata.ErrInvalidSupplier),
		errors.Is(err, masterdata.ErrInvalidProduct),
		errors.Is(err, masterdata.ErrInvalidWarehouse),
		errors.Is(err, procurement.ErrInvalidPurchaseOrder),
		errors.Is(err, procurement.ErrPurchaseOrderAlreadySubmitted),
		errors.Is(err, approval.ErrInvalidRequest),
		errors.Is(err, approval.ErrApprovalNotPending):
		presenter.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, shared.ErrPolicyDenied):
		presenter.Error(c, http.StatusForbidden, err.Error())
	case errors.Is(err, shared.ErrApprovalRequired):
		presenter.Error(c, http.StatusConflict, err.Error())
	default:
		presenter.Error(c, http.StatusInternalServerError, err.Error())
	}
}

func supplierResponse(supplier masterdata.Supplier) gin.H {
	return gin.H{
		"id":        supplier.ID,
		"tenant_id": supplier.TenantID,
		"code":      supplier.Code,
		"name":      supplier.Name,
	}
}

func productResponse(product masterdata.Product) gin.H {
	return gin.H{
		"id":        product.ID,
		"tenant_id": product.TenantID,
		"sku":       product.SKU,
		"name":      product.Name,
		"unit":      product.Unit,
	}
}

func warehouseResponse(warehouse masterdata.Warehouse) gin.H {
	return gin.H{
		"id":        warehouse.ID,
		"tenant_id": warehouse.TenantID,
		"code":      warehouse.Code,
		"name":      warehouse.Name,
	}
}

func purchaseOrderResponse(order procurement.PurchaseOrder) gin.H {
	lines := make([]gin.H, 0, len(order.Lines))
	for _, line := range order.Lines {
		lines = append(lines, gin.H{
			"product_id": line.ProductID,
			"quantity":   line.Quantity,
		})
	}
	return gin.H{
		"id":           order.ID,
		"tenant_id":    order.TenantID,
		"supplier_id":  order.SupplierID,
		"warehouse_id": order.WarehouseID,
		"status":       order.Status,
		"approval_id":  order.ApprovalID,
		"lines":        lines,
	}
}

func approvalResponse(request approval.Request) any {
	if request.ID == "" {
		return nil
	}
	return gin.H{
		"id":            request.ID,
		"tenant_id":     request.TenantID,
		"resource_type": request.ResourceType,
		"resource_id":   request.ResourceID,
		"status":        request.Status,
		"requested_by":  request.RequestedBy,
		"decided_by":    request.DecidedBy,
	}
}

func purchaseOrderDetailResponse(order procurement.PurchaseOrder, request approval.Request) gin.H {
	return gin.H{
		"order":    purchaseOrderResponse(order),
		"approval": approvalResponse(request),
	}
}
