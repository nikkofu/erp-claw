package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

	for _, path := range []string{
		"/api/platform/v1/agent/tasks/task-001/pause",
		"/api/platform/v1/agent/tasks/task-001/resume",
		"/api/platform/v1/agent/tasks/task-001/handoff",
	} {
		doJSONWithHeaders(t, h, http.MethodPost, path, map[string]any{}, http.StatusConflict, map[string]string{
			"X-Tenant-ID": "tenant-admin",
			"X-Actor-ID":  "system",
		})
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

	required := map[string]bool{
		"runtime.tasks.pause":   false,
		"runtime.tasks.resume":  false,
		"runtime.tasks.handoff": false,
	}
	for _, raw := range records {
		record, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		command, _ := record["command_name"].(string)
		if _, exists := required[command]; !exists {
			continue
		}
		required[command] = true
		if got := strings.ToLower(stringField(t, record, "decision")); got != "allow" {
			t.Fatalf("expected decision allow for %s, got %s", command, got)
		}
		if got := stringField(t, record, "outcome"); got != "failed" {
			t.Fatalf("expected outcome failed for %s, got %s", command, got)
		}
		if got := stringField(t, record, "correlation_id"); got == "" {
			t.Fatalf("expected correlation_id for %s", command)
		}
		if got := stringField(t, record, "resource_type"); got == "" {
			t.Fatalf("expected resource_type for %s", command)
		}
		if got := stringField(t, record, "resource_id"); got == "" {
			t.Fatalf("expected resource_id for %s", command)
		}
	}
	for command, seen := range required {
		if !seen {
			t.Fatalf("expected audit record for %s", command)
		}
	}
}

func TestControlPlaneSessionAndTaskQueriesAreActorScopedWithinTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	tenantID := "tenant-integration-actor-scope"
	actorA := "actor-integration-a"
	actorB := "actor-integration-b"

	for _, actorID := range []string{actorA, actorB} {
		postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
			"tenant_id":     tenantID,
			"actor_id":      actorID,
			"roles":         []string{"workspace_operator"},
			"department_id": "ops",
		})
	}

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-int-actor-a",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-int-actor-b",
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorB,
	})

	for _, taskID := range []string{"task-int-actor-a-1", "task-int-actor-a-2"} {
		doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-int-actor-a/tasks", map[string]any{
			"task_id":   taskID,
			"task_type": "tool.call",
			"input":     map[string]any{"tool": "search"},
		}, http.StatusOK, map[string]string{
			"X-Tenant-ID": tenantID,
			"X-Actor-ID":  actorA,
		})
		doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/"+taskID+"/start", map[string]any{}, http.StatusOK, map[string]string{
			"X-Tenant-ID": tenantID,
			"X-Actor-ID":  actorA,
		})
	}

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions/sess-int-actor-b/tasks", map[string]any{
		"task_id":   "task-int-actor-b",
		"task_type": "tool.call",
		"input":     map[string]any{"tool": "search"},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorB,
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/tasks/task-int-actor-b/start", map[string]any{}, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorB,
	})

	sessionsForA := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/sessions?status=open&limit=10", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	})
	sessionItems, ok := sessionsForA.Data["items"].([]any)
	if !ok {
		t.Fatalf("expected sessions items array, got %#v", sessionsForA.Data["items"])
	}
	if len(sessionItems) != 1 {
		t.Fatalf("expected 1 actor-scoped session, got %d", len(sessionItems))
	}
	sessionItem, ok := sessionItems[0].(map[string]any)
	if !ok {
		t.Fatalf("expected session object, got %#v", sessionItems[0])
	}
	if stringField(t, sessionItem, "id") != "sess-int-actor-a" {
		t.Fatalf("expected sess-int-actor-a, got %s", stringField(t, sessionItem, "id"))
	}

	page1 := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/tasks?status=running&limit=1", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	})
	page1Items, ok := page1.Data["items"].([]any)
	if !ok {
		t.Fatalf("expected tasks page1 items array, got %#v", page1.Data["items"])
	}
	if len(page1Items) != 1 {
		t.Fatalf("expected 1 actor-scoped task on page1, got %d", len(page1Items))
	}
	page1Task, ok := page1Items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected task object, got %#v", page1Items[0])
	}
	if strings.HasSuffix(stringField(t, page1Task, "id"), "actor-b") {
		t.Fatalf("expected actor-a task on page1, got %s", stringField(t, page1Task, "id"))
	}
	nextCursor := stringField(t, page1.Data, "next_cursor")
	if nextCursor == "" {
		t.Fatal("expected next_cursor for actor-a page1")
	}

	page2 := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/tasks?status=running&limit=1&cursor="+nextCursor, nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	})
	page2Items, ok := page2.Data["items"].([]any)
	if !ok {
		t.Fatalf("expected tasks page2 items array, got %#v", page2.Data["items"])
	}
	if len(page2Items) != 1 {
		t.Fatalf("expected 1 actor-scoped task on page2, got %d", len(page2Items))
	}
	page2Task, ok := page2Items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected task object, got %#v", page2Items[0])
	}
	if strings.HasSuffix(stringField(t, page2Task, "id"), "actor-b") {
		t.Fatalf("expected actor-a task on page2, got %s", stringField(t, page2Task, "id"))
	}
	if stringField(t, page2.Data, "next_cursor") != "" {
		t.Fatalf("expected empty next_cursor on actor-a final page, got %s", stringField(t, page2.Data, "next_cursor"))
	}

	doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/sessions/sess-int-actor-b", nil, http.StatusNotFound, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	})
	doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/tasks/task-int-actor-b", nil, http.StatusNotFound, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	})
	doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/sessions/sess-int-actor-b/tasks", nil, http.StatusNotFound, map[string]string{
		"X-Tenant-ID": tenantID,
		"X-Actor-ID":  actorA,
	})
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

func TestPlatformControlPlaneTaskRetryFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	postJSONData(t, h, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-retry-001",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	})
	postJSONData(t, h, "/api/platform/v1/agent/sessions/sess-retry-001/tasks", map[string]any{
		"task_id":   "task-retry-001",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	})
	postJSONData(t, h, "/api/platform/v1/agent/tasks/task-retry-001/start", map[string]any{})
	postJSONData(t, h, "/api/platform/v1/agent/tasks/task-retry-001/fail", map[string]any{
		"reason": "tool timeout",
	})

	retried := postJSONData(t, h, "/api/platform/v1/agent/tasks/task-retry-001/retry", map[string]any{})
	if got := stringField(t, retried, "status"); got != "pending" {
		t.Fatalf("expected pending after retry, got %s", got)
	}
	if got := stringField(t, retried, "failure_reason"); got != "" {
		t.Fatalf("expected empty failure reason after retry, got %s", got)
	}

	restarted := postJSONData(t, h, "/api/platform/v1/agent/tasks/task-retry-001/start", map[string]any{})
	if got := stringField(t, restarted, "status"); got != "running" {
		t.Fatalf("expected running after restart, got %s", got)
	}
	if got := intField(t, restarted, "attempts"); got != 2 {
		t.Fatalf("expected attempts 2 after retry/restart, got %d", got)
	}
}

