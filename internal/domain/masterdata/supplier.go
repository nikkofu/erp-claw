package masterdata

import (
	"errors"
	"strings"
)

var ErrInvalidSupplier = errors.New("invalid supplier")

type Supplier struct {
	ID       string
	TenantID string
	Code     string
	Name     string
}

func NewSupplier(id, tenantID, code, name string) (Supplier, error) {
	supplier := Supplier{
		ID:       strings.TrimSpace(id),
		TenantID: strings.TrimSpace(tenantID),
		Code:     strings.TrimSpace(code),
		Name:     strings.TrimSpace(name),
	}
	if supplier.ID == "" || supplier.TenantID == "" || supplier.Code == "" || supplier.Name == "" {
		return Supplier{}, ErrInvalidSupplier
	}
	return supplier, nil
}
