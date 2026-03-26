package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	controlcommand "github.com/nikkofu/erp-claw/internal/application/controlplane/command"
	sharedoutbox "github.com/nikkofu/erp-claw/internal/application/shared/outbox"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestControlPlaneMigrationContainsCatalogTables(t *testing.T) {
	data, err := os.ReadFile("../../migrations/000002_init_control_plane_catalog.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := strings.ToLower(string(data))

	required := []string{
		"create table if not exists organization",
		"create table if not exists iam_user",
		"create table if not exists iam_role",
		"create table if not exists iam_user_role_binding",
		"create table if not exists agent_profile",
	}
	for _, needle := range required {
		if !strings.Contains(sql, needle) {
			t.Fatalf("expected migration to contain %q", needle)
		}
	}

	iamUser := mustTableBlock(t, sql, "iam_user")
	requireContainsAll(t, iamUser, []string{
		"tenant_id text not null",
		"id text not null",
		"email text not null",
		"display_name text not null",
	})

	iamRole := mustTableBlock(t, sql, "iam_role")
	requireContainsAll(t, iamRole, []string{
		"tenant_id text not null",
		"id text not null",
		"name text not null",
		"description text not null",
	})

	agentProfile := mustTableBlock(t, sql, "agent_profile")
	requireContainsAll(t, agentProfile, []string{
		"tenant_id text not null",
		"id text not null",
		"name text not null",
		"model text not null",
	})

	forbidden := []string{
		"organization_id text not null",
		"profile_key text not null",
		"role_key text not null",
		"config jsonb",
		"owner_user_id text",
	}
	for _, needle := range forbidden {
		if strings.Contains(sql, needle) {
			t.Fatalf("did not expect migration to contain %q", needle)
		}
	}
}

func TestTenantIAMExtensionMigrationContainsDepartmentAndBindingTables(t *testing.T) {
	data, err := os.ReadFile("../../migrations/000007_phase1_tenant_iam_extension.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := strings.ToLower(string(data))

	required := []string{
		"create table if not exists iam_department",
		"create table if not exists iam_user_role",
		"create table if not exists iam_user_department",
	}
	for _, needle := range required {
		if !strings.Contains(sql, needle) {
			t.Fatalf("expected migration to contain %q", needle)
		}
	}

	department := mustTableBlock(t, sql, "iam_department")
	requireContainsAll(t, department, []string{
		"tenant_id text not null",
		"id text not null",
		"name text not null",
		"parent_department_id text",
	})

	userRole := mustTableBlock(t, sql, "iam_user_role")
	requireContainsAll(t, userRole, []string{
		"tenant_id text not null",
		"id text not null",
		"user_id text not null",
		"role_id text not null",
	})

	userDepartment := mustTableBlock(t, sql, "iam_user_department")
	requireContainsAll(t, userDepartment, []string{
		"tenant_id text not null",
		"id text not null",
		"user_id text not null",
		"department_id text not null",
	})
}

func mustTableBlock(t *testing.T, migrationSQL, table string) string {
	t.Helper()

	startNeedle := "create table if not exists " + table + " ("
	startIdx := strings.Index(migrationSQL, startNeedle)
	if startIdx == -1 {
		t.Fatalf("expected table block %q", table)
	}
	block := migrationSQL[startIdx:]

	endIdx := strings.Index(block, ");")
	if endIdx == -1 {
		t.Fatalf("expected table block %q to end with );", table)
	}

	return block[:endIdx]
}

func requireContainsAll(t *testing.T, block string, required []string) {
	t.Helper()
	for _, needle := range required {
		if !strings.Contains(block, needle) {
			t.Fatalf("expected table block to contain %q", needle)
		}
	}
}

func TestCreateTenantCommandRejectsEmptyCode(t *testing.T) {
	handler := controlcommand.CreateTenantHandler{}

	_, err := handler.Handle(context.Background(), controlcommand.CreateTenant{
		Code: "",
		Name: "Tenant A",
	})
	if err == nil {
		t.Fatal("expected error for empty tenant code")
	}
}

func TestAdminCreateTenantRoute(t *testing.T) {
	h := router.New(router.WithContainer(bootstrap.NewTestContainer()))
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/tenants", strings.NewReader(`{"code":"tenant-a","name":"Tenant A"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "platform-root")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestAdminRoleDepartmentLifecycleRoutes(t *testing.T) {
	h := router.New(router.WithContainer(bootstrap.NewTestContainer()))

	tenantID := createAdminEntityAndReadID(t, h, http.MethodPost, "/api/admin/v1/tenants", `{"code":"tenant-a","name":"Tenant A"}`, "platform-root")

	userID := createAdminEntityAndReadID(t, h, http.MethodPost, "/api/admin/v1/users", `{"tenant_id":"`+tenantID+`","email":"ada@example.com","display_name":"Ada"}`, tenantID)
	roleID := createAdminEntityAndReadID(t, h, http.MethodPost, "/api/admin/v1/roles", `{"name":"ops-admin","description":"Operations admin"}`, tenantID)
	departmentID := createAdminEntityAndReadID(t, h, http.MethodPost, "/api/admin/v1/departments", `{"name":"operations"}`, tenantID)

	createAdminEntityAndReadID(t, h, http.MethodPost, "/api/admin/v1/user-role-bindings", `{"user_id":"`+userID+`","role_id":"`+roleID+`"}`, tenantID)
	createAdminEntityAndReadID(t, h, http.MethodPost, "/api/admin/v1/user-department-bindings", `{"user_id":"`+userID+`","department_id":"`+departmentID+`"}`, tenantID)

	rolesReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/roles?tenant_id="+tenantID, nil)
	rolesReq.Header.Set("X-Tenant-ID", tenantID)
	rolesRec := httptest.NewRecorder()
	h.ServeHTTP(rolesRec, rolesReq)
	if rolesRec.Code != http.StatusOK {
		t.Fatalf("expected roles list 200, got %d", rolesRec.Code)
	}
	if !strings.Contains(rolesRec.Body.String(), "ops-admin") {
		t.Fatalf("expected roles list to contain created role, got %s", rolesRec.Body.String())
	}

	departmentsReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/departments?tenant_id="+tenantID, nil)
	departmentsReq.Header.Set("X-Tenant-ID", tenantID)
	departmentsRec := httptest.NewRecorder()
	h.ServeHTTP(departmentsRec, departmentsReq)
	if departmentsRec.Code != http.StatusOK {
		t.Fatalf("expected departments list 200, got %d", departmentsRec.Code)
	}
	if !strings.Contains(departmentsRec.Body.String(), "operations") {
		t.Fatalf("expected departments list to contain created department, got %s", departmentsRec.Body.String())
	}
}

func TestAdminCreateUserRouteRejectsUnknownTenant(t *testing.T) {
	h := router.New(router.WithContainer(bootstrap.NewTestContainer()))

	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/users", strings.NewReader(`{"tenant_id":"tenant-missing","email":"ada@example.com","display_name":"Ada"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-missing")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown tenant, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminApprovalLifecycleRoutes(t *testing.T) {
	h := router.New(router.WithContainer(bootstrap.NewTestContainer()))

	tenantID := createAdminEntityAndReadID(t, h, http.MethodPost, "/api/admin/v1/tenants", `{"code":"tenant-approval","name":"Tenant Approval"}`, "platform-root")

	definitionID := createAdminEntityAndReadID(
		t,
		h,
		http.MethodPost,
		"/api/admin/v1/approval-definitions",
		`{"tenant_id":"`+tenantID+`","id":"def-a","name":"purchase approval","approver_id":"manager-a","active":true}`,
		tenantID,
	)
	if definitionID != "def-a" {
		t.Fatalf("expected definition id def-a, got %s", definitionID)
	}

	instanceID := createAdminEntityAndReadID(
		t,
		h,
		http.MethodPost,
		"/api/admin/v1/approval-instances",
		`{"tenant_id":"`+tenantID+`","definition_id":"def-a","resource_type":"purchase_order","resource_id":"po-1","requested_by":"user-a"}`,
		tenantID,
	)
	if instanceID == "" {
		t.Fatal("expected approval instance id")
	}

	definitionsReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/approval-definitions?tenant_id="+tenantID, nil)
	definitionsReq.Header.Set("X-Tenant-ID", tenantID)
	definitionsRec := httptest.NewRecorder()
	h.ServeHTTP(definitionsRec, definitionsReq)
	if definitionsRec.Code != http.StatusOK {
		t.Fatalf("expected approval definitions 200, got %d: %s", definitionsRec.Code, definitionsRec.Body.String())
	}
	if !strings.Contains(definitionsRec.Body.String(), "purchase approval") {
		t.Fatalf("expected definitions response to contain purchase approval, got %s", definitionsRec.Body.String())
	}

	tasksReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/approval-tasks?tenant_id="+tenantID, nil)
	tasksReq.Header.Set("X-Tenant-ID", tenantID)
	tasksRec := httptest.NewRecorder()
	h.ServeHTTP(tasksRec, tasksReq)
	if tasksRec.Code != http.StatusOK {
		t.Fatalf("expected approval tasks 200, got %d: %s", tasksRec.Code, tasksRec.Body.String())
	}

	taskID := firstIDFromListResponse(t, tasksRec.Body.Bytes())
	approveReq := httptest.NewRequest(http.MethodPost, "/api/admin/v1/approval-tasks/"+taskID+"/approve", strings.NewReader(`{"actor_id":"manager-a","comment":"approved"}`))
	approveReq.Header.Set("Content-Type", "application/json")
	approveReq.Header.Set("X-Tenant-ID", tenantID)
	approveRec := httptest.NewRecorder()
	h.ServeHTTP(approveRec, approveReq)
	if approveRec.Code != http.StatusCreated {
		t.Fatalf("expected task approve 201, got %d: %s", approveRec.Code, approveRec.Body.String())
	}

	instancesReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/approval-instances?tenant_id="+tenantID, nil)
	instancesReq.Header.Set("X-Tenant-ID", tenantID)
	instancesRec := httptest.NewRecorder()
	h.ServeHTTP(instancesRec, instancesReq)
	if instancesRec.Code != http.StatusOK {
		t.Fatalf("expected approval instances 200, got %d: %s", instancesRec.Code, instancesRec.Body.String())
	}
	if !strings.Contains(instancesRec.Body.String(), instanceID) {
		t.Fatalf("expected instances response to contain %s, got %s", instanceID, instancesRec.Body.String())
	}
	if !strings.Contains(instancesRec.Body.String(), "approved") {
		t.Fatalf("expected instances response to contain approved status, got %s", instancesRec.Body.String())
	}
}

func TestAdminApprovalRejectRoutes(t *testing.T) {
	h := router.New(router.WithContainer(bootstrap.NewTestContainer()))

	tenantID := createAdminEntityAndReadID(t, h, http.MethodPost, "/api/admin/v1/tenants", `{"code":"tenant-reject","name":"Tenant Reject"}`, "platform-root")

	createAdminEntityAndReadID(
		t,
		h,
		http.MethodPost,
		"/api/admin/v1/approval-definitions",
		`{"tenant_id":"`+tenantID+`","id":"def-r","name":"expense approval","approver_id":"manager-r","active":true}`,
		tenantID,
	)

	createAdminEntityAndReadID(
		t,
		h,
		http.MethodPost,
		"/api/admin/v1/approval-instances",
		`{"tenant_id":"`+tenantID+`","definition_id":"def-r","resource_type":"expense_claim","resource_id":"exp-1","requested_by":"user-r"}`,
		tenantID,
	)

	tasksReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/approval-tasks?tenant_id="+tenantID, nil)
	tasksReq.Header.Set("X-Tenant-ID", tenantID)
	tasksRec := httptest.NewRecorder()
	h.ServeHTTP(tasksRec, tasksReq)
	if tasksRec.Code != http.StatusOK {
		t.Fatalf("expected approval tasks 200, got %d: %s", tasksRec.Code, tasksRec.Body.String())
	}

	taskID := firstIDFromListResponse(t, tasksRec.Body.Bytes())
	rejectReq := httptest.NewRequest(http.MethodPost, "/api/admin/v1/approval-tasks/"+taskID+"/reject", strings.NewReader(`{"actor_id":"manager-r","comment":"rejected"}`))
	rejectReq.Header.Set("Content-Type", "application/json")
	rejectReq.Header.Set("X-Tenant-ID", tenantID)
	rejectRec := httptest.NewRecorder()
	h.ServeHTTP(rejectRec, rejectReq)
	if rejectRec.Code != http.StatusCreated {
		t.Fatalf("expected task reject 201, got %d: %s", rejectRec.Code, rejectRec.Body.String())
	}

	instancesReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/approval-instances?tenant_id="+tenantID, nil)
	instancesReq.Header.Set("X-Tenant-ID", tenantID)
	instancesRec := httptest.NewRecorder()
	h.ServeHTTP(instancesRec, instancesReq)
	if instancesRec.Code != http.StatusOK {
		t.Fatalf("expected approval instances 200, got %d: %s", instancesRec.Code, instancesRec.Body.String())
	}
	if !strings.Contains(instancesRec.Body.String(), "rejected") {
		t.Fatalf("expected instances response to contain rejected status, got %s", instancesRec.Body.String())
	}
}

func TestAdminCapabilityCatalogRoutes(t *testing.T) {
	h := router.New(router.WithContainer(bootstrap.NewTestContainer()))

	tenantID := createAdminEntityAndReadID(t, h, http.MethodPost, "/api/admin/v1/tenants", `{"code":"tenant-capability","name":"Tenant Capability"}`, "platform-root")

	modelReq := httptest.NewRequest(http.MethodPost, "/api/admin/v1/model-catalog-entries", strings.NewReader(`{"entry_id":"model-1","model_key":"gpt-5.4","display_name":"GPT 5.4","provider":"openai","status":"active"}`))
	modelReq.Header.Set("Content-Type", "application/json")
	modelReq.Header.Set("X-Tenant-ID", tenantID)
	modelRec := httptest.NewRecorder()
	h.ServeHTTP(modelRec, modelReq)
	if modelRec.Code != http.StatusCreated {
		t.Fatalf("expected model catalog create 201, got %d: %s", modelRec.Code, modelRec.Body.String())
	}

	toolReq := httptest.NewRequest(http.MethodPost, "/api/admin/v1/tool-catalog-entries", strings.NewReader(`{"entry_id":"tool-1","tool_key":"purchase.submit","display_name":"Purchase Submit","risk_level":"medium","status":"active"}`))
	toolReq.Header.Set("Content-Type", "application/json")
	toolReq.Header.Set("X-Tenant-ID", tenantID)
	toolRec := httptest.NewRecorder()
	h.ServeHTTP(toolRec, toolReq)
	if toolRec.Code != http.StatusCreated {
		t.Fatalf("expected tool catalog create 201, got %d: %s", toolRec.Code, toolRec.Body.String())
	}

	modelsReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/model-catalog-entries?tenant_id="+tenantID, nil)
	modelsReq.Header.Set("X-Tenant-ID", tenantID)
	modelsRec := httptest.NewRecorder()
	h.ServeHTTP(modelsRec, modelsReq)
	if modelsRec.Code != http.StatusOK {
		t.Fatalf("expected model catalog list 200, got %d: %s", modelsRec.Code, modelsRec.Body.String())
	}
	if !strings.Contains(modelsRec.Body.String(), "gpt-5.4") {
		t.Fatalf("expected model catalog response to contain gpt-5.4, got %s", modelsRec.Body.String())
	}

	toolsReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/tool-catalog-entries?tenant_id="+tenantID, nil)
	toolsReq.Header.Set("X-Tenant-ID", tenantID)
	toolsRec := httptest.NewRecorder()
	h.ServeHTTP(toolsRec, toolsReq)
	if toolsRec.Code != http.StatusOK {
		t.Fatalf("expected tool catalog list 200, got %d: %s", toolsRec.Code, toolsRec.Body.String())
	}
	if !strings.Contains(toolsRec.Body.String(), "purchase.submit") {
		t.Fatalf("expected tool catalog response to contain purchase.submit, got %s", toolsRec.Body.String())
	}
}

func TestAdminGovernanceRoutes(t *testing.T) {
	container := bootstrap.NewTestContainer()
	if _, err := container.GovernanceCatalog.Append(context.Background(), audit.Record{
		TenantID:    "tenant-gov",
		ID:          "evt-1",
		CommandName: "purchase.submit",
		ActorID:     "actor-a",
		Decision:    policy.DecisionRequireApproval,
		Outcome:     "pending_approval",
		OccurredAt:  time.Date(2026, 3, 26, 10, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("seed audit event: %v", err)
	}

	h := router.New(router.WithContainer(container))

	ruleReq := httptest.NewRequest(http.MethodPost, "/api/admin/v1/policy-rules", strings.NewReader(`{"tenant_id":"tenant-gov","id":"rule-gov","command_name":"purchase.submit","actor_id":"*","decision":"REQUIRE_APPROVAL","priority":80,"active":true}`))
	ruleReq.Header.Set("Content-Type", "application/json")
	ruleReq.Header.Set("X-Tenant-ID", "tenant-gov")
	ruleRec := httptest.NewRecorder()
	h.ServeHTTP(ruleRec, ruleReq)
	if ruleRec.Code != http.StatusCreated {
		t.Fatalf("expected policy rule create 201, got %d: %s", ruleRec.Code, ruleRec.Body.String())
	}

	rulesReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/policy-rules?tenant_id=tenant-gov&command_name=purchase.submit&active_only=true", nil)
	rulesReq.Header.Set("X-Tenant-ID", "tenant-gov")
	rulesRec := httptest.NewRecorder()
	h.ServeHTTP(rulesRec, rulesReq)
	if rulesRec.Code != http.StatusOK {
		t.Fatalf("expected policy rules 200, got %d: %s", rulesRec.Code, rulesRec.Body.String())
	}
	if !strings.Contains(rulesRec.Body.String(), "rule-gov") {
		t.Fatalf("expected rules response to contain rule-gov, got %s", rulesRec.Body.String())
	}

	deactivateReq := httptest.NewRequest(http.MethodPost, "/api/admin/v1/policy-rules/rule-gov/deactivate", strings.NewReader(`{"tenant_id":"tenant-gov"}`))
	deactivateReq.Header.Set("Content-Type", "application/json")
	deactivateReq.Header.Set("X-Tenant-ID", "tenant-gov")
	deactivateRec := httptest.NewRecorder()
	h.ServeHTTP(deactivateRec, deactivateReq)
	if deactivateRec.Code != http.StatusCreated {
		t.Fatalf("expected policy rule deactivate 201, got %d: %s", deactivateRec.Code, deactivateRec.Body.String())
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/audit-events?tenant_id=tenant-gov&command_name=purchase.submit&actor_id=actor-a", nil)
	eventsReq.Header.Set("X-Tenant-ID", "tenant-gov")
	eventsRec := httptest.NewRecorder()
	h.ServeHTTP(eventsRec, eventsReq)
	if eventsRec.Code != http.StatusOK {
		t.Fatalf("expected audit events 200, got %d: %s", eventsRec.Code, eventsRec.Body.String())
	}
	if !strings.Contains(eventsRec.Body.String(), "pending_approval") {
		t.Fatalf("expected audit events response to contain pending_approval, got %s", eventsRec.Body.String())
	}
}

func TestAdminOutboxOperatorRoutes(t *testing.T) {
	container := bootstrap.NewTestContainer()
	outboxCatalog := bootstrap.NewInMemoryOutboxCatalogForTest()
	outboxCatalog.StoreMessage(sharedoutbox.Message{
		ID:          1001,
		TenantID:    "tenant-outbox",
		Topic:       "orders.created",
		EventType:   "orders.created",
		Attempts:    3,
		Status:      "failed",
		LastError:   "nats unavailable",
		AvailableAt: time.Date(2026, 3, 26, 11, 0, 0, 0, time.UTC),
	})
	outboxCatalog.StoreMessage(sharedoutbox.Message{
		ID:          1002,
		TenantID:    "tenant-outbox",
		Topic:       "orders.updated",
		EventType:   "orders.updated",
		Attempts:    1,
		Status:      "pending",
		AvailableAt: time.Date(2026, 3, 26, 11, 1, 0, 0, time.UTC),
	})
	container.OutboxCatalog = outboxCatalog

	h := router.New(router.WithContainer(container))

	listFailedReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/outbox/messages?tenant_id=tenant-outbox&status=failed&limit=10", nil)
	listFailedReq.Header.Set("X-Tenant-ID", "tenant-outbox")
	listFailedRec := httptest.NewRecorder()
	h.ServeHTTP(listFailedRec, listFailedReq)
	if listFailedRec.Code != http.StatusOK {
		t.Fatalf("expected outbox messages 200, got %d: %s", listFailedRec.Code, listFailedRec.Body.String())
	}
	if !strings.Contains(listFailedRec.Body.String(), "nats unavailable") {
		t.Fatalf("expected outbox messages to contain last_error, got %s", listFailedRec.Body.String())
	}

	requeueReq := httptest.NewRequest(http.MethodPost, "/api/admin/v1/outbox/messages/requeue-failed", strings.NewReader(`{"tenant_id":"tenant-outbox","ids":[1001]}`))
	requeueReq.Header.Set("Content-Type", "application/json")
	requeueReq.Header.Set("X-Tenant-ID", "tenant-outbox")
	requeueRec := httptest.NewRecorder()
	h.ServeHTTP(requeueRec, requeueReq)
	if requeueRec.Code != http.StatusCreated {
		t.Fatalf("expected outbox requeue 201, got %d: %s", requeueRec.Code, requeueRec.Body.String())
	}
	if !strings.Contains(requeueRec.Body.String(), `"requeued_count":1`) {
		t.Fatalf("expected outbox requeue count to be 1, got %s", requeueRec.Body.String())
	}

	listPendingReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/outbox/messages?tenant_id=tenant-outbox&status=pending&limit=10", nil)
	listPendingReq.Header.Set("X-Tenant-ID", "tenant-outbox")
	listPendingRec := httptest.NewRecorder()
	h.ServeHTTP(listPendingRec, listPendingReq)
	if listPendingRec.Code != http.StatusOK {
		t.Fatalf("expected pending outbox messages 200, got %d: %s", listPendingRec.Code, listPendingRec.Body.String())
	}
	if !strings.Contains(listPendingRec.Body.String(), `"ID":1001`) {
		t.Fatalf("expected pending outbox messages to contain requeued id, got %s", listPendingRec.Body.String())
	}
}

func TestAdminOutboxRequeueRejectsCrossTenantMessage(t *testing.T) {
	container := bootstrap.NewTestContainer()
	outboxCatalog := bootstrap.NewInMemoryOutboxCatalogForTest()
	outboxCatalog.StoreMessage(sharedoutbox.Message{
		ID:          2001,
		TenantID:    "tenant-a",
		Topic:       "orders.created",
		EventType:   "orders.created",
		Attempts:    2,
		Status:      "failed",
		LastError:   "broker timeout",
		AvailableAt: time.Date(2026, 3, 26, 11, 5, 0, 0, time.UTC),
	})
	container.OutboxCatalog = outboxCatalog

	h := router.New(router.WithContainer(container))

	requeueReq := httptest.NewRequest(http.MethodPost, "/api/admin/v1/outbox/messages/requeue-failed", strings.NewReader(`{"tenant_id":"tenant-b","ids":[2001]}`))
	requeueReq.Header.Set("Content-Type", "application/json")
	requeueReq.Header.Set("X-Tenant-ID", "tenant-b")
	requeueRec := httptest.NewRecorder()
	h.ServeHTTP(requeueRec, requeueReq)
	if requeueRec.Code != http.StatusBadRequest {
		t.Fatalf("expected cross-tenant outbox requeue 400, got %d: %s", requeueRec.Code, requeueRec.Body.String())
	}
}

func createAdminEntityAndReadID(t *testing.T, h http.Handler, method, path, payload, tenantID string) string {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for %s %s, got %d: %s", method, path, rec.Code, rec.Body.String())
	}

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	id, _ := envelope.Data["ID"].(string)
	if id == "" {
		t.Fatalf("expected ID in response, got %+v", envelope.Data)
	}
	return id
}

func firstIDFromListResponse(t *testing.T, payload []byte) string {
	t.Helper()

	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(envelope.Data) == 0 {
		t.Fatalf("expected at least one list item, got %s", string(payload))
	}

	id, _ := envelope.Data[0]["ID"].(string)
	if id == "" {
		t.Fatalf("expected ID in first list item, got %+v", envelope.Data[0])
	}
	return id
}
