package procurement

import (
	"context"
	"errors"
)

var ErrPurchaseOrderNotFound = errors.New("purchase order not found")

type Repository interface {
	Save(ctx context.Context, order PurchaseOrder) error
	Get(ctx context.Context, tenantID, orderID string) (PurchaseOrder, error)
}
