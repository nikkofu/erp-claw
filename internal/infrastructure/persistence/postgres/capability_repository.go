package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/capability"
)

var _ capability.ModelCatalogRepository = (*CapabilityRepository)(nil)
var _ capability.ToolCatalogRepository = (*CapabilityRepository)(nil)

type rowScanner interface {
	Close() error
	Next() bool
	Scan(dest ...any) error
	Err() error
}

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (rowScanner, error)
}

type CapabilityRepository struct {
	db dbExecutor
}

func NewCapabilityRepository(db dbExecutor) (*CapabilityRepository, error) {
	if db == nil {
		return nil, errors.New("db executor is required")
	}
	return &CapabilityRepository{db: db}, nil
}

func (r *CapabilityRepository) Save(ctx context.Context, entry *capability.ModelCatalogEntry) error {
	if entry == nil {
		return errors.New("entry is required")
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO model_catalog_entries
		(tenant_id, entry_id, model_key, display_name, provider, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tenant_id, entry_id) DO UPDATE
		SET model_key = EXCLUDED.model_key,
			display_name = EXCLUDED.display_name,
			provider = EXCLUDED.provider,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at`,
		entry.TenantID,
		entry.EntryID,
		entry.ModelKey,
		entry.DisplayName,
		entry.Provider,
		entry.Status,
		entry.CreatedAt,
		entry.UpdatedAt,
	)
	return err
}

func (r *CapabilityRepository) ListByTenant(ctx context.Context, tenantID string) ([]*capability.ModelCatalogEntry, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT tenant_id, entry_id, model_key, display_name, provider, status, created_at, updated_at
		FROM model_catalog_entries WHERE tenant_id = $1 ORDER BY entry_id`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*capability.ModelCatalogEntry
	for rows.Next() {
		var entry capability.ModelCatalogEntry
		if err := rows.Scan(&entry.TenantID, &entry.EntryID, &entry.ModelKey, &entry.DisplayName, &entry.Provider, &entry.Status, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *CapabilityRepository) SaveTool(ctx context.Context, entry *capability.ToolCatalogEntry) error {
	if entry == nil {
		return errors.New("entry is required")
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO tool_catalog_entries
		(tenant_id, entry_id, tool_key, display_name, risk_level, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tenant_id, entry_id) DO UPDATE
		SET tool_key = EXCLUDED.tool_key,
			display_name = EXCLUDED.display_name,
			risk_level = EXCLUDED.risk_level,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at`,
		entry.TenantID,
		entry.EntryID,
		entry.ToolKey,
		entry.DisplayName,
		entry.RiskLevel,
		entry.Status,
		entry.CreatedAt,
		entry.UpdatedAt,
	)
	return err
}

func (r *CapabilityRepository) ListToolsByTenant(ctx context.Context, tenantID string) ([]*capability.ToolCatalogEntry, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT tenant_id, entry_id, tool_key, display_name, risk_level, status, created_at, updated_at
		FROM tool_catalog_entries WHERE tenant_id = $1 ORDER BY entry_id`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*capability.ToolCatalogEntry
	for rows.Next() {
		var entry capability.ToolCatalogEntry
		if err := rows.Scan(&entry.TenantID, &entry.EntryID, &entry.ToolKey, &entry.DisplayName, &entry.RiskLevel, &entry.Status, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
