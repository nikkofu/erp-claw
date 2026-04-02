package postgres

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

type fakeExecCall struct {
	query string
	args  []any
}

type fakeSQLExecutor struct {
	result sql.Result
	err    error
	calls  []fakeExecCall
}

func (f *fakeSQLExecutor) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	f.calls = append(f.calls, fakeExecCall{
		query: query,
		args:  append([]any(nil), args...),
	})
	if f.err != nil {
		return nil, f.err
	}
	if f.result == nil {
		return rowsAffectedResult(1), nil
	}
	return f.result, nil
}

type rowsAffectedResult int64

func (r rowsAffectedResult) LastInsertId() (int64, error) {
	return 0, errors.New("not supported")
}

func (r rowsAffectedResult) RowsAffected() (int64, error) {
	return int64(r), nil
}

type rowsAffectedErrResult struct{}

func (rowsAffectedErrResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (rowsAffectedErrResult) RowsAffected() (int64, error) {
	return 0, errors.New("rows affected unavailable")
}

func TestNewInboxStoreRejectsNilExecutor(t *testing.T) {
	t.Parallel()

	_, err := NewInboxStore(nil)
	if err == nil {
		t.Fatal("expected nil executor error, got nil")
	}
}

func TestInboxStoreClaimMessageReturnsTrueWhenInserted(t *testing.T) {
	t.Parallel()

	exec := &fakeSQLExecutor{
		result: rowsAffectedResult(1),
	}
	store, err := NewInboxStore(exec)
	if err != nil {
		t.Fatalf("expected no constructor error, got %v", err)
	}

	now := time.Date(2026, 3, 27, 15, 30, 0, 0, time.UTC)
	store.now = func() time.Time { return now }

	inserted, err := store.ClaimMessage(context.Background(), 101, " msg-001 ", " platform.task.created ", []byte(`{"x":1}`))
	if err != nil {
		t.Fatalf("expected no claim error, got %v", err)
	}
	if !inserted {
		t.Fatal("expected inserted=true, got false")
	}
	if len(exec.calls) != 1 {
		t.Fatalf("expected one exec call, got %d", len(exec.calls))
	}
	call := exec.calls[0]
	if !strings.Contains(call.query, "insert into inbox") {
		t.Fatalf("expected insert query, got %s", call.query)
	}
	if len(call.args) != 6 {
		t.Fatalf("expected 6 args, got %d", len(call.args))
	}
	if got, ok := call.args[0].(int64); !ok || got != 101 {
		t.Fatalf("expected tenant id arg 101, got %v", call.args[0])
	}
	if got, ok := call.args[1].(string); !ok || got != "msg-001" {
		t.Fatalf("expected message key msg-001, got %v", call.args[1])
	}
	if got, ok := call.args[2].(string); !ok || got != "platform.task.created" {
		t.Fatalf("expected topic platform.task.created, got %v", call.args[2])
	}
	if got, ok := call.args[3].([]byte); !ok || !reflect.DeepEqual(got, []byte(`{"x":1}`)) {
		t.Fatalf("expected payload bytes %q, got %v", `{"x":1}`, call.args[3])
	}
	if got, ok := call.args[4].(string); !ok || got != inboxStatusReceived {
		t.Fatalf("expected inbox status %s, got %v", inboxStatusReceived, call.args[4])
	}
	if got, ok := call.args[5].(time.Time); !ok || !got.Equal(now) {
		t.Fatalf("expected receivedAt %s, got %v", now, call.args[5])
	}
}

func TestInboxStoreClaimMessageReturnsFalseWhenDuplicate(t *testing.T) {
	t.Parallel()

	exec := &fakeSQLExecutor{
		result: rowsAffectedResult(0),
	}
	store, err := NewInboxStore(exec)
	if err != nil {
		t.Fatalf("expected no constructor error, got %v", err)
	}

	inserted, err := store.ClaimMessage(context.Background(), 101, "msg-dup", "platform.task.created", []byte(`{"x":1}`))
	if err != nil {
		t.Fatalf("expected no claim error, got %v", err)
	}
	if inserted {
		t.Fatal("expected inserted=false for duplicate, got true")
	}
}

func TestInboxStoreClaimMessageRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	store, err := NewInboxStore(&fakeSQLExecutor{})
	if err != nil {
		t.Fatalf("expected no constructor error, got %v", err)
	}

	tests := []struct {
		name       string
		tenantID   int64
		messageKey string
		topic      string
	}{
		{name: "invalid tenant id", tenantID: 0, messageKey: "k1", topic: "t1"},
		{name: "empty message key", tenantID: 101, messageKey: "   ", topic: "t1"},
		{name: "empty topic", tenantID: 101, messageKey: "k1", topic: "   "},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := store.ClaimMessage(context.Background(), tc.tenantID, tc.messageKey, tc.topic, nil)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestInboxStoreClaimMessageReturnsRowsAffectedError(t *testing.T) {
	t.Parallel()

	exec := &fakeSQLExecutor{
		result: rowsAffectedErrResult{},
	}
	store, err := NewInboxStore(exec)
	if err != nil {
		t.Fatalf("expected no constructor error, got %v", err)
	}

	_, err = store.ClaimMessage(context.Background(), 101, "k1", "topic", nil)
	if err == nil {
		t.Fatal("expected rows affected error, got nil")
	}
}

func TestInboxStoreMarkProcessedWritesProcessedState(t *testing.T) {
	t.Parallel()

	exec := &fakeSQLExecutor{
		result: rowsAffectedResult(1),
	}
	store, err := NewInboxStore(exec)
	if err != nil {
		t.Fatalf("expected no constructor error, got %v", err)
	}

	now := time.Date(2026, 3, 27, 15, 31, 0, 0, time.UTC)
	store.now = func() time.Time { return now }

	if err := store.MarkProcessed(context.Background(), 102, " msg-processed "); err != nil {
		t.Fatalf("expected no mark processed error, got %v", err)
	}
	if len(exec.calls) != 1 {
		t.Fatalf("expected one exec call, got %d", len(exec.calls))
	}
	call := exec.calls[0]
	if !strings.Contains(call.query, "update inbox") {
		t.Fatalf("expected update query, got %s", call.query)
	}
	if got, ok := call.args[0].(int64); !ok || got != 102 {
		t.Fatalf("expected tenant id 102, got %v", call.args[0])
	}
	if got, ok := call.args[1].(string); !ok || got != "msg-processed" {
		t.Fatalf("expected message key msg-processed, got %v", call.args[1])
	}
	if got, ok := call.args[2].(string); !ok || got != inboxStatusProcessed {
		t.Fatalf("expected status %s, got %v", inboxStatusProcessed, call.args[2])
	}
	if got, ok := call.args[3].(time.Time); !ok || !got.Equal(now) {
		t.Fatalf("expected processedAt %s, got %v", now, call.args[3])
	}
}

func TestInboxStoreMarkFailedWritesFailedState(t *testing.T) {
	t.Parallel()

	exec := &fakeSQLExecutor{
		result: rowsAffectedResult(1),
	}
	store, err := NewInboxStore(exec)
	if err != nil {
		t.Fatalf("expected no constructor error, got %v", err)
	}

	if err := store.MarkFailed(context.Background(), 103, " msg-failed ", " nats timeout "); err != nil {
		t.Fatalf("expected no mark failed error, got %v", err)
	}
	if len(exec.calls) != 1 {
		t.Fatalf("expected one exec call, got %d", len(exec.calls))
	}
	call := exec.calls[0]
	if got, ok := call.args[0].(int64); !ok || got != 103 {
		t.Fatalf("expected tenant id 103, got %v", call.args[0])
	}
	if got, ok := call.args[1].(string); !ok || got != "msg-failed" {
		t.Fatalf("expected message key msg-failed, got %v", call.args[1])
	}
	if got, ok := call.args[2].(string); !ok || got != inboxStatusFailed {
		t.Fatalf("expected status %s, got %v", inboxStatusFailed, call.args[2])
	}
	if got, ok := call.args[3].(string); !ok || got != "nats timeout" {
		t.Fatalf("expected failure reason nats timeout, got %v", call.args[3])
	}
}

func TestInboxStoreMarkFailedUsesUnknownReasonWhenEmpty(t *testing.T) {
	t.Parallel()

	exec := &fakeSQLExecutor{
		result: rowsAffectedResult(1),
	}
	store, err := NewInboxStore(exec)
	if err != nil {
		t.Fatalf("expected no constructor error, got %v", err)
	}

	if err := store.MarkFailed(context.Background(), 104, "msg-failed", " "); err != nil {
		t.Fatalf("expected no mark failed error, got %v", err)
	}
	if len(exec.calls) != 1 {
		t.Fatalf("expected one exec call, got %d", len(exec.calls))
	}
	if got, ok := exec.calls[0].args[3].(string); !ok || got != "unknown error" {
		t.Fatalf("expected unknown error reason, got %v", exec.calls[0].args[3])
	}
}
