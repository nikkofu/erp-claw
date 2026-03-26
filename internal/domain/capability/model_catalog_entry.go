package capability

import (
	"errors"
	"time"
)

var (
	ErrTenantIDRequired = errors.New("tenant id is required")
	ErrEntryIDRequired  = errors.New("entry id is required")
	ErrModelKeyRequired = errors.New("model key is required")
)

type ModelCatalogEntry struct {
	TenantID    string
	EntryID     string
	ModelKey    string
	DisplayName string
	Provider    string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewModelCatalogEntry(tenantID, entryID, modelKey, displayName, provider, status string) (*ModelCatalogEntry, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}
	if entryID == "" {
		return nil, ErrEntryIDRequired
	}
	if modelKey == "" {
		return nil, ErrModelKeyRequired
	}

	now := time.Now().UTC()

	return &ModelCatalogEntry{
		TenantID:    tenantID,
		EntryID:     entryID,
		ModelKey:    modelKey,
		DisplayName: displayName,
		Provider:    provider,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
