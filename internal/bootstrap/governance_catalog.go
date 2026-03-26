package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/postgres"
	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func newGovernanceCatalog(cfg Config) GovernanceCatalog {
	if shouldUseInMemoryCatalogFallback(cfg) {
		return newInMemoryGovernanceCatalog()
	}

	catalog, err := newPostgresGovernanceCatalog(cfg.Database)
	if err == nil {
		return catalog
	}

	panic(fmt.Errorf("bootstrap: governance catalog init failed: %w", err))
}

func newPostgresGovernanceCatalog(cfg DatabaseConfig) (GovernanceCatalog, error) {
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

	repo, err := postgres.NewPolicyAuditRepository(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return repo, nil
}

type inMemoryGovernanceCatalog struct {
	rules  *policy.InMemoryRuleRepository
	events *audit.InMemoryStore
}

func newInMemoryGovernanceCatalog() *inMemoryGovernanceCatalog {
	return &inMemoryGovernanceCatalog{
		rules:  policy.NewInMemoryRuleRepository(),
		events: audit.NewInMemoryStore(),
	}
}

func (c *inMemoryGovernanceCatalog) UpsertRule(ctx context.Context, rule policy.Rule) (policy.Rule, error) {
	return c.rules.UpsertRule(ctx, rule)
}

func (c *inMemoryGovernanceCatalog) ListRules(ctx context.Context, filter policy.RuleFilter) ([]policy.Rule, error) {
	return c.rules.ListRules(ctx, filter)
}

func (c *inMemoryGovernanceCatalog) SetRuleActive(ctx context.Context, tenantID, ruleID string, active bool) (policy.Rule, error) {
	return c.rules.SetRuleActive(ctx, tenantID, ruleID, active)
}

func (c *inMemoryGovernanceCatalog) Append(ctx context.Context, record audit.Record) (audit.Record, error) {
	return c.events.Append(ctx, record)
}

func (c *inMemoryGovernanceCatalog) List(ctx context.Context, query audit.Query) ([]audit.Record, error) {
	return c.events.List(ctx, query)
}
