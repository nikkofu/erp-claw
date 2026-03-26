package capability

import (
	"context"
	"testing"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

func TestListModelCatalogEntriesHandlerDelegates(t *testing.T) {
	t.Parallel()

	repo := &fakeModelCatalogRepository{
		list: []*domaincap.ModelCatalogEntry{
			{EntryID: "entry-x"},
		},
	}
	handler, err := NewListModelCatalogEntriesHandler(repo)
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
