package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nikkofu/erp-claw/internal/bootstrap"
	agentruntime "github.com/nikkofu/erp-claw/internal/domain/agentruntime"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
	"github.com/nikkofu/erp-claw/internal/interfaces/ws"
	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
)

func TestWorkspaceRoutesExposeSessionsTasksAndEvents(t *testing.T) {
	catalog := bootstrap.NewInMemoryAgentRuntimeCatalogForTest()
	gateway := ws.NewWorkspaceGateway()
	if _, err := gateway.RegisterChannel("workspace-session", 8); err != nil {
		t.Fatalf("register channel: %v", err)
	}

	session, err := agentruntime.NewSession("tenant-a", "workspace-session", map[string]any{"source": "workspace-api-test"})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if _, err := catalog.CreateSession(context.Background(), session); err != nil {
		t.Fatalf("create session: %v", err)
	}

	task, err := agentruntime.NewTask("tenant-a", "workspace-session", "plan", map[string]any{"prompt": "query"})
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	createdTask, err := catalog.CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if err := gateway.AppendWorkspaceEvent(context.Background(), platformruntime.WorkspaceEvent{
		Type:      platformruntime.WorkspaceEventTypeTaskStatusChanged,
		TenantID:  "tenant-a",
		SessionID: "workspace-session",
		TaskID:    createdTask.ID,
		Payload:   map[string]any{"status": "running"},
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}

	container := bootstrap.NewTestContainer()
	container.AgentRuntimeCatalog = catalog
	container.WorkspaceGateway = gateway

	h := router.New(router.WithContainer(container))

	sessionsReq := httptest.NewRequest(http.MethodGet, "/api/workspace/v1/sessions?tenant_id=tenant-a", nil)
	sessionsReq.Header.Set("X-Tenant-ID", "tenant-a")
	sessionsRec := httptest.NewRecorder()
	h.ServeHTTP(sessionsRec, sessionsReq)
	if sessionsRec.Code != http.StatusOK {
		t.Fatalf("expected sessions 200, got %d: %s", sessionsRec.Code, sessionsRec.Body.String())
	}
	if !strings.Contains(sessionsRec.Body.String(), "workspace-session") {
		t.Fatalf("expected session response to contain workspace-session, got %s", sessionsRec.Body.String())
	}

	tasksReq := httptest.NewRequest(http.MethodGet, "/api/workspace/v1/tasks?tenant_id=tenant-a&session_id=workspace-session", nil)
	tasksReq.Header.Set("X-Tenant-ID", "tenant-a")
	tasksRec := httptest.NewRecorder()
	h.ServeHTTP(tasksRec, tasksReq)
	if tasksRec.Code != http.StatusOK {
		t.Fatalf("expected tasks 200, got %d: %s", tasksRec.Code, tasksRec.Body.String())
	}
	if !strings.Contains(tasksRec.Body.String(), "plan") {
		t.Fatalf("expected tasks response to contain task type, got %s", tasksRec.Body.String())
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/api/workspace/v1/events?tenant_id=tenant-a&session_id=workspace-session", nil)
	eventsReq.Header.Set("X-Tenant-ID", "tenant-a")
	eventsRec := httptest.NewRecorder()
	h.ServeHTTP(eventsRec, eventsReq)
	if eventsRec.Code != http.StatusOK {
		t.Fatalf("expected events 200, got %d: %s", eventsRec.Code, eventsRec.Body.String())
	}
	if !strings.Contains(eventsRec.Body.String(), "running") {
		t.Fatalf("expected events response to contain running status, got %s", eventsRec.Body.String())
	}
}

func TestWorkspaceRoutesSupportWriteFlow(t *testing.T) {
	container := bootstrap.NewTestContainer()
	container.AgentRuntimeCatalog = bootstrap.NewInMemoryAgentRuntimeCatalogForTest()
	container.WorkspaceGateway = ws.NewWorkspaceGateway()

	h := router.New(router.WithContainer(container))

	createSessionReq := httptest.NewRequest(http.MethodPost, "/api/workspace/v1/sessions", strings.NewReader(`{"session_key":"workspace-live","metadata":{"source":"workspace-write-test"}}`))
	createSessionReq.Header.Set("Content-Type", "application/json")
	createSessionReq.Header.Set("X-Tenant-ID", "tenant-a")
	createSessionRec := httptest.NewRecorder()
	h.ServeHTTP(createSessionRec, createSessionReq)
	if createSessionRec.Code != http.StatusCreated {
		t.Fatalf("expected create session 201, got %d: %s", createSessionRec.Code, createSessionRec.Body.String())
	}

	createTaskReq := httptest.NewRequest(http.MethodPost, "/api/workspace/v1/tasks", strings.NewReader(`{"session_id":"workspace-live","task_type":"plan","input":{"prompt":"month end close"}}`))
	createTaskReq.Header.Set("Content-Type", "application/json")
	createTaskReq.Header.Set("X-Tenant-ID", "tenant-a")
	createTaskRec := httptest.NewRecorder()
	h.ServeHTTP(createTaskRec, createTaskReq)
	if createTaskRec.Code != http.StatusCreated {
		t.Fatalf("expected create task 201, got %d: %s", createTaskRec.Code, createTaskRec.Body.String())
	}

	taskID := firstWorkspaceIDFromResponse(t, createTaskRec.Body.Bytes())

	startTaskReq := httptest.NewRequest(http.MethodPost, "/api/workspace/v1/tasks/"+taskID+"/start", strings.NewReader(`{}`))
	startTaskReq.Header.Set("Content-Type", "application/json")
	startTaskReq.Header.Set("X-Tenant-ID", "tenant-a")
	startTaskRec := httptest.NewRecorder()
	h.ServeHTTP(startTaskRec, startTaskReq)
	if startTaskRec.Code != http.StatusCreated {
		t.Fatalf("expected start task 201, got %d: %s", startTaskRec.Code, startTaskRec.Body.String())
	}

	completeTaskReq := httptest.NewRequest(http.MethodPost, "/api/workspace/v1/tasks/"+taskID+"/complete", strings.NewReader(`{"output":{"result":"done"}}`))
	completeTaskReq.Header.Set("Content-Type", "application/json")
	completeTaskReq.Header.Set("X-Tenant-ID", "tenant-a")
	completeTaskRec := httptest.NewRecorder()
	h.ServeHTTP(completeTaskRec, completeTaskReq)
	if completeTaskRec.Code != http.StatusCreated {
		t.Fatalf("expected complete task 201, got %d: %s", completeTaskRec.Code, completeTaskRec.Body.String())
	}

	closeSessionReq := httptest.NewRequest(http.MethodPost, "/api/workspace/v1/sessions/workspace-live/close", strings.NewReader(`{}`))
	closeSessionReq.Header.Set("Content-Type", "application/json")
	closeSessionReq.Header.Set("X-Tenant-ID", "tenant-a")
	closeSessionRec := httptest.NewRecorder()
	h.ServeHTTP(closeSessionRec, closeSessionReq)
	if closeSessionRec.Code != http.StatusCreated {
		t.Fatalf("expected close session 201, got %d: %s", closeSessionRec.Code, closeSessionRec.Body.String())
	}

	sessionsReq := httptest.NewRequest(http.MethodGet, "/api/workspace/v1/sessions?tenant_id=tenant-a", nil)
	sessionsReq.Header.Set("X-Tenant-ID", "tenant-a")
	sessionsRec := httptest.NewRecorder()
	h.ServeHTTP(sessionsRec, sessionsReq)
	if sessionsRec.Code != http.StatusOK {
		t.Fatalf("expected sessions 200, got %d: %s", sessionsRec.Code, sessionsRec.Body.String())
	}
	if !strings.Contains(sessionsRec.Body.String(), "closed") {
		t.Fatalf("expected sessions response to contain closed status, got %s", sessionsRec.Body.String())
	}

	tasksReq := httptest.NewRequest(http.MethodGet, "/api/workspace/v1/tasks?tenant_id=tenant-a&session_id=workspace-live", nil)
	tasksReq.Header.Set("X-Tenant-ID", "tenant-a")
	tasksRec := httptest.NewRecorder()
	h.ServeHTTP(tasksRec, tasksReq)
	if tasksRec.Code != http.StatusOK {
		t.Fatalf("expected tasks 200, got %d: %s", tasksRec.Code, tasksRec.Body.String())
	}
	if !strings.Contains(tasksRec.Body.String(), "succeeded") {
		t.Fatalf("expected tasks response to contain succeeded status, got %s", tasksRec.Body.String())
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/api/workspace/v1/events?tenant_id=tenant-a&session_id=workspace-live", nil)
	eventsReq.Header.Set("X-Tenant-ID", "tenant-a")
	eventsRec := httptest.NewRecorder()
	h.ServeHTTP(eventsRec, eventsReq)
	if eventsRec.Code != http.StatusOK {
		t.Fatalf("expected events 200, got %d: %s", eventsRec.Code, eventsRec.Body.String())
	}
	for _, needle := range []string{"running", "succeeded", "closed"} {
		if !strings.Contains(eventsRec.Body.String(), needle) {
			t.Fatalf("expected events response to contain %s, got %s", needle, eventsRec.Body.String())
		}
	}
}

func firstWorkspaceIDFromResponse(t *testing.T, payload []byte) string {
	t.Helper()

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("decode workspace response: %v", err)
	}
	id, _ := envelope.Data["ID"].(string)
	if id == "" {
		t.Fatalf("expected ID in workspace response, got %+v", envelope.Data)
	}
	return id
}
