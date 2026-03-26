package audit

import (
	"context"
	"strings"
	"sync"
)

// Recorder persists audit records emitted by the command pipeline.
type Recorder interface {
	Record(ctx context.Context, record Record) error
}

type Query struct {
	TenantID    string
	CommandName string
	Limit       int
}

type Reader interface {
	List(ctx context.Context, query Query) ([]Record, error)
}

type noopRecorder struct{}

// NoopRecorder drops all records and is suitable for bootstrap/test defaults.
func NoopRecorder() Recorder {
	return noopRecorder{}
}

func (noopRecorder) Record(context.Context, Record) error {
	return nil
}

// InMemoryRecorder stores records in-memory for tests.
type InMemoryRecorder struct {
	mu      sync.Mutex
	records []Record
}

func NewInMemoryRecorder() *InMemoryRecorder {
	return &InMemoryRecorder{}
}

func (r *InMemoryRecorder) Record(_ context.Context, record Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.records = append(r.records, record)
	return nil
}

func (r *InMemoryRecorder) Records() []Record {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]Record, len(r.records))
	copy(out, r.records)
	return out
}

func (r *InMemoryRecorder) List(_ context.Context, query Query) ([]Record, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	tenantID := strings.TrimSpace(query.TenantID)
	commandName := strings.TrimSpace(query.CommandName)

	out := make([]Record, 0, len(r.records))
	for idx := len(r.records) - 1; idx >= 0; idx-- {
		record := r.records[idx]
		if tenantID != "" && record.TenantID != tenantID {
			continue
		}
		if commandName != "" && record.CommandName != commandName {
			continue
		}
		out = append(out, record)
		if query.Limit > 0 && len(out) >= query.Limit {
			break
		}
	}
	return out, nil
}
