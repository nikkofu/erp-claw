package integration

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestPlatformAuditRecordsSupportFilterAndOffset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	tenantID := "tenant-audit-filter"
	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id": tenantID,
		"actor_id":  "actor-viewer",
		"roles":     []string{"viewer"},
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-AUD-1",
		"name": "Audit Supplier 1",
	}, http.StatusForbidden, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  "actor-viewer",
	})

	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id": tenantID,
		"actor_id":  "actor-viewer",
		"roles":     []string{"supplychain_operator"},
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-AUD-2",
		"name": "Audit Supplier 2",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  "actor-viewer",
	})

	filtered := doJSONWithHeaders(
		t,
		h,
		http.MethodGet,
		"/api/platform/v1/audit/records?actor_id=actor-viewer&decision=DENY&outcome=rejected&limit=10",
		nil,
		http.StatusOK,
		map[string]string{
			"X-Tenant-ID": tenantID,
		},
	)
	items, ok := filtered.Data["records"].([]any)
	if !ok {
		t.Fatalf("expected records array, got %#v", filtered.Data["records"])
	}
	if len(items) == 0 {
		t.Fatal("expected at least one filtered record")
	}

	for _, raw := range items {
		record, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("expected audit record object, got %#v", raw)
		}
		if record["actor_id"] != "actor-viewer" {
			t.Fatalf("expected actor filter actor-viewer, got %#v", record["actor_id"])
		}
		if record["decision"] != "DENY" {
			t.Fatalf("expected decision DENY, got %#v", record["decision"])
		}
		if record["outcome"] != "rejected" {
			t.Fatalf("expected outcome rejected, got %#v", record["outcome"])
		}
	}

	prefixFiltered := doJSONWithHeaders(
		t,
		h,
		http.MethodGet,
		"/api/platform/v1/audit/records?command_prefix=masterdata.suppliers.&limit=20",
		nil,
		http.StatusOK,
		map[string]string{
			"X-Tenant-ID": tenantID,
		},
	)
	prefixItems, ok := prefixFiltered.Data["records"].([]any)
	if !ok {
		t.Fatalf("expected records array, got %#v", prefixFiltered.Data["records"])
	}
	if len(prefixItems) == 0 {
		t.Fatal("expected at least one prefix-filtered record")
	}
	for _, raw := range prefixItems {
		record, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("expected audit record object, got %#v", raw)
		}
		commandName, ok := record["command_name"].(string)
		if !ok {
			t.Fatalf("expected command_name string, got %#v", record["command_name"])
		}
		if !strings.HasPrefix(commandName, "masterdata.suppliers.") {
			t.Fatalf("expected masterdata.suppliers.* command, got %s", commandName)
		}
	}
}
