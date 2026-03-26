package masterdata

import (
	"errors"
	"strings"
)

var ErrInvalidWarehouse = errors.New("invalid warehouse")

type Warehouse struct {
	ID       string
	TenantID string
	Code     string
	Name     string
}

func NewWarehouse(id, tenantID, code, name string) (Warehouse, error) {
	warehouse := Warehouse{
		ID:       strings.TrimSpace(id),
		TenantID: strings.TrimSpace(tenantID),
		Code:     strings.TrimSpace(code),
		Name:     strings.TrimSpace(name),
	}
	if warehouse.ID == "" || warehouse.TenantID == "" || warehouse.Code == "" || warehouse.Name == "" {
		return Warehouse{}, ErrInvalidWarehouse
	}
	return warehouse, nil
}
