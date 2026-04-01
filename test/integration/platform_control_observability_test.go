package integration

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestTimelineEvidenceReadModelFreshness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	tenantID := "tenant-e3-int"
	actorID := "actor-e3-int"

	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id":     tenantID,
		"actor_id":      actorID,
		"roles":         []string{"workspace_operator"},
		"department_id": "ops",
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-e3-int",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-e3-int/tasks", map[string]any{
		"task_id":   "task-e3-int-001",
		"task_type": "tool.call",
		"input":     map[string]any{"tool": "search"},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/task-e3-int-001/start", map[string]any{}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/task-e3-int-001/complete", map[string]any{
		"output": map[string]any{"result": "ok"},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	timelineBySession := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/timeline?session_id=sess-e3-int&limit=10", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})
	if stringField(t, timelineBySession.Data, "as_of") == "" {
		t.Fatal("expected timeline as_of")
	}
	timelineItems, ok := timelineBySession.Data["items"].([]any)
	if !ok || len(timelineItems) == 0 {
		t.Fatalf("expected timeline items, got %#v", timelineBySession.Data["items"])
	}

	timelineByTask := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/timeline?task_id=task-e3-int-001&limit=10", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})
	if stringField(t, timelineByTask.Data, "as_of") == "" {
		t.Fatal("expected task timeline as_of")
	}
	taskTimelineItems, ok := timelineByTask.Data["items"].([]any)
	if !ok || len(taskTimelineItems) == 0 {
		t.Fatalf("expected task timeline items, got %#v", timelineByTask.Data["items"])
	}

	evidencePage1 := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/evidence?task_id=task-e3-int-001&limit=1", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})
	if stringField(t, evidencePage1.Data, "as_of") == "" {
		t.Fatal("expected evidence as_of")
	}
	nextCursor := stringField(t, evidencePage1.Data, "next_cursor")
	if nextCursor == "" {
		t.Fatal("expected evidence next_cursor")
	}

	evidencePage2 := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/evidence?task_id=task-e3-int-001&limit=1&cursor="+nextCursor, nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})
	page2Items, ok := evidencePage2.Data["items"].([]any)
	if !ok || len(page2Items) == 0 {
		t.Fatalf("expected evidence page2 items, got %#v", evidencePage2.Data["items"])
	}
}
