package integration

import (
	"os"
	"strings"
	"testing"
)

func TestCapabilityGovernanceMigrationDefinesModelCatalogTable(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("../../migrations/000005_phase1_capability_governance.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	content := strings.ToLower(string(raw))

	if !strings.Contains(content, "create table if not exists model_catalog_entries") {
		t.Fatalf("migration missing model_catalog_entries table definition")
	}

	requiredColumns := []string{
		"tenant_id text not null",
		"entry_id text not null",
		"model_key text not null",
		"display_name text not null",
		"provider text not null",
		"status text not null",
	}
	for _, col := range requiredColumns {
		if !strings.Contains(content, col) {
			t.Fatalf("column %q missing from migration", col)
		}
	}

	if !strings.Contains(content, "primary key (tenant_id, entry_id)") {
		t.Fatalf("migration should declare composite primary key")
	}
}

func TestCapabilityGovernanceMigrationDefinesToolCatalogTable(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("../../migrations/000008_phase1_tool_catalog.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	content := strings.ToLower(string(raw))

	if !strings.Contains(content, "create table if not exists tool_catalog_entries") {
		t.Fatalf("migration missing tool_catalog_entries table definition")
	}

	requiredColumns := []string{
		"tenant_id text not null",
		"entry_id text not null",
		"tool_key text not null",
		"display_name text not null",
		"risk_level text not null",
		"status text not null",
	}
	for _, col := range requiredColumns {
		if !strings.Contains(content, col) {
			t.Fatalf("column %q missing from migration", col)
		}
	}

	if !strings.Contains(content, "primary key (tenant_id, entry_id)") {
		t.Fatalf("migration should declare composite primary key")
	}
}

func TestCapabilityGovernanceMigrationDefinesAgentCapabilityPolicyTables(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("../../migrations/000010_phase1_agent_capability_policy.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	content := strings.ToLower(string(raw))

	requiredTables := []string{
		"create table if not exists agent_profile_allowed_model",
		"create table if not exists agent_profile_allowed_tool",
	}
	for _, table := range requiredTables {
		if !strings.Contains(content, table) {
			t.Fatalf("migration missing %q table definition", table)
		}
	}

	requiredSnippets := []string{
		"agent_profile_id text not null",
		"model_entry_id text not null",
		"tool_entry_id text not null",
		"primary key (tenant_id, agent_profile_id, model_entry_id)",
		"primary key (tenant_id, agent_profile_id, tool_entry_id)",
		"foreign key (tenant_id, agent_profile_id) references agent_profile(tenant_id, id) on delete cascade",
		"foreign key (tenant_id, model_entry_id) references model_catalog_entries(tenant_id, entry_id) on delete cascade",
		"foreign key (tenant_id, tool_entry_id) references tool_catalog_entries(tenant_id, entry_id) on delete cascade",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(content, snippet) {
			t.Fatalf("migration missing %q", snippet)
		}
	}
}
