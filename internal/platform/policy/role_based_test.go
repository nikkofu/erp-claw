package policy

import (
	"context"
	"errors"
	"testing"
)

func TestRoleEvaluatorAllowsWhenActorHasRequiredRole(t *testing.T) {
	evaluator := NewRoleEvaluator(
		func(_ context.Context, _, actorID string) ([]string, error) {
			if actorID == "actor-admin" {
				return []string{"platform_admin"}, nil
			}
			return nil, nil
		},
		[]Rule{{
			CommandPrefix: "procurement.",
			AnyOfRoles:    []string{"platform_admin", "supplychain_operator"},
		}},
	)

	decision, err := evaluator.Evaluate(context.Background(), Input{
		CommandName: "procurement.purchase_orders.create",
		TenantID:    "tenant-admin",
		ActorID:     "actor-admin",
	})
	if err != nil {
		t.Fatalf("evaluate role policy: %v", err)
	}
	if decision != DecisionAllow {
		t.Fatalf("expected decision allow, got %s", decision)
	}
}

func TestRoleEvaluatorDeniesWhenRoleIsMissing(t *testing.T) {
	evaluator := NewRoleEvaluator(
		func(_ context.Context, _, _ string) ([]string, error) {
			return []string{"viewer"}, nil
		},
		[]Rule{{
			CommandPrefix: "procurement.",
			AnyOfRoles:    []string{"platform_admin", "supplychain_operator"},
		}},
	)

	decision, err := evaluator.Evaluate(context.Background(), Input{
		CommandName: "procurement.purchase_orders.create",
		TenantID:    "tenant-admin",
		ActorID:     "actor-viewer",
	})
	if err != nil {
		t.Fatalf("evaluate role policy: %v", err)
	}
	if decision != DecisionDeny {
		t.Fatalf("expected decision deny, got %s", decision)
	}
}

func TestRoleEvaluatorReturnsErrorWhenLookupFails(t *testing.T) {
	expectedErr := errors.New("lookup failed")
	evaluator := NewRoleEvaluator(
		func(_ context.Context, _, _ string) ([]string, error) {
			return nil, expectedErr
		},
		[]Rule{{
			CommandPrefix: "procurement.",
			AnyOfRoles:    []string{"platform_admin"},
		}},
	)

	_, err := evaluator.Evaluate(context.Background(), Input{
		CommandName: "procurement.purchase_orders.create",
		TenantID:    "tenant-admin",
		ActorID:     "actor-admin",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected lookup error, got %v", err)
	}
}
