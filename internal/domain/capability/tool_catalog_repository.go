package capability

import "context"

type ToolCatalogRepository interface {
	SaveTool(ctx context.Context, entry *ToolCatalogEntry) error
	ListToolsByTenant(ctx context.Context, tenantID string) ([]*ToolCatalogEntry, error)
}
