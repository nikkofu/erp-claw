package capability

import (
	"context"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

type ListModelCatalogEntriesHandler struct {
	repo domaincap.ModelCatalogRepository
}

func NewListModelCatalogEntriesHandler(repo domaincap.ModelCatalogRepository) (*ListModelCatalogEntriesHandler, error) {
	if repo == nil {
		return nil, ErrRepositoryRequired
	}
	return &ListModelCatalogEntriesHandler{repo: repo}, nil
}

func (h *ListModelCatalogEntriesHandler) Handle(ctx context.Context, tenantID string) ([]*domaincap.ModelCatalogEntry, error) {
	return h.repo.ListByTenant(ctx, tenantID)
}
