package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

var errPolicyAuditRepositoryNilDB = errors.New("postgres policy-audit repository requires non-nil db")

// PolicyAuditRepository provides persistent storage for policy rules and audit events.
type PolicyAuditRepository struct {
	db *sql.DB
}

func NewPolicyAuditRepository(db *sql.DB) (*PolicyAuditRepository, error) {
	if db == nil {
		return nil, errPolicyAuditRepositoryNilDB
	}
	return &PolicyAuditRepository{db: db}, nil
}

func (r *PolicyAuditRepository) UpsertRule(ctx context.Context, rule policy.Rule) (policy.Rule, error) {
	if strings.TrimSpace(rule.ID) == "" {
		rule.ID = uuid.NewString()
	}
	if rule.ActorID == "" {
		rule.ActorID = "*"
	}
	if rule.CommandName == "" {
		rule.CommandName = "*"
	}

	var decision string
	if err := r.db.QueryRowContext(
		ctx,
		`insert into policy_rule(tenant_id, id, command_name, actor_id, decision, priority, is_active, updated_at)
		 values ($1, $2, $3, $4, $5, $6, $7, now())
		 on conflict (tenant_id, id)
		 do update set
			command_name = excluded.command_name,
			actor_id = excluded.actor_id,
			decision = excluded.decision,
			priority = excluded.priority,
			is_active = excluded.is_active,
			updated_at = now()
		 returning tenant_id, id, command_name, actor_id, decision, priority, is_active, created_at, updated_at`,
		rule.TenantID,
		rule.ID,
		rule.CommandName,
		rule.ActorID,
		string(rule.Decision),
		rule.Priority,
		rule.Active,
	).Scan(
		&rule.TenantID,
		&rule.ID,
		&rule.CommandName,
		&rule.ActorID,
		&decision,
		&rule.Priority,
		&rule.Active,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	); err != nil {
		return policy.Rule{}, err
	}

	rule.Decision = policy.Decision(decision)
	return rule, nil
}

func (r *PolicyAuditRepository) ListRules(ctx context.Context, filter policy.RuleFilter) ([]policy.Rule, error) {
	if strings.TrimSpace(filter.TenantID) == "" {
		return []policy.Rule{}, nil
	}

	query := strings.Builder{}
	query.WriteString(
		`select tenant_id, id, command_name, actor_id, decision, priority, is_active, created_at, updated_at
		 from policy_rule
		 where tenant_id = $1`,
	)
	args := []any{filter.TenantID}

	if filter.CommandName != "" {
		args = append(args, filter.CommandName)
		query.WriteString(fmt.Sprintf(" and (command_name = $%d or command_name = '*')", len(args)))
	}
	if filter.ActorID != "" {
		args = append(args, filter.ActorID)
		query.WriteString(fmt.Sprintf(" and (actor_id = $%d or actor_id = '*')", len(args)))
	}
	if filter.ActiveOnly {
		query.WriteString(" and is_active = true")
	}

	query.WriteString(" order by priority desc, id asc")
	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		query.WriteString(fmt.Sprintf(" limit $%d", len(args)))
	}

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]policy.Rule, 0)
	for rows.Next() {
		var rule policy.Rule
		var decision string
		if err := rows.Scan(
			&rule.TenantID,
			&rule.ID,
			&rule.CommandName,
			&rule.ActorID,
			&decision,
			&rule.Priority,
			&rule.Active,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rule.Decision = policy.Decision(decision)
		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *PolicyAuditRepository) SetRuleActive(ctx context.Context, tenantID, ruleID string, active bool) (policy.Rule, error) {
	var rule policy.Rule
	var decision string
	if err := r.db.QueryRowContext(
		ctx,
		`update policy_rule
		 set is_active = $3, updated_at = now()
		 where tenant_id = $1 and id = $2
		 returning tenant_id, id, command_name, actor_id, decision, priority, is_active, created_at, updated_at`,
		tenantID,
		ruleID,
		active,
	).Scan(
		&rule.TenantID,
		&rule.ID,
		&rule.CommandName,
		&rule.ActorID,
		&decision,
		&rule.Priority,
		&rule.Active,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return policy.Rule{}, policy.ErrRuleNotFound
		}
		return policy.Rule{}, err
	}

	rule.Decision = policy.Decision(decision)
	return rule, nil
}

