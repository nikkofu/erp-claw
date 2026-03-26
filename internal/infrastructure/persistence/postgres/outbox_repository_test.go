package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/nikkofu/erp-claw/internal/application/shared/outbox"
)

var _ outbox.Repository = (*OutboxRepository)(nil)
var _ outbox.RecoveryRepository = (*OutboxRepository)(nil)
var _ outbox.MessageReader = (*OutboxRepository)(nil)

func TestNewOutboxRepositoryRejectsNilDB(t *testing.T) {
	t.Parallel()

	_, err := NewOutboxRepository(nil)
	if err == nil {
		t.Fatal("expected nil db to fail")
	}
}

func TestOutboxRepositoryRequeueFailedRejectsEmptyIDs(t *testing.T) {
	t.Parallel()

	repo := &OutboxRepository{}
	_, err := repo.RequeueFailed(context.Background(), nil, time.Date(2026, 3, 25, 14, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("expected empty ids to fail")
	}
}

func TestOutboxRepositoryRequeueFailedRejectsNonPositiveID(t *testing.T) {
	t.Parallel()

	repo := &OutboxRepository{}
	_, err := repo.RequeueFailed(context.Background(), []int64{101, 0}, time.Date(2026, 3, 25, 14, 1, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("expected non-positive id to fail")
	}
}
