package sales

import (
	"context"
	"errors"
)

var ErrOrderNotFound = errors.New("sales order not found")

type Repository interface {
	Save(ctx context.Context, order Order) error
	Get(ctx context.Context, tenantID, orderID string) (Order, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Order, error)
}
