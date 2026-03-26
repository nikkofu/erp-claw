package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
)

func TestPlatformControlPlaneActorPolicyAndAgentRuntimeFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	tenantID := "tenant-policy"

	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id":     tenantID,
		"actor_id":      "actor-buyer",
		"roles":         []string{"viewer"},
		"department_id": "ops",
	})

	forbidden := doJSONWithHeaders(t, h, http.MethodPost, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-001",
		"name": "Acme Supply",
	}, http.StatusForbidden, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  "actor-buyer",
	})
	if forbidden.Error["message"] == "" {
		t.Fatal("expected forbidden policy message")
	}

	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id":     tenantID,
		"actor_id":      "actor-buyer",
		"roles":         []string{"supplychain_operator"},
		"department_id": "ops",
	})

	supplier := doJSONWithHeaders(t, h, http.MethodPost, "/api/admin/v1/master-data/suppliers", map[string]any{
		"code": "SUP-002",
		"name": "Acme Supply 2",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  "actor-buyer",
	})
	if stringField(t, supplier.Data, "code") != "SUP-002" {
		t.Fatalf("expected supplier code SUP-002, got %s", stringField(t, supplier.Data, "code"))
	}

	session := postJSONData(t, h, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-001",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	})
	if got := stringField(t, session, "status"); got != "open" {
		t.Fatalf("expected open session, got %s", got)
	}

	task := postJSONData(t, h, "/api/platform/v1/agent/sessions/sess-001/tasks", map[string]any{
		"task_id":   "task-001",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	})
	if got := stringField(t, task, "status"); got != "pending" {
		t.Fatalf("expected pending task, got %s", got)
	}

	started := postJSONData(t, h, "/api/platform/v1/agent/tasks/task-001/start", map[string]any{})
	if got := stringField(t, started, "status"); got != "running" {
		t.Fatalf("expected running task, got %s", got)
	}

	gotSession := getJSONData(t, h, "/api/platform/v1/agent/sessions/sess-001")
	if got := stringField(t, gotSession, "id"); got != "sess-001" {
		t.Fatalf("expected session id sess-001, got %s", got)
	}

	sessionTasks := getJSONData(t, h, "/api/platform/v1/agent/sessions/sess-001/tasks")
	taskItems, ok := sessionTasks["tasks"].([]any)
	if !ok {
		t.Fatalf("expected tasks array, got %#v", sessionTasks["tasks"])
	}
	if len(taskItems) != 1 {
		t.Fatalf("expected 1 task item, got %d", len(taskItems))
	}

	completed := postJSONData(t, h, "/api/platform/v1/agent/tasks/task-001/complete", map[string]any{
		"output": map[string]any{
			"result": "ok",
		},
	})
	if got := stringField(t, completed, "status"); got != "succeeded" {
		t.Fatalf("expected succeeded task, got %s", got)
	}

	gotTask := getJSONData(t, h, "/api/platform/v1/agent/tasks/task-001")
	if got := stringField(t, gotTask, "status"); got != "succeeded" {
		t.Fatalf("expected task status succeeded, got %s", got)
	}

	auditResp := getJSONData(t, h, "/api/platform/v1/audit/records?limit=20")
	records, ok := auditResp["records"].([]any)
	if !ok {
		t.Fatalf("expected records array, got %#v", auditResp["records"])
	}
	if len(records) == 0 {
		t.Fatal("expected audit records")
	}
}

func TestPlatformControlPlaneSessionCloseAndListFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	headers := map[string]string{
		"X-Tenant-ID": "tenant-admin",
	}
	session := doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-close-001",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	}, http.StatusOK, headers).Data
	if got := stringField(t, session, "status"); got != "open" {
		t.Fatalf("expected open session, got %s", got)
	}

	listed := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/sessions", nil, http.StatusOK, headers).Data
	items, ok := listed["sessions"].([]any)
	if !ok {
		t.Fatalf("expected sessions array, got %#v", listed["sessions"])
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 session item, got %d", len(items))
	}
	item, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected session object, got %#v", items[0])
	}
	if got := stringField(t, item, "id"); got != "sess-close-001" {
		t.Fatalf("expected session id sess-close-001, got %s", got)
	}

	closed := doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-close-001/close", nil, http.StatusOK, headers).Data
	if got := stringField(t, closed, "status"); got != "closed" {
		t.Fatalf("expected closed session, got %s", got)
	}

	denied := doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-close-001/tasks", map[string]any{
		"task_id":   "task-after-close",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	}, http.StatusBadRequest, headers)
	if denied.Error["message"] == "" {
		t.Fatal("expected enqueue rejection message")
	}
}

func TestPlatformControlPlaneTaskCancelFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	postJSONData(t, h, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-cancel-001",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	})
	postJSONData(t, h, "/api/platform/v1/agent/sessions/sess-cancel-001/tasks", map[string]any{
		"task_id":   "task-cancel-001",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	})

	canceled := postJSONData(t, h, "/api/platform/v1/agent/tasks/task-cancel-001/cancel", map[string]any{
		"reason": "manual cancel",
	})
	if got := stringField(t, canceled, "status"); got != "canceled" {
		t.Fatalf("expected canceled task, got %s", got)
	}
	if got := stringField(t, canceled, "failure_reason"); got != "manual cancel" {
		t.Fatalf("expected cancel reason manual cancel, got %s", got)
	}

	gotTask := getJSONData(t, h, "/api/platform/v1/agent/tasks/task-cancel-001")
	if got := stringField(t, gotTask, "status"); got != "canceled" {
		t.Fatalf("expected persisted canceled task, got %s", got)
	}
}

func TestPlatformControlPlaneCloseSessionRejectedWhenTaskActive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	postJSONData(t, h, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-active-001",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	})
	postJSONData(t, h, "/api/platform/v1/agent/sessions/sess-active-001/tasks", map[string]any{
		"task_id":   "task-active-001",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	})

	rejected := doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-active-001/close", nil, http.StatusBadRequest, nil)
	if rejected.Error["message"] == "" {
		t.Fatal("expected session close rejection message")
	}

	current := getJSONData(t, h, "/api/platform/v1/agent/sessions/sess-active-001")
	if got := stringField(t, current, "status"); got != "open" {
		t.Fatalf("expected session to remain open, got %s", got)
	}
}

func TestPlatformControlPlaneListTasksSupportsSessionAndStatusFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	postJSONData(t, h, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-list-001",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	})
	postJSONData(t, h, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-list-002",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	})

	postJSONData(t, h, "/api/platform/v1/agent/sessions/sess-list-001/tasks", map[string]any{
		"task_id":   "task-pending-001",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	})
	postJSONData(t, h, "/api/platform/v1/agent/sessions/sess-list-002/tasks", map[string]any{
		"task_id":   "task-running-001",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	})
	postJSONData(t, h, "/api/platform/v1/agent/tasks/task-running-001/start", map[string]any{})

	runningList := getJSONData(t, h, "/api/platform/v1/agent/tasks?status=running")
	runningItems, ok := runningList["tasks"].([]any)
	if !ok {
		t.Fatalf("expected tasks array, got %#v", runningList["tasks"])
	}
	if len(runningItems) != 1 {
		t.Fatalf("expected 1 running task, got %d", len(runningItems))
	}
	runningTask, ok := runningItems[0].(map[string]any)
	if !ok {
		t.Fatalf("expected task object, got %#v", runningItems[0])
	}
	if got := stringField(t, runningTask, "id"); got != "task-running-001" {
		t.Fatalf("expected running task id task-running-001, got %s", got)
	}

	sessionList := getJSONData(t, h, "/api/platform/v1/agent/tasks?session_id=sess-list-001")
	sessionItems, ok := sessionList["tasks"].([]any)
	if !ok {
		t.Fatalf("expected tasks array, got %#v", sessionList["tasks"])
	}
	if len(sessionItems) != 1 {
		t.Fatalf("expected 1 session-filtered task, got %d", len(sessionItems))
	}
	sessionTask, ok := sessionItems[0].(map[string]any)
	if !ok {
		t.Fatalf("expected task object, got %#v", sessionItems[0])
	}
	if got := stringField(t, sessionTask, "id"); got != "task-pending-001" {
		t.Fatalf("expected session task id task-pending-001, got %s", got)
	}
}

func TestPlatformControlPlaneListSessionsSupportsStatusFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	postJSONData(t, h, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-filter-open",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	})
	postJSONData(t, h, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-filter-closed",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	})
	postJSONData(t, h, "/api/platform/v1/agent/sessions/sess-filter-closed/close", map[string]any{})

	filtered := getJSONData(t, h, "/api/platform/v1/agent/sessions?status=closed")
	items, ok := filtered["sessions"].([]any)
	if !ok {
		t.Fatalf("expected sessions array, got %#v", filtered["sessions"])
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 closed session, got %d", len(items))
	}
	session, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected session object, got %#v", items[0])
	}
	if got := stringField(t, session, "id"); got != "sess-filter-closed" {
		t.Fatalf("expected sess-filter-closed, got %s", got)
	}
}

func TestPlatformControlPlaneListTasksSupportsPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	postJSONData(t, h, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-page-001",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	})
	for _, taskID := range []string{"task-page-001", "task-page-002", "task-page-003"} {
		postJSONData(t, h, "/api/platform/v1/agent/sessions/sess-page-001/tasks", map[string]any{
			"task_id":   taskID,
			"task_type": "tool.call",
			"input": map[string]any{
				"tool": "search",
			},
		})
	}

	paged := getJSONData(t, h, "/api/platform/v1/agent/tasks?offset=1&limit=1")
	items, ok := paged["tasks"].([]any)
	if !ok {
		t.Fatalf("expected tasks array, got %#v", paged["tasks"])
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 paged task, got %d", len(items))
	}
	task, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected task object, got %#v", items[0])
	}
	if got := stringField(t, task, "id"); got != "task-page-002" {
		t.Fatalf("expected task-page-002, got %s", got)
	}
}

