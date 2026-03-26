package integration

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestControlPlaneCanListTenantsAndActors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	postJSONData(t, h, "/api/platform/v1/control/tenants", map[string]any{
		"code": "tenant-directory-a",
		"name": "Tenant Directory A",
	})
	postJSONData(t, h, "/api/platform/v1/control/tenants", map[string]any{
		"code": "tenant-directory-b",
		"name": "Tenant Directory B",
	})

	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id": "tenant-directory-a",
		"actor_id":  "actor-alpha",
		"roles":     []string{"viewer"},
	})
	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id": "tenant-directory-a",
		"actor_id":  "actor-beta",
		"roles":     []string{"supplychain_operator"},
	})

	tenantsResp := getJSONData(t, h, "/api/platform/v1/control/tenants")
	tenants, ok := tenantsResp["tenants"].([]any)
	if !ok {
		t.Fatalf("expected tenants array, got %#v", tenantsResp["tenants"])
	}
	if len(tenants) < 2 {
		t.Fatalf("expected at least 2 tenants, got %d", len(tenants))
	}

	tenantDetail := getJSONData(t, h, "/api/platform/v1/control/tenants/tenant-directory-a")
	if stringField(t, tenantDetail, "code") != "tenant-directory-a" {
		t.Fatalf("expected tenant code tenant-directory-a, got %s", stringField(t, tenantDetail, "code"))
	}

	actorsResp := getJSONData(t, h, "/api/platform/v1/control/actors?tenant_id=tenant-directory-a")
	actors, ok := actorsResp["actors"].([]any)
	if !ok {
		t.Fatalf("expected actors array, got %#v", actorsResp["actors"])
	}
	if len(actors) != 2 {
		t.Fatalf("expected 2 actors, got %d", len(actors))
	}

	actorDetail := getJSONData(t, h, "/api/platform/v1/control/actors/actor-alpha?tenant_id=tenant-directory-a")
	if stringField(t, actorDetail, "id") != "actor-alpha" {
		t.Fatalf("expected actor id actor-alpha, got %s", stringField(t, actorDetail, "id"))
	}
}
