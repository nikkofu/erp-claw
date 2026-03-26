package inventory

import (
	"errors"
	"strings"
)

var ErrInvalidReceipt = errors.New("invalid receipt")

type ReceiptStatus string

const (
	ReceiptStatusPosted ReceiptStatus = "posted"
)

type ReceiptLine struct {
	ProductID string
	Quantity  int
}

type Receipt struct {
	ID              string
	TenantID        string
	PurchaseOrderID string
	WarehouseID     string
	Status          ReceiptStatus
	CreatedBy       string
	Lines           []ReceiptLine
}

func NewReceipt(id, tenantID, purchaseOrderID, warehouseID, createdBy string, lines []ReceiptLine) (Receipt, error) {
	receipt := Receipt{
		ID:              strings.TrimSpace(id),
		TenantID:        strings.TrimSpace(tenantID),
		PurchaseOrderID: strings.TrimSpace(purchaseOrderID),
		WarehouseID:     strings.TrimSpace(warehouseID),
		Status:          ReceiptStatusPosted,
		CreatedBy:       strings.TrimSpace(createdBy),
		Lines:           append([]ReceiptLine(nil), lines...),
	}
	if receipt.ID == "" || receipt.TenantID == "" || receipt.PurchaseOrderID == "" || receipt.WarehouseID == "" || receipt.CreatedBy == "" {
		return Receipt{}, ErrInvalidReceipt
	}
	if err := validateReceiptLines(receipt.Lines); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func validateReceiptLines(lines []ReceiptLine) error {
	if len(lines) == 0 {
		return ErrInvalidReceipt
	}
	for _, line := range lines {
		if strings.TrimSpace(line.ProductID) == "" || line.Quantity <= 0 {
			return ErrInvalidReceipt
		}
	}
	return nil
}
