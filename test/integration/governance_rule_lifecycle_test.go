package integration

import (
	"context"
	"testing"

	governancecommand "github.com/nikkofu/erp-claw/internal/application/governance/command"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestGovernanceRuleLifecycleDeactivateThenActivate(t *testing.T) {
	repo := policy.NewInMemoryRuleRepository()
	ruleService, err := policy.NewRuleService(repo)
	if err != nil {
		t.Fatalf("new rule service: %v", err)
	}

	upsertHandler := governancecommand.UpsertPolicyRuleHandler{Rules: ruleService}
	setActiveHandler := governancecommand.SetPolicyRuleActiveHandler{Rules: ruleService}
	evaluator := policy.NewRuleEvaluator(repo, policy.DecisionAllow)

	_, err = upsertHandler.Handle(context.Background(), governancecommand.UpsertPolicyRule{
		TenantID:    "tenant-a",
		ID:          "rule-deny-adjust",
		CommandName: "inventory.adjust",
		ActorID:     "*",
		Decision:    policy.DecisionDeny,
		Priority:    100,
		Active:      true,
	})
	if err != nil {
		t.Fatalf("upsert policy rule: %v", err)
	}

	decision, err := evaluator.Evaluate(context.Background(), policy.Input{
		CommandName: "inventory.adjust",
		TenantID:    "tenant-a",
		ActorID:     "actor-a",
	})
	if err != nil {
		t.Fatalf("evaluate active rule: %v", err)
	}
	if decision != policy.DecisionDeny {
		t.Fatalf("expected deny before deactivation, got %s", decision)
	}

	_, err = setActiveHandler.Handle(context.Background(), governancecommand.SetPolicyRuleActive{
		TenantID: "tenant-a",
		RuleID:   "rule-deny-adjust",
		Active:   false,
	})
	if err != nil {
		t.Fatalf("deactivate policy rule: %v", err)
	}

	decision, err = evaluator.Evaluate(context.Background(), policy.Input{
		CommandName: "inventory.adjust",
		TenantID:    "tenant-a",
		ActorID:     "actor-a",
	})
	if err != nil {
		t.Fatalf("evaluate inactive rule: %v", err)
	}
	if decision != policy.DecisionAllow {
		t.Fatalf("expected fallback allow after deactivation, got %s", decision)
	}

	_, err = setActiveHandler.Handle(context.Background(), governancecommand.SetPolicyRuleActive{
		TenantID: "tenant-a",
		RuleID:   "rule-deny-adjust",
		Active:   true,
	})
	if err != nil {
		t.Fatalf("activate policy rule: %v", err)
	}

	decision, err = evaluator.Evaluate(context.Background(), policy.Input{
		CommandName: "inventory.adjust",
		TenantID:    "tenant-a",
		ActorID:     "actor-a",
	})
	if err != nil {
		t.Fatalf("evaluate reactivated rule: %v", err)
	}
	if decision != policy.DecisionDeny {
		t.Fatalf("expected deny after activation, got %s", decision)
	}
}
