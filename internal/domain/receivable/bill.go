package receivable

import (
	"errors"
	"strings"
)

var (
	ErrInvalidBill = errors.New("invalid receivable bill")
)

type BillStatus string

const (
	BillStatusOpen BillStatus = "open"
)

type Bill struct {
	ID          string
	TenantID    string
	ExternalRef string
	Status      BillStatus
	CreatedBy   string
}

func NewBill(id, tenantID, externalRef, createdBy string) (Bill, error) {
	bill := Bill{
		ID:          strings.TrimSpace(id),
		TenantID:    strings.TrimSpace(tenantID),
		ExternalRef: strings.TrimSpace(externalRef),
		Status:      BillStatusOpen,
		CreatedBy:   strings.TrimSpace(createdBy),
	}
	if bill.ID == "" || bill.TenantID == "" || bill.ExternalRef == "" || bill.CreatedBy == "" {
		return Bill{}, ErrInvalidBill
	}
	return bill, nil
}
