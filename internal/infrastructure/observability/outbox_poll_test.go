package observability

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestInstrumentedOutboxProcessorRecordsSuccess(t *testing.T) {
	t.Parallel()

	metrics := &outboxPollMetricsStub{}
	next := &outboxProcessorStub{}

	processor, err := NewInstrumentedOutboxProcessor(next, metrics)
	if err != nil {
		t.Fatalf("NewInstrumentedOutboxProcessor() error = %v, want nil", err)
	}

	if err := processor.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("ProcessOnce() error = %v, want nil", err)
	}

	if metrics.calls != 1 {
		t.Fatalf("metrics calls = %d, want 1", metrics.calls)
	}
	if metrics.lastFailed {
		t.Fatalf("lastFailed = true, want false")
	}
	if metrics.lastDuration < 0 {
		t.Fatalf("lastDuration = %s, want >= 0", metrics.lastDuration)
	}
}

func TestInstrumentedOutboxProcessorRecordsFailureAndReturnsError(t *testing.T) {
	t.Parallel()

	metrics := &outboxPollMetricsStub{}
	next := &outboxProcessorStub{
		err: errors.New("poll failed"),
	}

	processor, err := NewInstrumentedOutboxProcessor(next, metrics)
	if err != nil {
		t.Fatalf("NewInstrumentedOutboxProcessor() error = %v, want nil", err)
	}

	err = processor.ProcessOnce(context.Background())
	if err == nil {
		t.Fatal("expected poll error, got nil")
	}

	if metrics.calls != 1 {
		t.Fatalf("metrics calls = %d, want 1", metrics.calls)
	}
	if !metrics.lastFailed {
		t.Fatalf("lastFailed = false, want true")
	}
}

func TestNewInstrumentedOutboxProcessorRejectsNilProcessor(t *testing.T) {
	t.Parallel()

	_, err := NewInstrumentedOutboxProcessor(nil, &outboxPollMetricsStub{})
	if err == nil {
		t.Fatal("expected nil processor to fail")
	}
}

type outboxProcessorStub struct {
	err error
}

func (s *outboxProcessorStub) ProcessOnce(_ context.Context) error {
	return s.err
}

type outboxPollMetricsStub struct {
	calls        int
	lastDuration time.Duration
	lastFailed   bool
}

func (s *outboxPollMetricsStub) RecordOutboxPoll(duration time.Duration, failed bool) {
	s.calls++
	s.lastDuration = duration
	s.lastFailed = failed
}
