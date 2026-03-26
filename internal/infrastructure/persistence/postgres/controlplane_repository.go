package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var errControlPlaneRepositoryNilDB = errors.New("postgres control-plane repository requires non-nil db")

type ControlPlaneRepository struct {
	db *sql.DB
}

func NewControlPlaneRepository(db *sql.DB) (*ControlPlaneRepository, error) {
	if db == nil {
		return nil, errControlPlaneRepositoryNilDB
	}

	return &ControlPlaneRepository{db: db}, nil
}

func (r *ControlPlaneRepository) CreateTenant(ctx context.Context, tenant controlplane.Tenant) (controlplane.Tenant, error) {
	if strings.TrimSpace(tenant.ID) == "" {
		var dbID int64
		if err := r.db.QueryRowContext(
			ctx,
			`insert into tenant(code, name) values ($1, $2) returning id`,
			tenant.Code,
			tenant.Name,
		).Scan(&dbID); err != nil {
			return controlplane.Tenant{}, err
		}

		tenant.ID = strconv.FormatInt(dbID, 10)
		return tenant, nil
	}

	tenantID, err := strconv.ParseInt(tenant.ID, 10, 64)
	if err != nil {
		return controlplane.Tenant{}, err
	}

	var dbID int64
	if err := r.db.QueryRowContext(
		ctx,
		`insert into tenant(id, code, name) values ($1, $2, $3) returning id`,
		tenantID,
		tenant.Code,
		tenant.Name,
	).Scan(&dbID); err != nil {
		return controlplane.Tenant{}, err
	}

	tenant.ID = strconv.FormatInt(dbID, 10)
	return tenant, nil
}

func (r *ControlPlaneRepository) ListTenants(ctx context.Context) ([]controlplane.Tenant, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`select id, code, name from tenant order by code asc, id asc`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tenants := make([]controlplane.Tenant, 0)
	for rows.Next() {
		var dbID int64
		var tenant controlplane.Tenant
		if err := rows.Scan(&dbID, &tenant.Code, &tenant.Name); err != nil {
			return nil, err
		}
		tenant.ID = strconv.FormatInt(dbID, 10)
		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tenants, nil
}

func (r *ControlPlaneRepository) CreateUser(ctx context.Context, user controlplane.User) (controlplane.User, error) {
	if err := r.ensureTenantExists(ctx, user.TenantID); err != nil {
		return controlplane.User{}, err
	}

	if strings.TrimSpace(user.ID) == "" {
		user.ID = uuid.NewString()
	}

	if err := r.db.QueryRowContext(
		ctx,
		`insert into iam_user(tenant_id, id, email, display_name)
		 values ($1, $2, $3, $4)
		 returning id`,
		user.TenantID,
		user.ID,
		user.Email,
		user.DisplayName,
	).Scan(&user.ID); err != nil {
		return controlplane.User{}, err
	}

	return user, nil
}

func (r *ControlPlaneRepository) ListUsers(ctx context.Context, tenantID string) ([]controlplane.User, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`select id, email, display_name
		 from iam_user
		 where tenant_id = $1
		 order by display_name asc, id asc`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]controlplane.User, 0)
	for rows.Next() {
		user := controlplane.User{TenantID: tenantID}
		if err := rows.Scan(&user.ID, &user.Email, &user.DisplayName); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *ControlPlaneRepository) CreateRole(ctx context.Context, role controlplane.Role) (controlplane.Role, error) {
	if err := r.ensureTenantExists(ctx, role.TenantID); err != nil {
		return controlplane.Role{}, err
	}

	if strings.TrimSpace(role.ID) == "" {
		role.ID = uuid.NewString()
	}

	if err := r.db.QueryRowContext(
		ctx,
		`insert into iam_role(tenant_id, id, name, description)
		 values ($1, $2, $3, $4)
		 returning id`,
		role.TenantID,
		role.ID,
		role.Name,
		role.Description,
	).Scan(&role.ID); err != nil {
		return controlplane.Role{}, err
	}

	return role, nil
}

func (r *ControlPlaneRepository) ListRoles(ctx context.Context, tenantID string) ([]controlplane.Role, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`select id, name, description
		 from iam_role
		 where tenant_id = $1
		 order by name asc, id asc`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]controlplane.Role, 0)
	for rows.Next() {
		role := controlplane.Role{TenantID: tenantID}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *ControlPlaneRepository) CreateDepartment(ctx context.Context, department controlplane.Department) (controlplane.Department, error) {
	if err := r.ensureTenantExists(ctx, department.TenantID); err != nil {
		return controlplane.Department{}, err
	}

	if strings.TrimSpace(department.ID) == "" {
		department.ID = uuid.NewString()
	}

	if err := r.db.QueryRowContext(
		ctx,
		`insert into iam_department(tenant_id, id, name, parent_department_id)
		 values ($1, $2, $3, nullif($4, ''))
		 returning id`,
		department.TenantID,
		department.ID,
		department.Name,
		department.ParentDepartmentID,
	).Scan(&department.ID); err != nil {
		return controlplane.Department{}, err
	}

	return department, nil
}

