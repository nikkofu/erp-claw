package integration

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestWorkspaceInventoryQueriesReturnBalanceAndLedger(t *testing.T) {
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
		"reference_id":   "shp-workspace-001",
	})

	balance := doJSON(t, h, http.MethodGet, "/api/workspace/v1/inventory/balances?product_id="+productID+"&warehouse_id="+warehouseID, nil, http.StatusOK).Data
	if got := intField(t, balance, "on_hand"); got != 3 {
		t.Fatalf("expected on_hand 3, got %d", got)
	}
	if got := intField(t, balance, "reserved"); got != 0 {
		t.Fatalf("expected reserved 0, got %d", got)
	}
	if got := intField(t, balance, "available"); got != 3 {
		t.Fatalf("expected available 3, got %d", got)
	}

	ledgerEntries := doJSONForArray(t, h, http.MethodGet, "/api/workspace/v1/inventory/ledger?product_id="+productID+"&warehouse_id="+warehouseID, nil, http.StatusOK).Data
	if len(ledgerEntries) != 2 {
		t.Fatalf("expected 2 ledger entries, got %d", len(ledgerEntries))
	}
	if got := stringField(t, ledgerEntries[0], "movement_type"); got != "inbound" {
		t.Fatalf("expected first movement_type inbound, got %s", got)
	}
	if got := intField(t, ledgerEntries[0], "quantity_delta"); got != 5 {
		t.Fatalf("expected first quantity_delta 5, got %d", got)
	}
	if got := stringField(t, ledgerEntries[1], "movement_type"); got != "outbound" {
		t.Fatalf("expected second movement_type outbound, got %s", got)
	}
	if got := intField(t, ledgerEntries[1], "quantity_delta"); got != -2 {
		t.Fatalf("expected second quantity_delta -2, got %d", got)
	}
}
