package postgres

import (
	"strings"
	"testing"
	"time"

	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

var (
	_ policy.RuleRepository = (*PolicyAuditRepository)(nil)
	_ audit.EventStore      = (*PolicyAuditRepository)(nil)
	_ audit.Recorder        = (*PolicyAuditRepository)(nil)
)

func TestNewPolicyAuditRepositoryRejectsNilDB(t *testing.T) {
	_, err := NewPolicyAuditRepository(nil)
	if err == nil {
		t.Fatal("expected nil db to fail")
	}
}

func TestBuildAuditListQueryIncludesTimeWindowFilters(t *testing.T) {
	after := time.Date(2026, 3, 25, 8, 1, 0, 0, time.UTC)
	before := time.Date(2026, 3, 25, 8, 3, 0, 0, time.UTC)

	sql, args, ok := buildAuditListQuery(audit.Query{
		TenantID:       "tenant-a",
		CommandName:    "purchase.approve",
		ActorID:        "actor-a",
		OccurredAfter:  after,
		OccurredBefore: before,
		Limit:          5,
	})
	if !ok {
		t.Fatal("expected query to be buildable")
	}

	requiredClauses := []string{
		"where tenant_id = $1",
		"and command_name = $2",
		"and actor_id = $3",
		"and occurred_at >= $4",
		"and occurred_at < $5",
		"limit $6",
	}
	for _, clause := range requiredClauses {
		if !strings.Contains(sql, clause) {
			t.Fatalf("expected clause %q in sql: %s", clause, sql)
		}
	}

	if len(args) != 6 {
		t.Fatalf("expected 6 args, got %d", len(args))
	}
	if args[0] != "tenant-a" || args[1] != "purchase.approve" || args[2] != "actor-a" {
		t.Fatalf("unexpected first args: %#v", args[:3])
	}
}
