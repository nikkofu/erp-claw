package bootstrap

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/postgres"
)

func newCapabilityCatalog(cfg Config) CapabilityCatalog {
	if shouldUseInMemoryCatalogFallback(cfg) {
		return newInMemoryCapabilityCatalog()
	}

	catalog, err := newPostgresCapabilityCatalog(cfg.Database)
	if err == nil {
		return catalog
	}

	panic(fmt.Errorf("bootstrap: capability catalog init failed: %w", err))
}

func newPostgresCapabilityCatalog(cfg DatabaseConfig) (CapabilityCatalog, error) {
	db, err := postgres.New(postgres.Config{
		DSN:          cfg.DSN,
		MaxOpenConns: cfg.MaxOpenConns,
		MaxIdleConns: cfg.MaxIdleConns,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	repo, err := postgres.NewCapabilityRepositoryFromSQLDB(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return repo, nil
}

type inMemoryCapabilityCatalog struct {
	mu       sync.RWMutex
	models   map[string]*domaincap.ModelCatalogEntry
	tools    map[string]*domaincap.ToolCatalogEntry
	policies map[string]*domaincap.AgentCapabilityPolicy
}

func newInMemoryCapabilityCatalog() *inMemoryCapabilityCatalog {
	return &inMemoryCapabilityCatalog{
		models:   make(map[string]*domaincap.ModelCatalogEntry),
		tools:    make(map[string]*domaincap.ToolCatalogEntry),
		policies: make(map[string]*domaincap.AgentCapabilityPolicy),
	}
}

func (r *inMemoryCapabilityCatalog) Save(_ context.Context, entry *domaincap.ModelCatalogEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entry == nil {
		return fmt.Errorf("model catalog entry is required")
	}

	copied := *entry
	now := time.Now().UTC()
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = now
	}
	copied.UpdatedAt = now
	r.models[r.modelKey(copied.TenantID, copied.EntryID)] = &copied
	return nil
}

func (r *inMemoryCapabilityCatalog) ListByTenant(_ context.Context, tenantID string) ([]*domaincap.ModelCatalogEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]*domaincap.ModelCatalogEntry, 0)
	for _, entry := range r.models {
		if entry.TenantID != tenantID {
			continue
		}
		copied := *entry
		entries = append(entries, &copied)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].EntryID < entries[j].EntryID
	})
	return entries, nil
}

func (r *inMemoryCapabilityCatalog) SaveTool(_ context.Context, entry *domaincap.ToolCatalogEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entry == nil {
		return fmt.Errorf("tool catalog entry is required")
	}

	copied := *entry
	now := time.Now().UTC()
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = now
	}
	copied.UpdatedAt = now
	r.tools[r.toolKey(copied.TenantID, copied.EntryID)] = &copied
	return nil
}

func (r *inMemoryCapabilityCatalog) ListToolsByTenant(_ context.Context, tenantID string) ([]*domaincap.ToolCatalogEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]*domaincap.ToolCatalogEntry, 0)
	for _, entry := range r.tools {
		if entry.TenantID != tenantID {
			continue
		}
		copied := *entry
		entries = append(entries, &copied)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].EntryID < entries[j].EntryID
	})
	return entries, nil
}

func (r *inMemoryCapabilityCatalog) SaveAgentCapabilityPolicy(_ context.Context, policy *domaincap.AgentCapabilityPolicy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if policy == nil {
		return fmt.Errorf("agent capability policy is required")
	}

	copied := *policy
	copied.AllowedModelEntryIDs = cloneStringSlice(policy.AllowedModelEntryIDs)
	copied.AllowedToolEntryIDs = cloneStringSlice(policy.AllowedToolEntryIDs)
	r.policies[r.policyKey(copied.TenantID, copied.AgentProfileID)] = &copied
	return nil
}

func (r *inMemoryCapabilityCatalog) GetAgentCapabilityPolicy(_ context.Context, tenantID, agentProfileID string) (*domaincap.AgentCapabilityPolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := r.policyKey(tenantID, agentProfileID)
	if existing, ok := r.policies[key]; ok {
		copied := *existing
		copied.AllowedModelEntryIDs = cloneStringSlice(existing.AllowedModelEntryIDs)
		copied.AllowedToolEntryIDs = cloneStringSlice(existing.AllowedToolEntryIDs)
		return &copied, nil
	}

	return domaincap.NewAgentCapabilityPolicy(tenantID, agentProfileID, nil, nil)
}

func (r *inMemoryCapabilityCatalog) modelKey(tenantID, entryID string) string {
	return tenantID + "|" + entryID
}

func (r *inMemoryCapabilityCatalog) toolKey(tenantID, entryID string) string {
	return tenantID + "|" + entryID
}

func (r *inMemoryCapabilityCatalog) policyKey(tenantID, agentProfileID string) string {
	return tenantID + "|" + agentProfileID
}

func cloneStringSlice(values []string) []string {
	return append([]string{}, values...)
}
