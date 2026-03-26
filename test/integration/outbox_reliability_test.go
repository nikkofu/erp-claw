package integration

import (
	"os"
	"strings"
	"testing"
)

func TestOutboxReliabilityMigrationContract(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../migrations/000006_phase1_reliability_hardening.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}

	sql := strings.ToLower(string(data))
	required := []string{
		"alter table outbox add column if not exists attempts integer not null default 0",
		"alter table outbox add column if not exists last_error text",
		"alter table outbox add column if not exists processing_at timestamptz",
		"create index if not exists idx_outbox_pending_available on outbox(status, available_at, id)",
	}

	for _, needle := range required {
		if !strings.Contains(sql, needle) {
			t.Fatalf("expected migration to contain %q", needle)
		}
	}
}
