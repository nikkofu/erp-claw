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

	doJSONWithHeaders(
		t,
		h,
		"DELETE",
		"/api/platform/v1/control/actors/actor-beta?tenant_id=tenant-directory-a",
		nil,
		200,
		nil,
	)

	actorsRespAfterDelete := getJSONData(t, h, "/api/platform/v1/control/actors?tenant_id=tenant-directory-a")
	actorsAfterDelete, ok := actorsRespAfterDelete["actors"].([]any)
	if !ok {
		t.Fatalf("expected actors array, got %#v", actorsRespAfterDelete["actors"])
	}
	if len(actorsAfterDelete) != 1 {
		t.Fatalf("expected 1 actor after delete, got %d", len(actorsAfterDelete))
	}

	doJSONWithHeaders(
		t,
		h,
		"GET",
		"/api/platform/v1/control/actors/actor-beta?tenant_id=tenant-directory-a",
		nil,
		404,
		nil,
	)

	doJSONWithHeaders(
		t,
		h,
		"DELETE",
		"/api/platform/v1/control/tenants/tenant-directory-b",
		nil,
		200,
		nil,
	)

	doJSONWithHeaders(
		t,
		h,
		"GET",
		"/api/platform/v1/control/tenants/tenant-directory-b",
		nil,
		404,
		nil,
	)

	tenantsAfterDelete := getJSONData(t, h, "/api/platform/v1/control/tenants")
	tenantsItemsAfterDelete, ok := tenantsAfterDelete["tenants"].([]any)
	if !ok {
		t.Fatalf("expected tenants array, got %#v", tenantsAfterDelete["tenants"])
	}
	if len(tenantsItemsAfterDelete) != 1 {
		t.Fatalf("expected 1 tenant after delete, got %d", len(tenantsItemsAfterDelete))
	}
}

func TestControlPlaneCanFilterAndPaginateActors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	for _, payload := range []map[string]any{
		{
			"tenant_id":     "tenant-directory-filter",
			"actor_id":      "actor-alpha",
			"roles":         []string{"viewer"},
			"department_id": "ops",
		},
		{
			"tenant_id":     "tenant-directory-filter",
			"actor_id":      "actor-beta",
			"roles":         []string{"supplychain_operator"},
			"department_id": "ops",
		},
		{
			"tenant_id":     "tenant-directory-filter",
			"actor_id":      "actor-gamma",
			"roles":         []string{"viewer"},
			"department_id": "finance",
		},
	} {
		postJSONData(t, h, "/api/platform/v1/control/actors", payload)
	}

	viewers := getJSONData(t, h, "/api/platform/v1/control/actors?tenant_id=tenant-directory-filter&role=viewer")
	viewerItems, ok := viewers["actors"].([]any)
	if !ok {
		t.Fatalf("expected actors array, got %#v", viewers["actors"])
	}
	if len(viewerItems) != 2 {
		t.Fatalf("expected 2 viewer actors, got %d", len(viewerItems))
	}

	opsDept := getJSONData(t, h, "/api/platform/v1/control/actors?tenant_id=tenant-directory-filter&department_id=ops")
	opsItems, ok := opsDept["actors"].([]any)
	if !ok {
		t.Fatalf("expected actors array, got %#v", opsDept["actors"])
	}
	if len(opsItems) != 2 {
		t.Fatalf("expected 2 ops actors, got %d", len(opsItems))
	}

	paged := getJSONData(t, h, "/api/platform/v1/control/actors?tenant_id=tenant-directory-filter&offset=1&limit=1")
	pagedItems, ok := paged["actors"].([]any)
	if !ok {
		t.Fatalf("expected actors array, got %#v", paged["actors"])
	}
	if len(pagedItems) != 1 {
		t.Fatalf("expected 1 paged actor, got %d", len(pagedItems))
	}
	pagedActor, ok := pagedItems[0].(map[string]any)
	if !ok {
		t.Fatalf("expected actor object, got %#v", pagedItems[0])
	}
	if got := stringField(t, pagedActor, "id"); got != "actor-beta" {
		t.Fatalf("expected actor-beta in paged result, got %s", got)
	}

	filtered := getJSONData(t, h, "/api/platform/v1/control/actors?tenant_id=tenant-directory-filter&role=viewer&department_id=ops")
	filteredItems, ok := filtered["actors"].([]any)
	if !ok {
		t.Fatalf("expected actors array, got %#v", filtered["actors"])
	}
	if len(filteredItems) != 1 {
		t.Fatalf("expected 1 filtered actor, got %d", len(filteredItems))
	}
	filteredActor, ok := filteredItems[0].(map[string]any)
	if !ok {
		t.Fatalf("expected actor object, got %#v", filteredItems[0])
	}
	if got := stringField(t, filteredActor, "id"); got != "actor-alpha" {
		t.Fatalf("expected actor-alpha after role+department filter, got %s", got)
	}
}
