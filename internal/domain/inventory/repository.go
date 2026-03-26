package inventory

import "context"

type Repository interface {
	SaveReceipt(ctx context.Context, receipt Receipt) error
	AppendLedgerEntries(ctx context.Context, entries []LedgerEntry) error
	ListLedgerEntries(ctx context.Context, tenantID, productID, warehouseID string) ([]LedgerEntry, error)
}
