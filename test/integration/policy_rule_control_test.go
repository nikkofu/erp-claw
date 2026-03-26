package integration

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestControlPlanePolicyRuleCanGrantTenantRoleAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	tenantID := "tenant-policy-rules"

	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id": tenantID,
		"actor_id":  "actor-viewer",
		"roles":     []string{"viewer"},
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-POLICY-1",
		"name": "Policy Supplier",
	}, http.StatusForbidden, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  "actor-viewer",
	})

	postJSONData(t, h, "/api/platform/v1/control/policy/rules", map[string]any{
		"tenant_id":      tenantID,
		"command_prefix": "masterdata.",
		"any_of_roles":   []string{"viewer"},
	})
	postJSONData(t, h, "/api/platform/v1/control/policy/rules", map[string]any{
		"tenant_id":      tenantID,
		"command_prefix": "masterdata.products.",
		"any_of_roles":   []string{"viewer"},
	})
	postJSONData(t, h, "/api/platform/v1/control/policy/rules", map[string]any{
		"tenant_id":      tenantID,
		"command_prefix": "masterdata.suppliers.",
		"any_of_roles":   []string{"viewer"},
	})

	okResp := doJSONWithHeaders(t, h, http.MethodPost, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-POLICY-2",
		"name": "Policy Supplier 2",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  "actor-viewer",
	})
	if stringField(t, okResp.Data, "code") != "SUP-POLICY-2" {
		t.Fatalf("expected supplier code SUP-POLICY-2, got %s", stringField(t, okResp.Data, "code"))
	}

	rulesResp := getJSONData(t, h, "/api/platform/v1/control/policy/rules?tenant_id="+tenantID)
	items, ok := rulesResp["rules"].([]any)
	if !ok {
		t.Fatalf("expected rules array, got %#v", rulesResp["rules"])
	}
	if len(items) == 0 {
		t.Fatal("expected at least one policy rule")
	}

	filtered := getJSONData(t, h, "/api/platform/v1/control/policy/rules?tenant_id="+tenantID+"&command_prefix=masterdata.&offset=1&limit=1")
	filteredItems, ok := filtered["rules"].([]any)
	if !ok {
		t.Fatalf("expected rules array, got %#v", filtered["rules"])
	}
	if len(filteredItems) != 1 {
		t.Fatalf("expected 1 filtered rule, got %d", len(filteredItems))
	}
	filteredRule, ok := filteredItems[0].(map[string]any)
	if !ok {
		t.Fatalf("expected rule object, got %#v", filteredItems[0])
	}
	if stringField(t, filteredRule, "command_prefix") != "masterdata.products." {
		t.Fatalf("expected paged masterdata.products. rule, got %s", stringField(t, filteredRule, "command_prefix"))
	}

	for _, prefix := range []string{"masterdata.", "masterdata.products.", "masterdata.suppliers."} {
		doJSONWithHeaders(
			t,
			h,
			http.MethodDelete,
			"/api/platform/v1/control/policy/rules?tenant_id="+tenantID+"&command_prefix="+prefix,
			nil,
			http.StatusOK,
			map[string]string{
				"X-Tenant-ID": "tenant-admin",
			},
		)
	}

	doJSONWithHeaders(t, h, http.MethodPost, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-POLICY-3",
		"name": "Policy Supplier 3",
	}, http.StatusForbidden, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  "actor-viewer",
	})
}
