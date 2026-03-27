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
