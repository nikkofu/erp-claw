package audit

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"
)

// Query defines tenant-scoped filters for audit event retrieval.
type Query struct {
	TenantID       string
	CommandName    string
	ActorID        string
	OccurredAfter  time.Time
	OccurredBefore time.Time
	Limit          int
}

// EventStore persists and queries audit events.
type EventStore interface {
	Append(ctx context.Context, record Record) (Record, error)
	List(ctx context.Context, query Query) ([]Record, error)
}

// InMemoryStore stores events in-memory for tests and bootstrap defaults.
type InMemoryStore struct {
	mu      sync.Mutex
	nextID  int
	records []Record
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{}
}

func (s *InMemoryStore) Append(_ context.Context, record Record) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	if record.ID == "" {
		record.ID = strconv.Itoa(s.nextID)
	}

	now := time.Now().UTC()
	if record.OccurredAt.IsZero() {
		record.OccurredAt = now
	}
	if record.RecordedAt.IsZero() {
		record.RecordedAt = now
	}

	s.records = append(s.records, record)
	return record, nil
}

func (s *InMemoryStore) List(_ context.Context, query Query) ([]Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Record, 0)
	for _, record := range s.records {
		if query.TenantID != "" && record.TenantID != query.TenantID {
			continue
		}
		if query.CommandName != "" && record.CommandName != query.CommandName {
			continue
		}
		if query.ActorID != "" && record.ActorID != query.ActorID {
			continue
		}
		if !query.OccurredAfter.IsZero() && record.OccurredAt.Before(query.OccurredAfter) {
			continue
		}
		if !query.OccurredBefore.IsZero() && !record.OccurredAt.Before(query.OccurredBefore) {
			continue
		}
		out = append(out, record)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].OccurredAt.Equal(out[j].OccurredAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].OccurredAt.After(out[j].OccurredAt)
	})

	if query.Limit > 0 && len(out) > query.Limit {
		return out[:query.Limit], nil
	}

	return out, nil
}
