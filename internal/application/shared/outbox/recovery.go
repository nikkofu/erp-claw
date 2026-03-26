package outbox

import (
	"context"
	"errors"
	"fmt"
)

var errOutboxRecoveryEmptyIDs = errors.New("outbox recovery requires at least one message id")
var errOutboxRecoveryRepositoryRequired = errors.New("outbox recovery requires repository")

type RecoveryConfig struct {
	Clock Clock
}

type RecoveryService struct {
	repository RecoveryRepository
	clock      Clock
}

func NewRecoveryService(repository RecoveryRepository, cfg RecoveryConfig) *RecoveryService {
	clock := cfg.Clock
	if clock == nil {
		clock = systemClock{}
	}

	return &RecoveryService{
		repository: repository,
		clock:      clock,
	}
}

func (s *RecoveryService) RequeueFailed(ctx context.Context, ids []int64) (int, error) {
	if s.repository == nil {
		return 0, errOutboxRecoveryRepositoryRequired
	}
	if len(ids) == 0 {
		return 0, errOutboxRecoveryEmptyIDs
	}
	for _, id := range ids {
		if id <= 0 {
			return 0, fmt.Errorf("outbox recovery id must be positive: %d", id)
		}
	}

	return s.repository.RequeueFailed(ctx, ids, s.clock.Now())
}
