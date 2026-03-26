package query

import (
	"context"
	"testing"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

func TestListRolesHandlerUsesTenantScope(t *testing.T) {
	repo := &stubRoleRepository{
		list: []controlplane.Role{{TenantID: "tenant-a", Name: "ops-admin"}},
	}
	handler := ListRolesHandler{Roles: repo}

	roles, err := handler.Handle(context.Background(), ListRoles{TenantID: "tenant-a"})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if repo.lastTenantID != "tenant-a" {
		t.Fatalf("expected tenant scope forwarded, got %q", repo.lastTenantID)
	}
	if len(roles) != 1 || roles[0].Name != "ops-admin" {
		t.Fatalf("unexpected roles: %+v", roles)
	}
}

func TestListDepartmentsHandlerUsesTenantScope(t *testing.T) {
	repo := &stubDepartmentRepository{
		list: []controlplane.Department{{TenantID: "tenant-a", Name: "operations"}},
	}
	handler := ListDepartmentsHandler{Departments: repo}

	departments, err := handler.Handle(context.Background(), ListDepartments{TenantID: "tenant-a"})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if repo.lastTenantID != "tenant-a" {
		t.Fatalf("expected tenant scope forwarded, got %q", repo.lastTenantID)
	}
	if len(departments) != 1 || departments[0].Name != "operations" {
		t.Fatalf("unexpected departments: %+v", departments)
	}
}

type stubRoleRepository struct {
	list         []controlplane.Role
	lastTenantID string
}

func (s *stubRoleRepository) CreateRole(_ context.Context, _ controlplane.Role) (controlplane.Role, error) {
	return controlplane.Role{}, nil
}

func (s *stubRoleRepository) ListRoles(_ context.Context, tenantID string) ([]controlplane.Role, error) {
	s.lastTenantID = tenantID
	return s.list, nil
}

type stubDepartmentRepository struct {
	list         []controlplane.Department
	lastTenantID string
}

func (s *stubDepartmentRepository) CreateDepartment(_ context.Context, _ controlplane.Department) (controlplane.Department, error) {
	return controlplane.Department{}, nil
}

func (s *stubDepartmentRepository) ListDepartments(_ context.Context, tenantID string) ([]controlplane.Department, error) {
	s.lastTenantID = tenantID
	return s.list, nil
}
