package capability

import (
	"context"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

type ListToolCatalogEntriesHandler struct {
	repo domaincap.ToolCatalogRepository
}

func NewListToolCatalogEntriesHandler(repo domaincap.ToolCatalogRepository) (*ListToolCatalogEntriesHandler, error) {
	if repo == nil {
		return nil, ErrToolRepositoryRequired
	}
	return &ListToolCatalogEntriesHandler{repo: repo}, nil
}

func (h *ListToolCatalogEntriesHandler) Handle(ctx context.Context, tenantID string) ([]*domaincap.ToolCatalogEntry, error) {
	return h.repo.ListToolsByTenant(ctx, tenantID)
}
