package capability

import (
	"context"
	"fmt"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

type SetModelCatalogEntryStatusHandler struct {
	repo domaincap.ModelCatalogRepository
}

func NewSetModelCatalogEntryStatusHandler(repo domaincap.ModelCatalogRepository) (*SetModelCatalogEntryStatusHandler, error) {
	if repo == nil {
		return nil, ErrRepositoryRequired
	}
	return &SetModelCatalogEntryStatusHandler{repo: repo}, nil
}

func (h *SetModelCatalogEntryStatusHandler) Handle(ctx context.Context, tenantID, entryID string, active bool) (*domaincap.ModelCatalogEntry, error) {
	entries, err := h.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.EntryID != entryID {
			continue
		}
		status := domaincap.CatalogStatusInactive
		if active {
			status = domaincap.CatalogStatusActive
		}
		if err := entry.SetStatus(status); err != nil {
			return nil, err
		}
		if err := h.repo.Save(ctx, entry); err != nil {
			return nil, err
		}
		return entry, nil
	}

	return nil, fmt.Errorf("model catalog entry %q not found", entryID)
}
