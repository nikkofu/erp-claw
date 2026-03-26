package command

import (
	"context"
	"testing"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

func TestCreateRoleHandlerRejectsEmptyName(t *testing.T) {
	repo := &stubRoleRepository{}
	handler := CreateRoleHandler{Roles: repo}

	_, err := handler.Handle(context.Background(), CreateRole{
		TenantID:    "tenant-a",
		Name:        "",
		Description: "platform admin",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if len(repo.created) != 0 {
		t.Fatalf("repository should not be called on validation failure")
	}
}

func TestCreateRoleHandlerPersistsRole(t *testing.T) {
	repo := &stubRoleRepository{}
	handler := CreateRoleHandler{Roles: repo}

	role, err := handler.Handle(context.Background(), CreateRole{
		TenantID:    "tenant-a",
		Name:        "ops-admin",
		Description: "platform admin",
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if len(repo.created) != 1 {
		t.Fatalf("expected repository create call, got %d", len(repo.created))
	}
	if role.Name != "ops-admin" || role.TenantID != "tenant-a" {
		t.Fatalf("unexpected role: %+v", role)
	}
}

func TestCreateDepartmentHandlerPersistsDepartment(t *testing.T) {
	repo := &stubDepartmentRepository{}
	handler := CreateDepartmentHandler{Departments: repo}

	department, err := handler.Handle(context.Background(), CreateDepartment{
		TenantID:           "tenant-a",
		Name:               "operations",
		ParentDepartmentID: "",
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if len(repo.created) != 1 {
		t.Fatalf("expected repository create call, got %d", len(repo.created))
	}
	if department.Name != "operations" || department.TenantID != "tenant-a" {
		t.Fatalf("unexpected department: %+v", department)
	}
}

func TestAssignUserRoleHandlerPersistsBinding(t *testing.T) {
	repo := &stubUserRoleBindingRepository{}
	handler := AssignUserRoleHandler{Bindings: repo}

	binding, err := handler.Handle(context.Background(), AssignUserRole{
		TenantID: "tenant-a",
		UserID:   "user-a",
		RoleID:   "role-a",
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if len(repo.created) != 1 {
		t.Fatalf("expected repository create call, got %d", len(repo.created))
	}
	if binding.UserID != "user-a" || binding.RoleID != "role-a" {
		t.Fatalf("unexpected binding: %+v", binding)
	}
}

func TestAssignUserDepartmentHandlerPersistsBinding(t *testing.T) {
	repo := &stubUserDepartmentBindingRepository{}
	handler := AssignUserDepartmentHandler{Bindings: repo}

	binding, err := handler.Handle(context.Background(), AssignUserDepartment{
		TenantID:     "tenant-a",
		UserID:       "user-a",
		DepartmentID: "dept-a",
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if len(repo.created) != 1 {
		t.Fatalf("expected repository create call, got %d", len(repo.created))
	}
	if binding.UserID != "user-a" || binding.DepartmentID != "dept-a" {
		t.Fatalf("unexpected binding: %+v", binding)
	}
}

type stubRoleRepository struct {
	created []controlplane.Role
}

func (s *stubRoleRepository) CreateRole(_ context.Context, role controlplane.Role) (controlplane.Role, error) {
	s.created = append(s.created, role)
	return role, nil
}

func (s *stubRoleRepository) ListRoles(_ context.Context, _ string) ([]controlplane.Role, error) {
	return nil, nil
}

type stubDepartmentRepository struct {
	created []controlplane.Department
}

func (s *stubDepartmentRepository) CreateDepartment(_ context.Context, department controlplane.Department) (controlplane.Department, error) {
	s.created = append(s.created, department)
	return department, nil
}

func (s *stubDepartmentRepository) ListDepartments(_ context.Context, _ string) ([]controlplane.Department, error) {
	return nil, nil
}

type stubUserRoleBindingRepository struct {
	created []controlplane.UserRoleBinding
}

func (s *stubUserRoleBindingRepository) AssignUserRole(_ context.Context, binding controlplane.UserRoleBinding) (controlplane.UserRoleBinding, error) {
	s.created = append(s.created, binding)
	return binding, nil
}

type stubUserDepartmentBindingRepository struct {
	created []controlplane.UserDepartmentBinding
}

func (s *stubUserDepartmentBindingRepository) AssignUserDepartment(_ context.Context, binding controlplane.UserDepartmentBinding) (controlplane.UserDepartmentBinding, error) {
	s.created = append(s.created, binding)
	return binding, nil
}