func TestPlatformControlPlaneTaskRetryRejectsWhenLimitExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	postJSONData(t, h, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-retry-limit-001",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	})
	postJSONData(t, h, "/api/platform/v1/agent/sessions/sess-retry-limit-001/tasks", map[string]any{
		"task_id":   "task-retry-limit-001",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	})

	for i := 0; i < 3; i++ {
		postJSONData(t, h, "/api/platform/v1/agent/tasks/task-retry-limit-001/start", map[string]any{})
		postJSONData(t, h, "/api/platform/v1/agent/tasks/task-retry-limit-001/fail", map[string]any{
			"reason": "tool timeout",
		})
		if i < 2 {
			postJSONData(t, h, "/api/platform/v1/agent/tasks/task-retry-limit-001/retry", map[string]any{})
		}
	}

	rejected := doJSONWithHeaders(
		t,
		h,
		http.MethodPost,
		"/api/platform/v1/agent/tasks/task-retry-limit-001/retry",
		map[string]any{},
		http.StatusConflict,
		nil,
	)
	if rejected.Error["message"] == "" {
		t.Fatal("expected retry limit rejection message")
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

	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id": "tenant-admin",
		"actor_id":  "actor-alice",
		"roles":     []string{"workspace_operator"},
	})
	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id": "tenant-admin",
		"actor_id":  "actor-bob",
		"roles":     []string{"workspace_operator"},
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-actor-filter-001",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": "tenant-admin",
		"X-Actor-ID":  "actor-alice",
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-actor-filter-002",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": "tenant-admin",
		"X-Actor-ID":  "actor-bob",
	})

	postJSONData(t, h, "/api/platform/v1/agent/sessions/sess-actor-filter-001/tasks", map[string]any{
		"task_id":   "task-actor-001",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	})
	postJSONData(t, h, "/api/platform/v1/agent/sessions/sess-actor-filter-002/tasks", map[string]any{
		"task_id":   "task-actor-002",
		"task_type": "tool.call",
		"input": map[string]any{
			"tool": "search",
		},
	})

	actorFiltered := getJSONData(t, h, "/api/platform/v1/agent/tasks?actor_id=actor-alice")
	actorItems, ok := actorFiltered["tasks"].([]any)
	if !ok {
		t.Fatalf("expected tasks array, got %#v", actorFiltered["tasks"])
	}
	if len(actorItems) != 1 {
		t.Fatalf("expected 1 actor-filtered task, got %d", len(actorItems))
	}
	actorTask, ok := actorItems[0].(map[string]any)
	if !ok {
		t.Fatalf("expected task object, got %#v", actorItems[0])
	}
	if got := stringField(t, actorTask, "id"); got != "task-actor-001" {
		t.Fatalf("expected actor-filtered task task-actor-001, got %s", got)
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

func TestPlatformControlPlaneListSessionsSupportsActorFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
	h := router.New(router.WithContainer(container))

	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id": "tenant-admin",
		"actor_id":  "actor-alice",
		"roles":     []string{"workspace_operator"},
	})
	postJSONData(t, h, "/api/platform/v1/control/actors", map[string]any{
		"tenant_id": "tenant-admin",
		"actor_id":  "actor-bob",
		"roles":     []string{"workspace_operator"},
	})

	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-actor-001",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": "tenant-admin",
		"X-Actor-ID":  "actor-alice",
	})
	doJSONWithHeaders(t, h, http.MethodPost, "/api/platform/v1/agent/sessions", map[string]any{
		"session_id": "sess-actor-002",
		"metadata": map[string]any{
			"channel": "workspace",
		},
	}, http.StatusOK, map[string]string{
		"X-Tenant-ID": "tenant-admin",
		"X-Actor-ID":  "actor-bob",
	})

	filtered := doJSONWithHeaders(t, h, http.MethodGet, "/api/platform/v1/agent/sessions?actor_id=actor-alice", nil, http.StatusOK, map[string]string{
		"X-Tenant-ID": "tenant-admin",
	}).Data
	items, ok := filtered["sessions"].([]any)
	if !ok {
		t.Fatalf("expected sessions array, got %#v", filtered["sessions"])
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 actor-filtered session, got %d", len(items))
	}
	session, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected session object, got %#v", items[0])
	}
	if got := stringField(t, session, "id"); got != "sess-actor-001" {
		t.Fatalf("expected sess-actor-001, got %s", got)
	}
	if got := stringField(t, session, "actor_id"); got != "actor-alice" {
		t.Fatalf("expected actor-alice, got %s", got)
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
