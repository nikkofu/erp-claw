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
	claimLimit        int
	claimReadyBefore  time.Time
	claimLeaseUntil   time.Time
	publishedIDs      []int64
	retryIDs          []int64
	retryAvailableAts []time.Time
	retryAttempts     []int
	retryErrors       []string
	failedIDs         []int64
	failedAttempts    []int
	failedAts         []time.Time
	failedErrors      []string
}

func (f *fakeOutboxStore) ClaimReady(_ context.Context, limit int, readyBefore time.Time, leaseUntil time.Time) ([]outboxRecord, error) {
	if f.claimErr != nil {
		return nil, f.claimErr
	}
	f.claimLimit = limit
	f.claimReadyBefore = readyBefore
	f.claimLeaseUntil = leaseUntil
	return append([]outboxRecord(nil), f.claimRecords...), nil
}

func (f *fakeOutboxStore) MarkPublished(_ context.Context, id int64) error {
	f.publishedIDs = append(f.publishedIDs, id)
	return nil
}

func (f *fakeOutboxStore) MarkPendingRetry(_ context.Context, id int64, nextAvailableAt time.Time, attempts int, lastError string) error {
	f.retryIDs = append(f.retryIDs, id)
	f.retryAvailableAts = append(f.retryAvailableAts, nextAvailableAt)
	f.retryAttempts = append(f.retryAttempts, attempts)
	f.retryErrors = append(f.retryErrors, lastError)
	return nil
}

func (f *fakeOutboxStore) MarkFailed(_ context.Context, id int64, attempts int, failedAt time.Time, lastError string) error {
	f.failedIDs = append(f.failedIDs, id)
	f.failedAttempts = append(f.failedAttempts, attempts)
	f.failedAts = append(f.failedAts, failedAt)
	f.failedErrors = append(f.failedErrors, lastError)
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
			{ID: 10, TenantID: 1001, Topic: "platform.audit.created", EventType: "audit.created", Payload: []byte(`{"x":1}`), Attempts: 0},
			{ID: 11, TenantID: 1002, Topic: "platform.task.created", EventType: "task.created", Payload: []byte(`{"x":2}`), Attempts: 0},
		},
	}
	bus := &fakeBus{}
	now := time.Date(2026, 3, 27, 9, 45, 0, 0, time.UTC)

	err := pollOutboxBatchWithStore(context.Background(), store, bus, now, 100, 5*time.Second, 3, 30*time.Second)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if store.claimLimit != 100 {
		t.Fatalf("expected claim limit 100, got %d", store.claimLimit)
	}
	if !store.claimReadyBefore.Equal(now) {
		t.Fatalf("expected claim readyBefore %s, got %s", now, store.claimReadyBefore)
	}
	expectedLeaseUntil := now.Add(30 * time.Second)
	if !store.claimLeaseUntil.Equal(expectedLeaseUntil) {
		t.Fatalf("expected leaseUntil %s, got %s", expectedLeaseUntil, store.claimLeaseUntil)
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
	if got := bus.published[0].MessageID; got != "outbox:10" {
		t.Fatalf("expected message id outbox:10, got %s", got)
	}
	if len(store.publishedIDs) != 2 || store.publishedIDs[0] != 10 || store.publishedIDs[1] != 11 {
		t.Fatalf("expected published IDs [10 11], got %v", store.publishedIDs)
	}
	if len(store.retryIDs) != 0 {
		t.Fatalf("expected no retry IDs, got %v", store.retryIDs)
	}
	if len(store.failedIDs) != 0 {
		t.Fatalf("expected no failed IDs, got %v", store.failedIDs)
	}
}

func TestPollOutboxBatchWithStoreRetriesWhenPublishFails(t *testing.T) {
	store := &fakeOutboxStore{
		claimRecords: []outboxRecord{
			{ID: 20, TenantID: 1003, Topic: "platform.task.failed", EventType: "task.failed", Payload: []byte(`{"x":3}`), Attempts: 0},
		},
	}
	bus := &fakeBus{
		publishErrByTopic: map[string]error{
			"platform.task.failed": errors.New("nats down"),
		},
	}
	now := time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC)
	retryDelay := 7 * time.Second

	err := pollOutboxBatchWithStore(context.Background(), store, bus, now, 100, retryDelay, 3, 30*time.Second)
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
	if len(store.retryAttempts) != 1 || store.retryAttempts[0] != 1 {
		t.Fatalf("expected retry attempts [1], got %v", store.retryAttempts)
	}
	if len(store.failedIDs) != 0 {
		t.Fatalf("expected no failed IDs for first failure, got %v", store.failedIDs)
	}
}

