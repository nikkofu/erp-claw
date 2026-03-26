package integration

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestAdminInventoryReceiptFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	supplierID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-001",
		"name": "Acme Supply",
	}), "id")
	productID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/products", map[string]any{
		"sku":  "SKU-001",
		"name": "Copper Wire",
		"unit": "roll",
	}), "id")
	warehouseID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/warehouses", map[string]any{
		"code": "WH-SH",
		"name": "Shanghai Warehouse",
	}), "id")

	orderID := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  supplierID,
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	}), "id")

	submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/submit", map[string]any{})
	approvalID := stringField(t, nestedMap(t, submitResp, "approval"), "id")
	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})

	receiptResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/receive", map[string]any{
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	})
	if got := stringField(t, nestedMap(t, receiptResp, "order"), "status"); got != "received" {
		t.Fatalf("expected received order status, got %s", got)
	}

	balanceResp := getJSONData(t, h, "/api/admin/v1/inventory/balances?product_id="+productID+"&warehouse_id="+warehouseID)
	if got := intField(t, balanceResp, "on_hand"); got != 5 {
		t.Fatalf("expected on_hand 5, got %d", got)
	}
	if got := intField(t, balanceResp, "reserved"); got != 0 {
		t.Fatalf("expected reserved 0, got %d", got)
	}
	if got := intField(t, balanceResp, "available"); got != 5 {
		t.Fatalf("expected available 5, got %d", got)
	}
}

func TestAdminInventoryReceiptRequiresApprovedOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	supplierID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-001",
		"name": "Acme Supply",
	}), "id")
	productID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/products", map[string]any{
		"sku":  "SKU-001",
		"name": "Copper Wire",
		"unit": "roll",
	}), "id")
	warehouseID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/warehouses", map[string]any{
		"code": "WH-SH",
		"name": "Shanghai Warehouse",
	}), "id")

	orderID := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  supplierID,
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	}), "id")

	env := doJSON(t, h, http.MethodPost, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/receive", map[string]any{
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	}, http.StatusBadRequest)
	if env.Meta["request_id"] == "" {
		t.Fatal("expected request_id metadata")
	}
}

func TestAdminInventoryReservationFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	supplierID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-001",
		"name": "Acme Supply",
	}), "id")
	productID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/products", map[string]any{
		"sku":  "SKU-001",
		"name": "Copper Wire",
		"unit": "roll",
	}), "id")
	warehouseID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/warehouses", map[string]any{
		"code": "WH-SH",
		"name": "Shanghai Warehouse",
	}), "id")

	orderID := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  supplierID,
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	}), "id")
	submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/submit", map[string]any{})
	approvalID := stringField(t, nestedMap(t, submitResp, "approval"), "id")
	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})
	postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/receive", map[string]any{
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	})

	reservationResp := postJSONData(t, h, "/api/admin/v1/inventory/reservations", map[string]any{
		"product_id":     productID,
		"warehouse_id":   warehouseID,
		"quantity":       2,
		"reference_type": "sales_order",
		"reference_id":   "so-001",
	})
	if got := intField(t, reservationResp, "quantity"); got != 2 {
		t.Fatalf("expected reservation quantity 2, got %d", got)
	}
	if got := stringField(t, reservationResp, "status"); got != "active" {
		t.Fatalf("expected reservation status active, got %s", got)
	}

	balanceResp := getJSONData(t, h, "/api/admin/v1/inventory/balances?product_id="+productID+"&warehouse_id="+warehouseID)
	if got := intField(t, balanceResp, "on_hand"); got != 5 {
		t.Fatalf("expected on_hand 5, got %d", got)
	}
	if got := intField(t, balanceResp, "reserved"); got != 2 {
		t.Fatalf("expected reserved 2, got %d", got)
	}
	if got := intField(t, balanceResp, "available"); got != 3 {
		t.Fatalf("expected available 3, got %d", got)
	}
}

