package capability

import (
	"context"
	"testing"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

func TestCreateToolCatalogEntryHandlerRejectsEmptyToolKey(t *testing.T) {
	t.Parallel()

	repo := &fakeToolCatalogRepository{}
	handler, err := NewCreateToolCatalogEntryHandler(repo)
	if err != nil {
		t.Fatalf("handler init failed: %v", err)
	}

	err = handler.Handle(context.Background(), "tenant-a", CreateToolCatalogEntryPayload{
		ID:          "entry-1",
		ToolKey:     "",
		DisplayName: "Tool",
		RiskLevel:   "high",
		Status:      "active",
	})
	if err != domaincap.ErrToolKeyRequired {
		t.Fatalf("expected tool key error, got %v", err)
	}
	if len(repo.saved) != 0 {
		t.Fatalf("repository should not be called on validation failure")
	}
}

func TestCreateToolCatalogEntryHandlerSavesEntry(t *testing.T) {
	t.Parallel()

	repo := &fakeToolCatalogRepository{}
	handler, err := NewCreateToolCatalogEntryHandler(repo)
	if err != nil {
		t.Fatalf("handler init failed: %v", err)
	}

	payload := CreateToolCatalogEntryPayload{
		ID:          "entry-2",
		ToolKey:     "tool-key",
		DisplayName: "Tool Name",
		RiskLevel:   "high",
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
	if saved.ToolKey != payload.ToolKey {
		t.Fatalf("tool key mismatch: %s", saved.ToolKey)
	}
}

func TestListToolCatalogEntriesHandlerDelegates(t *testing.T) {
	t.Parallel()

	repo := &fakeToolCatalogRepository{
		list: []*domaincap.ToolCatalogEntry{
			{EntryID: "entry-x"},
		},
	}
	handler, err := NewListToolCatalogEntriesHandler(repo)
	if err != nil {
		t.Fatalf("handler init failed: %v", err)
	}

	entries, err := handler.Handle(context.Background(), "tenant-x")
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if repo.lastTenant != "tenant-x" {
		t.Fatalf("tenant not propagated: %s", repo.lastTenant)
	}
}

type fakeToolCatalogRepository struct {
	saved      []*domaincap.ToolCatalogEntry
	list       []*domaincap.ToolCatalogEntry
	lastTenant string
}

func (f *fakeToolCatalogRepository) SaveTool(ctx context.Context, entry *domaincap.ToolCatalogEntry) error {
	f.saved = append(f.saved, entry)
	return nil
}

func (f *fakeToolCatalogRepository) ListToolsByTenant(ctx context.Context, tenantID string) ([]*domaincap.ToolCatalogEntry, error) {
	f.lastTenant = tenantID
	return f.list, nil
}
