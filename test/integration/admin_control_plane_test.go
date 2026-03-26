package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	controlcommand "github.com/nikkofu/erp-claw/internal/application/controlplane/command"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
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
