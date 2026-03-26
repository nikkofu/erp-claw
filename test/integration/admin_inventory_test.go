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

func TestAdminInventoryTransferFlow(t *testing.T) {
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
	sourceWarehouseID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/warehouses", map[string]any{
		"code": "WH-SH",
		"name": "Shanghai Warehouse",
	}), "id")
	targetWarehouseID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/warehouses", map[string]any{
		"code": "WH-BJ",
		"name": "Beijing Warehouse",
	}), "id")

	orderID := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  supplierID,
		"warehouse_id": sourceWarehouseID,
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

	transferResp := postJSONData(t, h, "/api/admin/v1/inventory/transfers", map[string]any{
		"product_id":        productID,
		"from_warehouse_id": sourceWarehouseID,
		"to_warehouse_id":   targetWarehouseID,
		"quantity":          2,
		"reference_type":    "transfer_order",
		"reference_id":      "trf-001",
	})
	rawEntries, ok := transferResp["ledger_entries"].([]any)
	if !ok {
		t.Fatalf("expected ledger_entries array, got %#v", transferResp["ledger_entries"])
	}
	if len(rawEntries) != 2 {
		t.Fatalf("expected 2 transfer ledger entries, got %d", len(rawEntries))
	}
	first, ok := rawEntries[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first ledger entry object, got %#v", rawEntries[0])
	}
	second, ok := rawEntries[1].(map[string]any)
	if !ok {
		t.Fatalf("expected second ledger entry object, got %#v", rawEntries[1])
	}
	if got := stringField(t, first, "movement_type"); got != "outbound" {
		t.Fatalf("expected first movement_type outbound, got %s", got)
	}
	if got := intField(t, first, "quantity_delta"); got != -2 {
		t.Fatalf("expected first quantity_delta -2, got %d", got)
	}
	if got := stringField(t, second, "movement_type"); got != "inbound" {
		t.Fatalf("expected second movement_type inbound, got %s", got)
	}
	if got := intField(t, second, "quantity_delta"); got != 2 {
		t.Fatalf("expected second quantity_delta 2, got %d", got)
	}

	sourceBalanceResp := getJSONData(t, h, "/api/admin/v1/inventory/balances?product_id="+productID+"&warehouse_id="+sourceWarehouseID)
	if got := intField(t, sourceBalanceResp, "on_hand"); got != 3 {
		t.Fatalf("expected source on_hand 3, got %d", got)
	}
	if got := intField(t, sourceBalanceResp, "available"); got != 3 {
		t.Fatalf("expected source available 3, got %d", got)
	}

	targetBalanceResp := getJSONData(t, h, "/api/admin/v1/inventory/balances?product_id="+productID+"&warehouse_id="+targetWarehouseID)
	if got := intField(t, targetBalanceResp, "on_hand"); got != 2 {
		t.Fatalf("expected target on_hand 2, got %d", got)
	}
	if got := intField(t, targetBalanceResp, "available"); got != 2 {
		t.Fatalf("expected target available 2, got %d", got)
	}
}

func TestAdminInventoryTransferRejectsExcessQuantity(t *testing.T) {
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
	sourceWarehouseID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/warehouses", map[string]any{
		"code": "WH-SH",
		"name": "Shanghai Warehouse",
	}), "id")
	targetWarehouseID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/warehouses", map[string]any{
		"code": "WH-BJ",
		"name": "Beijing Warehouse",
	}), "id")

	orderID := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  supplierID,
		"warehouse_id": sourceWarehouseID,
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

	env := doJSON(t, h, http.MethodPost, "/api/admin/v1/inventory/transfers", map[string]any{
		"product_id":        productID,
		"from_warehouse_id": sourceWarehouseID,
		"to_warehouse_id":   targetWarehouseID,
		"quantity":          6,
		"reference_type":    "transfer_order",
		"reference_id":      "trf-002",
	}, http.StatusBadRequest)
	if env.Meta["request_id"] == "" {
		t.Fatal("expected request_id metadata")
	}
}
