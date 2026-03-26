package capability

import (
	"context"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

type fakeModelCatalogRepository struct {
	saved      []*domaincap.ModelCatalogEntry
	list       []*domaincap.ModelCatalogEntry
	lastTenant string
}

func (f *fakeModelCatalogRepository) Save(ctx context.Context, entry *domaincap.ModelCatalogEntry) error {
	f.saved = append(f.saved, entry)
	return nil
}

func (f *fakeModelCatalogRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domaincap.ModelCatalogEntry, error) {
	f.lastTenant = tenantID
	return f.list, nil
}
