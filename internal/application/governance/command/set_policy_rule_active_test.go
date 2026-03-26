package command

import (
	"context"
	"testing"

	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestSetPolicyRuleActiveHandlerRequiresRulesService(t *testing.T) {
	handler := SetPolicyRuleActiveHandler{}

	_, err := handler.Handle(context.Background(), SetPolicyRuleActive{
		TenantID: "tenant-a",
		RuleID:   "rule-a",
		Active:   false,
	})
	if err == nil {
		t.Fatal("expected nil rules service to fail")
	}
}

func TestSetPolicyRuleActiveHandlerUsesAuthorizeAndAuditHooks(t *testing.T) {
	rules := &stubRuleActivationManager{
		rule: policy.Rule{
			TenantID:    "tenant-a",
			ID:          "rule-a",
			CommandName: "purchase.approve",
			ActorID:     "*",
			Decision:    policy.DecisionRequireApproval,
			Priority:    50,
			Active:      false,
		},
	}

	authorizeCalled := false
	auditCalled := false
	handler := SetPolicyRuleActiveHandler{
		Rules: rules,
		Authorize: func(_ context.Context, cmd SetPolicyRuleActive) error {
			authorizeCalled = true
			if cmd.TenantID != "tenant-a" {
				t.Fatalf("unexpected tenant id: %s", cmd.TenantID)
			}
			if cmd.RuleID != "rule-a" {
				t.Fatalf("unexpected rule id: %s", cmd.RuleID)
			}
			if cmd.Active {
				t.Fatal("expected deactivate command")
			}
			return nil
		},
		Audit: func(_ context.Context, rule policy.Rule) error {
			auditCalled = true
			if rule.Active {
				t.Fatal("expected rule to be inactive")
			}
			return nil
		},
	}

	out, err := handler.Handle(context.Background(), SetPolicyRuleActive{
		TenantID: "tenant-a",
		RuleID:   "rule-a",
		Active:   false,
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
	if rules.tenantID != "tenant-a" || rules.ruleID != "rule-a" || rules.active {
		t.Fatalf("unexpected activation call: tenant=%s rule=%s active=%t", rules.tenantID, rules.ruleID, rules.active)
	}
}

type stubRuleActivationManager struct {
	rule     policy.Rule
	tenantID string
	ruleID   string
	active   bool
}

func (s *stubRuleActivationManager) SetActive(_ context.Context, tenantID, ruleID string, active bool) (policy.Rule, error) {
	s.tenantID = tenantID
	s.ruleID = ruleID
	s.active = active
	s.rule.Active = active
	return s.rule, nil
}
