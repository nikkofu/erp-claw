package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nikkofu/erp-claw/internal/domain/capability"
)

type fakeDB struct {
	execCalls  []execCall
	queryCalls []queryCall
	queryRows  rowScanner
	querySets  []rowScanner
	beginCalls int
	tx         *fakeTx
}

type execCall struct {
	query string
	args  []any
}

type queryCall struct {
	query string
	args  []any
}

func (f *fakeDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	f.execCalls = append(f.execCalls, execCall{query: query, args: args})
	return fakeResult{}, nil
}

func (f *fakeDB) QueryContext(ctx context.Context, query string, args ...any) (rowScanner, error) {
	f.queryCalls = append(f.queryCalls, queryCall{query: query, args: args})
	if len(f.querySets) > 0 {
		rows := f.querySets[0]
		f.querySets = f.querySets[1:]
		return rows, nil
	}
	return f.queryRows, nil
}

func (f *fakeDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (transactionExecutor, error) {
	f.beginCalls++
	f.tx = &fakeTx{parent: f}
	return f.tx, nil
}

type fakeResult struct{}

func (fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (fakeResult) RowsAffected() (int64, error) { return 0, nil }

type fakeTx struct {
	parent        *fakeDB
	commitCalls   int
	rollbackCalls int
}

func (f *fakeTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	f.parent.execCalls = append(f.parent.execCalls, execCall{query: query, args: args})
	return fakeResult{}, nil
}

func (f *fakeTx) QueryContext(ctx context.Context, query string, args ...any) (rowScanner, error) {
	f.parent.queryCalls = append(f.parent.queryCalls, queryCall{query: query, args: args})
	if len(f.parent.querySets) > 0 {
		rows := f.parent.querySets[0]
		f.parent.querySets = f.parent.querySets[1:]
		return rows, nil
	}
	return f.parent.queryRows, nil
}

func (f *fakeTx) Commit() error {
	f.commitCalls++
	return nil
}

func (f *fakeTx) Rollback() error {
	f.rollbackCalls++
	return nil
}

func TestNewCapabilityRepositoryRejectsNilDB(t *testing.T) {
	t.Parallel()

	if _, err := NewCapabilityRepository(nil); err == nil {
		t.Fatalf("expected error when db executor is nil")
	}
}

func TestNewCapabilityRepositoryFromSQLDBRejectsNilDB(t *testing.T) {
	t.Parallel()

	if _, err := NewCapabilityRepositoryFromSQLDB(nil); err == nil {
		t.Fatalf("expected error when sql db is nil")
	}
}

func TestCapabilityRepositorySaveWritesCatalogColumns(t *testing.T) {
	t.Parallel()

	db := &fakeDB{}
	repo, err := NewCapabilityRepository(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry, err := capability.NewModelCatalogEntry("tenant-a", "entry-1", "model-key", "Model A", "provider", "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := repo.Save(context.Background(), entry); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if len(db.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(db.execCalls))
	}
	call := db.execCalls[0]
	if !strings.Contains(call.query, "model_catalog_entries") {
		t.Fatalf("unexpected query: %s", call.query)
	}
	if !strings.Contains(call.query, "model_key") || !strings.Contains(call.query, "display_name") {
		t.Fatalf("query missing catalog columns: %s", call.query)
	}
	if len(call.args) != 8 {
		t.Fatalf("expected 8 args, got %d", len(call.args))
	}
	if call.args[2] != "model-key" {
		t.Fatalf("unexpected model key arg: %v", call.args[2])
	}
}

func TestCapabilityRepositoryListsTenantEntries(t *testing.T) {
	t.Parallel()

	rows := newFakeRows(
		[]any{"tenant-b", "entry-2", "model-key", "Model B", "provider", "active", time.Date(2026, 3, 25, 12, 10, 0, 0, time.UTC), time.Date(2026, 3, 25, 12, 10, 0, 0, time.UTC)},
	)

	db := &fakeDB{queryRows: rows}
	repo, err := NewCapabilityRepository(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries, err := repo.ListByTenant(context.Background(), "tenant-b")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].EntryID != "entry-2" {
		t.Fatalf("unexpected entry id: %s", entries[0].EntryID)
	}
	if entries[0].ModelKey != "model-key" {
		t.Fatalf("unexpected model key: %s", entries[0].ModelKey)
	}

	if len(db.queryCalls) != 1 {
		t.Fatalf("expected 1 query call, got %d", len(db.queryCalls))
	}
	if db.queryCalls[0].args[0] != "tenant-b" {
		t.Fatalf("unexpected tenant arg: %v", db.queryCalls[0].args)
	}
}

func TestCapabilityRepositorySaveToolWritesCatalogColumns(t *testing.T) {
	t.Parallel()

	db := &fakeDB{}
	repo, err := NewCapabilityRepository(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry, err := capability.NewToolCatalogEntry("tenant-a", "tool-1", "tool-key", "Tool A", "high", "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := repo.SaveTool(context.Background(), entry); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if len(db.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(db.execCalls))
	}
	call := db.execCalls[0]
	if !strings.Contains(call.query, "tool_catalog_entries") {
		t.Fatalf("unexpected query: %s", call.query)
	}
	if !strings.Contains(call.query, "tool_key") || !strings.Contains(call.query, "risk_level") {
		t.Fatalf("query missing tool catalog columns: %s", call.query)
	}
	if len(call.args) != 8 {
		t.Fatalf("expected 8 args, got %d", len(call.args))
	}
	if call.args[2] != "tool-key" {
		t.Fatalf("unexpected tool key arg: %v", call.args[2])
	}
}

func TestCapabilityRepositoryListsTenantTools(t *testing.T) {
	t.Parallel()

	rows := newFakeRows(
		[]any{"tenant-b", "tool-2", "tool-key", "Tool B", "high", "active", time.Date(2026, 3, 25, 12, 10, 0, 0, time.UTC), time.Date(2026, 3, 25, 12, 10, 0, 0, time.UTC)},
	)

	db := &fakeDB{queryRows: rows}
	repo, err := NewCapabilityRepository(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries, err := repo.ListToolsByTenant(context.Background(), "tenant-b")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].EntryID != "tool-2" {
		t.Fatalf("unexpected entry id: %s", entries[0].EntryID)
	}
	if entries[0].ToolKey != "tool-key" {
		t.Fatalf("unexpected tool key: %s", entries[0].ToolKey)
	}

	if len(db.queryCalls) != 1 {
		t.Fatalf("expected 1 query call, got %d", len(db.queryCalls))
	}
	if db.queryCalls[0].args[0] != "tenant-b" {
		t.Fatalf("unexpected tenant arg: %v", db.queryCalls[0].args)
	}
}

func TestCapabilityRepositorySaveAgentCapabilityPolicyReplacesBindings(t *testing.T) {
	t.Parallel()

	db := &fakeDB{}
	repo, err := NewCapabilityRepository(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	policy, err := capability.NewAgentCapabilityPolicy("tenant-a", "profile-1", []string{"model-2", "model-1"}, []string{"tool-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := repo.SaveAgentCapabilityPolicy(context.Background(), policy); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if db.beginCalls != 1 {
		t.Fatalf("expected 1 transaction begin, got %d", db.beginCalls)
	}
	if db.tx == nil || db.tx.commitCalls != 1 {
		t.Fatalf("expected transaction commit once")
	}
	if len(db.execCalls) != 5 {
		t.Fatalf("expected 5 exec calls, got %d", len(db.execCalls))
	}
	if !strings.Contains(strings.ToLower(db.execCalls[0].query), "delete from agent_profile_allowed_model") {
		t.Fatalf("unexpected first query: %s", db.execCalls[0].query)
	}
	if !strings.Contains(strings.ToLower(db.execCalls[1].query), "delete from agent_profile_allowed_tool") {
		t.Fatalf("unexpected second query: %s", db.execCalls[1].query)
	}
	if !strings.Contains(strings.ToLower(db.execCalls[2].query), "insert into agent_profile_allowed_model") {
		t.Fatalf("unexpected third query: %s", db.execCalls[2].query)
	}
	if !strings.Contains(strings.ToLower(db.execCalls[4].query), "insert into agent_profile_allowed_tool") {
		t.Fatalf("unexpected final query: %s", db.execCalls[4].query)
	}
}

func TestCapabilityRepositoryGetsAgentCapabilityPolicy(t *testing.T) {
	t.Parallel()

	db := &fakeDB{
		querySets: []rowScanner{
			newFakeRows([]any{"model-1"}, []any{"model-2"}),
			newFakeRows([]any{"tool-1"}),
		},
	}
	repo, err := NewCapabilityRepository(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	policy, err := repo.GetAgentCapabilityPolicy(context.Background(), "tenant-a", "profile-1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if policy.TenantID != "tenant-a" || policy.AgentProfileID != "profile-1" {
		t.Fatalf("unexpected policy identity: %+v", policy)
	}
	if !reflect.DeepEqual(policy.AllowedModelEntryIDs, []string{"model-1", "model-2"}) {
		t.Fatalf("unexpected model ids: %v", policy.AllowedModelEntryIDs)
	}
	if !reflect.DeepEqual(policy.AllowedToolEntryIDs, []string{"tool-1"}) {
		t.Fatalf("unexpected tool ids: %v", policy.AllowedToolEntryIDs)
	}
	if len(db.queryCalls) != 2 {
		t.Fatalf("expected 2 query calls, got %d", len(db.queryCalls))
	}
}

type fakeRows struct {
	data [][]any
	idx  int
}

func newFakeRows(rows ...[]any) *fakeRows {
	return &fakeRows{data: rows}
}

func (f *fakeRows) Close() error {
	return nil
}

func (f *fakeRows) Next() bool {
	if f.idx >= len(f.data) {
		return false
	}
	f.idx++
	return true
}

func (f *fakeRows) Scan(dest ...any) error {
	if f.idx == 0 || f.idx > len(f.data) {
		return fmt.Errorf("no row to scan")
	}
	row := f.data[f.idx-1]
	if len(dest) != len(row) {
		return fmt.Errorf("column mismatch: want %d got %d", len(row), len(dest))
	}
	for i := range dest {
		switch d := dest[i].(type) {
		case *string:
			val, _ := row[i].(string)
			*d = val
		case *time.Time:
			if v, ok := row[i].(time.Time); ok {
				*d = v
			}
		default:
			return fmt.Errorf("unsupported scan destination of type %T", dest[i])
		}
	}
	return nil
}

func (f *fakeRows) Err() error {
	return nil
}
