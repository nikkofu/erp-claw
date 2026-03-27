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

func TestWorkspaceSalesOrderListSupportsStatusSortAndPagination(t *testing.T) {
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
			"quantity":   8,
		}},
	}), "id")
	submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+purchaseOrderID+"/submit", map[string]any{})
	approvalID := stringField(t, nestedMap(t, submitResp, "approval"), "id")
	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})
	postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+purchaseOrderID+"/receive", map[string]any{
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   8,
		}},
	})

	createSalesOrder := func(externalRef string) string {
		return stringField(t, postJSONData(t, h, "/api/admin/v1/sales-orders", map[string]any{
			"warehouse_id": warehouseID,
			"external_ref": externalRef,
			"lines": []map[string]any{{
				"product_id": productID,
				"quantity":   2,
			}},
		}), "id")
	}

	orderA := createSalesOrder("SO-WS-LIST-001")
	orderB := createSalesOrder("SO-WS-LIST-002")
	orderC := createSalesOrder("SO-WS-LIST-003")

	postJSONData(t, h, "/api/admin/v1/sales-orders/"+orderB+"/ship", map[string]any{})

	page1 := doJSONForArray(t, h, http.MethodGet, "/api/workspace/v1/sales-orders?sort=id_asc&page=1&page_size=2", nil, http.StatusOK).Data
	if len(page1) != 2 {
		t.Fatalf("expected 2 sales orders in workspace page1, got %d", len(page1))
	}
	if got := stringField(t, page1[0], "id"); got != orderA {
		t.Fatalf("expected workspace page1 first id %s, got %s", orderA, got)
	}
	if got := stringField(t, page1[1], "id"); got != orderB {
		t.Fatalf("expected workspace page1 second id %s, got %s", orderB, got)
	}

	page2 := doJSONForArray(t, h, http.MethodGet, "/api/workspace/v1/sales-orders?sort=id_asc&page=2&page_size=2", nil, http.StatusOK).Data
	if len(page2) != 1 {
		t.Fatalf("expected 1 sales order in workspace page2, got %d", len(page2))
	}
	if got := stringField(t, page2[0], "id"); got != orderC {
		t.Fatalf("expected workspace page2 id %s, got %s", orderC, got)
	}

	shipped := doJSONForArray(t, h, http.MethodGet, "/api/workspace/v1/sales-orders?status=shipped", nil, http.StatusOK).Data
	if len(shipped) != 1 {
		t.Fatalf("expected 1 shipped workspace sales order, got %d", len(shipped))
	}
	if got := stringField(t, shipped[0], "id"); got != orderB {
		t.Fatalf("expected shipped workspace sales order id %s, got %s", orderB, got)
	}
}

func TestWorkspaceSalesOrderListRejectsInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	cases := []string{
		"/api/workspace/v1/sales-orders?status=unknown",
		"/api/workspace/v1/sales-orders?sort=unknown",
		"/api/workspace/v1/sales-orders?page=0",
		"/api/workspace/v1/sales-orders?page_size=0",
	}
	for _, path := range cases {
		env := doJSON(t, h, http.MethodGet, path, nil, http.StatusBadRequest)
		if env.Meta["request_id"] == "" {
			t.Fatalf("expected request_id metadata in bad request response for %s", path)
		}
	}
}
