package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
)

func TestListTasksSupportsStatusAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := New(WithContainer(container))

	tenantID := "tenant-e1"
	actorID := "actor-e1"

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id":     tenantID,
		"actor_id":      actorID,
		"roles":         []string{"workspace_operator"},
		"department_id": "ops",
	}, http.StatusOK, nil)

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-e1",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	for _, taskID := range []string{"task-e1-001", "task-e1-002", "task-e1-003"} {
		doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-e1/tasks", map[string]any{
			"task_id":   taskID,
			"task_type": "tool.call",
			"input": map[string]any{
				"tool": "search",
			},
		}, http.StatusOK, map[string]string{
			"X-Tenant-ID": tenantID,
			"X-Actor-ID":  actorID,
		})
		doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/"+taskID+"/start", map[string]any{}, http.StatusOK, map[string]string{
			"X-Tenant-ID": tenantID,
			"X-Actor-ID":  actorID,
		})
	}

	firstPage := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/tasks?status=running&limit=2", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data

	items, ok := firstPage["items"].([]any)
	if !ok {
		t.Fatalf("expected items array, got %#v", firstPage["items"])
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	nextCursor := stringField(t, firstPage, "next_cursor")
	if nextCursor == "" {
		t.Fatal("expected next_cursor to be populated")
	}
	if stringField(t, firstPage, "as_of") == "" {
		t.Fatal("expected as_of to be populated")
	}

	secondPage := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/tasks?status=running&limit=2&cursor="+nextCursor, nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data

	secondItems, ok := secondPage["items"].([]any)
	if !ok {
		t.Fatalf("expected items array, got %#v", secondPage["items"])
	}
	if len(secondItems) != 1 {
		t.Fatalf("expected 1 item, got %d", len(secondItems))
	}
	if stringField(t, secondPage, "next_cursor") != "" {
		t.Fatalf("expected empty next_cursor on final page, got %s", stringField(t, secondPage, "next_cursor"))
	}
}

func TestSessionAndTaskDetailQueries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := New(WithContainer(container))

	tenantID := "tenant-e1-queries"
	actorID := "actor-e1-queries"

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id":     tenantID,
		"actor_id":      actorID,
		"roles":         []string{"workspace_operator"},
		"department_id": "ops",
	}, http.StatusOK, nil)

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-open-001",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})
	secondSession := doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-open-002",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data

	task := doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-open-001/tasks", map[string]any{
		"task_id":   "task-detail-001",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data

	sessionList := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/sessions?status=open&limit=10", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data

	sessionItems, ok := sessionList["items"].([]any)
	if !ok {
		t.Fatalf("expected items array, got %#v", sessionList["items"])
	}
	if len(sessionItems) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessionItems))
	}
	if stringField(t, sessionList, "as_of") == "" {
		t.Fatal("expected as_of to be populated")
	}

	gotSession := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/sessions/"+stringField(t, secondSession, "id"), nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data
	if stringField(t, gotSession, "id") != "sess-open-002" {
		t.Fatalf("expected session id sess-open-002, got %s", stringField(t, gotSession, "id"))
	}

	gotTask := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/tasks/"+stringField(t, task, "id"), nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data
	if stringField(t, gotTask, "id") != "task-detail-001" {
		t.Fatalf("expected task id task-detail-001, got %s", stringField(t, gotTask, "id"))
	}
}

func TestListTasksSupportsSessionAndStatusFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := New(WithContainer(container))

	tenantID := "tenant-e1-session-filter"
	actorID := "actor-e1-session-filter"

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id":     tenantID,
		"actor_id":      actorID,
		"roles":         []string{"workspace_operator"},
		"department_id": "ops",
	}, http.StatusOK, nil)

	for _, sessionID := range []string{"sess-filter-001", "sess-filter-002"} {
		doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
			"session_id": sessionID,
		}, http.StatusOK, map[string]string{
			"X-Tenant-ID": tenantID,
			"X-Actor-ID":  actorID,
		})
	}

	for _, taskID := range []string{"task-filter-001", "task-filter-002"} {
		doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-filter-001/tasks", map[string]any{
			"task_id":   taskID,
			"task_type": "tool.call",
			"input": map[string]any{
				"tool": "search",
			},
		}, http.StatusOK, map[string]string{
			"X-Tenant-ID": tenantID,
			"X-Actor-ID":  actorID,
		})
		doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/"+taskID+"/start", map[string]any{}, http.StatusOK, map[string]string{
			"X-Tenant-ID": tenantID,
			"X-Actor-ID":  actorID,
		})
	}

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-filter-002/tasks", map[string]any{
		"task_id":   "task-filter-003",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	filtered := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/tasks?session_id=sess-filter-001&status=running&limit=10", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	}).Data

	items, ok := filtered["items"].([]any)
	if !ok {
		t.Fatalf("expected filtered items array, got %#v", filtered["items"])
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 running tasks for target session, got %d", len(items))
	}
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("expected object task item, got %#v", raw)
		}
		if stringField(t, item, "session_id") != "sess-filter-001" {
			t.Fatalf("expected only sess-filter-001 items, got %s", stringField(t, item, "session_id"))
		}
		if stringField(t, item, "status") != "running" {
			t.Fatalf("expected only running items, got %s", stringField(t, item, "status"))
		}
	}
}

func TestSessionAndTaskListsAreIsolatedByActorWithinTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := New(WithContainer(container))

	tenantID := "tenant-e1-actor-scope"
	actorA := "actor-a"
	actorB := "actor-b"

	for _, actorID := range []string{actorA, actorB} {
		doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/control/actors", map[string]any{
			"tenant_id":     tenantID,
			"actor_id":      actorID,
			"roles":         []string{"workspace_operator"},
			"department_id": "ops",
		}, http.StatusOK, nil)
	}

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{"session_id": "sess-actor-a"}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{"session_id": "sess-actor-b"}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorB,
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-actor-a/tasks", map[string]any{
		"task_id":   "task-actor-a",
		"task_type": "tool.call",
		"input":     map[string]any{"tool": "search"},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/task-actor-a/start", map[string]any{}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-actor-b/tasks", map[string]any{
		"task_id":   "task-actor-b",
		"task_type": "tool.call",
		"input":     map[string]any{"tool": "search"},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorB,
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/task-actor-b/start", map[string]any{}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorB,
	})

	sessionsForA := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/sessions?status=open&limit=10", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	}).Data
	sessionItems, ok := sessionsForA["items"].([]any)
	if !ok {
		t.Fatalf("expected sessions items array, got %#v", sessionsForA["items"])
	}
	if len(sessionItems) != 1 {
		t.Fatalf("expected 1 actor-scoped session, got %d", len(sessionItems))
	}
	sessionItem, ok := sessionItems[0].(map[string]any)
	if !ok {
		t.Fatalf("expected session object, got %#v", sessionItems[0])
	}
	if stringField(t, sessionItem, "id") != "sess-actor-a" {
		t.Fatalf("expected sess-actor-a, got %s", stringField(t, sessionItem, "id"))
	}

	tasksForA := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/tasks?status=running&limit=10", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	}).Data
	taskItems, ok := tasksForA["items"].([]any)
	if !ok {
		t.Fatalf("expected tasks items array, got %#v", tasksForA["items"])
	}
	if len(taskItems) != 1 {
		t.Fatalf("expected 1 actor-scoped task, got %d", len(taskItems))
	}
	taskItem, ok := taskItems[0].(map[string]any)
	if !ok {
		t.Fatalf("expected task object, got %#v", taskItems[0])
	}
	if stringField(t, taskItem, "id") != "task-actor-a" {
		t.Fatalf("expected task-actor-a, got %s", stringField(t, taskItem, "id"))
	}
}

type routerEnvelope struct {
	Data  map[string]any `json:"data"`
	Error map[string]any `json:"error"`
	Meta  map[string]any `json:"meta"`
}

func doJSONWithHeaders(
	t *testing.T,
	h http.Handler,
	method, path string,
	body any,
	expectedStatus int,
	headers map[string]string,
) routerEnvelope {
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

	var env routerEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if env.Meta["request_id"] == "" {
		t.Fatal("expected meta.request_id")
	}
	return env
}

func stringField(t *testing.T, payload map[string]any, field string) string {
	t.Helper()
	value, ok := payload[field]
	if !ok {
		t.Fatalf("missing field %q in payload %#v", field, payload)
	}
	str, ok := value.(string)
	if !ok {
		t.Fatalf("expected field %q to be a string, got %#v", field, value)
	}
	return str
}
