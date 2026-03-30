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

func TestControlPlaneGovernanceCommandsReturnExplicitRejection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	tenantID := "tenant-governance-reject"
	actorID := "actor-governance"

	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id":     tenantID,
		"actor_id":      actorID,
		"roles":         []string{"workspace_operator"},
		"department_id": "ops",
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-governance",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-governance/tasks", map[string]any{
		"task_id":   "task-governance",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	for _, path := range []string{
		"/api/platform/v1/agent/tasks/task-governance/pause",
		"/api/platform/v1/agent/tasks/task-governance/resume",
		"/api/platform/v1/agent/tasks/task-governance/handoff",
	} {
		rejected := doJSONWithHeaders(t, h, http.MethodPost, path, map[string]any{}, http.StatusConflict, map[string]string{
			"X-Tenant-ID": tenantID,
			"X-Actor-ID":  actorID,
		})
		if rejected.Error["message"] == "" {
			t.Fatalf("expected explicit rejection message for %s", path)
		}
	}

	gotTask := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/tasks/task-governance", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})
	if got := stringField(t, gotTask.Data, "status"); got != "pending" {
		t.Fatalf("expected task status pending after governance rejection, got %s", got)
	}
}

func TestControlPlaneCommandRequiresActorContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	tenantID := "tenant-actor-required"
	actorID := "actor-runtime"

	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id":     tenantID,
		"actor_id":      actorID,
		"roles":         []string{"workspace_operator"},
		"department_id": "ops",
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-require-actor",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-require-actor/tasks", map[string]any{
		"task_id":   "task-require-actor",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorID,
	})

	for _, path := range []string{
		"/api/platform/v1/agent/tasks/task-require-actor/pause",
		"/api/platform/v1/agent/tasks/task-require-actor/resume",
		"/api/platform/v1/agent/tasks/task-require-actor/handoff",
	} {
		forbidden := doJSONWithHeaders(t, h, http.MethodPost, path, map[string]any{}, http.StatusForbidden, map[string]string{
			"X-Tenant-ID": tenantID,
		})
		if forbidden.Error["message"] == "" {
			t.Fatalf("expected forbidden error message for %s", path)
		}
	}
}

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
		"roles":         []string{"supplychain_operator", "workspace_operator"},
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

	started := doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/task-001/start", map[string]any{}, http.StatusOK, map[string]string{
		"X-Tenant-ID": "tenant-admin",
		"X-Actor-ID":  "system",
	})
	if got := stringField(t, started.Data, "status"); got != "running" {
		t.Fatalf("expected running task, got %s", got)
	}

	gotSession := getJSONData(t, h, "/api/platform/v1/agent/sessions/sess-001")
	if got := stringField(t, gotSession, "id"); got != "sess-001" {
		t.Fatalf("expected session id sess-001, got %s", got)
	}

	sessionList := getJSONData(t, h, "/api/platform/v1/agent/sessions?status=open&limit=10")
	sessionItems, ok := sessionList["items"].([]any)
	if !ok {
		t.Fatalf("expected session items array, got %#v", sessionList["items"])
	}
	if len(sessionItems) != 1 {
		t.Fatalf("expected 1 open session, got %d", len(sessionItems))
	}
	if stringField(t, sessionList, "as_of") == "" {
		t.Fatal("expected session list as_of")
	}

	sessionTasks := getJSONData(t, h, "/api/platform/v1/agent/sessions/sess-001/tasks")
	taskItems, ok := sessionTasks["tasks"].([]any)
	if !ok {
		t.Fatalf("expected tasks array, got %#v", sessionTasks["tasks"])
	}
	if len(taskItems) != 1 {
		t.Fatalf("expected 1 task item, got %d", len(taskItems))
	}

	runningTasks := getJSONData(t, h, "/api/platform/v1/agent/tasks?status=running&limit=1")
	runningItems, ok := runningTasks["items"].([]any)
	if !ok {
		t.Fatalf("expected running items array, got %#v", runningTasks["items"])
	}
	if len(runningItems) != 1 {
		t.Fatalf("expected 1 running task item, got %d", len(runningItems))
	}
	if stringField(t, runningTasks, "as_of") == "" {
		t.Fatal("expected task list as_of")
	}

	completed := doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/task-001/complete", map[string]any{
		"output": map[string]any{
			"result": "ok",
		},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": "tenant-admin",
		"X-Actor-ID":  "system",
	})
	if got := stringField(t, completed.Data, "status"); got != "succeeded" {
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
