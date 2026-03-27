package integration

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestIntegrationFinanceQueriesReturnPayableAndReceivableReadModels(t *testing.T) {
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

	payable := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+purchaseOrderID+"/payable-bills", map[string]any{})
	payableID := stringField(t, payable, "id")
	postJSONData(t, h, "/api/admin/v1/payables/"+payableID+"/payment-plans", map[string]any{
		"due_date": "2026-04-01",
	})

	receivable := postJSONData(t, h, "/api/admin/v1/receivables", map[string]any{
		"external_ref": "SO-INT-FIN-001",
	})
	receivableID := stringField(t, receivable, "id")

	payables := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/payables", nil, http.StatusOK).Data
	if len(payables) != 1 {
		t.Fatalf("expected 1 integration payable bill, got %d", len(payables))
	}
	if got := stringField(t, payables[0], "id"); got != payableID {
		t.Fatalf("expected integration payable id %s, got %s", payableID, got)
	}

	payableDetail := doJSON(t, h, http.MethodGet, "/api/integration/v1/payables/"+payableID, nil, http.StatusOK).Data
	if got := stringField(t, payableDetail, "id"); got != payableID {
		t.Fatalf("expected integration payable detail id %s, got %s", payableID, got)
	}
	rawPlans, ok := payableDetail["payment_plans"].([]any)
	if !ok {
		t.Fatalf("expected integration payable detail payment_plans array, got %#v", payableDetail["payment_plans"])
	}
	if len(rawPlans) != 1 {
		t.Fatalf("expected 1 integration payable payment plan, got %d", len(rawPlans))
	}

	receivables := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/receivables", nil, http.StatusOK).Data
	if len(receivables) != 1 {
		t.Fatalf("expected 1 integration receivable bill, got %d", len(receivables))
	}
	if got := stringField(t, receivables[0], "id"); got != receivableID {
		t.Fatalf("expected integration receivable id %s, got %s", receivableID, got)
	}

	receivableDetail := doJSON(t, h, http.MethodGet, "/api/integration/v1/receivables/"+receivableID, nil, http.StatusOK).Data
	if got := stringField(t, receivableDetail, "id"); got != receivableID {
		t.Fatalf("expected integration receivable detail id %s, got %s", receivableID, got)
	}
}

func TestIntegrationFinanceListSupportsStatusSortAndPagination(t *testing.T) {
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

	createPayable := func(quantity int) string {
		orderID := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
			"supplier_id":  supplierID,
			"warehouse_id": warehouseID,
			"lines": []map[string]any{{
				"product_id": productID,
				"quantity":   quantity,
			}},
		}), "id")
		submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/submit", map[string]any{})
		approvalID := stringField(t, nestedMap(t, submitResp, "approval"), "id")
		postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})
		postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/receive", map[string]any{
			"lines": []map[string]any{{
				"product_id": productID,
				"quantity":   quantity,
			}},
		})
		resp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/payable-bills", map[string]any{})
		return stringField(t, resp, "id")
	}
	createReceivable := func(externalRef string) string {
		resp := postJSONData(t, h, "/api/admin/v1/receivables", map[string]any{
			"external_ref": externalRef,
		})
		return stringField(t, resp, "id")
	}

	payableA := createPayable(2)
	payableB := createPayable(3)
	payableC := createPayable(4)
	receivableA := createReceivable("SO-INT-FIN-LIST-001")
	receivableB := createReceivable("SO-INT-FIN-LIST-002")
	receivableC := createReceivable("SO-INT-FIN-LIST-003")

	payablePage1 := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/payables?sort=id_asc&page=1&page_size=2", nil, http.StatusOK).Data
	if len(payablePage1) != 2 {
		t.Fatalf("expected 2 integration payable bills in page1, got %d", len(payablePage1))
	}
	if got := stringField(t, payablePage1[0], "id"); got != payableA {
		t.Fatalf("expected integration payable page1 first id %s, got %s", payableA, got)
	}
	if got := stringField(t, payablePage1[1], "id"); got != payableB {
		t.Fatalf("expected integration payable page1 second id %s, got %s", payableB, got)
	}
	payablePage2 := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/payables?sort=id_asc&page=2&page_size=2", nil, http.StatusOK).Data
	if len(payablePage2) != 1 {
		t.Fatalf("expected 1 integration payable bill in page2, got %d", len(payablePage2))
	}
	if got := stringField(t, payablePage2[0], "id"); got != payableC {
		t.Fatalf("expected integration payable page2 id %s, got %s", payableC, got)
	}
	payableOpen := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/payables?status=open", nil, http.StatusOK).Data
	if len(payableOpen) != 3 {
		t.Fatalf("expected 3 open integration payable bills, got %d", len(payableOpen))
	}

	receivablePage1 := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/receivables?sort=id_asc&page=1&page_size=2", nil, http.StatusOK).Data
	if len(receivablePage1) != 2 {
		t.Fatalf("expected 2 integration receivable bills in page1, got %d", len(receivablePage1))
	}
	if got := stringField(t, receivablePage1[0], "id"); got != receivableA {
		t.Fatalf("expected integration receivable page1 first id %s, got %s", receivableA, got)
	}
	if got := stringField(t, receivablePage1[1], "id"); got != receivableB {
		t.Fatalf("expected integration receivable page1 second id %s, got %s", receivableB, got)
	}
	receivablePage2 := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/receivables?sort=id_asc&page=2&page_size=2", nil, http.StatusOK).Data
	if len(receivablePage2) != 1 {
		t.Fatalf("expected 1 integration receivable bill in page2, got %d", len(receivablePage2))
	}
	if got := stringField(t, receivablePage2[0], "id"); got != receivableC {
		t.Fatalf("expected integration receivable page2 id %s, got %s", receivableC, got)
	}
	receivableOpen := doJSONForArray(t, h, http.MethodGet, "/api/integration/v1/receivables?status=open", nil, http.StatusOK).Data
	if len(receivableOpen) != 3 {
		t.Fatalf("expected 3 open integration receivable bills, got %d", len(receivableOpen))
	}
}

func TestIntegrationFinanceListRejectsInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	cases := []string{
		"/api/integration/v1/payables?status=closed",
		"/api/integration/v1/payables?sort=unknown",
		"/api/integration/v1/payables?page=0",
		"/api/integration/v1/payables?page_size=0",
		"/api/integration/v1/receivables?status=closed",
		"/api/integration/v1/receivables?sort=unknown",
		"/api/integration/v1/receivables?page=0",
		"/api/integration/v1/receivables?page_size=0",
	}
	for _, path := range cases {
		env := doJSON(t, h, http.MethodGet, path, nil, http.StatusBadRequest)
		if env.Meta["request_id"] == "" {
			t.Fatalf("expected request_id metadata in bad request response for %s", path)
		}
	}
}
