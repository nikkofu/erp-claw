package integration

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestAdminBackofficeOverviewReadModel(t *testing.T) {
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

	postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+purchaseOrderID+"/payable-bills", map[string]any{})
	postJSONData(t, h, "/api/admin/v1/receivables", map[string]any{
		"external_ref": "SO-RM-001",
	})

	shipOrderID := stringField(t, postJSONData(t, h, "/api/admin/v1/sales-orders", map[string]any{
		"warehouse_id": warehouseID,
		"external_ref": "SO-RM-002",
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   2,
		}},
	}), "id")
	postJSONData(t, h, "/api/admin/v1/sales-orders/"+shipOrderID+"/ship", map[string]any{})
	postJSONData(t, h, "/api/admin/v1/sales-orders", map[string]any{
		"warehouse_id": warehouseID,
		"external_ref": "SO-RM-003",
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   1,
		}},
	})

	overview := getJSONData(t, h, "/api/admin/v1/read-models/overview")
	payable := nestedMap(t, overview, "payable")
	if got := intField(t, payable, "open_count"); got != 1 {
		t.Fatalf("expected payable open_count 1, got %d", got)
	}
	receivable := nestedMap(t, overview, "receivable")
	if got := intField(t, receivable, "open_count"); got != 1 {
		t.Fatalf("expected receivable open_count 1, got %d", got)
	}
	sales := nestedMap(t, overview, "sales")
	if got := intField(t, sales, "draft_count"); got != 1 {
		t.Fatalf("expected sales draft_count 1, got %d", got)
	}
	if got := intField(t, sales, "shipped_count"); got != 1 {
		t.Fatalf("expected sales shipped_count 1, got %d", got)
	}
	if got := intField(t, sales, "total_count"); got != 2 {
		t.Fatalf("expected sales total_count 2, got %d", got)
	}
}
