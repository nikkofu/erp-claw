package controlplane

import "context"

// TenantRepository defines persistence operations for tenant catalog entries.
type TenantRepository interface {
	CreateTenant(ctx context.Context, tenant Tenant) (Tenant, error)
	ListTenants(ctx context.Context) ([]Tenant, error)
}

// UserRepository defines persistence operations for IAM users.
type UserRepository interface {
	CreateUser(ctx context.Context, user User) (User, error)
	ListUsers(ctx context.Context, tenantID string) ([]User, error)
}

// RoleRepository defines persistence operations for IAM roles.
type RoleRepository interface {
	CreateRole(ctx context.Context, role Role) (Role, error)
	ListRoles(ctx context.Context, tenantID string) ([]Role, error)
}

// DepartmentRepository defines persistence operations for IAM departments.
type DepartmentRepository interface {
	CreateDepartment(ctx context.Context, department Department) (Department, error)
	ListDepartments(ctx context.Context, tenantID string) ([]Department, error)
}

// UserRoleBindingRepository defines persistence operations for user-role assignments.
type UserRoleBindingRepository interface {
	AssignUserRole(ctx context.Context, binding UserRoleBinding) (UserRoleBinding, error)
}

// UserDepartmentBindingRepository defines persistence operations for user-department assignments.
type UserDepartmentBindingRepository interface {
	AssignUserDepartment(ctx context.Context, binding UserDepartmentBinding) (UserDepartmentBinding, error)
}

// AgentProfileRepository defines persistence operations for agent profile catalog entries.
type AgentProfileRepository interface {
	CreateAgentProfile(ctx context.Context, profile AgentProfile) (AgentProfile, error)
	ListAgentProfiles(ctx context.Context, tenantID string) ([]AgentProfile, error)
}
