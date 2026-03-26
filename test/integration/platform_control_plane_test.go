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

	completed := postJSONData(t, h, "/api/platform/v1/agent/tasks/task-001/complete", map[string]any{
		"output": map[string]any{
			"result": "ok",
		},
	})
	if got := stringField(t, completed, "status"); got != "succeeded" {
		t.Fatalf("expected succeeded task, got %s", got)
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
