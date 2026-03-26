package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nikkofu/erp-claw/internal/application/shared/outbox"
)

func TestOutboxRecoveryServiceRequeuesOnlyFailedMessages(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 25, 15, 0, 0, 0, time.UTC)
	repo := &inMemoryOutboxRecoveryRepository{
		rows: map[int64]outboxRow{
			10: {status: "failed", availableAt: now.Add(-time.Hour), lastError: "broker timeout"},
			11: {status: "pending", availableAt: now.Add(time.Minute)},
			12: {status: "failed", availableAt: now.Add(-2 * time.Hour), lastError: "nats unavailable"},
		},
	}
	service := outbox.NewRecoveryService(repo, outbox.RecoveryConfig{
		Clock: fixedIntegrationClock{at: now},
	})

	count, err := service.RequeueFailed(context.Background(), []int64{10, 11, 999})
	if err != nil {
		t.Fatalf("RequeueFailed() error = %v, want nil", err)
	}
	if count != 1 {
		t.Fatalf("RequeueFailed() count = %d, want 1", count)
	}

	row10 := repo.rows[10]
	if row10.status != "pending" {
		t.Fatalf("row 10 status = %q, want pending", row10.status)
	}
	if row10.availableAt != now {
		t.Fatalf("row 10 available_at = %s, want %s", row10.availableAt, now)
	}
	if row10.lastError != "" {
		t.Fatalf("row 10 last_error = %q, want empty", row10.lastError)
	}

	row12 := repo.rows[12]
	if row12.status != "failed" {
		t.Fatalf("row 12 status = %q, want failed", row12.status)
	}
}

func TestOutboxRecoveryServiceReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	repo := &inMemoryOutboxRecoveryRepository{
		err: errors.New("simulated storage failure"),
	}
	service := outbox.NewRecoveryService(repo, outbox.RecoveryConfig{
		Clock: fixedIntegrationClock{at: time.Date(2026, 3, 25, 15, 0, 0, 0, time.UTC)},
	})

	_, err := service.RequeueFailed(context.Background(), []int64{10})
	if err == nil {
		t.Fatal("expected repository error, got nil")
	}
}

type inMemoryOutboxRecoveryRepository struct {
	rows map[int64]outboxRow
	err  error
}

type outboxRow struct {
	status      string
	availableAt time.Time
	lastError   string
}

func (r *inMemoryOutboxRecoveryRepository) RequeueFailed(_ context.Context, ids []int64, availableAt time.Time) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	requeued := 0
	for _, id := range ids {
		row, ok := r.rows[id]
		if !ok {
			continue
		}
		if row.status != "failed" {
			continue
		}

		row.status = "pending"
		row.availableAt = availableAt
		row.lastError = ""
		r.rows[id] = row
		requeued++
	}
	return requeued, nil
}

type fixedIntegrationClock struct {
	at time.Time
}

func (c fixedIntegrationClock) Now() time.Time {
	return c.at
}
