package outbox

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRecoveryServiceRequeueFailedUsesClockNow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)
	repo := &recoveryRepositoryStub{requeueCount: 2}
	service := NewRecoveryService(repo, RecoveryConfig{
		Clock: fixedClock{at: now},
	})

	count, err := service.RequeueFailed(context.Background(), []int64{101, 202, 303})
	if err != nil {
		t.Fatalf("RequeueFailed() error = %v, want nil", err)
	}
	if count != 2 {
		t.Fatalf("RequeueFailed() count = %d, want 2", count)
	}

	if len(repo.requeueCalls) != 1 {
		t.Fatalf("requeue call count = %d, want 1", len(repo.requeueCalls))
	}
	call := repo.requeueCalls[0]
	if call.at != now {
		t.Fatalf("requeue at = %s, want %s", call.at, now)
	}
	if len(call.ids) != 3 {
		t.Fatalf("requeue ids len = %d, want 3", len(call.ids))
	}
}

func TestRecoveryServiceRequeueFailedReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	repo := &recoveryRepositoryStub{
		requeueErr: errors.New("db unavailable"),
	}
	service := NewRecoveryService(repo, RecoveryConfig{
		Clock: fixedClock{at: time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)},
	})

	_, err := service.RequeueFailed(context.Background(), []int64{777})
	if err == nil {
		t.Fatal("expected repository error, got nil")
	}
}

func TestRecoveryServiceRequeueFailedRejectsEmptyIDs(t *testing.T) {
	t.Parallel()

	repo := &recoveryRepositoryStub{}
	service := NewRecoveryService(repo, RecoveryConfig{
		Clock: fixedClock{at: time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)},
	})

	_, err := service.RequeueFailed(context.Background(), nil)
	if err == nil {
		t.Fatal("expected empty ids to fail")
	}
}

func TestRecoveryServiceRequeueFailedRejectsNilRepository(t *testing.T) {
	t.Parallel()

	service := NewRecoveryService(nil, RecoveryConfig{
		Clock: fixedClock{at: time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)},
	})

	_, err := service.RequeueFailed(context.Background(), []int64{1})
	if !errors.Is(err, errOutboxRecoveryRepositoryRequired) {
		t.Fatalf("RequeueFailed() error = %v, want %v", err, errOutboxRecoveryRepositoryRequired)
	}
}

type recoveryRepositoryStub struct {
	requeueCount int
	requeueErr   error
	requeueCalls []requeueCall
}

type requeueCall struct {
	ids []int64
	at  time.Time
}

func (r *recoveryRepositoryStub) RequeueFailed(_ context.Context, ids []int64, availableAt time.Time) (int, error) {
	idsCopy := append([]int64(nil), ids...)
	r.requeueCalls = append(r.requeueCalls, requeueCall{
		ids: idsCopy,
		at:  availableAt,
	})
	return r.requeueCount, r.requeueErr
}
