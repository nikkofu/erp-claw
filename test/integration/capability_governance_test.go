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
