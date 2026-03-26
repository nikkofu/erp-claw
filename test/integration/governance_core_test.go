package integration

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestGovernanceCoreMigrationContainsPolicyRuleAndAuditEventTables(t *testing.T) {
	data, err := os.ReadFile("../../migrations/000003_phase1_governance_core.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := strings.ToLower(string(data))

	required := []string{
		"create table if not exists policy_rule",
		"create table if not exists audit_event",
	}
	for _, needle := range required {
		if !strings.Contains(sql, needle) {
			t.Fatalf("expected migration to contain %q", needle)
		}
	}

	policyRule := mustGovernanceTableBlock(t, sql, "policy_rule")
	requireGovernanceContainsAll(t, policyRule, []string{
		"tenant_id text not null",
		"id text not null",
		"command_name text not null",
		"actor_id text not null",
		"decision text not null",
	})

	auditEvent := mustGovernanceTableBlock(t, sql, "audit_event")
	requireGovernanceContainsAll(t, auditEvent, []string{
		"tenant_id text not null",
		"id text not null",
		"command_name text not null",
		"actor_id text not null",
		"decision text not null",
		"outcome text not null",
		"occurred_at timestamptz not null",
	})
}

func TestGovernanceCorePolicyAndAuditSeams(t *testing.T) {
	rules := policy.NewInMemoryRuleRepository()
	_, err := rules.UpsertRule(context.Background(), policy.Rule{
		TenantID:    "tenant-a",
		ID:          "deny-all-adjust",
		CommandName: "inventory.adjust",
		ActorID:     "*",
		Decision:    policy.DecisionDeny,
		Priority:    100,
		Active:      true,
	})
	if err != nil {
		t.Fatalf("upsert rule: %v", err)
	}

	auditStore := audit.NewInMemoryStore()
	auditService, err := audit.NewService(auditStore)
	if err != nil {
		t.Fatalf("new audit service: %v", err)
	}

	pipeline := shared.NewPipeline(shared.PipelineDeps{
		Policy: policy.NewRuleEvaluator(rules, policy.DecisionAllow),
		Audit:  auditService,
	})

	err = pipeline.Execute(context.Background(), shared.Command{
		Name:     "inventory.adjust",
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Payload:  map[string]any{"sku": "A-1", "delta": 3},
	})
	if err == nil {
		t.Fatal("expected denied policy to block command")
	}

	events, err := auditService.List(context.Background(), audit.Query{
		TenantID:    "tenant-a",
		CommandName: "inventory.adjust",
		ActorID:     "actor-a",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(events))
	}
	if events[0].Decision != policy.DecisionDeny {
		t.Fatalf("expected deny decision, got %s", events[0].Decision)
	}
	if events[0].Outcome != "rejected" {
		t.Fatalf("expected rejected outcome, got %s", events[0].Outcome)
	}
}

func mustGovernanceTableBlock(t *testing.T, migrationSQL, table string) string {
	t.Helper()

	startNeedle := "create table if not exists " + table + " ("
	startIdx := strings.Index(migrationSQL, startNeedle)
	if startIdx == -1 {
		t.Fatalf("expected table block %q", table)
	}
	block := migrationSQL[startIdx:]

	endIdx := strings.Index(block, ");")
	if endIdx == -1 {
		t.Fatalf("expected table block %q to end with );", table)
	}

	return block[:endIdx]
}

func requireGovernanceContainsAll(t *testing.T, block string, required []string) {
	t.Helper()
	for _, needle := range required {
		if !strings.Contains(block, needle) {
			t.Fatalf("expected table block to contain %q", needle)
		}
	}
}
