package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestAdminSupplyChainFlow(t *testing.T) {
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

	orderResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  supplierID,
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	})
	orderID := stringField(t, orderResp, "id")
	if got := stringField(t, orderResp, "status"); got != "draft" {
		t.Fatalf("expected draft order, got %s", got)
	}

	submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/submit", map[string]any{})
	submittedOrder := nestedMap(t, submitResp, "order")
	if got := stringField(t, submittedOrder, "status"); got != "pending_approval" {
		t.Fatalf("expected pending approval order, got %s", got)
	}
	approvalResp := nestedMap(t, submitResp, "approval")
	approvalID := stringField(t, approvalResp, "id")
	if got := stringField(t, approvalResp, "status"); got != "pending" {
		t.Fatalf("expected pending approval request, got %s", got)
	}

	detailResp := getJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID)
	detailOrder := nestedMap(t, detailResp, "order")
	if got := stringField(t, detailOrder, "id"); got != orderID {
		t.Fatalf("expected detail order %s, got %s", orderID, got)
	}
	detailApproval := nestedMap(t, detailResp, "approval")
	if got := stringField(t, detailApproval, "id"); got != approvalID {
		t.Fatalf("expected detail approval %s, got %s", approvalID, got)
	}

	approveResp := postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})
	approvedOrder := nestedMap(t, approveResp, "order")
	if got := stringField(t, approvedOrder, "status"); got != "approved" {
		t.Fatalf("expected approved order, got %s", got)
	}
	approvedRequest := nestedMap(t, approveResp, "approval")
	if got := stringField(t, approvedRequest, "status"); got != "approved" {
		t.Fatalf("expected approved request, got %s", got)
	}
}

func TestAdminSupplyChainCreateOrderUnknownSupplierReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	productID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/products", map[string]any{
		"sku":  "SKU-001",
		"name": "Copper Wire",
		"unit": "roll",
	}), "id")

	warehouseID := stringField(t, postJSONData(t, h, "/api/admin/v1/master-data/warehouses", map[string]any{
		"code": "WH-SH",
		"name": "Shanghai Warehouse",
	}), "id")

	env := doJSON(t, h, http.MethodPost, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  "sup-missing",
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	}, http.StatusNotFound)

	if env.Meta["request_id"] == "" {
		t.Fatal("expected request_id metadata in not found response")
	}
}

func TestAdminSupplyChainRejectFlow(t *testing.T) {
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

	orderResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  supplierID,
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	})

	submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+stringField(t, orderResp, "id")+"/submit", map[string]any{})
	approvalID := stringField(t, nestedMap(t, submitResp, "approval"), "id")

	rejectResp := postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/reject", map[string]any{})
	rejectedOrder := nestedMap(t, rejectResp, "order")
	if got := stringField(t, rejectedOrder, "status"); got != "rejected" {
		t.Fatalf("expected rejected order, got %s", got)
	}
	rejectedRequest := nestedMap(t, rejectResp, "approval")
	if got := stringField(t, rejectedRequest, "status"); got != "rejected" {
		t.Fatalf("expected rejected request, got %s", got)
	}
}

func TestAdminApprovalListSupportsStatusFilter(t *testing.T) {
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

	orderA := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  supplierID,
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	}), "id")
	submitA := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderA+"/submit", map[string]any{})
	approvalA := stringField(t, nestedMap(t, submitA, "approval"), "id")

	orderB := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
		"supplier_id":  supplierID,
		"warehouse_id": warehouseID,
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   6,
		}},
	}), "id")
	submitB := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderB+"/submit", map[string]any{})
	approvalB := stringField(t, nestedMap(t, submitB, "approval"), "id")

	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalB+"/approve", map[string]any{})

	all := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/approvals", nil, http.StatusOK).Data
	if len(all) != 2 {
		t.Fatalf("expected 2 approvals in list, got %d", len(all))
	}

	pending := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/approvals?status=pending", nil, http.StatusOK).Data
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending approval, got %d", len(pending))
	}
	if got := stringField(t, pending[0], "id"); got != approvalA {
		t.Fatalf("expected pending approval id %s, got %s", approvalA, got)
	}

	approved := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/approvals?status=approved", nil, http.StatusOK).Data
	if len(approved) != 1 {
		t.Fatalf("expected 1 approved approval, got %d", len(approved))
	}
	if got := stringField(t, approved[0], "id"); got != approvalB {
		t.Fatalf("expected approved approval id %s, got %s", approvalB, got)
	}
}

