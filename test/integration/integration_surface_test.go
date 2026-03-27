package integration

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestIntegrationReadModelAndSalesQueries(t *testing.T) {
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
		"external_ref": "SO-INT-001",
	})

	salesOrderID := stringField(t, postJSONData(t, h, "/api/admin/v1/sales-orders", map[string]any{
		"warehouse_id": warehouseID,
		"external_ref": "SO-INT-002",
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   2,
		}},
	}), "id")
	postJSONData(t, h, "/api/admin/v1/sales-orders/"+salesOrderID+"/ship", map[string]any{})

	overview := doJSON(t, h, http.MethodGet, "/api/integration/v1/read-models/overview", nil, http.StatusOK).Data
	sales := nestedMap(t, overview, "sales")
	if got := intField(t, sales, "total_count"); got != 1 {
		t.Fatalf("expected integration sales total_count 1, got %d", got)
	}
	payable := nestedMap(t, overview, "payable")
	if got := intField(t, payable, "open_count"); got != 1 {
		t.Fatalf("expected integration payable open_count 1, got %d", got)
	}

	salesOrders := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/sales-orders", nil, http.StatusOK).Data
	if len(salesOrders) != 1 {
		t.Fatalf("expected 1 integration sales order, got %d", len(salesOrders))
	}
	if got := stringField(t, salesOrders[0], "id"); got != salesOrderID {
		t.Fatalf("expected integration sales order id %s, got %s", salesOrderID, got)
	}

	detail := doJSON(t, h, http.MethodGet, "/api/integration/v1/sales-orders/"+salesOrderID, nil, http.StatusOK).Data
	if got := stringField(t, detail, "status"); got != "shipped" {
		t.Fatalf("expected integration detail status shipped, got %s", got)
	}
}

func TestIntegrationSalesOrderListSupportsStatusSortAndPagination(t *testing.T) {
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

	orderA := createSalesOrder("SO-INT-LIST-001")
	orderB := createSalesOrder("SO-INT-LIST-002")
	orderC := createSalesOrder("SO-INT-LIST-003")

	postJSONData(t, h, "/api/admin/v1/sales-orders/"+orderB+"/ship", map[string]any{})

	page1 := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/sales-orders?sort=id_asc&page=1&page_size=2", nil, http.StatusOK).Data
	if len(page1) != 2 {
		t.Fatalf("expected 2 sales orders in integration page1, got %d", len(page1))
	}
	if got := stringField(t, page1[0], "id"); got != orderA {
		t.Fatalf("expected integration page1 first id %s, got %s", orderA, got)
	}
	if got := stringField(t, page1[1], "id"); got != orderB {
		t.Fatalf("expected integration page1 second id %s, got %s", orderB, got)
	}

	page2 := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/sales-orders?sort=id_asc&page=2&page_size=2", nil, http.StatusOK).Data
	if len(page2) != 1 {
		t.Fatalf("expected 1 sales order in integration page2, got %d", len(page2))
	}
	if got := stringField(t, page2[0], "id"); got != orderC {
		t.Fatalf("expected integration page2 id %s, got %s", orderC, got)
	}

	shipped := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/sales-orders?status=shipped", nil, http.StatusOK).Data
	if len(shipped) != 1 {
		t.Fatalf("expected 1 shipped integration sales order, got %d", len(shipped))
	}
	if got := stringField(t, shipped[0], "id"); got != orderB {
		t.Fatalf("expected shipped integration sales order id %s, got %s", orderB, got)
	}
}

func TestIntegrationSalesOrderListRejectsInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	cases := []string{
		"/api/integration/v1/sales-orders?status=unknown",
		"/api/integration/v1/sales-orders?sort=unknown",
		"/api/integration/v1/sales-orders?page=0",
		"/api/integration/v1/sales-orders?page_size=0",
	}
	for _, path := range cases {
		env := doJSON(t, h, http.MethodGet, path, nil, http.StatusBadRequest)
		if env.Meta["request_id"] == "" {
			t.Fatalf("expected request_id metadata in bad request response for %s", path)
		}
	}
}