func TestPlatformControlPlaneListSessionsSupportsPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	for _, sessionID := range []string{"sess-page-001", "sess-page-002", "sess-page-003"} {
		postJSONData(t, h, "/api/platform/v1/agent/sessions", map[string]any{
			"session_id": sessionID,
			"metadata": map[string]any{
				"channel": "workspace",
			},
		})
	}

	paged := getJSONData(t, h, "/api/platform/v1/agent/sessions?offset=1&limit=1")
	items, ok := paged["sessions"].([]any)
	if !ok {
		t.Fatalf("expected sessions array, got %#v", paged["sessions"])
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 paged session, got %d", len(items))
	}
	session, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected session object, got %#v", items[0])
	}
	if got := stringField(t, session, "id"); got != "sess-page-002" {
		t.Fatalf("expected sess-page-002, got %s", got)
	}
}

func TestPlatformControlPlaneListSessionTasksSupportsStatusAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	postJSONData(t, h, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-filter-001",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	})
	for _, taskID := range []string{"task-001", "task-002", "task-003"} {
		postJSONData(t, h, "/api/platform/v1/agent/sessions/sess-filter-001/tasks", map[string]any{
			"task_id":   taskID,
			"task_type": "tool.call",
			"input": map[string]any{
				"tool": "search",
			},
		})
	}
	postJSONData(t, h, "/api/platform/v1/agent/tasks/task-002/start", map[string]any{})

	running := getJSONData(t, h, "/api/platform/v1/agent/sessions/sess-filter-001/tasks?status=running")
	runningItems, ok := running["tasks"].([]any)
	if !ok {
		t.Fatalf("expected tasks array, got %#v", running["tasks"])
	}
	if len(runningItems) != 1 {
		t.Fatalf("expected 1 running task, got %d", len(runningItems))
	}
	runningTask, ok := runningItems[0].(map[string]any)
	if !ok {
		t.Fatalf("expected task object, got %#v", runningItems[0])
	}
	if got := stringField(t, runningTask, "id"); got != "task-002" {
		t.Fatalf("expected task-002, got %s", got)
	}

	paged := getJSONData(t, h, "/api/platform/v1/agent/sessions/sess-filter-001/tasks?offset=1&limit=1")
	pagedItems, ok := paged["tasks"].([]any)
	if !ok {
		t.Fatalf("expected tasks array, got %#v", paged["tasks"])
	}
	if len(pagedItems) != 1 {
		t.Fatalf("expected 1 paged task, got %d", len(pagedItems))
	}
	pagedTask, ok := pagedItems[0].(map[string]any)
	if !ok {
		t.Fatalf("expected task object, got %#v", pagedItems[0])
	}
	if got := stringField(t, pagedTask, "id"); got != "task-002" {
		t.Fatalf("expected task-002 in paged result, got %s", got)
	}
}

func doJSONWithHeaders(
	t *testing.T,
	h http.Handler,
	method, path string,
	body any,
	expectedStatus int,
	headers map[string]string,
) envelope {
	t.Helper()

	var reqBody *bytes.Reader
	if body == nil {
		reqBody = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if headers == nil {
		headers = map[string]string{}
	}
	if _, ok := headers["X-Tenant-ID"]; !ok {
		headers["X-Tenant-ID"] = "tenant-admin"
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d with body %s", expectedStatus, rec.Code, rec.Body.String())
	}

	var env envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if env.Meta["request_id"] == "" {
		t.Fatal("expected meta.request_id")
	}
	return env
}
