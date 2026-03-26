package integration

import (
	"os"
	"strings"
	"testing"
)

func TestApprovalBaselineMigrationContainsDefinitionInstanceAndTaskTables(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("../../migrations/000009_phase1_approval_baseline.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	content := strings.ToLower(string(raw))

	requiredTables := []string{
		"create table if not exists approval_definition",
		"create table if not exists approval_instance",
		"create table if not exists approval_task",
	}
	for _, needle := range requiredTables {
		if !strings.Contains(content, needle) {
			t.Fatalf("migration missing %q", needle)
		}
	}

	definition := mustGovernanceTableBlock(t, content, "approval_definition")
	requireContainsAll(t, definition, []string{
		"tenant_id text not null",
		"id text not null",
		"name text not null",
		"approver_id text not null",
		"active boolean not null default true",
	})

	instance := mustGovernanceTableBlock(t, content, "approval_instance")
	requireContainsAll(t, instance, []string{
		"tenant_id text not null",
		"id text not null",
		"definition_id text not null",
		"resource_type text not null",
		"resource_id text not null",
		"requested_by text not null",
		"status text not null",
	})

	task := mustGovernanceTableBlock(t, content, "approval_task")
	requireContainsAll(t, task, []string{
		"tenant_id text not null",
		"id text not null",
		"instance_id text not null",
		"approver_id text not null",
		"status text not null",
		"decided_by text",
		"comment text",
	})
}
