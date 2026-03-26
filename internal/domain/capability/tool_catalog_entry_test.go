package capability

import (
	"testing"
	"time"
)

func TestNewToolCatalogEntryValidatesTenantEntryAndToolKey(t *testing.T) {
	t.Parallel()

	_, err := NewToolCatalogEntry("", "entry-1", "tool-key", "Tool", "high", "active")
	if err != ErrTenantIDRequired {
		t.Fatalf("expected tenant ID error, got %v", err)
	}

	_, err = NewToolCatalogEntry("tenant", "", "tool-key", "Tool", "high", "active")
	if err != ErrEntryIDRequired {
		t.Fatalf("expected entry ID error, got %v", err)
	}

	_, err = NewToolCatalogEntry("tenant", "entry-1", "", "Tool", "high", "active")
	if err != ErrToolKeyRequired {
		t.Fatalf("expected tool key error, got %v", err)
	}
}

func TestNewToolCatalogEntryPopulatesFields(t *testing.T) {
	t.Parallel()

	entry, err := NewToolCatalogEntry("tenant", "entry-1", "tool-key", "Tool Name", "high", "disabled")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.TenantID != "tenant" || entry.EntryID != "entry-1" || entry.ToolKey != "tool-key" {
		t.Fatalf("fields not set correctly: %+v", entry)
	}
	if entry.DisplayName != "Tool Name" {
		t.Fatalf("unexpected display name: %s", entry.DisplayName)
	}
	if entry.RiskLevel != "high" {
		t.Fatalf("unexpected risk level: %s", entry.RiskLevel)
	}
	if entry.Status != "disabled" {
		t.Fatalf("unexpected status: %s", entry.Status)
	}
	if entry.CreatedAt.IsZero() || entry.UpdatedAt.IsZero() {
		t.Fatalf("timestamps should be set")
	}
	if entry.UpdatedAt.Before(entry.CreatedAt) {
		t.Fatalf("updated at should not be before created at")
	}
	if time.Since(entry.CreatedAt) > time.Minute {
		t.Fatalf("created at seems too old: %s", entry.CreatedAt)
	}
}

func TestToolCatalogEntryStatusHelpers(t *testing.T) {
	t.Parallel()

	entry, err := NewToolCatalogEntry("tenant", "entry-1", "tool-key", "Tool Name", "high", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !entry.IsActive() {
		t.Fatalf("expected empty status to normalize to active")
	}

	updatedAt := entry.UpdatedAt
	if err := entry.SetStatus(CatalogStatusInactive); err != nil {
		t.Fatalf("unexpected set status error: %v", err)
	}
	if entry.IsActive() {
		t.Fatalf("expected entry to become inactive")
	}
	if entry.Status != CatalogStatusInactive {
		t.Fatalf("unexpected status: %s", entry.Status)
	}
	if !entry.UpdatedAt.After(updatedAt) {
		t.Fatalf("expected updated at to advance")
	}

	if err := entry.SetStatus("unsupported"); err != ErrCatalogEntryStatusInvalid {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}
