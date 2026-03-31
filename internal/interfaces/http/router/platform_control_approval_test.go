package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
)

func TestApprovalActionsDeniedWithoutRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := New(WithContainer(container))

	tenantID := "tenant-approval-deny"
	operatorActorID := "actor-operator"

	doApprovalJSON(t, h, http.MethodPost, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id":     tenantID,
		"actor_id":      operatorActorID,
		"roles":         []string{"workspace_operator"},
		"department_id": "ops",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  "system",
	})

	doApprovalJSON(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-approval-deny",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  operatorActorID,
	})

	doApprovalJSON(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-approval-deny/tasks", map[string]any{
		"task_id":   "task-approval-deny",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  operatorActorID,
	})

	for _, path := range []string{
		"/api/platform/v1/agent/tasks/task-approval-deny/pause",
		"/api/platform/v1/agent/tasks/task-approval-deny/resume",
		"/api/platform/v1/agent/tasks/task-approval-deny/handoff",
	} {
		resp := doApprovalJSON(t, h, http.MethodPost, path, map[string]any{}, http.StatusForbidden, map[string]string{
			"X-Tenant-ID": tenantID,
		})
		if resp.Error["message"] != shared.ErrPolicyDenied.Error() {
			t.Fatalf("expected policy denied message for %s, got %q", path, resp.Error["message"])
		}
	}

	audit := doApprovalJSON(t, h, http.MethodGet, "/api/platform/v1/audit/records?limit=50", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  "system",
	})
	records, ok := audit.Data["records"].([]any)
	if !ok {
		t.Fatalf("expected records array, got %#v", audit.Data["records"])
	}

	for _, command := range []string{"runtime.tasks.pause", "runtime.tasks.resume", "runtime.tasks.handoff"} {
		if !hasDeniedAuditRecord(records, command) {
			t.Fatalf("expected denied audit record for %s", command)
		}
	}
}

func hasDeniedAuditRecord(records []any, command string) bool {
	for _, item := range records {
		record, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if record["command_name"] != command {
			continue
		}
		if record["outcome"] == "rejected" && record["error"] == shared.ErrPolicyDenied.Error() {
			return true
		}
	}
	return false
}

type approvalEnvelope struct {
	Data  map[string]any    `json:"data"`
	Error map[string]string `json:"error"`
	Meta  map[string]string `json:"meta"`
}

func doApprovalJSON(
	t *testing.T,
	h http.Handler,
	method, path string,
	body any,
	expectedStatus int,
	headers map[string]string,
) approvalEnvelope {
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
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d with body %s", expectedStatus, rec.Code, rec.Body.String())
	}

	var env approvalEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if env.Meta["request_id"] == "" {
		t.Fatal("expected meta.request_id")
	}
	return env
}
