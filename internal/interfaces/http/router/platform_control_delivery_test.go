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

	recoveredPage := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/deliveries?status=recovered&session_id=sess-e2-deliveries&task_id=task-e2-deliveries&limit=20", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data

	items, ok := recoveredPage["items"].([]any)
	if !ok {
		t.Fatalf("expected deliveries items array, got %#v", recoveredPage["items"])
	}
	if len(items) == 0 {
		t.Fatal("expected recovered deliveries")
	}
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("expected delivery item object, got %#v", raw)
		}
		if stringField(t, item, "status") != "recovered" {
			t.Fatalf("expected recovered delivery status, got %s", stringField(t, item, "status"))
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
	if stringField(t, recoveredPage, "as_of") == "" {
		t.Fatal("expected deliveries as_of")
	}
}
