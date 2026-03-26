package capability

import (
	"context"
	"testing"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

func TestSetModelCatalogEntryStatusHandlerUpdatesEntry(t *testing.T) {
	t.Parallel()

	repo := &fakeModelCatalogRepository{
		list: []*domaincap.ModelCatalogEntry{
			{TenantID: "tenant-a", EntryID: "model-1", Status: domaincap.CatalogStatusActive},
		},
	}
	handler, err := NewSetModelCatalogEntryStatusHandler(repo)
	if err != nil {
		t.Fatalf("handler init failed: %v", err)
	}

	entry, err := handler.Handle(context.Background(), "tenant-a", "model-1", false)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	if entry.Status != domaincap.CatalogStatusInactive {
		t.Fatalf("expected inactive status, got %s", entry.Status)
	}
	if len(repo.saved) != 1 || repo.saved[0].Status != domaincap.CatalogStatusInactive {
		t.Fatalf("expected repository save with inactive status, got %+v", repo.saved)
	}
}

func TestSetModelCatalogEntryStatusHandlerRejectsMissingEntry(t *testing.T) {
	t.Parallel()

	repo := &fakeModelCatalogRepository{}
	handler, err := NewSetModelCatalogEntryStatusHandler(repo)
	if err != nil {
		t.Fatalf("handler init failed: %v", err)
	}

	if _, err := handler.Handle(context.Background(), "tenant-a", "model-missing", false); err == nil {
		t.Fatal("expected missing model error")
	}
}

func TestSetToolCatalogEntryStatusHandlerUpdatesEntry(t *testing.T) {
	t.Parallel()

	repo := &fakeToolCatalogRepository{
		list: []*domaincap.ToolCatalogEntry{
			{TenantID: "tenant-a", EntryID: "tool-1", Status: domaincap.CatalogStatusActive},
		},
	}
	handler, err := NewSetToolCatalogEntryStatusHandler(repo)
	if err != nil {
		t.Fatalf("handler init failed: %v", err)
	}

	entry, err := handler.Handle(context.Background(), "tenant-a", "tool-1", false)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	if entry.Status != domaincap.CatalogStatusInactive {
		t.Fatalf("expected inactive status, got %s", entry.Status)
	}
	if len(repo.saved) != 1 || repo.saved[0].Status != domaincap.CatalogStatusInactive {
		t.Fatalf("expected repository save with inactive status, got %+v", repo.saved)
	}
}
