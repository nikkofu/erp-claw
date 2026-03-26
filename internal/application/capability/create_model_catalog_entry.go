package capability

import (
	"context"
	"errors"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

var ErrRepositoryRequired = errors.New("model catalog repository is required")

type CreateModelCatalogEntryPayload struct {
	ID          string
	ModelKey    string
	DisplayName string
	Provider    string
	Status      string
}

type CreateModelCatalogEntryHandler struct {
	repo domaincap.ModelCatalogRepository
}

func NewCreateModelCatalogEntryHandler(repo domaincap.ModelCatalogRepository) (*CreateModelCatalogEntryHandler, error) {
	if repo == nil {
		return nil, ErrRepositoryRequired
	}
	return &CreateModelCatalogEntryHandler{repo: repo}, nil
}

func (h *CreateModelCatalogEntryHandler) Handle(ctx context.Context, tenantID string, payload CreateModelCatalogEntryPayload) error {
	if payload.ModelKey == "" {
		return domaincap.ErrModelKeyRequired
	}

	entry, err := domaincap.NewModelCatalogEntry(
		tenantID,
		payload.ID,
		payload.ModelKey,
		payload.DisplayName,
		payload.Provider,
		payload.Status,
	)
	if err != nil {
		return err
	}

	return h.repo.Save(ctx, entry)
}
