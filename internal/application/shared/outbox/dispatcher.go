package outbox

import (
	"context"
	"strings"
	"time"
)

const (
	defaultBatchSize   = 50
	defaultRetryDelay  = 15 * time.Second
	defaultMaxAttempts = 5
)

type DispatcherConfig struct {
	BatchSize   int
	RetryDelay  time.Duration
	MaxAttempts int
	Clock       Clock
}

type Dispatcher struct {
	repository  Repository
	publisher   Publisher
	batchSize   int
	retryDelay  time.Duration
	maxAttempts int
	clock       Clock
}

func NewDispatcher(repository Repository, publisher Publisher, cfg DispatcherConfig) *Dispatcher {
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	retryDelay := cfg.RetryDelay
	if retryDelay <= 0 {
		retryDelay = defaultRetryDelay
	}

	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxAttempts
	}

	clock := cfg.Clock
	if clock == nil {
		clock = systemClock{}
	}

	return &Dispatcher{
		repository:  repository,
		publisher:   publisher,
		batchSize:   batchSize,
		retryDelay:  retryDelay,
		maxAttempts: maxAttempts,
		clock:       clock,
	}
}

func (d *Dispatcher) ProcessOnce(ctx context.Context) error {
	now := d.clock.Now()

	messages, err := d.repository.FetchPublishable(ctx, d.batchSize, now)
	if err != nil {
		return err
	}

	for _, message := range messages {
		if err := d.publisher.Publish(ctx, message); err != nil {
			reason := strings.TrimSpace(err.Error())
			if reason == "" {
				reason = "outbox publish failed"
			}
			if message.Attempts >= d.maxAttempts {
				if markErr := d.repository.MarkFailed(ctx, message.ID, now, reason); markErr != nil {
					return markErr
				}
				continue
			}
			if markErr := d.repository.MarkForRetry(ctx, message.ID, now.Add(d.retryDelay), reason); markErr != nil {
				return markErr
			}
			continue
		}

		if err := d.repository.MarkPublished(ctx, message.ID, now); err != nil {
			return err
		}
	}

	return nil
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}
