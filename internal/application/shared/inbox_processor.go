package shared

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInboxStoreRequired    = errors.New("inbox store is required")
	ErrInboxHandlerRequired  = errors.New("inbox handler is required")
	ErrInboxMessageKeyNeeded = errors.New("inbox message key is required")
	ErrInboxTopicNeeded      = errors.New("inbox topic is required")
)

type InboxStore interface {
	ClaimMessage(ctx context.Context, tenantID int64, messageKey string, topic string, payload []byte) (bool, error)
	MarkProcessed(ctx context.Context, tenantID int64, messageKey string) error
	MarkFailed(ctx context.Context, tenantID int64, messageKey string, failureReason string) error
}

type InboxMessage struct {
	TenantID   int64
	MessageKey string
	Topic      string
	Payload    []byte
}

type InboxHandler func(ctx context.Context, msg InboxMessage) error

// ProcessInboxMessage executes an idempotent inbox processing flow:
// claim once -> run handler -> mark processed/failed.
// It returns claimed=false for duplicate messages that were already processed/claimed.
func ProcessInboxMessage(
	ctx context.Context,
	store InboxStore,
	msg InboxMessage,
	handler InboxHandler,
) (claimed bool, err error) {
	if store == nil {
		return false, ErrInboxStoreRequired
	}
	if handler == nil {
		return false, ErrInboxHandlerRequired
	}
	if strings.TrimSpace(msg.MessageKey) == "" {
		return false, ErrInboxMessageKeyNeeded
	}
	if strings.TrimSpace(msg.Topic) == "" {
		return false, ErrInboxTopicNeeded
	}

	inserted, err := store.ClaimMessage(ctx, msg.TenantID, msg.MessageKey, msg.Topic, msg.Payload)
	if err != nil {
		return false, fmt.Errorf("claim inbox message: %w", err)
	}
	if !inserted {
		return false, nil
	}

	if err := handler(ctx, msg); err != nil {
		handlerErr := err
		if markErr := store.MarkFailed(ctx, msg.TenantID, msg.MessageKey, err.Error()); markErr != nil {
			return true, fmt.Errorf("handle inbox message: %v; mark failed: %w", handlerErr, markErr)
		}
		return true, fmt.Errorf("handle inbox message: %w", handlerErr)
	}

	if err := store.MarkProcessed(ctx, msg.TenantID, msg.MessageKey); err != nil {
		return true, fmt.Errorf("mark inbox processed: %w", err)
	}
	return true, nil
}