func TestAdminApprovalListRejectsInvalidStatusFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	env := doJSON(t, h, http.MethodGet, "/api/admin/v1/approvals?status=invalid", nil, http.StatusBadRequest)
	if env.Meta["request_id"] == "" {
		t.Fatal("expected request_id metadata in bad request response")
	}
}

func TestAdminApprovalListSupportsSortAndPagination(t *testing.T) {
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

	createAndSubmit := func(quantity int) string {
		orderID := stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
			"supplier_id":  supplierID,
			"warehouse_id": warehouseID,
			"lines": []map[string]any{{
				"product_id": productID,
				"quantity":   quantity,
			}},
		}), "id")
		resp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/submit", map[string]any{})
		return stringField(t, nestedMap(t, resp, "approval"), "id")
	}

	approvalA := createAndSubmit(1)
	approvalB := createAndSubmit(2)
	approvalC := createAndSubmit(3)

	page1 := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/approvals?sort=id_asc&page=1&page_size=2", nil, http.StatusOK).Data
	if len(page1) != 2 {
		t.Fatalf("expected 2 approvals in page1, got %d", len(page1))
	}
	if got := stringField(t, page1[0], "id"); got != approvalA {
		t.Fatalf("expected page1 first id %s, got %s", approvalA, got)
	}
	if got := stringField(t, page1[1], "id"); got != approvalB {
		t.Fatalf("expected page1 second id %s, got %s", approvalB, got)
	}

	page2 := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/approvals?sort=id_asc&page=2&page_size=2", nil, http.StatusOK).Data
	if len(page2) != 1 {
		t.Fatalf("expected 1 approval in page2, got %d", len(page2))
	}
	if got := stringField(t, page2[0], "id"); got != approvalC {
		t.Fatalf("expected page2 id %s, got %s", approvalC, got)
	}
}

func TestAdminApprovalListRejectsInvalidSortAndPaginationQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	cases := []string{
		"/api/admin/v1/approvals?sort=unknown",
		"/api/admin/v1/approvals?page=0",
		"/api/admin/v1/approvals?page_size=0",
	}

	for _, path := range cases {
		env := doJSON(t, h, http.MethodGet, path, nil, http.StatusBadRequest)
		if env.Meta["request_id"] == "" {
			t.Fatalf("expected request_id metadata in bad request response for %s", path)
		}
	}
}

func TestAdminPurchaseOrderListSupportsStatusSortAndPagination(t *testing.T) {
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

	createOrder := func(quantity int) string {
		return stringField(t, postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders", map[string]any{
			"supplier_id":  supplierID,
			"warehouse_id": warehouseID,
			"lines": []map[string]any{{
				"product_id": productID,
				"quantity":   quantity,
			}},
		}), "id")
	}

	orderA := createOrder(1)
	orderB := createOrder(2)
	orderC := createOrder(3)

	submitB := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderB+"/submit", map[string]any{})
	approvalB := stringField(t, nestedMap(t, submitB, "approval"), "id")
	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalB+"/reject", map[string]any{})

	submitC := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderC+"/submit", map[string]any{})
	approvalC := stringField(t, nestedMap(t, submitC, "approval"), "id")
	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalC+"/approve", map[string]any{})

	page1 := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/procurement/purchase-orders?sort=id_asc&page=1&page_size=2", nil, http.StatusOK).Data
	if len(page1) != 2 {
		t.Fatalf("expected 2 orders in page1, got %d", len(page1))
	}
	if got := stringField(t, page1[0], "id"); got != orderA {
		t.Fatalf("expected page1 first order %s, got %s", orderA, got)
	}
	if got := stringField(t, page1[1], "id"); got != orderB {
		t.Fatalf("expected page1 second order %s, got %s", orderB, got)
	}

	page2 := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/procurement/purchase-orders?sort=id_asc&page=2&page_size=2", nil, http.StatusOK).Data
	if len(page2) != 1 {
		t.Fatalf("expected 1 order in page2, got %d", len(page2))
	}
	if got := stringField(t, page2[0], "id"); got != orderC {
		t.Fatalf("expected page2 order %s, got %s", orderC, got)
	}

	rejected := doJSONForArray(t, h, http.MethodGet, "/api/admin/v1/procurement/purchase-orders?status=rejected", nil, http.StatusOK).Data
	if len(rejected) != 1 {
		t.Fatalf("expected 1 rejected order, got %d", len(rejected))
	}
	if got := stringField(t, rejected[0], "id"); got != orderB {
		t.Fatalf("expected rejected order %s, got %s", orderB, got)
	}
}

func TestAdminPurchaseOrderListRejectsInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	cases := []string{
		"/api/admin/v1/procurement/purchase-orders?status=unknown",
		"/api/admin/v1/procurement/purchase-orders?sort=unknown",
		"/api/admin/v1/procurement/purchase-orders?page=0",
		"/api/admin/v1/procurement/purchase-orders?page_size=0",
	}
	for _, path := range cases {
		env := doJSON(t, h, http.MethodGet, path, nil, http.StatusBadRequest)
		if env.Meta["request_id"] == "" {
			t.Fatalf("expected request_id metadata in bad request response for %s", path)
		}
	}
}

type envelope struct {
	Data  map[string]any `json:"data"`
	Error map[string]any `json:"error"`
	Meta  map[string]any `json:"meta"`
}

type arrayEnvelope struct {
	Data  []map[string]any `json:"data"`
	Error map[string]any   `json:"error"`
	Meta  map[string]any   `json:"meta"`
}

func postJSONData(t *testing.T, h http.Handler, path string, body any) map[string]any {
	t.Helper()
	return doJSON(t, h, http.MethodPost, path, body, http.StatusOK).Data
}

func getJSONData(t *testing.T, h http.Handler, path string) map[string]any {
	t.Helper()
	return doJSON(t, h, http.MethodGet, path, nil, http.StatusOK).Data
}

func getJSONArrayData(t *testing.T, h http.Handler, path string) []map[string]any {
	t.Helper()
	return doJSONForArray(t, h, http.MethodGet, path, nil, http.StatusOK).Data
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any, expectedStatus int) envelope {
	t.Helper()

	var reqBody *bytes.Reader
	if body == nil {
		reqBody = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-admin")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d with body %s", expectedStatus, rec.Code, rec.Body.String())
	}

	var env envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if env.Meta["request_id"] == "" {
		t.Fatal("expected meta.request_id")
	}
	return env
}

func doJSONForArray(t *testing.T, h http.Handler, method, path string, body any, expectedStatus int) arrayEnvelope {
	t.Helper()

	var reqBody *bytes.Reader
	if body == nil {
		reqBody = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-admin")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d with body %s", expectedStatus, rec.Code, rec.Body.String())
	}

	var env arrayEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if env.Meta["request_id"] == "" {
		t.Fatal("expected meta.request_id")
	}
	return env
}

func nestedMap(t *testing.T, payload map[string]any, field string) map[string]any {
	t.Helper()
	value, ok := payload[field]
	if !ok {
		t.Fatalf("missing field %q in payload %#v", field, payload)
	}
	nested, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected field %q to be an object, got %#v", field, value)
	}
	return nested
}

func stringField(t *testing.T, payload map[string]any, field string) string {
	t.Helper()
	value, ok := payload[field]
	if !ok {
		t.Fatalf("missing field %q in payload %#v", field, payload)
	}
	str, ok := value.(string)
	if !ok {
		t.Fatalf("expected field %q to be a string, got %#v", field, value)
	}
	return str
}

func intField(t *testing.T, payload map[string]any, field string) int {
	t.Helper()
	value, ok := payload[field]
	if !ok {
		t.Fatalf("missing field %q in payload %#v", field, payload)
	}
	number, ok := value.(float64)
	if !ok {
		t.Fatalf("expected field %q to be a number, got %#v", field, value)
	}
	return int(number)
}
