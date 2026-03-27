package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nikkofu/erp-claw/internal/platform/eventbus"
)

type fakeOutboxStore struct {
	claimRecords      []outboxRecord
	claimErr          error
	publishedIDs      []int64
	retryIDs          []int64
	retryAvailableAts []time.Time
}

func (f *fakeOutboxStore) ClaimPending(_ context.Context, _ int) ([]outboxRecord, error) {
	if f.claimErr != nil {
		return nil, f.claimErr
	}
	return append([]outboxRecord(nil), f.claimRecords...), nil
}

func (f *fakeOutboxStore) MarkPublished(_ context.Context, id int64) error {
	f.publishedIDs = append(f.publishedIDs, id)
	return nil
}

func (f *fakeOutboxStore) MarkPendingRetry(_ context.Context, id int64, nextAvailableAt time.Time) error {
	f.retryIDs = append(f.retryIDs, id)
	f.retryAvailableAts = append(f.retryAvailableAts, nextAvailableAt)
	return nil
}

type fakeBus struct {
	publishErrByTopic map[string]error
	published         []eventbus.Event
}

func (b *fakeBus) Publish(_ context.Context, evt eventbus.Event) error {
	if err, ok := b.publishErrByTopic[evt.Topic]; ok {
		return err
	}
	b.published = append(b.published, evt)
	return nil
}

func TestPollOutboxBatchWithStorePublishesAndMarksPublished(t *testing.T) {
	store := &fakeOutboxStore{
		claimRecords: []outboxRecord{
			{ID: 10, TenantID: 1001, Topic: "platform.audit.created", EventType: "audit.created", Payload: []byte(`{"x":1}`)},
			{ID: 11, TenantID: 1002, Topic: "platform.task.created", EventType: "task.created", Payload: []byte(`{"x":2}`)},
		},
	}
	bus := &fakeBus{}
	now := time.Date(2026, 3, 27, 9, 45, 0, 0, time.UTC)

	err := pollOutboxBatchWithStore(context.Background(), store, bus, now, 100, 5*time.Second)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bus.published) != 2 {
		t.Fatalf("expected 2 published events, got %d", len(bus.published))
	}
	if got := bus.published[0].TenantID; got != "1001" {
		t.Fatalf("expected tenant id 1001, got %s", got)
	}
	if got := bus.published[0].Correlation; got != "outbox:10" {
		t.Fatalf("expected correlation outbox:10, got %s", got)
	}
	if len(store.publishedIDs) != 2 || store.publishedIDs[0] != 10 || store.publishedIDs[1] != 11 {
		t.Fatalf("expected published IDs [10 11], got %v", store.publishedIDs)
	}
	if len(store.retryIDs) != 0 {
		t.Fatalf("expected no retry IDs, got %v", store.retryIDs)
	}
}

func TestPollOutboxBatchWithStoreRetriesWhenPublishFails(t *testing.T) {
	store := &fakeOutboxStore{
		claimRecords: []outboxRecord{
			{ID: 20, TenantID: 1003, Topic: "platform.task.failed", EventType: "task.failed", Payload: []byte(`{"x":3}`)},
		},
	}
	bus := &fakeBus{
		publishErrByTopic: map[string]error{
			"platform.task.failed": errors.New("nats down"),
		},
	}
	now := time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC)
	retryDelay := 7 * time.Second

	err := pollOutboxBatchWithStore(context.Background(), store, bus, now, 100, retryDelay)
	if err == nil {
		t.Fatal("expected error when publish fails")
	}
	if len(store.publishedIDs) != 0 {
		t.Fatalf("expected no published IDs, got %v", store.publishedIDs)
	}
	if len(store.retryIDs) != 1 || store.retryIDs[0] != 20 {
		t.Fatalf("expected retry ID [20], got %v", store.retryIDs)
	}
	if len(store.retryAvailableAts) != 1 {
		t.Fatalf("expected one retry timestamp, got %d", len(store.retryAvailableAts))
	}
	expectedRetryAt := now.Add(retryDelay)
	if !store.retryAvailableAts[0].Equal(expectedRetryAt) {
		t.Fatalf("expected retry at %s, got %s", expectedRetryAt, store.retryAvailableAts[0])
	}
}

func TestPollOutboxBatchWithStoreReturnsClaimError(t *testing.T) {
	store := &fakeOutboxStore{
		claimErr: errors.New("db unavailable"),
	}
	bus := &fakeBus{}

	err := pollOutboxBatchWithStore(context.Background(), store, bus, time.Now(), 100, 5*time.Second)
	if err == nil {
		t.Fatal("expected claim error")
	}
}
