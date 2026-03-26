package capability

import (
	"testing"
	"time"
)

func TestNewModelCatalogEntryValidatesTenantEntryAndModelKey(t *testing.T) {
	t.Parallel()

	_, err := NewModelCatalogEntry("", "entry-1", "model-key", "Model", "provider", "active")
	if err != ErrTenantIDRequired {
		t.Fatalf("expected tenant ID error, got %v", err)
	}

	_, err = NewModelCatalogEntry("tenant", "", "model-key", "Model", "provider", "active")
	if err != ErrEntryIDRequired {
		t.Fatalf("expected entry ID error, got %v", err)
	}

	_, err = NewModelCatalogEntry("tenant", "entry-1", "", "Model", "provider", "active")
	if err != ErrModelKeyRequired {
		t.Fatalf("expected model key error, got %v", err)
	}
}

func TestNewModelCatalogEntryPopulatesFields(t *testing.T) {
	t.Parallel()

	entry, err := NewModelCatalogEntry("tenant", "entry-1", "model-key", "Model Name", "provider", "disabled")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.TenantID != "tenant" || entry.EntryID != "entry-1" || entry.ModelKey != "model-key" {
		t.Fatalf("fields not set correctly: %+v", entry)
	}
	if entry.DisplayName != "Model Name" {
		t.Fatalf("unexpected display name: %s", entry.DisplayName)
	}
	if entry.Provider != "provider" {
		t.Fatalf("unexpected provider: %s", entry.Provider)
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