func TestPollOutboxBatchWithStoreMarksFailedAndPublishesDeadLetterWhenAttemptsExhausted(t *testing.T) {
	store := &fakeOutboxStore{
		claimRecords: []outboxRecord{
			{ID: 21, TenantID: 1004, Topic: "platform.task.failed.final", EventType: "task.failed", Payload: []byte(`{"x":4}`), Attempts: 2},
		},
	}
	bus := &fakeBus{
		publishErrByTopic: map[string]error{
			"platform.task.failed.final": errors.New("nats down hard"),
		},
	}
	now := time.Date(2026, 3, 27, 10, 5, 0, 0, time.UTC)

	err := pollOutboxBatchWithStore(context.Background(), store, bus, now, 100, 5*time.Second, 3, 30*time.Second)
	if err == nil {
		t.Fatal("expected error when terminal publish failure occurs")
	}
	if len(store.retryIDs) != 0 {
		t.Fatalf("expected no retry IDs when exhausted, got %v", store.retryIDs)
	}
	if len(store.failedIDs) != 1 || store.failedIDs[0] != 21 {
		t.Fatalf("expected failed ID [21], got %v", store.failedIDs)
	}
	if len(store.failedAttempts) != 1 || store.failedAttempts[0] != 3 {
		t.Fatalf("expected failed attempts [3], got %v", store.failedAttempts)
	}
	if len(store.failedAts) != 1 || !store.failedAts[0].Equal(now) {
		t.Fatalf("expected failedAt %s, got %v", now, store.failedAts)
	}
	if len(bus.published) != 1 {
		t.Fatalf("expected one dead-letter event, got %d", len(bus.published))
	}
	if bus.published[0].Topic != defaultOutboxDLQTopic {
		t.Fatalf("expected dead-letter topic %s, got %s", defaultOutboxDLQTopic, bus.published[0].Topic)
	}
	if bus.published[0].Correlation != "outbox:21:dead-letter" {
		t.Fatalf("expected dead-letter correlation outbox:21:dead-letter, got %s", bus.published[0].Correlation)
	}
	if bus.published[0].MessageID != "outbox:21:dead-letter" {
		t.Fatalf("expected dead-letter message id outbox:21:dead-letter, got %s", bus.published[0].MessageID)
	}
	payload, ok := bus.published[0].Payload.(outboxDeadLetterPayload)
	if !ok {
		t.Fatalf("expected dead-letter payload type %T, got %T", outboxDeadLetterPayload{}, bus.published[0].Payload)
	}
	if payload.OutboxID != 21 || payload.TenantID != 1004 {
		t.Fatalf("unexpected dead-letter identity payload: %+v", payload)
	}
	if payload.Topic != "platform.task.failed.final" || payload.EventType != "task.failed" {
		t.Fatalf("unexpected dead-letter routing payload: %+v", payload)
	}
	if payload.Attempts != 3 || payload.LastError != "nats down hard" {
		t.Fatalf("unexpected dead-letter retry metadata: %+v", payload)
	}
	if !payload.FailedAt.Equal(now) {
		t.Fatalf("expected dead-letter failedAt %s, got %s", now, payload.FailedAt)
	}
	if string(payload.Payload) != `{"x":4}` {
		t.Fatalf("expected dead-letter payload body %s, got %s", `{"x":4}`, string(payload.Payload))
	}
	if payload.OccurredAt.IsZero() {
		t.Fatal("expected dead-letter occurredAt to be set")
	}
}

func TestPollOutboxBatchWithStoreStillMarksFailedWhenDeadLetterPublishFails(t *testing.T) {
	store := &fakeOutboxStore{
		claimRecords: []outboxRecord{
			{ID: 22, TenantID: 1005, Topic: "platform.task.failed.final", EventType: "task.failed", Payload: []byte(`{"x":5}`), Attempts: 2},
		},
	}
	bus := &fakeBus{
		publishErrByTopic: map[string]error{
			"platform.task.failed.final": errors.New("nats down hard"),
			defaultOutboxDLQTopic:        errors.New("dlq down"),
		},
	}
	now := time.Date(2026, 3, 27, 10, 6, 0, 0, time.UTC)

	err := pollOutboxBatchWithStore(context.Background(), store, bus, now, 100, 5*time.Second, 3, 30*time.Second)
	if err == nil {
		t.Fatal("expected error when terminal publish failure and dead-letter publish failure occur")
	}
	if len(store.failedIDs) != 1 || store.failedIDs[0] != 22 {
		t.Fatalf("expected failed ID [22], got %v", store.failedIDs)
	}
	if len(bus.published) != 0 {
		t.Fatalf("expected no successful publish events, got %d", len(bus.published))
	}
}

func TestPollOutboxBatchWithStoreReturnsClaimError(t *testing.T) {
	store := &fakeOutboxStore{
		claimErr: errors.New("db unavailable"),
	}
	bus := &fakeBus{}

	err := pollOutboxBatchWithStore(context.Background(), store, bus, time.Now(), 100, 5*time.Second, 3, 30*time.Second)
	if err == nil {
		t.Fatal("expected claim error")
	}
}
