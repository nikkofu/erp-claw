package approval

import (
	"context"
	"errors"
)

var ErrRequestNotFound = errors.New("approval request not found")

type Repository interface {
	Save(ctx context.Context, request Request) error
	Get(ctx context.Context, tenantID, requestID string) (Request, error)
}
