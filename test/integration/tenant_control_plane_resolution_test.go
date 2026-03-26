package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestTenantResolutionUsesControlPlaneCatalogWhenPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	postJSONData(t, h, "/api/platform/v1/control/tenants", map[string]any{
		"code": "tenant-catalog",
		"name": "Catalog Tenant",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/platform/v1/health/livez", nil)
	req.Header.Set("X-Tenant-ID", "tenant-catalog")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected health endpoint status 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Tenant-ID") != "tenant-catalog" {
		t.Fatalf("expected tenant header tenant-catalog, got %s", rec.Header().Get("X-Tenant-ID"))
	}
	if rec.Header().Get("X-Tenant-Isolation") != "logical_cell" {
		t.Fatalf("expected logical_cell isolation header, got %s", rec.Header().Get("X-Tenant-Isolation"))
	}
}
