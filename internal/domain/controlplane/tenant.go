package controlplane

import (
	"errors"
	"strings"
)

var (
	errTenantCodeRequired = errors.New("tenant code is required")
)

// Tenant is the catalog root for a control-plane tenant.
type Tenant struct {
	ID   string
	Code string
	Name string
}

// NewTenant validates required catalog fields and returns a tenant aggregate root.
func NewTenant(code, name string) (Tenant, error) {
	if strings.TrimSpace(code) == "" {
		return Tenant{}, errTenantCodeRequired
	}

	return Tenant{
		Code: code,
		Name: name,
	}, nil
}
