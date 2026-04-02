package approval

import (
	"context"
	"errors"
)

var ErrRequestNotFound = errors.New("approval request not found")
var ErrInvalidRequestQuery = errors.New("invalid approval request query")

type Repository interface {
	Save(ctx context.Context, request Request) error
	Get(ctx context.Context, tenantID, requestID string) (Request, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Request, error)
}
