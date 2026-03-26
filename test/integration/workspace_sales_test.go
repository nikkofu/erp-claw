package integration

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestWorkspaceSalesOrderQueriesReturnListAndDetail(t *testing.T) {
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

	purchaseOrderID := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  supplierID,
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	}), "id")
	submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+purchaseOrderID+"/submit", map[string]any{})
	approvalID := stringField(t, nestedMap(t, submitResp, "approval"), "id")
	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})
	postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+purchaseOrderID+"/receive", map[string]any{
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	})

	salesOrderID := stringField(t, postJSONData(t, h, "/api/admin/v1/sales-orders", map[string]any{
		"warehouse_id": warehouseID,
		"external_ref": "SO-WS-001",
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   2,
		}},
	}), "id")
	postJSONData(t, h, "/api/admin/v1/sales-orders/"+salesOrderID+"/ship", map[string]any{})

	list := doJSONForArray(t, h, http.MethodGet, "/api/workspace/v1/sales-orders", nil, http.StatusOK).Data
	if len(list) != 1 {
		t.Fatalf("expected 1 sales order in workspace list, got %d", len(list))
	}
	if got := stringField(t, list[0], "id"); got != salesOrderID {
		t.Fatalf("expected workspace list sales order id %s, got %s", salesOrderID, got)
	}
	if got := stringField(t, list[0], "status"); got != "shipped" {
		t.Fatalf("expected workspace list sales order status shipped, got %s", got)
	}

	detail := doJSON(t, h, http.MethodGet, "/api/workspace/v1/sales-orders/"+salesOrderID, nil, http.StatusOK).Data
	if got := stringField(t, detail, "id"); got != salesOrderID {
		t.Fatalf("expected workspace detail sales order id %s, got %s", salesOrderID, got)
	}
	if got := stringField(t, detail, "status"); got != "shipped" {
		t.Fatalf("expected workspace detail sales order status shipped, got %s", got)
	}
}
