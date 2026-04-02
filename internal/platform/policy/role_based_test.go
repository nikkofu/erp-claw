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

func TestRoleEvaluatorDeniesApprovalCommandsWithoutRequiredRoles(t *testing.T) {
	evaluator := NewRoleEvaluator(
		func(_ context.Context, _, _ string) ([]string, error) {
			return []string{"viewer"}, nil
		},
		[]Rule{
			{CommandPrefix: "runtime.tasks.pause", AnyOfRoles: []string{"platform_admin", "workspace_operator"}},
			{CommandPrefix: "runtime.tasks.resume", AnyOfRoles: []string{"platform_admin", "workspace_operator"}},
			{CommandPrefix: "runtime.tasks.handoff", AnyOfRoles: []string{"platform_admin", "workspace_operator"}},
		},
	)

	for _, command := range []string{"runtime.tasks.pause", "runtime.tasks.resume", "runtime.tasks.handoff"} {
		decision, err := evaluator.Evaluate(context.Background(), Input{
			CommandName: command,
			TenantID:    "tenant-admin",
			ActorID:     "actor-viewer",
		})
		if err != nil {
			t.Fatalf("evaluate role policy for %s: %v", command, err)
		}
		if decision != DecisionDeny {
			t.Fatalf("expected decision deny for %s, got %s", command, decision)
		}
	}
}

func TestRoleEvaluatorCommandRuleMatchesExactCommandName(t *testing.T) {
	evaluator := NewRoleEvaluator(
		func(_ context.Context, _, _ string) ([]string, error) {
			return []string{"workspace_operator"}, nil
		},
		[]Rule{{
			CommandPrefix: "runtime.tasks.pause",
			AnyOfRoles:    []string{"workspace_operator"},
		}},
	)

	decision, err := evaluator.Evaluate(context.Background(), Input{
		CommandName: "runtime.tasks.pause.extra",
		TenantID:    "tenant-admin",
		ActorID:     "actor-operator",
	})
	if err != nil {
		t.Fatalf("evaluate role policy: %v", err)
	}
	if decision != DecisionAllow {
		t.Fatalf("expected allow for non-exact command match, got %s", decision)
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

func TestRoleEvaluatorAllowsTenantOverrideRule(t *testing.T) {
	evaluator := NewRoleEvaluatorWithTenantRules(
		func(_ context.Context, _, _ string) ([]string, error) {
			return []string{"viewer"}, nil
		},
		[]Rule{{
			CommandPrefix: "masterdata.",
			AnyOfRoles:    []string{"platform_admin"},
		}},
		func(_ context.Context, _ string) ([]Rule, error) {
			return []Rule{{
				CommandPrefix: "masterdata.",
				AnyOfRoles:    []string{"viewer"},
			}}, nil
		},
	)

	decision, err := evaluator.Evaluate(context.Background(), Input{
		CommandName: "masterdata.suppliers.create",
		TenantID:    "tenant-a",
		ActorID:     "actor-viewer",
	})
	if err != nil {
		t.Fatalf("evaluate role policy: %v", err)
	}
	if decision != DecisionAllow {
		t.Fatalf("expected decision allow from tenant override, got %s", decision)
	}
}
