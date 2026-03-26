package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nikkofu/erp-claw/internal/application/shared/outbox"
)

const defaultOutboxFetchLimit = 50

var errOutboxRepositoryNilDB = errors.New("postgres outbox repository requires non-nil db")

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) (*OutboxRepository, error) {
	if db == nil {
		return nil, errOutboxRepositoryNilDB
	}

	return &OutboxRepository{db: db}, nil
}

func (r *OutboxRepository) FetchPublishable(ctx context.Context, limit int, now time.Time) ([]outbox.Message, error) {
	if limit <= 0 {
		limit = defaultOutboxFetchLimit
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(
		ctx,
		`select id, tenant_id, topic, event_type, payload, attempts
		 from outbox
		 where status = 'pending' and available_at <= $1
		 order by available_at asc, id asc
		 for update skip locked
		 limit $2`,
		now,
		limit,
	)
	if err != nil {
		return nil, rollbackWithError(tx, err)
	}
	defer rows.Close()

	messages := make([]outbox.Message, 0)
	for rows.Next() {
		var message outbox.Message
		var tenantID int64
		if err := rows.Scan(&message.ID, &tenantID, &message.Topic, &message.EventType, &message.Payload, &message.Attempts); err != nil {
			return nil, rollbackWithError(tx, err)
		}
		message.TenantID = strconv.FormatInt(tenantID, 10)
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, rollbackWithError(tx, err)
	}

	claimed := make([]outbox.Message, 0, len(messages))
	for _, message := range messages {
		result, err := tx.ExecContext(
			ctx,
			`update outbox
			 set status = 'publishing',
			     attempts = attempts + 1,
			     processing_at = $2
			 where id = $1 and status = 'pending'`,
			message.ID,
			now,
		)
		if err != nil {
			return nil, rollbackWithError(tx, err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return nil, rollbackWithError(tx, err)
		}
		if affected == 1 {
			message.Attempts++
			claimed = append(claimed, message)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return claimed, nil
}

func (r *OutboxRepository) MarkPublished(ctx context.Context, id int64, publishedAt time.Time) error {
	_, err := r.db.ExecContext(
		ctx,
		`update outbox
		 set status = 'published',
		     published_at = $2,
		     last_error = null,
		     processing_at = null
		 where id = $1`,
		id,
		publishedAt,
	)
	return err
}

func (r *OutboxRepository) MarkForRetry(ctx context.Context, id int64, nextAvailableAt time.Time, reason string) error {
	_, err := r.db.ExecContext(
		ctx,
		`update outbox
		 set status = 'pending',
		     available_at = $2,
		     last_error = $3,
		     processing_at = null
		 where id = $1`,
		id,
		nextAvailableAt,
		reason,
	)
	return err
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id int64, failedAt time.Time, reason string) error {
	_, err := r.db.ExecContext(
		ctx,
		`update outbox
		 set status = 'failed',
		     available_at = $2,
		     last_error = $3,
		     processing_at = null
		 where id = $1`,
		id,
		failedAt,
		reason,
	)
	return err
}

func (r *OutboxRepository) RequeueFailed(ctx context.Context, ids []int64, availableAt time.Time) (int, error) {
	if len(ids) == 0 {
		return 0, errors.New("outbox requeue requires at least one message id")
	}

	args := make([]any, 0, len(ids)+1)
	placeholders := make([]string, 0, len(ids))
	for i, id := range ids {
		if id <= 0 {
			return 0, fmt.Errorf("outbox requeue id must be positive: %d", id)
		}
		args = append(args, id)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
	}

	args = append(args, availableAt)
	availableAtArgPos := len(args)

	result, err := r.db.ExecContext(
		ctx,
		fmt.Sprintf(
			`update outbox
			 set status = 'pending',
			     available_at = $%d,
			     last_error = null,
			     processing_at = null
			 where status = 'failed' and id in (%s)`,
			availableAtArgPos,
			strings.Join(placeholders, ", "),
		),
		args...,
	)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(affected), nil
}

func rollbackWithError(tx *sql.Tx, err error) error {
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return fmt.Errorf("%w (rollback failed: %v)", err, rollbackErr)
	}
	return err
}
