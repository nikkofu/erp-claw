package inventory

import (
	"errors"
	"strings"
)

var (
	ErrInvalidTransferOrder       = errors.New("invalid transfer order")
	ErrInvalidTransferOrderQuery  = errors.New("invalid transfer order query")
	ErrTransferOrderNotFound      = errors.New("transfer order not found")
	ErrTransferOrderNotExecutable = errors.New("transfer order not executable")
)

type TransferOrderStatus string

const (
	TransferOrderStatusPlanned  TransferOrderStatus = "planned"
	TransferOrderStatusExecuted TransferOrderStatus = "executed"
)

type TransferOrder struct {
	ID              string
	TenantID        string
	ProductID       string
	FromWarehouseID string
	ToWarehouseID   string
	Quantity        int
	Status          TransferOrderStatus
	CreatedBy       string
	ExecutedBy      string
}

func NewTransferOrder(id, tenantID, productID, fromWarehouseID, toWarehouseID, createdBy string, quantity int) (TransferOrder, error) {
	order := TransferOrder{
		ID:              strings.TrimSpace(id),
		TenantID:        strings.TrimSpace(tenantID),
		ProductID:       strings.TrimSpace(productID),
		FromWarehouseID: strings.TrimSpace(fromWarehouseID),
		ToWarehouseID:   strings.TrimSpace(toWarehouseID),
		Quantity:        quantity,
		Status:          TransferOrderStatusPlanned,
		CreatedBy:       strings.TrimSpace(createdBy),
	}
	if order.ID == "" ||
		order.TenantID == "" ||
		order.ProductID == "" ||
		order.FromWarehouseID == "" ||
		order.ToWarehouseID == "" ||
		order.CreatedBy == "" ||
		order.Quantity <= 0 ||
		order.FromWarehouseID == order.ToWarehouseID {
		return TransferOrder{}, ErrInvalidTransferOrder
	}
	return order, nil
}

func (o *TransferOrder) MarkExecuted(actorID string) error {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" || o.Status != TransferOrderStatusPlanned {
		return ErrTransferOrderNotExecutable
	}
	o.Status = TransferOrderStatusExecuted
	o.ExecutedBy = actorID
	return nil
}
