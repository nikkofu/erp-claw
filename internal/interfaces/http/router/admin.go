package router

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/application/admin/supplychain"
	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/domain/approval"
	"github.com/nikkofu/erp-claw/internal/domain/inventory"
	"github.com/nikkofu/erp-claw/internal/domain/masterdata"
	"github.com/nikkofu/erp-claw/internal/domain/payable"
	"github.com/nikkofu/erp-claw/internal/domain/procurement"
	"github.com/nikkofu/erp-claw/internal/domain/receivable"
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

	procurementGroup.POST("/:id/receive", func(c *gin.Context) {
		var req receivePurchaseOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		lines := make([]supplychain.ReceivePurchaseOrderLine, 0, len(req.Lines))
		for _, line := range req.Lines {
			lines = append(lines, supplychain.ReceivePurchaseOrderLine{
				ProductID: line.ProductID,
				Quantity:  line.Quantity,
			})
		}

		receipt, ledgerEntries, order, err := container.SupplyChain.ReceivePurchaseOrder(c.Request.Context(), supplychain.ReceivePurchaseOrderInput{
			TenantID:        tenantIDFromContext(c),
			ActorID:         actorIDFromContext(c),
			PurchaseOrderID: c.Param("id"),
			Lines:           lines,
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, purchaseOrderReceiptResponse(receipt, ledgerEntries, order))
	})

	procurementGroup.POST("/:id/payable-bills", func(c *gin.Context) {
		bill, err := container.SupplyChain.CreatePayableBill(c.Request.Context(), supplychain.CreatePayableBillInput{
			TenantID:        tenantIDFromContext(c),
			ActorID:         actorIDFromContext(c),
			PurchaseOrderID: c.Param("id"),
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, payableBillResponse(bill))
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

	inventoryGroup := rg.Group("/inventory")
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
	inventoryGroup.POST("/reservations", func(c *gin.Context) {
		var req createInventoryReservationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		reservation, err := container.SupplyChain.ReserveInventory(c.Request.Context(), supplychain.ReserveInventoryInput{
			TenantID:      tenantIDFromContext(c),
			ActorID:       actorIDFromContext(c),
			ProductID:     req.ProductID,
			WarehouseID:   req.WarehouseID,
			Quantity:      req.Quantity,
			ReferenceType: req.ReferenceType,
			ReferenceID:   req.ReferenceID,
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, inventoryReservationResponse(reservation))
	})
	inventoryGroup.POST("/outbounds", func(c *gin.Context) {
		var req createInventoryOutboundRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		outbound, err := container.SupplyChain.IssueInventory(c.Request.Context(), supplychain.IssueInventoryInput{
			TenantID:      tenantIDFromContext(c),
			ActorID:       actorIDFromContext(c),
			ProductID:     req.ProductID,
			WarehouseID:   req.WarehouseID,
			Quantity:      req.Quantity,
			ReferenceType: req.ReferenceType,
			ReferenceID:   req.ReferenceID,
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, inventoryOutboundResponse(outbound))
	})
	inventoryGroup.POST("/transfers", func(c *gin.Context) {
		var req createInventoryTransferRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		entries, err := container.SupplyChain.TransferInventory(c.Request.Context(), supplychain.TransferInventoryInput{
			TenantID:        tenantIDFromContext(c),
			ActorID:         actorIDFromContext(c),
			ProductID:       req.ProductID,
			FromWarehouseID: req.FromWarehouseID,
			ToWarehouseID:   req.ToWarehouseID,
			Quantity:        req.Quantity,
			ReferenceType:   req.ReferenceType,
			ReferenceID:     req.ReferenceID,
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, inventoryTransferResponse(entries))
	})

	receivableGroup := rg.Group("/receivables")
	receivableGroup.POST("", func(c *gin.Context) {
		var req createReceivableBillRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		bill, err := container.SupplyChain.CreateReceivableBill(c.Request.Context(), supplychain.CreateReceivableBillInput{
			TenantID:    tenantIDFromContext(c),
			ActorID:     actorIDFromContext(c),
			ExternalRef: req.ExternalRef,
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, receivableBillResponse(bill))
	})
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
	payableGroup.POST("/:id/payment-plans", func(c *gin.Context) {
		var req createPayablePaymentPlanRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			presenter.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		plan, err := container.SupplyChain.CreatePayablePaymentPlan(c.Request.Context(), supplychain.CreatePayablePaymentPlanInput{
			TenantID:       tenantIDFromContext(c),
			ActorID:        actorIDFromContext(c),
			PayableBillID:  c.Param("id"),
			DueDateISO8601: req.DueDate,
		})
		if err != nil {
			renderSupplyChainError(c, err)
			return
		}
		presenter.OK(c, payablePaymentPlanResponse(plan))
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

type receivePurchaseOrderRequest struct {
	Lines []receivePurchaseOrderLineRequest `json:"lines"`
}

type receivePurchaseOrderLineRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type createInventoryReservationRequest struct {
	ProductID     string `json:"product_id"`
	WarehouseID   string `json:"warehouse_id"`
	Quantity      int    `json:"quantity"`
	ReferenceType string `json:"reference_type"`
	ReferenceID   string `json:"reference_id"`
}

type createInventoryOutboundRequest struct {
	ProductID     string `json:"product_id"`
	WarehouseID   string `json:"warehouse_id"`
	Quantity      int    `json:"quantity"`
	ReferenceType string `json:"reference_type"`
	ReferenceID   string `json:"reference_id"`
}

type createInventoryTransferRequest struct {
	ProductID       string `json:"product_id"`
	FromWarehouseID string `json:"from_warehouse_id"`
	ToWarehouseID   string `json:"to_warehouse_id"`
	Quantity        int    `json:"quantity"`
	ReferenceType   string `json:"reference_type"`
	ReferenceID     string `json:"reference_id"`
}

type createPayablePaymentPlanRequest struct {
	DueDate string `json:"due_date"`
}

type createReceivableBillRequest struct {
	ExternalRef string `json:"external_ref"`
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
		errors.Is(err, approval.ErrRequestNotFound),
		errors.Is(err, payable.ErrBillNotFound),
		errors.Is(err, receivable.ErrBillNotFound):
		presenter.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, masterdata.ErrInvalidSupplier),
		errors.Is(err, masterdata.ErrInvalidProduct),
		errors.Is(err, masterdata.ErrInvalidWarehouse),
		errors.Is(err, inventory.ErrInvalidReceipt),
		errors.Is(err, inventory.ErrInvalidReservation),
		errors.Is(err, inventory.ErrInsufficientAvailableInventory),
		errors.Is(err, procurement.ErrInvalidPurchaseOrder),
		errors.Is(err, procurement.ErrPurchaseOrderAlreadySubmitted),
		errors.Is(err, procurement.ErrPurchaseOrderNotReceivable),
		errors.Is(err, payable.ErrInvalidBill),
		errors.Is(err, payable.ErrBillAlreadyExists),
		errors.Is(err, payable.ErrOrderNotBillable),
		errors.Is(err, payable.ErrInvalidPaymentPlan),
		errors.Is(err, receivable.ErrInvalidBill),
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

func receiptResponse(receipt inventory.Receipt) gin.H {
	lines := make([]gin.H, 0, len(receipt.Lines))
	for _, line := range receipt.Lines {
		lines = append(lines, gin.H{
			"product_id": line.ProductID,
			"quantity":   line.Quantity,
		})
	}
	return gin.H{
		"id":                receipt.ID,
		"tenant_id":         receipt.TenantID,
		"purchase_order_id": receipt.PurchaseOrderID,
		"warehouse_id":      receipt.WarehouseID,
		"status":            receipt.Status,
		"created_by":        receipt.CreatedBy,
		"lines":             lines,
	}
}

func ledgerEntriesResponse(entries []inventory.LedgerEntry) []gin.H {
	out := make([]gin.H, 0, len(entries))
	for _, entry := range entries {
		out = append(out, gin.H{
			"id":             entry.ID,
			"tenant_id":      entry.TenantID,
			"product_id":     entry.ProductID,
			"warehouse_id":   entry.WarehouseID,
			"movement_type":  entry.MovementType,
			"quantity_delta": entry.QuantityDelta,
			"reference_type": entry.ReferenceType,
			"reference_id":   entry.ReferenceID,
		})
	}
	return out
}

func purchaseOrderReceiptResponse(receipt inventory.Receipt, entries []inventory.LedgerEntry, order procurement.PurchaseOrder) gin.H {
	return gin.H{
		"receipt":        receiptResponse(receipt),
		"ledger_entries": ledgerEntriesResponse(entries),
		"order":          purchaseOrderResponse(order),
	}
}

func inventoryBalanceResponse(balance inventory.Balance) gin.H {
	return gin.H{
		"tenant_id":    balance.TenantID,
		"product_id":   balance.ProductID,
		"warehouse_id": balance.WarehouseID,
		"on_hand":      balance.OnHand,
		"reserved":     balance.Reserved,
		"available":    balance.Available,
	}
}

func inventoryReservationResponse(reservation inventory.Reservation) gin.H {
	return gin.H{
		"id":             reservation.ID,
		"tenant_id":      reservation.TenantID,
		"product_id":     reservation.ProductID,
		"warehouse_id":   reservation.WarehouseID,
		"reference_type": reservation.ReferenceType,
		"reference_id":   reservation.ReferenceID,
		"status":         reservation.Status,
		"created_by":     reservation.CreatedBy,
		"quantity":       reservation.Quantity,
	}
}

func inventoryOutboundResponse(entry inventory.LedgerEntry) gin.H {
	return gin.H{
		"id":             entry.ID,
		"tenant_id":      entry.TenantID,
		"product_id":     entry.ProductID,
		"warehouse_id":   entry.WarehouseID,
		"movement_type":  entry.MovementType,
		"quantity_delta": entry.QuantityDelta,
		"reference_type": entry.ReferenceType,
		"reference_id":   entry.ReferenceID,
	}
}

func inventoryTransferResponse(entries []inventory.LedgerEntry) gin.H {
	return gin.H{
		"ledger_entries": ledgerEntriesResponse(entries),
	}
}

func payableBillResponse(bill payable.Bill) gin.H {
	return gin.H{
		"id":                bill.ID,
		"tenant_id":         bill.TenantID,
		"purchase_order_id": bill.PurchaseOrderID,
		"status":            bill.Status,
		"created_by":        bill.CreatedBy,
	}
}

func payablePaymentPlanResponse(plan payable.PaymentPlan) gin.H {
	return gin.H{
		"id":              plan.ID,
		"tenant_id":       plan.TenantID,
		"payable_bill_id": plan.PayableBillID,
		"status":          plan.Status,
		"due_date":        plan.DueDateISO8601,
		"created_by":      plan.CreatedBy,
	}
}

func payablePaymentPlansResponse(plans []payable.PaymentPlan) []gin.H {
	out := make([]gin.H, 0, len(plans))
	for _, plan := range plans {
		out = append(out, payablePaymentPlanResponse(plan))
	}
	return out
}

func payableBillDetailResponse(bill payable.Bill, plans []payable.PaymentPlan) gin.H {
	resp := payableBillResponse(bill)
	resp["payment_plans"] = payablePaymentPlansResponse(plans)
	return resp
}

func payableBillsResponse(bills []payable.Bill) []gin.H {
	out := make([]gin.H, 0, len(bills))
	for _, bill := range bills {
		out = append(out, payableBillResponse(bill))
	}
	return out
}

func receivableBillResponse(bill receivable.Bill) gin.H {
	return gin.H{
		"id":           bill.ID,
		"tenant_id":    bill.TenantID,
		"external_ref": bill.ExternalRef,
		"status":       bill.Status,
		"created_by":   bill.CreatedBy,
	}
}

func receivableBillsResponse(bills []receivable.Bill) []gin.H {
	out := make([]gin.H, 0, len(bills))
	for _, bill := range bills {
		out = append(out, receivableBillResponse(bill))
	}
	return out
}