func (r *PolicyAuditRepository) Record(ctx context.Context, record audit.Record) error {
	_, err := r.Append(ctx, record)
	return err
}

func (r *PolicyAuditRepository) Append(ctx context.Context, record audit.Record) (audit.Record, error) {
	if strings.TrimSpace(record.ID) == "" {
		record.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if record.OccurredAt.IsZero() {
		record.OccurredAt = now
	}
	if record.RecordedAt.IsZero() {
		record.RecordedAt = now
	}

	var decision string
	if err := r.db.QueryRowContext(
		ctx,
		`insert into audit_event(tenant_id, id, command_name, actor_id, decision, outcome, error, occurred_at, recorded_at)
		 values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 returning tenant_id, id, command_name, actor_id, decision, outcome, error, occurred_at, recorded_at`,
		record.TenantID,
		record.ID,
		record.CommandName,
		record.ActorID,
		string(record.Decision),
		record.Outcome,
		record.Error,
		record.OccurredAt,
		record.RecordedAt,
	).Scan(
		&record.TenantID,
		&record.ID,
		&record.CommandName,
		&record.ActorID,
		&decision,
		&record.Outcome,
		&record.Error,
		&record.OccurredAt,
		&record.RecordedAt,
	); err != nil {
		return audit.Record{}, err
	}

	record.Decision = policy.Decision(decision)
	return record, nil
}

func (r *PolicyAuditRepository) List(ctx context.Context, query audit.Query) ([]audit.Record, error) {
	sqlText, args, ok := buildAuditListQuery(query)
	if !ok {
		return []audit.Record{}, nil
	}

	rows, err := r.db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]audit.Record, 0)
	for rows.Next() {
		var record audit.Record
		var decision string
		if err := rows.Scan(
			&record.TenantID,
			&record.ID,
			&record.CommandName,
			&record.ActorID,
			&decision,
			&record.Outcome,
			&record.Error,
			&record.OccurredAt,
			&record.RecordedAt,
		); err != nil {
			return nil, err
		}
		record.Decision = policy.Decision(decision)
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func buildAuditListQuery(query audit.Query) (string, []any, bool) {
	if strings.TrimSpace(query.TenantID) == "" {
		return "", nil, false
	}

	sqlBuilder := strings.Builder{}
	sqlBuilder.WriteString(
		`select tenant_id, id, command_name, actor_id, decision, outcome, error, occurred_at, recorded_at
		 from audit_event
		 where tenant_id = $1`,
	)
	args := []any{query.TenantID}

	if query.CommandName != "" {
		args = append(args, query.CommandName)
		sqlBuilder.WriteString(fmt.Sprintf(" and command_name = $%d", len(args)))
	}
	if query.ActorID != "" {
		args = append(args, query.ActorID)
		sqlBuilder.WriteString(fmt.Sprintf(" and actor_id = $%d", len(args)))
	}
	if !query.OccurredAfter.IsZero() {
		args = append(args, query.OccurredAfter)
		sqlBuilder.WriteString(fmt.Sprintf(" and occurred_at >= $%d", len(args)))
	}
	if !query.OccurredBefore.IsZero() {
		args = append(args, query.OccurredBefore)
		sqlBuilder.WriteString(fmt.Sprintf(" and occurred_at < $%d", len(args)))
	}

	sqlBuilder.WriteString(" order by occurred_at desc, id asc")
	if query.Limit > 0 {
		args = append(args, query.Limit)
		sqlBuilder.WriteString(fmt.Sprintf(" limit $%d", len(args)))
	}

	return sqlBuilder.String(), args, true
}