func TestAdminInventoryReservationRejectsExcessQuantity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	supplierID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-001",
		"name": "Acme Supply",
	}), "id")
	productID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/products", map[string]any{
		"sku":  "SKU-001",
		"name": "Copper Wire",
		"unit": "roll",
	}), "id")
	warehouseID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/warehouses", map[string]any{
		"code": "WH-SH",
		"name": "Shanghai Warehouse",
	}), "id")

	orderID := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  supplierID,
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	}), "id")
	submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/submit", map[string]any{})
	approvalID := stringField(t, nestedMap(t, submitResp, "approval"), "id")
	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})
	postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/receive", map[string]any{
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	})

	env := doJSON(t, h, http.MethodPost, "/api/admin/v1/inventory/reservations", map[string]any{
		"product_id":     productID,
		"warehouse_id":   warehouseID,
		"quantity":       6,
		"reference_type": "sales_order",
		"reference_id":   "so-002",
	}, http.StatusBadRequest)
	if env.Meta["request_id"] == "" {
		t.Fatal("expected request_id metadata")
	}
}

func TestAdminInventoryOutboundFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	supplierID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-001",
		"name": "Acme Supply",
	}), "id")
	productID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/products", map[string]any{
		"sku":  "SKU-001",
		"name": "Copper Wire",
		"unit": "roll",
	}), "id")
	warehouseID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/warehouses", map[string]any{
		"code": "WH-SH",
		"name": "Shanghai Warehouse",
	}), "id")

	orderID := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  supplierID,
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	}), "id")
	submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/submit", map[string]any{})
	approvalID := stringField(t, nestedMap(t, submitResp, "approval"), "id")
	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})
	postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/receive", map[string]any{
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	})

	outboundResp := postJSONData(t, h, "/api/admin/v1/inventory/outbounds", map[string]any{
		"product_id":     productID,
		"warehouse_id":   warehouseID,
		"quantity":       2,
		"reference_type": "shipment",
		"reference_id":   "shp-001",
	})
	if got := stringField(t, outboundResp, "movement_type"); got != "outbound" {
		t.Fatalf("expected movement_type outbound, got %s", got)
	}
	if got := intField(t, outboundResp, "quantity_delta"); got != -2 {
		t.Fatalf("expected quantity_delta -2, got %d", got)
	}

	balanceResp := getJSONData(t, h, "/api/admin/v1/inventory/balances?product_id="+productID+"&warehouse_id="+warehouseID)
	if got := intField(t, balanceResp, "on_hand"); got != 3 {
		t.Fatalf("expected on_hand 3, got %d", got)
	}
	if got := intField(t, balanceResp, "reserved"); got != 0 {
		t.Fatalf("expected reserved 0, got %d", got)
	}
	if got := intField(t, balanceResp, "available"); got != 3 {
		t.Fatalf("expected available 3, got %d", got)
	}
}

func TestAdminInventoryOutboundRejectsExcessQuantity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	supplierID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-001",
		"name": "Acme Supply",
	}), "id")
	productID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/products", map[string]any{
		"sku":  "SKU-001",
		"name": "Copper Wire",
		"unit": "roll",
	}), "id")
	warehouseID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/warehouses", map[string]any{
		"code": "WH-SH",
		"name": "Shanghai Warehouse",
	}), "id")

	orderID := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  supplierID,
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	}), "id")
	submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/submit", map[string]any{})
	approvalID := stringField(t, nestedMap(t, submitResp, "approval"), "id")
	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})
	postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/receive", map[string]any{
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	})

	env := doJSON(t, h, http.MethodPost, "/api/admin/v1/inventory/outbounds", map[string]any{
		"product_id":     productID,
		"warehouse_id":   warehouseID,
		"quantity":       6,
		"reference_type": "shipment",
		"reference_id":   "shp-002",
	}, http.StatusBadRequest)
	if env.Meta["request_id"] == "" {
		t.Fatal("expected request_id metadata")
	}
}
