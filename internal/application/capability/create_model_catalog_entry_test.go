package capability

import (
	"context"
	"testing"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

func TestCreateModelCatalogEntryHandlerRejectsEmptyModelKey(t *testing.T) {
	t.Parallel()

	repo := &fakeModelCatalogRepository{}
	handler, err := NewCreateModelCatalogEntryHandler(repo)
	if err != nil {
		t.Fatalf("handler init failed: %v", err)
	}

	err = handler.Handle(context.Background(), "tenant-a", CreateModelCatalogEntryPayload{
		ID:          "entry-1",
		ModelKey:    "",
		DisplayName: "Model",
		Provider:    "provider",
		Status:      "active",
	})
	if err != domaincap.ErrModelKeyRequired {
		t.Fatalf("expected model key error, got %v", err)
	}
	if len(repo.saved) != 0 {
		t.Fatalf("repository should not be called on validation failure")
	}
}

func TestCreateModelCatalogEntryHandlerSavesEntry(t *testing.T) {
	t.Parallel()

	repo := &fakeModelCatalogRepository{}
	handler, err := NewCreateModelCatalogEntryHandler(repo)
	if err != nil {
		t.Fatalf("handler init failed: %v", err)
	}

	payload := CreateModelCatalogEntryPayload{
		ID:          "entry-2",
		ModelKey:    "model-key",
		DisplayName: "Model Name",
		Provider:    "provider",
		Status:      "active",
	}
	if err := handler.Handle(context.Background(), "tenant-b", payload); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if len(repo.saved) != 1 {
		t.Fatalf("expected repository save called once, got %d", len(repo.saved))
	}
	saved := repo.saved[0]
	if saved.EntryID != "entry-2" || saved.TenantID != "tenant-b" {
		t.Fatalf("saved entry missing identifiers: %+v", saved)
	}
	if saved.ModelKey != payload.ModelKey {
		t.Fatalf("model key mismatch: %s", saved.ModelKey)
	}
}
