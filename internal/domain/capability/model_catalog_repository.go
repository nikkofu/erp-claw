package capability

import "context"

type ModelCatalogRepository interface {
	Save(ctx context.Context, entry *ModelCatalogEntry) error
	ListByTenant(ctx context.Context, tenantID string) ([]*ModelCatalogEntry, error)
}
