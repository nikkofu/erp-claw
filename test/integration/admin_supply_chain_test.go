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

type envelope struct {
	Data  map[string]any `json:"data"`
	Error map[string]any `json:"error"`
	Meta  map[string]any `json:"meta"`
}

func postJSONData(t *testing.T, h http.Handler, path string, body any) map[string]any {
	t.Helper()
	return doJSON(t, h, http.MethodPost, path, body, http.StatusOK).Data
}

func getJSONData(t *testing.T, h http.Handler, path string) map[string]any {
	t.Helper()
	return doJSON(t, h, http.MethodGet, path, nil, http.StatusOK).Data
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
