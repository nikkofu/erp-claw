package capability

import (
	"errors"
	"time"
)

var (
	ErrToolKeyRequired = errors.New("tool key is required")
)

type ToolCatalogEntry struct {
	TenantID    string
	EntryID     string
	ToolKey     string
	DisplayName string
	RiskLevel   string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewToolCatalogEntry(tenantID, entryID, toolKey, displayName, riskLevel, status string) (*ToolCatalogEntry, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}
	if entryID == "" {
		return nil, ErrEntryIDRequired
	}
	if toolKey == "" {
		return nil, ErrToolKeyRequired
	}

	now := time.Now().UTC()

	return &ToolCatalogEntry{
		TenantID:    tenantID,
		EntryID:     entryID,
		ToolKey:     toolKey,
		DisplayName: displayName,
		RiskLevel:   riskLevel,
		Status:      normalizeCatalogEntryStatus(status),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (e *ToolCatalogEntry) IsActive() bool {
	if e == nil {
		return false
	}
	return normalizeCatalogEntryStatus(e.Status) == CatalogStatusActive
}

func (e *ToolCatalogEntry) SetStatus(status string) error {
	if e == nil {
		return errors.New("tool catalog entry is required")
	}
	normalized, err := validateCatalogEntryStatus(status)
	if err != nil {
		return err
	}
	e.Status = normalized
	e.UpdatedAt = time.Now().UTC()
	return nil
}
