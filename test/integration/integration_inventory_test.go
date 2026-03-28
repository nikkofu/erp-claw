package integration

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestIntegrationInventoryQueriesReturnBalanceAndLedger(t *testing.T) {
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
	postJSONData(t, h, "/api/admin/v1/inventory/outbounds", map[string]any{
		"product_id":     productID,
		"warehouse_id":   warehouseID,
		"quantity":       2,
		"reference_type": "shipment",
		"reference_id":   "shp-int-inv-001",
	})

	balance := doJSON(t, h, http.MethodGet, "/api/integration/v1/inventory/balances?product_id="+productID+"&warehouse_id="+warehouseID, nil, http.StatusOK).Data
	if got := intField(t, balance, "on_hand"); got != 3 {
		t.Fatalf("expected integration on_hand 3, got %d", got)
	}
	if got := intField(t, balance, "reserved"); got != 0 {
		t.Fatalf("expected integration reserved 0, got %d", got)
	}
	if got := intField(t, balance, "available"); got != 3 {
		t.Fatalf("expected integration available 3, got %d", got)
	}

	ledger := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/inventory/ledger?product_id="+productID+"&warehouse_id="+warehouseID, nil, http.StatusOK).Data
	if len(ledger) != 2 {
		t.Fatalf("expected 2 integration ledger entries, got %d", len(ledger))
	}
	if got := stringField(t, ledger[0], "movement_type"); got != "inbound" {
		t.Fatalf("expected first integration movement_type inbound, got %s", got)
	}
	if got := intField(t, ledger[0], "quantity_delta"); got != 5 {
		t.Fatalf("expected first integration quantity_delta 5, got %d", got)
	}
	if got := stringField(t, ledger[1], "movement_type"); got != "outbound" {
		t.Fatalf("expected second integration movement_type outbound, got %s", got)
	}
	if got := intField(t, ledger[1], "quantity_delta"); got != -2 {
		t.Fatalf("expected second integration quantity_delta -2, got %d", got)
	}
}

func TestIntegrationInventoryLedgerListSupportsSortAndPagination(t *testing.T) {
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
	postJSONData(t, h, "/api/admin/v1/inventory/outbounds", map[string]any{
		"product_id":     productID,
		"warehouse_id":   warehouseID,
		"quantity":       2,
		"reference_type": "shipment",
		"reference_id":   "shp-int-inv-sort-001",
	})
	postJSONData(t, h, "/api/admin/v1/inventory/outbounds", map[string]any{
		"product_id":     productID,
		"warehouse_id":   warehouseID,
		"quantity":       1,
		"reference_type": "shipment",
		"reference_id":   "shp-int-inv-sort-002",
	})

	descPage1 := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/inventory/ledger?product_id="+productID+"&warehouse_id="+warehouseID+"&sort=id_desc&page=1&page_size=1", nil, http.StatusOK).Data
	if len(descPage1) != 1 {
		t.Fatalf("expected 1 integration ledger entry in desc page1, got %d", len(descPage1))
	}
	if got := intField(t, descPage1[0], "quantity_delta"); got != -1 {
		t.Fatalf("expected integration desc page1 quantity_delta -1, got %d", got)
	}

	ascPage2 := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/inventory/ledger?product_id="+productID+"&warehouse_id="+warehouseID+"&sort=id_asc&page=2&page_size=2", nil, http.StatusOK).Data
	if len(ascPage2) != 1 {
		t.Fatalf("expected 1 integration ledger entry in asc page2, got %d", len(ascPage2))
	}
	if got := intField(t, ascPage2[0], "quantity_delta"); got != -1 {
		t.Fatalf("expected integration asc page2 quantity_delta -1, got %d", got)
	}
}

func TestIntegrationInventoryQueriesRejectInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	cases := []string{
		"/api/integration/v1/inventory/balances",
		"/api/integration/v1/inventory/balances?product_id=prd-001",
		"/api/integration/v1/inventory/balances?warehouse_id=wh-001",
		"/api/integration/v1/inventory/ledger?product_id=prd-001&warehouse_id=wh-001&sort=unknown",
		"/api/integration/v1/inventory/ledger?product_id=prd-001&warehouse_id=wh-001&page=0",
		"/api/integration/v1/inventory/ledger?product_id=prd-001&warehouse_id=wh-001&page_size=0",
	}
	for _, path := range cases {
		env := doJSON(t, h, http.MethodGet, path, nil, http.StatusBadRequest)
		if env.Meta["request_id"] == "" {
			t.Fatalf("expected request_id metadata for %s", path)
		}
	}
}
