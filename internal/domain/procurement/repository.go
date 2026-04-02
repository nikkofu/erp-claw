package procurement

import (
	"context"
	"errors"
)

var ErrPurchaseOrderNotFound = errors.New("purchase order not found")
var ErrInvalidPurchaseOrderQuery = errors.New("invalid purchase order query")

type Repository interface {
	Save(ctx context.Context, order PurchaseOrder) error
	Get(ctx context.Context, tenantID, orderID string) (PurchaseOrder, error)
	ListByTenant(ctx context.Context, tenantID string) ([]PurchaseOrder, error)
}
