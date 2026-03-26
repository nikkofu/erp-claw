package bootstrap

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/postgres"
)

func newControlPlaneCatalog(cfg Config) ControlPlaneCatalog {
	catalog, err := newPostgresControlPlaneCatalog(cfg.Database)
	if err == nil {
		return catalog
	}

	return newInMemoryControlPlaneCatalog()
}

func newPostgresControlPlaneCatalog(cfg DatabaseConfig) (ControlPlaneCatalog, error) {
	db, err := postgres.New(postgres.Config{
		DSN:          cfg.DSN,
		MaxOpenConns: cfg.MaxOpenConns,
		MaxIdleConns: cfg.MaxIdleConns,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	repo, err := postgres.NewControlPlaneRepository(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return repo, nil
}

type inMemoryControlPlaneCatalog struct {
	mu               sync.RWMutex
	nextTenantID     int64
	nextUserID       int64
	nextRoleID       int64
	nextDepartmentID int64
	nextUserRoleID   int64
	nextUserDeptID   int64
	nextProfileID    int64
	tenants          []controlplane.Tenant
	usersByTenant    map[string][]controlplane.User
	rolesByTenant    map[string][]controlplane.Role
	deptsByTenant    map[string][]controlplane.Department
	userRoles        map[string][]controlplane.UserRoleBinding
	userDepts        map[string][]controlplane.UserDepartmentBinding
	profilesByTenant map[string][]controlplane.AgentProfile
}

func newInMemoryControlPlaneCatalog() *inMemoryControlPlaneCatalog {
	return &inMemoryControlPlaneCatalog{
		tenants:          make([]controlplane.Tenant, 0),
		usersByTenant:    make(map[string][]controlplane.User),
		rolesByTenant:    make(map[string][]controlplane.Role),
		deptsByTenant:    make(map[string][]controlplane.Department),
		userRoles:        make(map[string][]controlplane.UserRoleBinding),
		userDepts:        make(map[string][]controlplane.UserDepartmentBinding),
		profilesByTenant: make(map[string][]controlplane.AgentProfile),
	}
}

func (r *inMemoryControlPlaneCatalog) CreateTenant(_ context.Context, tenant controlplane.Tenant) (controlplane.Tenant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(tenant.ID) == "" {
		r.nextTenantID++
		tenant.ID = strconv.FormatInt(r.nextTenantID, 10)
	}

	r.tenants = append(r.tenants, tenant)
	return tenant, nil
}

func (r *inMemoryControlPlaneCatalog) ListTenants(_ context.Context) ([]controlplane.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenants := make([]controlplane.Tenant, len(r.tenants))
	copy(tenants, r.tenants)
	return tenants, nil
}

func (r *inMemoryControlPlaneCatalog) CreateUser(_ context.Context, user controlplane.User) (controlplane.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(user.ID) == "" {
		r.nextUserID++
		user.ID = "user-" + strconv.FormatInt(r.nextUserID, 10)
	}

	r.usersByTenant[user.TenantID] = append(r.usersByTenant[user.TenantID], user)
	return user, nil
}

func (r *inMemoryControlPlaneCatalog) ListUsers(_ context.Context, tenantID string) ([]controlplane.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := r.usersByTenant[tenantID]
	list := make([]controlplane.User, len(users))
	copy(list, users)
	return list, nil
}

func (r *inMemoryControlPlaneCatalog) CreateRole(_ context.Context, role controlplane.Role) (controlplane.Role, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(role.ID) == "" {
		r.nextRoleID++
		role.ID = "role-" + strconv.FormatInt(r.nextRoleID, 10)
	}

	r.rolesByTenant[role.TenantID] = append(r.rolesByTenant[role.TenantID], role)
	return role, nil
}

func (r *inMemoryControlPlaneCatalog) ListRoles(_ context.Context, tenantID string) ([]controlplane.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roles := r.rolesByTenant[tenantID]
	list := make([]controlplane.Role, len(roles))
	copy(list, roles)
	return list, nil
}

func (r *inMemoryControlPlaneCatalog) CreateDepartment(_ context.Context, department controlplane.Department) (controlplane.Department, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(department.ID) == "" {
		r.nextDepartmentID++
		department.ID = "department-" + strconv.FormatInt(r.nextDepartmentID, 10)
	}

	r.deptsByTenant[department.TenantID] = append(r.deptsByTenant[department.TenantID], department)
	return department, nil
}

func (r *inMemoryControlPlaneCatalog) ListDepartments(_ context.Context, tenantID string) ([]controlplane.Department, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	departments := r.deptsByTenant[tenantID]
	list := make([]controlplane.Department, len(departments))
	copy(list, departments)
	return list, nil
}

func (r *inMemoryControlPlaneCatalog) AssignUserRole(_ context.Context, binding controlplane.UserRoleBinding) (controlplane.UserRoleBinding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(binding.ID) == "" {
		r.nextUserRoleID++
		binding.ID = "user-role-" + strconv.FormatInt(r.nextUserRoleID, 10)
	}

	r.userRoles[binding.TenantID] = append(r.userRoles[binding.TenantID], binding)
	return binding, nil
}

func (r *inMemoryControlPlaneCatalog) AssignUserDepartment(_ context.Context, binding controlplane.UserDepartmentBinding) (controlplane.UserDepartmentBinding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(binding.ID) == "" {
		r.nextUserDeptID++
		binding.ID = "user-department-" + strconv.FormatInt(r.nextUserDeptID, 10)
	}

	r.userDepts[binding.TenantID] = append(r.userDepts[binding.TenantID], binding)
	return binding, nil
}

func (r *inMemoryControlPlaneCatalog) CreateAgentProfile(_ context.Context, profile controlplane.AgentProfile) (controlplane.AgentProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(profile.ID) == "" {
		r.nextProfileID++
		profile.ID = "profile-" + strconv.FormatInt(r.nextProfileID, 10)
	}

	r.profilesByTenant[profile.TenantID] = append(r.profilesByTenant[profile.TenantID], profile)
	return profile, nil
}

func (r *inMemoryControlPlaneCatalog) ListAgentProfiles(_ context.Context, tenantID string) ([]controlplane.AgentProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	profiles := r.profilesByTenant[tenantID]
	list := make([]controlplane.AgentProfile, len(profiles))
	copy(list, profiles)
	return list, nil
}
