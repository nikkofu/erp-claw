package integration

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestAdminSalesOrderShipFlow(t *testing.T) {
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

	salesOrderResp := postJSONData(t, h, "/api/admin/v1/sales-orders", map[string]any{
		"warehouse_id": warehouseID,
		"external_ref": "SO-001",
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   2,
		}},
	})
	salesOrderID := stringField(t, salesOrderResp, "id")
	if got := stringField(t, salesOrderResp, "status"); got != "draft" {
		t.Fatalf("expected draft sales order, got %s", got)
	}

	shipResp := postJSONData(t, h, "/api/admin/v1/sales-orders/"+salesOrderID+"/ship", map[string]any{})
	order := nestedMap(t, shipResp, "order")
	if got := stringField(t, order, "status"); got != "shipped" {
		t.Fatalf("expected shipped sales order, got %s", got)
	}
	entriesRaw, ok := shipResp["ledger_entries"].([]any)
	if !ok {
		t.Fatalf("expected ledger_entries array, got %#v", shipResp["ledger_entries"])
	}
	if len(entriesRaw) != 1 {
		t.Fatalf("expected 1 shipment ledger entry, got %d", len(entriesRaw))
	}
	entry, ok := entriesRaw[0].(map[string]any)
	if !ok {
		t.Fatalf("expected shipment ledger entry object, got %#v", entriesRaw[0])
	}
	if got := stringField(t, entry, "movement_type"); got != "outbound" {
		t.Fatalf("expected outbound shipment movement, got %s", got)
	}
	if got := intField(t, entry, "quantity_delta"); got != -2 {
		t.Fatalf("expected quantity_delta -2, got %d", got)
	}

	listResp := getJSONArrayData(t, h, "/api/admin/v1/sales-orders")
	if len(listResp) != 1 {
		t.Fatalf("expected 1 sales order in list, got %d", len(listResp))
	}
	if got := stringField(t, listResp[0], "id"); got != salesOrderID {
		t.Fatalf("expected listed sales order id %s, got %s", salesOrderID, got)
	}

	detailResp := getJSONData(t, h, "/api/admin/v1/sales-orders/"+salesOrderID)
	if got := stringField(t, detailResp, "status"); got != "shipped" {
		t.Fatalf("expected detail status shipped, got %s", got)
	}

	balanceResp := getJSONData(t, h, "/api/admin/v1/inventory/balances?product_id="+productID+"&warehouse_id="+warehouseID)
	if got := intField(t, balanceResp, "on_hand"); got != 3 {
		t.Fatalf("expected on_hand 3 after shipment, got %d", got)
	}
}

func TestAdminSalesOrderListSupportsStatusSortAndPagination(t *testing.T) {
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

	createSalesOrder := func(externalRef string) string {
		return stringField(t, postJSONData(t, h, "/api/admin/v1/sales-orders", map[string]any{
			"warehouse_id": warehouseID,
			"external_ref": externalRef,
			"lines": []map[string]any{{
				"product_id": productID,
				"quantity":   1,
			}},
		}), "id")
	}

	orderA := createSalesOrder("SO-LIST-001")
	orderB := createSalesOrder("SO-LIST-002")
	orderC := createSalesOrder("SO-LIST-003")

	postJSONData(t, h, "/api/admin/v1/sales-orders/"+orderB+"/ship", map[string]any{})

	page1 := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/sales-orders?sort=id_asc&page=1&page_size=2", nil, http.StatusOK).Data
	if len(page1) != 2 {
		t.Fatalf("expected 2 sales orders in page1, got %d", len(page1))
	}
	if got := stringField(t, page1[0], "id"); got != orderA {
		t.Fatalf("expected page1 first id %s, got %s", orderA, got)
	}
	if got := stringField(t, page1[1], "id"); got != orderB {
		t.Fatalf("expected page1 second id %s, got %s", orderB, got)
	}

	page2 := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/sales-orders?sort=id_asc&page=2&page_size=2", nil, http.StatusOK).Data
	if len(page2) != 1 {
		t.Fatalf("expected 1 sales order in page2, got %d", len(page2))
	}
	if got := stringField(t, page2[0], "id"); got != orderC {
		t.Fatalf("expected page2 id %s, got %s", orderC, got)
	}

	shipped := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/sales-orders?status=shipped", nil, http.StatusOK).Data
	if len(shipped) != 1 {
		t.Fatalf("expected 1 shipped sales order, got %d", len(shipped))
	}
	if got := stringField(t, shipped[0], "id"); got != orderB {
		t.Fatalf("expected shipped sales order id %s, got %s", orderB, got)
	}
}

func TestAdminSalesOrderListRejectsInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	cases := []string{
		"/api/admin/v1/sales-orders?status=unknown",
		"/api/admin/v1/sales-orders?sort=unknown",
		"/api/admin/v1/sales-orders?page=0",
		"/api/admin/v1/sales-orders?page_size=0",
	}

	for _, path := range cases {
		env := doJSON(t, h, http.MethodGet, path, nil, http.StatusBadRequest)
		if env.Meta["request_id"] == "" {
			t.Fatalf("expected request_id metadata in bad request response for %s", path)
		}
	}
}

func TestAdminSalesOrderShipRejectsInsufficientInventory(t *testing.T) {
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
		"external_ref": "SO-002",
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   6,
		}},
	}), "id")

	env := doJSON(t, h, http.MethodPost, "/api/admin/v1/sales-orders/"+salesOrderID+"/ship", map[string]any{}, http.StatusBadRequest)
	if env.Meta["request_id"] == "" {
		t.Fatal("expected request_id metadata")
	}
}
