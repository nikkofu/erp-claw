package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nikkofu/erp-claw/internal/domain/capability"
)

var _ capability.ModelCatalogRepository = (*CapabilityRepository)(nil)
var _ capability.ToolCatalogRepository = (*CapabilityRepository)(nil)
var _ capability.AgentCapabilityPolicyRepository = (*CapabilityRepository)(nil)

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

type transactionExecutor interface {
	dbExecutor
	Commit() error
	Rollback() error
}

type transactionalDB interface {
	dbExecutor
	BeginTx(ctx context.Context, opts *sql.TxOptions) (transactionExecutor, error)
}

type CapabilityRepository struct {
	db transactionalDB
}

func NewCapabilityRepository(db transactionalDB) (*CapabilityRepository, error) {
	if db == nil {
		return nil, errors.New("db executor is required")
	}
	return &CapabilityRepository{db: db}, nil
}

type sqlDBAdapter struct {
	db *sql.DB
}

type sqlTxAdapter struct {
	tx *sql.Tx
}

func (a sqlDBAdapter) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return a.db.ExecContext(ctx, query, args...)
}

func (a sqlDBAdapter) QueryContext(ctx context.Context, query string, args ...any) (rowScanner, error) {
	return a.db.QueryContext(ctx, query, args...)
}

func (a sqlDBAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (transactionExecutor, error) {
	tx, err := a.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return sqlTxAdapter{tx: tx}, nil
}

func (a sqlTxAdapter) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return a.tx.ExecContext(ctx, query, args...)
}

func (a sqlTxAdapter) QueryContext(ctx context.Context, query string, args ...any) (rowScanner, error) {
	return a.tx.QueryContext(ctx, query, args...)
}

func (a sqlTxAdapter) Commit() error {
	return a.tx.Commit()
}

func (a sqlTxAdapter) Rollback() error {
	return a.tx.Rollback()
}

func NewCapabilityRepositoryFromSQLDB(db *sql.DB) (*CapabilityRepository, error) {
	if db == nil {
		return nil, errors.New("db executor is required")
	}
	return NewCapabilityRepository(sqlDBAdapter{db: db})
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

func (r *CapabilityRepository) SaveAgentCapabilityPolicy(ctx context.Context, policy *capability.AgentCapabilityPolicy) error {
	if policy == nil {
		return errors.New("policy is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := replaceAgentCapabilityPolicy(ctx, tx, policy); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func replaceAgentCapabilityPolicy(ctx context.Context, tx transactionExecutor, policy *capability.AgentCapabilityPolicy) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM agent_profile_allowed_model
		WHERE tenant_id = $1 AND agent_profile_id = $2`, policy.TenantID, policy.AgentProfileID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM agent_profile_allowed_tool
		WHERE tenant_id = $1 AND agent_profile_id = $2`, policy.TenantID, policy.AgentProfileID); err != nil {
		return err
	}

	for _, entryID := range policy.AllowedModelEntryIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO agent_profile_allowed_model
			(tenant_id, agent_profile_id, model_entry_id, created_at)
			VALUES ($1, $2, $3, NOW())`,
			policy.TenantID,
			policy.AgentProfileID,
			entryID,
		); err != nil {
			return err
		}
	}

	for _, entryID := range policy.AllowedToolEntryIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO agent_profile_allowed_tool
			(tenant_id, agent_profile_id, tool_entry_id, created_at)
			VALUES ($1, $2, $3, NOW())`,
			policy.TenantID,
			policy.AgentProfileID,
			entryID,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *CapabilityRepository) GetAgentCapabilityPolicy(ctx context.Context, tenantID, agentProfileID string) (*capability.AgentCapabilityPolicy, error) {
	modelRows, err := r.db.QueryContext(ctx, `SELECT model_entry_id
		FROM agent_profile_allowed_model
		WHERE tenant_id = $1 AND agent_profile_id = $2
		ORDER BY model_entry_id`, tenantID, agentProfileID)
	if err != nil {
		return nil, err
	}
	defer modelRows.Close()

	modelIDs := make([]string, 0)
	for modelRows.Next() {
		var entryID string
		if err := modelRows.Scan(&entryID); err != nil {
			return nil, err
		}
		modelIDs = append(modelIDs, entryID)
	}
	if err := modelRows.Err(); err != nil {
		return nil, err
	}

	toolRows, err := r.db.QueryContext(ctx, `SELECT tool_entry_id
		FROM agent_profile_allowed_tool
		WHERE tenant_id = $1 AND agent_profile_id = $2
		ORDER BY tool_entry_id`, tenantID, agentProfileID)
	if err != nil {
		return nil, err
	}
	defer toolRows.Close()

	toolIDs := make([]string, 0)
	for toolRows.Next() {
		var entryID string
		if err := toolRows.Scan(&entryID); err != nil {
			return nil, err
		}
		toolIDs = append(toolIDs, entryID)
	}
	if err := toolRows.Err(); err != nil {
		return nil, err
	}

	return capability.NewAgentCapabilityPolicy(tenantID, agentProfileID, modelIDs, toolIDs)
}
