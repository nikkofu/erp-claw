package capability

import (
	"context"
	"fmt"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

type SetToolCatalogEntryStatusHandler struct {
	repo domaincap.ToolCatalogRepository
}

func NewSetToolCatalogEntryStatusHandler(repo domaincap.ToolCatalogRepository) (*SetToolCatalogEntryStatusHandler, error) {
	if repo == nil {
		return nil, ErrToolRepositoryRequired
	}
	return &SetToolCatalogEntryStatusHandler{repo: repo}, nil
}

func (h *SetToolCatalogEntryStatusHandler) Handle(ctx context.Context, tenantID, entryID string, active bool) (*domaincap.ToolCatalogEntry, error) {
	entries, err := h.repo.ListToolsByTenant(ctx, tenantID)
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
		if err := h.repo.SaveTool(ctx, entry); err != nil {
			return nil, err
		}
		return entry, nil
	}

	return nil, fmt.Errorf("tool catalog entry %q not found", entryID)
}
