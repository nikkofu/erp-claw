package policy

import (
	"context"
	"errors"
	"testing"
)

func TestInMemoryRuleStoreUpsertsAndListsRulesByTenant(t *testing.T) {
	store := NewInMemoryRuleStore()

	if err := store.Upsert(context.Background(), "tenant-a", Rule{
		CommandPrefix: "masterdata.",
		AnyOfRoles:    []string{"viewer"},
	}); err != nil {
		t.Fatalf("upsert first rule: %v", err)
	}
	if err := store.Upsert(context.Background(), "tenant-a", Rule{
		CommandPrefix: "procurement.",
		AnyOfRoles:    []string{"buyer"},
	}); err != nil {
		t.Fatalf("upsert second rule: %v", err)
	}

	rules, err := store.List(context.Background(), "tenant-a")
	if err != nil {
		t.Fatalf("list rules: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
}

func TestInMemoryRuleStoreRejectsInvalidRule(t *testing.T) {
	store := NewInMemoryRuleStore()

	err := store.Upsert(context.Background(), "tenant-a", Rule{
		CommandPrefix: "",
		AnyOfRoles:    []string{"viewer"},
	})
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("expected ErrInvalidRule, got %v", err)
	}
}
