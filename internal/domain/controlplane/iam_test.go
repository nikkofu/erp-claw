package controlplane

import "testing"

func TestNewUserRejectsEmptyEmail(t *testing.T) {
	_, err := NewUser("tenant-a", "", "Ada")
	if err == nil {
		t.Fatal("expected empty email to fail")
	}
}

func TestNewRoleRejectsEmptyName(t *testing.T) {
	_, err := NewRole("tenant-a", "", "platform admin")
	if err == nil {
		t.Fatal("expected empty role name to fail")
	}
}

func TestNewDepartmentRejectsEmptyName(t *testing.T) {
	_, err := NewDepartment("tenant-a", "", "")
	if err == nil {
		t.Fatal("expected empty department name to fail")
	}
}

func TestNewUserRoleBindingRejectsEmptyRoleID(t *testing.T) {
	_, err := NewUserRoleBinding("tenant-a", "user-a", "")
	if err == nil {
		t.Fatal("expected empty role id to fail")
	}
}

func TestNewUserDepartmentBindingRejectsEmptyDepartmentID(t *testing.T) {
	_, err := NewUserDepartmentBinding("tenant-a", "user-a", "")
	if err == nil {
		t.Fatal("expected empty department id to fail")
	}
}
