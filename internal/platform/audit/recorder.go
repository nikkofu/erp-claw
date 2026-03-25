package audit

import (
	"context"
	"sync"
)

// Recorder persists audit records emitted by the command pipeline.
type Recorder interface {
	Record(ctx context.Context, record Record) error
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
