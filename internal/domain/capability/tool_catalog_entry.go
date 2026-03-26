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
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
