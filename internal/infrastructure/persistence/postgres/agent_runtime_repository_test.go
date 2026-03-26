package postgres

import (
	"context"
	"database/sql"
	"testing"

	application "github.com/nikkofu/erp-claw/internal/application/agentruntime"
	"github.com/nikkofu/erp-claw/internal/domain/agentruntime"
)

var (
	_ agentruntime.SessionRepository = (*AgentRuntimeRepository)(nil)
	_ agentruntime.TaskRepository    = (*AgentRuntimeRepository)(nil)
	_ application.SessionReader      = (*AgentRuntimeRepository)(nil)
	_ application.TaskReader         = (*AgentRuntimeRepository)(nil)
)

func TestNewAgentRuntimeRepositoryRejectsNilDB(t *testing.T) {
	_, err := NewAgentRuntimeRepository(nil)
	if err == nil {
		t.Fatal("expected nil db to fail")
	}
}

func TestAgentRuntimeRepositoryCreateSessionRejectsNonNumericTenantID(t *testing.T) {
	db := openAgentRuntimeTestDB(t)
	repo, err := NewAgentRuntimeRepository(db)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	session, err := agentruntime.NewSession("tenant-a", "session-a", nil)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	_, err = repo.CreateSession(context.Background(), session)
	if err == nil {
		t.Fatal("expected create session to reject non-numeric tenant id")
	}
}

func TestAgentRuntimeRepositoryGetTaskByIDRejectsNonNumericTaskID(t *testing.T) {
	db := openAgentRuntimeTestDB(t)
	repo, err := NewAgentRuntimeRepository(db)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	_, err = repo.GetTaskByID(context.Background(), "1001", "task-a")
	if err == nil {
		t.Fatal("expected get task by id to reject non-numeric task id")
	}
}

func TestAgentRuntimeRepositoryListSessionsRejectsNonNumericTenantID(t *testing.T) {
	db := openAgentRuntimeTestDB(t)
	repo, err := NewAgentRuntimeRepository(db)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	_, err = repo.ListSessions(context.Background(), "tenant-a")
	if err == nil {
		t.Fatal("expected list sessions to reject non-numeric tenant id")
	}
}

func TestAgentRuntimeRepositoryListTasksRejectsNonNumericSessionID(t *testing.T) {
	db := openAgentRuntimeTestDB(t)
	repo, err := NewAgentRuntimeRepository(db)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	_, err = repo.ListTasks(context.Background(), "1001", "session-a")
	if err == nil {
		t.Fatal("expected list tasks to reject non-numeric session id")
	}
}

func openAgentRuntimeTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", "postgres://invalid:invalid@127.0.0.1:1/erp_claw_test?sslmode=disable")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
