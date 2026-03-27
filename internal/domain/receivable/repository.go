package receivable

import (
	"context"
	"errors"
)

var ErrBillNotFound = errors.New("receivable bill not found")
var ErrInvalidBillQuery = errors.New("invalid receivable bill query")

type Repository interface {
	Save(ctx context.Context, bill Bill) error
	Get(ctx context.Context, tenantID, billID string) (Bill, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Bill, error)
}
