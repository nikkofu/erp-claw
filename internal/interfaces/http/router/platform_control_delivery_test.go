package router

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
)

func TestDeliveriesQuerySupportsStatusAndFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := New(WithContainer(container))

	tenantID := "tenant-e2-deliveries"
	actorID := "actor-e2-deliveries"

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id":     tenantID,
		"actor_id":      actorID,
		"roles":         []string{"workspace_operator"},
		"department_id": "ops",
	}, http.StatusOK, nil)

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-e2-deliveries",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-e2-deliveries/tasks", map[string]any{
		"task_id":   "task-e2-deliveries",
		"task_type": "tool.call",
		"input":     map[string]any{"tool": "search"},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/task-e2-deliveries/start", map[string]any{}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/task-e2-deliveries/complete", map[string]any{
		"output": map[string]any{"result": "ok"},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	pendingPage := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/deliveries?status=delivered&session_id=sess-e2-deliveries&task_id=task-e2-deliveries&limit=20", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data

	items, ok := pendingPage["items"].([]any)
	if !ok {
		t.Fatalf("expected deliveries items array, got %#v", pendingPage["items"])
	}
	if len(items) == 0 {
		t.Fatal("expected delivered deliveries")
	}
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("expected delivery item object, got %#v", raw)
		}
		if stringField(t, item, "status") != "delivered" {
			t.Fatalf("expected delivered delivery status, got %s", stringField(t, item, "status"))
		}
		if stringField(t, item, "session_id") != "sess-e2-deliveries" {
			t.Fatalf("expected session_id sess-e2-deliveries, got %s", stringField(t, item, "session_id"))
		}
		if stringField(t, item, "task_id") != "task-e2-deliveries" {
			t.Fatalf("expected task_id task-e2-deliveries, got %s", stringField(t, item, "task_id"))
		}
		if stringField(t, item, "updated_at") == "" {
			t.Fatal("expected updated_at in delivery record")
		}
	}
	if stringField(t, pendingPage, "as_of") == "" {
		t.Fatal("expected deliveries as_of")
	}
}

func TestDeliveriesAreActorScopedWithinTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := New(WithContainer(container))

	tenantID := "tenant-e2-actor-scope"
	actorA := "actor-e2-a"
	actorB := "actor-e2-b"

	for _, actorID := range []string{actorA, actorB} {
		doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/control/actors", map[string]any{
			"tenant_id":     tenantID,
			"actor_id":      actorID,
			"roles":         []string{"workspace_operator"},
			"department_id": "ops",
		}, http.StatusOK, nil)
	}

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-e2-actor-a",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-e2-actor-a/tasks", map[string]any{
		"task_id":   "task-e2-actor-a",
		"task_type": "tool.call",
		"input":     map[string]any{"tool": "search"},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/task-e2-actor-a/start", map[string]any{}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	})

	pageForB := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/deliveries?session_id=sess-e2-actor-a&task_id=task-e2-actor-a&limit=20", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorB,
	}).Data
	itemsForB, ok := pageForB["items"].([]any)
	if !ok {
		t.Fatalf("expected deliveries items array, got %#v", pageForB["items"])
	}
	if len(itemsForB) != 0 {
		t.Fatalf("expected actor B to see 0 delivery items for actor A, got %d", len(itemsForB))
	}

	pageForA := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/deliveries?session_id=sess-e2-actor-a&task_id=task-e2-actor-a&limit=20", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	}).Data
	itemsForA, ok := pageForA["items"].([]any)
	if !ok {
		t.Fatalf("expected deliveries items array, got %#v", pageForA["items"])
	}
	if len(itemsForA) == 0 {
		t.Fatal("expected actor A to see own delivery items")
	}
}
