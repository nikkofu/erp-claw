package inventory

import (
	"errors"
	"strings"
)

var (
	ErrInvalidReservation             = errors.New("invalid inventory reservation")
	ErrInsufficientAvailableInventory = errors.New("insufficient available inventory")
)

type ReservationStatus string

const (
	ReservationStatusActive ReservationStatus = "active"
)

type Reservation struct {
	ID            string
	TenantID      string
	ProductID     string
	WarehouseID   string
	ReferenceType string
	ReferenceID   string
	Status        ReservationStatus
	CreatedBy     string
	Quantity      int
}

func NewReservation(id, tenantID, productID, warehouseID, referenceType, referenceID, createdBy string, quantity int) (Reservation, error) {
	reservation := Reservation{
		ID:            strings.TrimSpace(id),
		TenantID:      strings.TrimSpace(tenantID),
		ProductID:     strings.TrimSpace(productID),
		WarehouseID:   strings.TrimSpace(warehouseID),
		ReferenceType: strings.TrimSpace(referenceType),
		ReferenceID:   strings.TrimSpace(referenceID),
		Status:        ReservationStatusActive,
		CreatedBy:     strings.TrimSpace(createdBy),
		Quantity:      quantity,
	}
	if reservation.ID == "" ||
		reservation.TenantID == "" ||
		reservation.ProductID == "" ||
		reservation.WarehouseID == "" ||
		reservation.ReferenceType == "" ||
		reservation.ReferenceID == "" ||
		reservation.CreatedBy == "" ||
		reservation.Quantity <= 0 {
		return Reservation{}, ErrInvalidReservation
	}
	return reservation, nil
}
