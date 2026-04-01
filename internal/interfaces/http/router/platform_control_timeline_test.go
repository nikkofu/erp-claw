package router

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
)

func TestTimelineQuerySupportsSessionOrTaskAggregation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := New(WithContainer(container))

	tenantID := "tenant-e3-timeline"
	actorID := "actor-e3-timeline"

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id":     tenantID,
		"actor_id":      actorID,
		"roles":         []string{"workspace_operator"},
		"department_id": "ops",
	}, http.StatusOK, nil)

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-e3-001",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-e3-001/tasks", map[string]any{
		"task_id":   "task-e3-001",
		"task_type": "tool.call",
		"input":     map[string]any{"tool": "search"},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/task-e3-001/start", map[string]any{}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/task-e3-001/complete", map[string]any{
		"output": map[string]any{"result": "ok"},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	sessionTimeline := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/timeline?session_id=sess-e3-001&limit=10", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data
	sessionItems, ok := sessionTimeline["items"].([]any)
	if !ok {
		t.Fatalf("expected session timeline items array, got %#v", sessionTimeline["items"])
	}
	if len(sessionItems) == 0 {
		t.Fatal("expected non-empty session timeline")
	}
	firstSessionItem, ok := sessionItems[0].(map[string]any)
	if !ok {
		t.Fatalf("expected session timeline object item, got %#v", sessionItems[0])
	}
	if stringField(t, firstSessionItem, "session_id") != "sess-e3-001" {
		t.Fatalf("expected session_id sess-e3-001, got %s", stringField(t, firstSessionItem, "session_id"))
	}

	taskTimeline := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/timeline?task_id=task-e3-001&limit=10", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data
	taskItems, ok := taskTimeline["items"].([]any)
	if !ok {
		t.Fatalf("expected task timeline items array, got %#v", taskTimeline["items"])
	}
	if len(taskItems) == 0 {
		t.Fatal("expected non-empty task timeline")
	}
	for _, raw := range taskItems {
		item, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("expected task timeline object item, got %#v", raw)
		}
		if stringField(t, item, "task_id") != "task-e3-001" {
			t.Fatalf("expected task_id task-e3-001, got %s", stringField(t, item, "task_id"))
		}
	}
}

func TestEvidenceAPIReturnsAsOfAndCursor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := New(WithContainer(container))

	tenantID := "tenant-e3-evidence"
	actorID := "actor-e3-evidence"

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id":     tenantID,
		"actor_id":      actorID,
		"roles":         []string{"workspace_operator"},
		"department_id": "ops",
	}, http.StatusOK, nil)

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-e3-evidence",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	for _, taskID := range []string{"task-e3-evidence-001", "task-e3-evidence-002", "task-e3-evidence-003"} {
		doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-e3-evidence/tasks", map[string]any{
			"task_id":   taskID,
			"task_type": "tool.call",
			"input":     map[string]any{"tool": "search"},
		}, http.StatusOK, map[string]string{
			"X-Tenant-ID": tenantID,
			"X-Actor-ID":  actorID,
		})
		doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/"+taskID+"/start", map[string]any{}, http.StatusOK, map[string]string{
			"X-Tenant-ID": tenantID,
			"X-Actor-ID":  actorID,
		})
	}

	page1 := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/evidence?task_id=task-e3-evidence-001&limit=1", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data
	items1, ok := page1["items"].([]any)
	if !ok {
		t.Fatalf("expected evidence items array, got %#v", page1["items"])
	}
	if len(items1) != 1 {
		t.Fatalf("expected 1 evidence item on page1, got %d", len(items1))
	}
	if stringField(t, page1, "as_of") == "" {
		t.Fatal("expected evidence as_of")
	}
	nextCursor := stringField(t, page1, "next_cursor")
	if nextCursor == "" {
		t.Fatal("expected evidence next_cursor")
	}

	page2 := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/evidence?task_id=task-e3-evidence-001&limit=1&cursor="+nextCursor, nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data
	items2, ok := page2["items"].([]any)
	if !ok {
		t.Fatalf("expected evidence items array, got %#v", page2["items"])
	}
	if len(items2) == 0 {
		t.Fatal("expected non-empty evidence page2")
	}
}