func (r *ControlPlaneRepository) ListDepartments(ctx context.Context, tenantID string) ([]controlplane.Department, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`select id, name, coalesce(parent_department_id, '')
		 from iam_department
		 where tenant_id = $1
		 order by name asc, id asc`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	departments := make([]controlplane.Department, 0)
	for rows.Next() {
		department := controlplane.Department{TenantID: tenantID}
		if err := rows.Scan(&department.ID, &department.Name, &department.ParentDepartmentID); err != nil {
			return nil, err
		}
		departments = append(departments, department)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return departments, nil
}

func (r *ControlPlaneRepository) AssignUserRole(ctx context.Context, binding controlplane.UserRoleBinding) (controlplane.UserRoleBinding, error) {
	if err := r.ensureTenantExists(ctx, binding.TenantID); err != nil {
		return controlplane.UserRoleBinding{}, err
	}

	if strings.TrimSpace(binding.ID) == "" {
		binding.ID = uuid.NewString()
	}

	if err := r.db.QueryRowContext(
		ctx,
		`insert into iam_user_role(tenant_id, id, user_id, role_id)
		 values ($1, $2, $3, $4)
		 returning id`,
		binding.TenantID,
		binding.ID,
		binding.UserID,
		binding.RoleID,
	).Scan(&binding.ID); err != nil {
		return controlplane.UserRoleBinding{}, err
	}

	return binding, nil
}

func (r *ControlPlaneRepository) AssignUserDepartment(ctx context.Context, binding controlplane.UserDepartmentBinding) (controlplane.UserDepartmentBinding, error) {
	if err := r.ensureTenantExists(ctx, binding.TenantID); err != nil {
		return controlplane.UserDepartmentBinding{}, err
	}

	if strings.TrimSpace(binding.ID) == "" {
		binding.ID = uuid.NewString()
	}

	if err := r.db.QueryRowContext(
		ctx,
		`insert into iam_user_department(tenant_id, id, user_id, department_id)
		 values ($1, $2, $3, $4)
		 returning id`,
		binding.TenantID,
		binding.ID,
		binding.UserID,
		binding.DepartmentID,
	).Scan(&binding.ID); err != nil {
		return controlplane.UserDepartmentBinding{}, err
	}

	return binding, nil
}

func (r *ControlPlaneRepository) CreateAgentProfile(ctx context.Context, profile controlplane.AgentProfile) (controlplane.AgentProfile, error) {
	if err := r.ensureTenantExists(ctx, profile.TenantID); err != nil {
		return controlplane.AgentProfile{}, err
	}

	if strings.TrimSpace(profile.ID) == "" {
		profile.ID = uuid.NewString()
	}

	if err := r.db.QueryRowContext(
		ctx,
		`insert into agent_profile(tenant_id, id, name, model)
		 values ($1, $2, $3, $4)
		 returning id`,
		profile.TenantID,
		profile.ID,
		profile.Name,
		profile.Model,
	).Scan(&profile.ID); err != nil {
		return controlplane.AgentProfile{}, err
	}

	return profile, nil
}

func (r *ControlPlaneRepository) ListAgentProfiles(ctx context.Context, tenantID string) ([]controlplane.AgentProfile, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`select id, name, model
		 from agent_profile
		 where tenant_id = $1
		 order by name asc, id asc`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make([]controlplane.AgentProfile, 0)
	for rows.Next() {
		profile := controlplane.AgentProfile{TenantID: tenantID}
		if err := rows.Scan(&profile.ID, &profile.Name, &profile.Model); err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}

func (r *ControlPlaneRepository) ensureTenantExists(ctx context.Context, tenantID string) error {
	parsedTenantID, err := strconv.ParseInt(strings.TrimSpace(tenantID), 10, 64)
	if err != nil {
		return err
	}

	var exists int
	if err := r.db.QueryRowContext(
		ctx,
		`select 1 from tenant where id = $1`,
		parsedTenantID,
	).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return controlplane.ErrTenantNotFound
		}
		return err
	}

	return nil
}
