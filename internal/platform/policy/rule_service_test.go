package policy

import (
	"context"
	"testing"
)

func TestNewRuleServiceRequiresRepository(t *testing.T) {
	_, err := NewRuleService(nil)
	if err == nil {
		t.Fatal("expected nil repository to fail")
	}
}

func TestRuleServiceUpsertAndList(t *testing.T) {
	repo := NewInMemoryRuleRepository()
	service, err := NewRuleService(repo)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	rule, err := service.Upsert(context.Background(), Rule{
		TenantID:    "tenant-a",
		ID:          "rule-1",
		CommandName: "purchase.approve",
		ActorID:     "*",
		Decision:    DecisionRequireApproval,
		Priority:    50,
		Active:      true,
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if rule.ID == "" {
		t.Fatal("expected rule id to be set")
	}

	rules, err := service.List(context.Background(), RuleFilter{
		TenantID: "tenant-a",
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
}

func TestRuleServiceActivateAndDeactivate(t *testing.T) {
	repo := NewInMemoryRuleRepository()
	service, err := NewRuleService(repo)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.Upsert(context.Background(), Rule{
		TenantID:    "tenant-a",
		ID:          "rule-1",
		CommandName: "purchase.approve",
		ActorID:     "*",
		Decision:    DecisionRequireApproval,
		Priority:    50,
		Active:      true,
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	rule, err := service.Deactivate(context.Background(), "tenant-a", "rule-1")
	if err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if rule.Active {
		t.Fatal("expected rule to be inactive")
	}

	rule, err = service.Activate(context.Background(), "tenant-a", "rule-1")
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !rule.Active {
		t.Fatal("expected rule to be active")
	}
}
