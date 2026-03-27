package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	inboxStatusReceived  = "received"
	inboxStatusProcessed = "processed"
	inboxStatusFailed    = "failed"
)

const claimInboxMessageSQL = `
insert into inbox (tenant_id, message_key, topic, payload, status, received_at)
values ($1, $2, $3, $4, $5, $6)
on conflict (tenant_id, message_key) do nothing;
`

const markInboxProcessedSQL = `
update inbox
set status = $3,
    processed_at = $4,
    error = null
where tenant_id = $1
  and message_key = $2;
`

const markInboxFailedSQL = `
update inbox
set status = $3,
    error = $4
where tenant_id = $1
  and message_key = $2;
`

type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type InboxStore struct {
	exec SQLExecutor
	now  func() time.Time
}

func NewInboxStore(exec SQLExecutor) (*InboxStore, error) {
	if exec == nil {
		return nil, errors.New("sql executor is required")
	}
	return &InboxStore{
		exec: exec,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}, nil
}

func (s *InboxStore) ClaimMessage(ctx context.Context, tenantID int64, messageKey string, topic string, payload []byte) (bool, error) {
	if tenantID <= 0 {
		return false, errors.New("tenant id must be positive")
	}
	key := strings.TrimSpace(messageKey)
	if key == "" {
		return false, errors.New("message key is required")
	}
	eventTopic := strings.TrimSpace(topic)
	if eventTopic == "" {
		return false, errors.New("topic is required")
	}

	receivedAt := s.now().UTC()
	result, err := s.exec.ExecContext(
		ctx,
		claimInboxMessageSQL,
		tenantID,
		key,
		eventTopic,
		payload,
		inboxStatusReceived,
		receivedAt,
	)
	if err != nil {
		return false, fmt.Errorf("claim inbox message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("claim inbox message rows affected: %w", err)
	}
	return rowsAffected > 0, nil
}

func (s *InboxStore) MarkProcessed(ctx context.Context, tenantID int64, messageKey string) error {
	if tenantID <= 0 {
		return errors.New("tenant id must be positive")
	}
	key := strings.TrimSpace(messageKey)
	if key == "" {
		return errors.New("message key is required")
	}

	processedAt := s.now().UTC()
	if _, err := s.exec.ExecContext(ctx, markInboxProcessedSQL, tenantID, key, inboxStatusProcessed, processedAt); err != nil {
		return fmt.Errorf("mark inbox processed: %w", err)
	}
	return nil
}

func (s *InboxStore) MarkFailed(ctx context.Context, tenantID int64, messageKey string, failureReason string) error {
	if tenantID <= 0 {
		return errors.New("tenant id must be positive")
	}
	key := strings.TrimSpace(messageKey)
	if key == "" {
		return errors.New("message key is required")
	}

	reason := strings.TrimSpace(failureReason)
	if reason == "" {
		reason = "unknown error"
	}

	if _, err := s.exec.ExecContext(ctx, markInboxFailedSQL, tenantID, key, inboxStatusFailed, reason); err != nil {
		return fmt.Errorf("mark inbox failed: %w", err)
	}
	return nil
}
