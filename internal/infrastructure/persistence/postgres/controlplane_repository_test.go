package postgres

import (
	"testing"

	"github.com/nikkofu/erp-claw/internal/domain/controlplane"
)

var (
	_ controlplane.TenantRepository                = (*ControlPlaneRepository)(nil)
	_ controlplane.UserRepository                  = (*ControlPlaneRepository)(nil)
	_ controlplane.RoleRepository                  = (*ControlPlaneRepository)(nil)
	_ controlplane.DepartmentRepository            = (*ControlPlaneRepository)(nil)
	_ controlplane.UserRoleBindingRepository       = (*ControlPlaneRepository)(nil)
	_ controlplane.UserDepartmentBindingRepository = (*ControlPlaneRepository)(nil)
	_ controlplane.AgentProfileRepository          = (*ControlPlaneRepository)(nil)
)

func TestNewControlPlaneRepositoryRejectsNilDB(t *testing.T) {
	_, err := NewControlPlaneRepository(nil)
	if err == nil {
		t.Fatal("expected nil db to fail")
	}
}
