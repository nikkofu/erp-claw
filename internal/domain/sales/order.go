package sales

import (
	"errors"
	"strings"
)

var (
	ErrInvalidOrder      = errors.New("invalid sales order")
	ErrOrderNotShippable = errors.New("sales order is not shippable")
)

type OrderStatus string

const (
	OrderStatusDraft   OrderStatus = "draft"
	OrderStatusShipped OrderStatus = "shipped"
)

type Line struct {
	ProductID string
	Quantity  int
}

type Order struct {
	ID          string
	TenantID    string
	WarehouseID string
	ExternalRef string
	Status      OrderStatus
	CreatedBy   string
	Lines       []Line
}

func NewOrder(id, tenantID, warehouseID, externalRef, createdBy string, lines []Line) (Order, error) {
	order := Order{
		ID:          strings.TrimSpace(id),
		TenantID:    strings.TrimSpace(tenantID),
		WarehouseID: strings.TrimSpace(warehouseID),
		ExternalRef: strings.TrimSpace(externalRef),
		Status:      OrderStatusDraft,
		CreatedBy:   strings.TrimSpace(createdBy),
		Lines:       append([]Line(nil), lines...),
	}
	if order.ID == "" || order.TenantID == "" || order.WarehouseID == "" || order.ExternalRef == "" || order.CreatedBy == "" {
		return Order{}, ErrInvalidOrder
	}
	if err := validateLines(order.Lines); err != nil {
		return Order{}, err
	}
	return order, nil
}

func (o *Order) MarkShipped() error {
	if o.Status != OrderStatusDraft {
		return ErrOrderNotShippable
	}
	o.Status = OrderStatusShipped
	return nil
}

func validateLines(lines []Line) error {
	if len(lines) == 0 {
		return ErrInvalidOrder
	}
	for _, line := range lines {
		if strings.TrimSpace(line.ProductID) == "" || line.Quantity <= 0 {
			return ErrInvalidOrder
		}
	}
	return nil
}
