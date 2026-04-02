package payable

import (
	"errors"
	"strings"
)

var (
	ErrInvalidBill       = errors.New("invalid payable bill")
	ErrBillAlreadyExists = errors.New("payable bill already exists")
	ErrOrderNotBillable  = errors.New("purchase order is not billable")
)

type BillStatus string

const (
	BillStatusOpen BillStatus = "open"
)

type Bill struct {
	ID              string
	TenantID        string
	PurchaseOrderID string
	Status          BillStatus
	CreatedBy       string
}

func NewBill(id, tenantID, purchaseOrderID, createdBy string) (Bill, error) {
	bill := Bill{
		ID:              strings.TrimSpace(id),
		TenantID:        strings.TrimSpace(tenantID),
		PurchaseOrderID: strings.TrimSpace(purchaseOrderID),
		Status:          BillStatusOpen,
		CreatedBy:       strings.TrimSpace(createdBy),
	}
	if bill.ID == "" || bill.TenantID == "" || bill.PurchaseOrderID == "" || bill.CreatedBy == "" {
		return Bill{}, ErrInvalidBill
	}
	return bill, nil
}
