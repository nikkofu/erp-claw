package inventory

import "strings"

type MovementType string

const (
	MovementTypeInbound MovementType = "inbound"
)

type LedgerEntry struct {
	ID            string
	TenantID      string
	ProductID     string
	WarehouseID   string
	MovementType  MovementType
	QuantityDelta int
	ReferenceType string
	ReferenceID   string
}

type Balance struct {
	TenantID    string
	ProductID   string
	WarehouseID string
	OnHand      int
}

func NewInboundLedgerEntry(id, tenantID, productID, warehouseID, referenceType, referenceID string, quantity int) (LedgerEntry, error) {
	entry := LedgerEntry{
		ID:            strings.TrimSpace(id),
		TenantID:      strings.TrimSpace(tenantID),
		ProductID:     strings.TrimSpace(productID),
		WarehouseID:   strings.TrimSpace(warehouseID),
		MovementType:  MovementTypeInbound,
		QuantityDelta: quantity,
		ReferenceType: strings.TrimSpace(referenceType),
		ReferenceID:   strings.TrimSpace(referenceID),
	}
	if entry.ID == "" || entry.TenantID == "" || entry.ProductID == "" || entry.WarehouseID == "" || entry.ReferenceType == "" || entry.ReferenceID == "" || entry.QuantityDelta <= 0 {
		return LedgerEntry{}, ErrInvalidReceipt
	}
	return entry, nil
}
