package tenant

// CellRoute describes a tenant's routing information for data isolation.
type CellRoute struct {
	TenantID      string
	Isolation     string
	DatabaseDSN   string
	CachePrefix   string
	StoragePrefix string
}
