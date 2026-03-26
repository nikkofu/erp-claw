package policy

import (
	"context"
	"testing"
)

func TestRuleEvaluatorMatchesTenantCommandActorWithPriority(t *testing.T) {
	repo := NewInMemoryRuleRepository()
	_, err := repo.UpsertRule(context.Background(), Rule{
		TenantID:    "tenant-a",
		ID:          "rule-low",
		CommandName: "inventory.adjust",
		ActorID:     "*",
		Decision:    DecisionAllow,
		Priority:    10,
		Active:      true,
	})
	if err != nil {
		t.Fatalf("upsert low rule: %v", err)
	}
	_, err = repo.UpsertRule(context.Background(), Rule{
		TenantID:    "tenant-a",
		ID:          "rule-high",
		CommandName: "inventory.adjust",
		ActorID:     "actor-a",
		Decision:    DecisionDeny,
		Priority:    100,
		Active:      true,
	})
	if err != nil {
		t.Fatalf("upsert high rule: %v", err)
	}

	evaluator := NewRuleEvaluator(repo, DecisionAllow)
	decision, err := evaluator.Evaluate(context.Background(), Input{
		CommandName: "inventory.adjust",
		TenantID:    "tenant-a",
		ActorID:     "actor-a",
	})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}

	if decision != DecisionDeny {
		t.Fatalf("expected deny decision, got %s", decision)
	}
}

func TestRuleEvaluatorFallsBackWhenNoRuleMatches(t *testing.T) {
	repo := NewInMemoryRuleRepository()
	evaluator := NewRuleEvaluator(repo, DecisionRequireApproval)

	decision, err := evaluator.Evaluate(context.Background(), Input{
		CommandName: "inventory.adjust",
		TenantID:    "tenant-a",
		ActorID:     "actor-a",
	})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if decision != DecisionRequireApproval {
		t.Fatalf("expected fallback decision, got %s", decision)
	}
}
