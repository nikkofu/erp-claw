package observability

import (
	"context"
	"errors"
	"time"
)

var errNilOutboxProcessor = errors.New("observability outbox processor requires non-nil processor")

type OutboxProcessor interface {
	ProcessOnce(ctx context.Context) error
}

type OutboxPollMetrics interface {
	RecordOutboxPoll(duration time.Duration, failed bool)
}

type InstrumentedOutboxProcessor struct {
	next    OutboxProcessor
	metrics OutboxPollMetrics
	now     func() time.Time
}

func NewInstrumentedOutboxProcessor(next OutboxProcessor, metrics OutboxPollMetrics) (*InstrumentedOutboxProcessor, error) {
	if next == nil {
		return nil, errNilOutboxProcessor
	}
	if metrics == nil {
		metrics = NewNoopOutboxPollMetrics()
	}

	return &InstrumentedOutboxProcessor{
		next:    next,
		metrics: metrics,
		now:     time.Now,
	}, nil
}

func (p *InstrumentedOutboxProcessor) ProcessOnce(ctx context.Context) error {
	start := p.now()

	err := p.next.ProcessOnce(ctx)
	p.metrics.RecordOutboxPoll(time.Since(start), err != nil)

	return err
}

type NoopOutboxPollMetrics struct{}

func NewNoopOutboxPollMetrics() *NoopOutboxPollMetrics {
	return &NoopOutboxPollMetrics{}
}

func (*NoopOutboxPollMetrics) RecordOutboxPoll(_ time.Duration, _ bool) {}
