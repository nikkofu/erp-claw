package integration

import (
	"context"
	"testing"
	"time"

	governancecommand "github.com/nikkofu/erp-claw/internal/application/governance/command"
	governancequery "github.com/nikkofu/erp-claw/internal/application/governance/query"
	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestGovernanceApplicationHandlersManagePolicyRulesAndAuditQueries(t *testing.T) {
	ruleRepo := policy.NewInMemoryRuleRepository()
	ruleService, err := policy.NewRuleService(ruleRepo)
	if err != nil {
		t.Fatalf("new rule service: %v", err)
	}
	auditStore := audit.NewInMemoryStore()
	auditService, err := audit.NewService(auditStore)
	if err != nil {
		t.Fatalf("new audit service: %v", err)
	}

	upsertHandler := governancecommand.UpsertPolicyRuleHandler{Rules: ruleService}
	listRulesHandler := governancequery.ListPolicyRulesHandler{Rules: ruleService}
	listAuditHandler := governancequery.ListAuditEventsHandler{AuditEvents: auditService}

	upserted, err := upsertHandler.Handle(context.Background(), governancecommand.UpsertPolicyRule{
		TenantID:    "tenant-a",
		ID:          "rule-a",
		CommandName: "purchase.approve",
		ActorID:     "*",
		Decision:    policy.DecisionRequireApproval,
		Priority:    90,
		Active:      true,
	})
	if err != nil {
		t.Fatalf("upsert policy rule: %v", err)
	}
	if upserted.ID == "" {
		t.Fatal("expected rule id")
	}

	rules, err := listRulesHandler.Handle(context.Background(), governancequery.ListPolicyRules{
		TenantID:    "tenant-a",
		CommandName: "purchase.approve",
		ActiveOnly:  true,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("list policy rules: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected one rule, got %d", len(rules))
	}
	if rules[0].Decision != policy.DecisionRequireApproval {
		t.Fatalf("unexpected rule decision: %s", rules[0].Decision)
	}

	err = auditService.Record(context.Background(), audit.Record{
		ID:          "evt-1",
		TenantID:    "tenant-a",
		CommandName: "purchase.approve",
		ActorID:     "actor-a",
		Decision:    policy.DecisionRequireApproval,
		Outcome:     "pending_approval",
		OccurredAt:  time.Date(2026, 3, 25, 9, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("record audit event: %v", err)
	}

	events, err := listAuditHandler.Handle(context.Background(), governancequery.ListAuditEvents{
		TenantID:       "tenant-a",
		CommandName:    "purchase.approve",
		ActorID:        "actor-a",
		OccurredAfter:  time.Date(2026, 3, 25, 8, 59, 0, 0, time.UTC),
		OccurredBefore: time.Date(2026, 3, 25, 9, 1, 0, 0, time.UTC),
		Limit:          5,
	})
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(events))
	}
	if events[0].Outcome != "pending_approval" {
		t.Fatalf("unexpected audit outcome: %s", events[0].Outcome)
	}
}
