package shared

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeInboxStore struct {
	claimInserted bool
	claimErr      error
	markProcErr   error
	markFailErr   error

	claimedTenantID int64
	claimedKey      string
	claimedTopic    string
	claimedPayload  []byte

	processedTenantID int64
	processedKey      string

	failedTenantID int64
	failedKey      string
	failedReason   string
}

func (f *fakeInboxStore) ClaimMessage(_ context.Context, tenantID int64, messageKey string, topic string, payload []byte) (bool, error) {
	f.claimedTenantID = tenantID
	f.claimedKey = messageKey
	f.claimedTopic = topic
	f.claimedPayload = append([]byte(nil), payload...)
	if f.claimErr != nil {
		return false, f.claimErr
	}
	return f.claimInserted, nil
}

func (f *fakeInboxStore) MarkProcessed(_ context.Context, tenantID int64, messageKey string) error {
	f.processedTenantID = tenantID
	f.processedKey = messageKey
	return f.markProcErr
}

func (f *fakeInboxStore) MarkFailed(_ context.Context, tenantID int64, messageKey string, failureReason string) error {
	f.failedTenantID = tenantID
	f.failedKey = messageKey
	f.failedReason = failureReason
	return f.markFailErr
}

func TestProcessInboxMessageSuccess(t *testing.T) {
	t.Parallel()

	store := &fakeInboxStore{claimInserted: true}
	handlerCalled := false
	msg := InboxMessage{
		TenantID:   101,
		MessageKey: "msg-1",
		Topic:      "platform.task.created",
		Payload:    []byte(`{"x":1}`),
	}

	claimed, err := ProcessInboxMessage(context.Background(), store, msg, func(_ context.Context, got InboxMessage) error {
		handlerCalled = true
		if got.MessageKey != "msg-1" {
			t.Fatalf("expected message key msg-1, got %s", got.MessageKey)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !claimed {
		t.Fatal("expected claimed=true, got false")
	}
	if !handlerCalled {
		t.Fatal("expected handler to be called")
	}
	if store.processedTenantID != 101 || store.processedKey != "msg-1" {
		t.Fatalf("expected mark processed on tenant/key 101/msg-1, got %d/%s", store.processedTenantID, store.processedKey)
	}
	if store.failedKey != "" {
		t.Fatalf("expected no mark failed call, got key=%s", store.failedKey)
	}
}

func TestProcessInboxMessageDuplicateReturnsWithoutHandling(t *testing.T) {
	t.Parallel()

	store := &fakeInboxStore{claimInserted: false}
	handlerCalled := false

	claimed, err := ProcessInboxMessage(context.Background(), store, InboxMessage{
		TenantID:   101,
		MessageKey: "dup-1",
		Topic:      "platform.task.created",
	}, func(_ context.Context, _ InboxMessage) error {
		handlerCalled = true
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error for duplicate, got %v", err)
	}
	if claimed {
		t.Fatal("expected claimed=false for duplicate message")
	}
	if handlerCalled {
		t.Fatal("expected handler not to be called for duplicate message")
	}
	if store.processedKey != "" || store.failedKey != "" {
		t.Fatalf("expected no state transition calls, got processed=%s failed=%s", store.processedKey, store.failedKey)
	}
}

func TestProcessInboxMessageClaimError(t *testing.T) {
	t.Parallel()

	store := &fakeInboxStore{
		claimErr: errors.New("db unavailable"),
	}
	claimed, err := ProcessInboxMessage(context.Background(), store, InboxMessage{
		TenantID:   101,
		MessageKey: "m1",
		Topic:      "t1",
	}, func(_ context.Context, _ InboxMessage) error { return nil })
	if err == nil {
		t.Fatal("expected claim error, got nil")
	}
	if claimed {
		t.Fatal("expected claimed=false when claim fails")
	}
}

func TestProcessInboxMessageHandlerFailureMarksFailed(t *testing.T) {
	t.Parallel()

	store := &fakeInboxStore{claimInserted: true}
	claimed, err := ProcessInboxMessage(context.Background(), store, InboxMessage{
		TenantID:   102,
		MessageKey: "m2",
		Topic:      "t2",
	}, func(_ context.Context, _ InboxMessage) error {
		return errors.New("handler failed")
	})
	if err == nil {
		t.Fatal("expected handler error, got nil")
	}
	if !claimed {
		t.Fatal("expected claimed=true when inserted")
	}
	if store.failedTenantID != 102 || store.failedKey != "m2" {
		t.Fatalf("expected mark failed tenant/key 102/m2, got %d/%s", store.failedTenantID, store.failedKey)
	}
	if !strings.Contains(store.failedReason, "handler failed") {
		t.Fatalf("expected failure reason to include handler failed, got %s", store.failedReason)
	}
	if store.processedKey != "" {
		t.Fatalf("expected no processed mark on handler failure, got %s", store.processedKey)
	}
}

func TestProcessInboxMessageHandlerFailureAndMarkFailedError(t *testing.T) {
	t.Parallel()

	store := &fakeInboxStore{
		claimInserted: true,
		markFailErr:   errors.New("mark failed unavailable"),
	}

	claimed, err := ProcessInboxMessage(context.Background(), store, InboxMessage{
		TenantID:   103,
		MessageKey: "m3",
		Topic:      "t3",
	}, func(_ context.Context, _ InboxMessage) error {
		return errors.New("handler failed")
	})
	if err == nil {
		t.Fatal("expected composite error, got nil")
	}
	if !claimed {
		t.Fatal("expected claimed=true")
	}
	if !strings.Contains(err.Error(), "mark failed") {
		t.Fatalf("expected error to include mark failed, got %v", err)
	}
}

func TestProcessInboxMessageMarkProcessedError(t *testing.T) {
	t.Parallel()

	store := &fakeInboxStore{
		claimInserted: true,
		markProcErr:   errors.New("mark processed unavailable"),
	}

	claimed, err := ProcessInboxMessage(context.Background(), store, InboxMessage{
		TenantID:   104,
		MessageKey: "m4",
		Topic:      "t4",
	}, func(_ context.Context, _ InboxMessage) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected mark processed error, got nil")
	}
	if !claimed {
		t.Fatal("expected claimed=true")
	}
	if store.processedKey != "m4" {
		t.Fatalf("expected processed key m4, got %s", store.processedKey)
	}
}

func TestProcessInboxMessageValidation(t *testing.T) {
	t.Parallel()

	_, err := ProcessInboxMessage(context.Background(), nil, InboxMessage{}, func(_ context.Context, _ InboxMessage) error { return nil })
	if !errors.Is(err, ErrInboxStoreRequired) {
		t.Fatalf("expected ErrInboxStoreRequired, got %v", err)
	}

	store := &fakeInboxStore{}
	_, err = ProcessInboxMessage(context.Background(), store, InboxMessage{}, nil)
	if !errors.Is(err, ErrInboxHandlerRequired) {
		t.Fatalf("expected ErrInboxHandlerRequired, got %v", err)
	}

	_, err = ProcessInboxMessage(context.Background(), store, InboxMessage{
		TenantID:   1,
		MessageKey: " ",
		Topic:      "t1",
	}, func(_ context.Context, _ InboxMessage) error { return nil })
	if !errors.Is(err, ErrInboxMessageKeyNeeded) {
		t.Fatalf("expected ErrInboxMessageKeyNeeded, got %v", err)
	}

	_, err = ProcessInboxMessage(context.Background(), store, InboxMessage{
		TenantID:   1,
		MessageKey: "m1",
		Topic:      " ",
	}, func(_ context.Context, _ InboxMessage) error { return nil })
	if !errors.Is(err, ErrInboxTopicNeeded) {
		t.Fatalf("expected ErrInboxTopicNeeded, got %v", err)
	}
}
