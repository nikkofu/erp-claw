package payable

import (
	"context"
	"errors"
)

var ErrBillNotFound = errors.New("payable bill not found")

type Repository interface {
	Save(ctx context.Context, bill Bill) error
	Get(ctx context.Context, tenantID, billID string) (Bill, error)
	GetByPurchaseOrder(ctx context.Context, tenantID, purchaseOrderID string) (Bill, error)
}
