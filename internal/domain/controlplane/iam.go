package controlplane

import (
	"errors"
	"strings"
)

var (
	errUserTenantIDRequired              = errors.New("user tenant id is required")
	errUserEmailRequired                 = errors.New("user email is required")
	errRoleTenantIDRequired              = errors.New("role tenant id is required")
	errRoleNameRequired                  = errors.New("role name is required")
	errDepartmentTenantIDRequired        = errors.New("department tenant id is required")
	errDepartmentNameRequired            = errors.New("department name is required")
	errUserRoleBindingTenantIDRequired   = errors.New("user role binding tenant id is required")
	errUserRoleBindingUserIDRequired     = errors.New("user role binding user id is required")
	errUserRoleBindingRoleIDRequired     = errors.New("user role binding role id is required")
	errUserDeptBindingTenantIDRequired   = errors.New("user department binding tenant id is required")
	errUserDeptBindingUserIDRequired     = errors.New("user department binding user id is required")
	errUserDeptBindingDepartmentRequired = errors.New("user department binding department id is required")
)

// Organization represents an IAM-scoped business unit in a tenant.
type Organization struct {
	ID       string
	TenantID string
	Name     string
}

// Role represents an IAM role in a tenant.
type Role struct {
	ID          string
	TenantID    string
	Name        string
	Description string
}

// Department represents a tenant-scoped organizational department.
type Department struct {
	ID                 string
	TenantID           string
	Name               string
	ParentDepartmentID string
}

// User is an IAM identity in the tenant catalog.
type User struct {
	ID          string
	TenantID    string
	Email       string
	DisplayName string
}

// UserRoleBinding assigns a user to a role inside one tenant.
type UserRoleBinding struct {
	ID       string
	TenantID string
	UserID   string
	RoleID   string
}

// UserDepartmentBinding assigns a user to a department inside one tenant.
type UserDepartmentBinding struct {
	ID           string
	TenantID     string
	UserID       string
	DepartmentID string
}

// NewUser validates required IAM identity attributes.
func NewUser(tenantID, email, displayName string) (User, error) {
	if strings.TrimSpace(tenantID) == "" {
		return User{}, errUserTenantIDRequired
	}
	if strings.TrimSpace(email) == "" {
		return User{}, errUserEmailRequired
	}

	return User{
		TenantID:    tenantID,
		Email:       email,
		DisplayName: displayName,
	}, nil
}

// NewRole validates required IAM role attributes.
func NewRole(tenantID, name, description string) (Role, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Role{}, errRoleTenantIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return Role{}, errRoleNameRequired
	}

	return Role{
		TenantID:    tenantID,
		Name:        name,
		Description: description,
	}, nil
}

// NewDepartment validates required department attributes.
func NewDepartment(tenantID, name, parentDepartmentID string) (Department, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Department{}, errDepartmentTenantIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return Department{}, errDepartmentNameRequired
	}

	return Department{
		TenantID:           tenantID,
		Name:               name,
		ParentDepartmentID: strings.TrimSpace(parentDepartmentID),
	}, nil
}

// NewUserRoleBinding validates a user-role assignment request.
func NewUserRoleBinding(tenantID, userID, roleID string) (UserRoleBinding, error) {
	if strings.TrimSpace(tenantID) == "" {
		return UserRoleBinding{}, errUserRoleBindingTenantIDRequired
	}
	if strings.TrimSpace(userID) == "" {
		return UserRoleBinding{}, errUserRoleBindingUserIDRequired
	}
	if strings.TrimSpace(roleID) == "" {
		return UserRoleBinding{}, errUserRoleBindingRoleIDRequired
	}

	return UserRoleBinding{
		TenantID: tenantID,
		UserID:   userID,
		RoleID:   roleID,
	}, nil
}

// NewUserDepartmentBinding validates a user-department assignment request.
func NewUserDepartmentBinding(tenantID, userID, departmentID string) (UserDepartmentBinding, error) {
	if strings.TrimSpace(tenantID) == "" {
		return UserDepartmentBinding{}, errUserDeptBindingTenantIDRequired
	}
	if strings.TrimSpace(userID) == "" {
		return UserDepartmentBinding{}, errUserDeptBindingUserIDRequired
	}
	if strings.TrimSpace(departmentID) == "" {
		return UserDepartmentBinding{}, errUserDeptBindingDepartmentRequired
	}

	return UserDepartmentBinding{
		TenantID:     tenantID,
		UserID:       userID,
		DepartmentID: departmentID,
	}, nil
}
