package inventory

import "strings"

type MovementType string

const (
	MovementTypeInbound  MovementType = "inbound"
	MovementTypeOutbound MovementType = "outbound"
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
	Reserved    int
	Available   int
}

func NewInboundLedgerEntry(id, tenantID, productID, warehouseID, referenceType, referenceID string, quantity int) (LedgerEntry, error) {
	if quantity <= 0 {
		return LedgerEntry{}, ErrInvalidReceipt
	}
	return newLedgerEntry(id, tenantID, productID, warehouseID, referenceType, referenceID, MovementTypeInbound, quantity)
}

func NewOutboundLedgerEntry(id, tenantID, productID, warehouseID, referenceType, referenceID string, quantity int) (LedgerEntry, error) {
	if quantity <= 0 {
		return LedgerEntry{}, ErrInvalidReceipt
	}
	return newLedgerEntry(id, tenantID, productID, warehouseID, referenceType, referenceID, MovementTypeOutbound, -quantity)
}

func newLedgerEntry(id, tenantID, productID, warehouseID, referenceType, referenceID string, movementType MovementType, quantityDelta int) (LedgerEntry, error) {
	entry := LedgerEntry{
		ID:            strings.TrimSpace(id),
		TenantID:      strings.TrimSpace(tenantID),
		ProductID:     strings.TrimSpace(productID),
		WarehouseID:   strings.TrimSpace(warehouseID),
		MovementType:  movementType,
		QuantityDelta: quantityDelta,
		ReferenceType: strings.TrimSpace(referenceType),
		ReferenceID:   strings.TrimSpace(referenceID),
	}
	if entry.ID == "" ||
		entry.TenantID == "" ||
		entry.ProductID == "" ||
		entry.WarehouseID == "" ||
		entry.ReferenceType == "" ||
		entry.ReferenceID == "" ||
		entry.MovementType == "" ||
		entry.QuantityDelta == 0 {
		return LedgerEntry{}, ErrInvalidReceipt
	}
	return entry, nil
}
