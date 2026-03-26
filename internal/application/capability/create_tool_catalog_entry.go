package capability

import (
	"context"
	"errors"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

var ErrToolRepositoryRequired = errors.New("tool catalog repository is required")

type CreateToolCatalogEntryPayload struct {
	ID          string
	ToolKey     string
	DisplayName string
	RiskLevel   string
	Status      string
}

type CreateToolCatalogEntryHandler struct {
	repo domaincap.ToolCatalogRepository
}

func NewCreateToolCatalogEntryHandler(repo domaincap.ToolCatalogRepository) (*CreateToolCatalogEntryHandler, error) {
	if repo == nil {
		return nil, ErrToolRepositoryRequired
	}
	return &CreateToolCatalogEntryHandler{repo: repo}, nil
}

func (h *CreateToolCatalogEntryHandler) Handle(ctx context.Context, tenantID string, payload CreateToolCatalogEntryPayload) error {
	if payload.ToolKey == "" {
		return domaincap.ErrToolKeyRequired
	}

	entry, err := domaincap.NewToolCatalogEntry(
		tenantID,
		payload.ID,
		payload.ToolKey,
		payload.DisplayName,
		payload.RiskLevel,
		payload.Status,
	)
	if err != nil {
		return err
	}

	return h.repo.SaveTool(ctx, entry)
}
