package integration

import (
	"context"
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

	container := bootstrap.NewContainer(bootstrap.DefaultConfig())
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
