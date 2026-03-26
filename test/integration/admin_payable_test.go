package integration

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestAdminPayableFlow(t *testing.T) {
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

	submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/submit", map[string]any{})
	approvalID := stringField(t, nestedMap(t, submitResp, "approval"), "id")

	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})
	postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/receive", map[string]any{
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	})

	payableResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/payable-bills", map[string]any{})
	payableID := stringField(t, payableResp, "id")
	if got := stringField(t, payableResp, "purchase_order_id"); got != orderID {
		t.Fatalf("expected payable purchase_order_id %s, got %s", orderID, got)
	}
	if got := stringField(t, payableResp, "status"); got != "open" {
		t.Fatalf("expected open payable status, got %s", got)
	}

	payableDetail := getJSONData(t, h, "/api/admin/v1/payables/"+payableID)
	if got := stringField(t, payableDetail, "id"); got != payableID {
		t.Fatalf("expected payable id %s, got %s", payableID, got)
	}
	if got := stringField(t, payableDetail, "purchase_order_id"); got != orderID {
		t.Fatalf("expected payable purchase_order_id %s, got %s", orderID, got)
	}

	planResp := postJSONData(t, h, "/api/admin/v1/payables/"+payableID+"/payment-plans", map[string]any{
		"due_date": "2026-04-01",
	})
	if got := stringField(t, planResp, "payable_bill_id"); got != payableID {
		t.Fatalf("expected payment plan payable_bill_id %s, got %s", payableID, got)
	}
	if got := stringField(t, planResp, "status"); got != "planned" {
		t.Fatalf("expected payment plan status planned, got %s", got)
	}
	if got := stringField(t, planResp, "due_date"); got != "2026-04-01" {
		t.Fatalf("expected payment plan due date 2026-04-01, got %s", got)
	}

	payableDetail = getJSONData(t, h, "/api/admin/v1/payables/"+payableID)
	rawPlans, ok := payableDetail["payment_plans"]
	if !ok {
		t.Fatalf("expected payable detail to include payment_plans, got %#v", payableDetail)
	}
	plans, ok := rawPlans.([]any)
	if !ok {
		t.Fatalf("expected payment_plans to be an array, got %#v", rawPlans)
	}
	if len(plans) != 1 {
		t.Fatalf("expected 1 payment plan, got %d", len(plans))
	}
	planObj, ok := plans[0].(map[string]any)
	if !ok {
		t.Fatalf("expected payment plan item to be an object, got %#v", plans[0])
	}
	if got := stringField(t, planObj, "due_date"); got != "2026-04-01" {
		t.Fatalf("expected payment plan due date 2026-04-01, got %s", got)
	}
}

func TestAdminPayableCreateBeforeReceiveReturnsBadRequest(t *testing.T) {
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

	submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/submit", map[string]any{})
	approvalID := stringField(t, nestedMap(t, submitResp, "approval"), "id")
	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})

	env := doJSON(t, h, http.MethodPost, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/payable-bills", map[string]any{}, http.StatusBadRequest)
	if env.Meta["request_id"] == "" {
		t.Fatal("expected request_id metadata in bad request response")
	}
}

func TestAdminPayableListReturnsTenantScopedBills(t *testing.T) {
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
	submitResp := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/submit", map[string]any{})
	approvalID := stringField(t, nestedMap(t, submitResp, "approval"), "id")
	postJSONData(t, h, "/api/admin/v1/approvals/"+approvalID+"/approve", map[string]any{})
	postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/receive", map[string]any{
		"lines": []map[string]any{{
			"product_id": productID,
			"quantity":   5,
		}},
	})
	createdPayable := postJSONData(t, h, "/api/admin/v1/procurement/purchase-orders/"+orderID+"/payable-bills", map[string]any{})
	createdPayableID := stringField(t, createdPayable, "id")

	payables := getJSONArrayData(t, h, "/api/admin/v1/payables")
	if len(payables) != 1 {
		t.Fatalf("expected 1 payable bill in list, got %d", len(payables))
	}
	if got := stringField(t, payables[0], "id"); got != createdPayableID {
		t.Fatalf("expected payable id %s, got %s", createdPayableID, got)
	}
}
