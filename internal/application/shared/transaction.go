package shared

import "context"

// TransactionManager defines the application-level mutation boundary.
type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(context.Context) error) error
}

type noopTransactionManager struct{}

// NoopTransactionManager is used until infrastructure-backed transactions are wired in.
func NoopTransactionManager() TransactionManager {
	return noopTransactionManager{}
}

func (noopTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	if fn == nil {
		return nil
	}
	return fn(ctx)
}
