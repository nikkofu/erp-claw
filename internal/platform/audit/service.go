package audit

import (
	"context"
	"errors"
)

var errServiceStoreRequired = errors.New("audit service requires event store")

// Service records and queries audit events through an EventStore.
type Service struct {
	store EventStore
}

func NewService(store EventStore) (*Service, error) {
	if store == nil {
		return nil, errServiceStoreRequired
	}
	return &Service{store: store}, nil
}

func (s *Service) Record(ctx context.Context, record Record) error {
	_, err := s.store.Append(ctx, record)
	return err
}

func (s *Service) List(ctx context.Context, query Query) ([]Record, error) {
	return s.store.List(ctx, query)
}
