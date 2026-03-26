package outbox

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDispatcherProcessOncePublishesAndMarksPublished(t *testing.T) {
	t.Parallel()

	repo := &repositoryStub{
		fetchMessages: []Message{
			{ID: 101, TenantID: "1", Topic: "orders.created", EventType: "orders.created", Payload: []byte(`{"id":"o-1"}`)},
			{ID: 102, TenantID: "2", Topic: "orders.approved", EventType: "orders.approved", Payload: []byte(`{"id":"o-2"}`)},
		},
	}
	publisher := &publisherStub{}
	clock := fixedClock{at: time.Date(2026, 3, 25, 10, 0, 0, 0, time.UTC)}
	dispatcher := NewDispatcher(repo, publisher, DispatcherConfig{
		BatchSize:  10,
		RetryDelay: 30 * time.Second,
		Clock:      clock,
	})

	if err := dispatcher.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("ProcessOnce() error = %v, want nil", err)
	}

	if len(publisher.published) != 2 {
		t.Fatalf("published count = %d, want 2", len(publisher.published))
	}

	if len(repo.published) != 2 {
		t.Fatalf("mark published count = %d, want 2", len(repo.published))
	}

	if len(repo.retried) != 0 {
		t.Fatalf("retry count = %d, want 0", len(repo.retried))
	}
}

func TestDispatcherProcessOnceMarksRetryOnPublishFailure(t *testing.T) {
	t.Parallel()

	repo := &repositoryStub{
		fetchMessages: []Message{
			{ID: 201, TenantID: "3", Topic: "inventory.adjusted", EventType: "inventory.adjusted", Payload: []byte(`{"id":"i-1"}`), Attempts: 1},
		},
	}
	publisher := &publisherStub{
		errByID: map[int64]error{
			201: errors.New("nats unavailable"),
		},
	}
	clock := fixedClock{at: time.Date(2026, 3, 25, 10, 1, 0, 0, time.UTC)}
	dispatcher := NewDispatcher(repo, publisher, DispatcherConfig{
		BatchSize:   10,
		RetryDelay:  45 * time.Second,
		MaxAttempts: 3,
		Clock:       clock,
	})

	if err := dispatcher.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("ProcessOnce() error = %v, want nil", err)
	}

	if len(repo.published) != 0 {
		t.Fatalf("mark published count = %d, want 0", len(repo.published))
	}

	if len(repo.retried) != 1 {
		t.Fatalf("retry count = %d, want 1", len(repo.retried))
	}
	if len(repo.failed) != 0 {
		t.Fatalf("failed count = %d, want 0", len(repo.failed))
	}

	retry := repo.retried[0]
	if retry.id != 201 {
		t.Fatalf("retry id = %d, want 201", retry.id)
	}
	if retry.at != clock.at.Add(45*time.Second) {
		t.Fatalf("retry at = %s, want %s", retry.at, clock.at.Add(45*time.Second))
	}
	if retry.reason == "" {
		t.Fatalf("retry reason should not be empty")
	}
}

func TestDispatcherProcessOnceMarksFailedWhenMaxAttemptsReached(t *testing.T) {
	t.Parallel()

	repo := &repositoryStub{
		fetchMessages: []Message{
			{ID: 301, TenantID: "7", Topic: "invoice.posted", EventType: "invoice.posted", Payload: []byte(`{"id":"inv-1"}`), Attempts: 3},
		},
	}
	publisher := &publisherStub{
		errByID: map[int64]error{
			301: errors.New("publish timeout"),
		},
	}
	clock := fixedClock{at: time.Date(2026, 3, 25, 10, 3, 0, 0, time.UTC)}
	dispatcher := NewDispatcher(repo, publisher, DispatcherConfig{
		BatchSize:   5,
		RetryDelay:  30 * time.Second,
		MaxAttempts: 3,
		Clock:       clock,
	})

	if err := dispatcher.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("ProcessOnce() error = %v, want nil", err)
	}

	if len(repo.retried) != 0 {
		t.Fatalf("retry count = %d, want 0", len(repo.retried))
	}

	if len(repo.failed) != 1 {
		t.Fatalf("failed count = %d, want 1", len(repo.failed))
	}

	failed := repo.failed[0]
	if failed.id != 301 {
		t.Fatalf("failed id = %d, want 301", failed.id)
	}
	if failed.at != clock.at {
		t.Fatalf("failed at = %s, want %s", failed.at, clock.at)
	}
	if failed.reason == "" {
		t.Fatalf("failed reason should not be empty")
	}
}

func TestDispatcherProcessOnceReturnsFetchError(t *testing.T) {
	t.Parallel()

	repo := &repositoryStub{
		fetchErr: errors.New("db down"),
	}
	dispatcher := NewDispatcher(repo, &publisherStub{}, DispatcherConfig{
		BatchSize:  10,
		RetryDelay: time.Second,
		Clock:      fixedClock{at: time.Date(2026, 3, 25, 10, 2, 0, 0, time.UTC)},
	})

	err := dispatcher.ProcessOnce(context.Background())
	if err == nil {
		t.Fatalf("expected fetch error, got nil")
	}
}

type repositoryStub struct {
	fetchMessages []Message
	fetchErr      error
	published     []publishedCall
	retried       []retryCall
	failed        []failedCall
}

type publishedCall struct {
	id int64
	at time.Time
}

type retryCall struct {
	id     int64
	at     time.Time
	reason string
}

type failedCall struct {
	id     int64
	at     time.Time
	reason string
}

func (r *repositoryStub) FetchPublishable(_ context.Context, _ int, _ time.Time) ([]Message, error) {
	if r.fetchErr != nil {
		return nil, r.fetchErr
	}
	return r.fetchMessages, nil
}

func (r *repositoryStub) MarkPublished(_ context.Context, id int64, publishedAt time.Time) error {
	r.published = append(r.published, publishedCall{id: id, at: publishedAt})
	return nil
}

func (r *repositoryStub) MarkForRetry(_ context.Context, id int64, nextAvailableAt time.Time, reason string) error {
	r.retried = append(r.retried, retryCall{id: id, at: nextAvailableAt, reason: reason})
	return nil
}

func (r *repositoryStub) MarkFailed(_ context.Context, id int64, failedAt time.Time, reason string) error {
	r.failed = append(r.failed, failedCall{id: id, at: failedAt, reason: reason})
	return nil
}

type publisherStub struct {
	published []Message
	errByID   map[int64]error
}

func (p *publisherStub) Publish(_ context.Context, message Message) error {
	p.published = append(p.published, message)
	if p.errByID == nil {
		return nil
	}
	return p.errByID[message.ID]
}

type fixedClock struct {
	at time.Time
}

func (c fixedClock) Now() time.Time {
	return c.at
}
