package controlplane

import "testing"

func TestNewTenantRejectsEmptyCode(t *testing.T) {
	_, err := NewTenant("", "ERP Claw")
	if err == nil {
		t.Fatal("expected empty tenant code to fail")
	}
}
