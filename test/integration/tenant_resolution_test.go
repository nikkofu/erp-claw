package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestTenantResolutionFromHeader(t *testing.T) {
	h := router.New(router.WithContainer(bootstrap.NewTestContainer()))
	req := httptest.NewRequest(http.MethodGet, "/api/platform/v1/health/livez", nil)
	req.Header.Set("X-Tenant-ID", "tenant-a")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Header().Get("X-Tenant-ID") != "tenant-a" {
		t.Fatalf("expected tenant header to round-trip through middleware")
	}
}
