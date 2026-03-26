package outbox

import (
	"context"
	"time"
)

type Message struct {
	ID        int64
	TenantID  string
	Topic     string
	EventType string
	Payload   []byte
	Attempts  int
}

type Repository interface {
	FetchPublishable(ctx context.Context, limit int, now time.Time) ([]Message, error)
	MarkPublished(ctx context.Context, id int64, publishedAt time.Time) error
	MarkForRetry(ctx context.Context, id int64, nextAvailableAt time.Time, reason string) error
	MarkFailed(ctx context.Context, id int64, failedAt time.Time, reason string) error
}

type RecoveryRepository interface {
	RequeueFailed(ctx context.Context, ids []int64, availableAt time.Time) (int, error)
}

type Publisher interface {
	Publish(ctx context.Context, message Message) error
}

type Clock interface {
	Now() time.Time
}
