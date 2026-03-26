package command

import (
	"context"
	"testing"

	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestUpsertPolicyRuleHandlerRequiresRulesService(t *testing.T) {
	handler := UpsertPolicyRuleHandler{}

	_, err := handler.Handle(context.Background(), UpsertPolicyRule{
		TenantID:    "tenant-a",
		ID:          "rule-a",
		CommandName: "purchase.approve",
		ActorID:     "actor-a",
		Decision:    policy.DecisionAllow,
		Priority:    10,
		Active:      true,
	})
	if err == nil {
		t.Fatal("expected nil rules service to fail")
	}
}

func TestUpsertPolicyRuleHandlerUsesAuthorizeAndAuditHooks(t *testing.T) {
	rules := &stubRuleManager{
		rule: policy.Rule{
			TenantID:    "tenant-a",
			ID:          "rule-a",
			CommandName: "purchase.approve",
			ActorID:     "actor-a",
			Decision:    policy.DecisionRequireApproval,
			Priority:    50,
			Active:      true,
		},
	}

	authorizeCalled := false
	auditCalled := false
	handler := UpsertPolicyRuleHandler{
		Rules: rules,
		Authorize: func(_ context.Context, cmd UpsertPolicyRule) error {
			authorizeCalled = true
			if cmd.TenantID != "tenant-a" {
				t.Fatalf("unexpected tenant id: %s", cmd.TenantID)
			}
			return nil
		},
		Audit: func(_ context.Context, rule policy.Rule) error {
			auditCalled = true
			if rule.Decision != policy.DecisionRequireApproval {
				t.Fatalf("unexpected decision: %s", rule.Decision)
			}
			return nil
		},
	}

	out, err := handler.Handle(context.Background(), UpsertPolicyRule{
		TenantID:    "tenant-a",
		ID:          "rule-a",
		CommandName: "purchase.approve",
		ActorID:     "actor-a",
		Decision:    policy.DecisionRequireApproval,
		Priority:    50,
		Active:      true,
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !authorizeCalled {
		t.Fatal("expected authorize hook to be called")
	}
	if !auditCalled {
		t.Fatal("expected audit hook to be called")
	}
	if out.ID != "rule-a" {
		t.Fatalf("unexpected rule id: %s", out.ID)
	}
}

type stubRuleManager struct {
	rule policy.Rule
}

func (s *stubRuleManager) Upsert(_ context.Context, _ policy.Rule) (policy.Rule, error) {
	return s.rule, nil
}

func (s *stubRuleManager) List(_ context.Context, _ policy.RuleFilter) ([]policy.Rule, error) {
	return []policy.Rule{s.rule}, nil
}
