package capability

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrTenantIDRequired = errors.New("tenant id is required")
	ErrEntryIDRequired  = errors.New("entry id is required")
	ErrModelKeyRequired = errors.New("model key is required")
	ErrCatalogEntryStatusInvalid = errors.New("catalog entry status must be active or inactive")
)

const (
	CatalogStatusActive   = "active"
	CatalogStatusInactive = "inactive"
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
		Status:      normalizeCatalogEntryStatus(status),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (e *ModelCatalogEntry) IsActive() bool {
	if e == nil {
		return false
	}
	return normalizeCatalogEntryStatus(e.Status) == CatalogStatusActive
}

func (e *ModelCatalogEntry) SetStatus(status string) error {
	if e == nil {
		return errors.New("model catalog entry is required")
	}
	normalized, err := validateCatalogEntryStatus(status)
	if err != nil {
		return err
	}
	e.Status = normalized
	e.UpdatedAt = time.Now().UTC()
	return nil
}

func normalizeCatalogEntryStatus(status string) string {
	trimmed := strings.ToLower(strings.TrimSpace(status))
	if trimmed == "" {
		return CatalogStatusActive
	}
	return trimmed
}

func validateCatalogEntryStatus(status string) (string, error) {
	normalized := normalizeCatalogEntryStatus(status)
	switch normalized {
	case CatalogStatusActive, CatalogStatusInactive:
		return normalized, nil
	default:
		return "", ErrCatalogEntryStatusInvalid
	}
}
