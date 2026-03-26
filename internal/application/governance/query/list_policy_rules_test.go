package query

import (
	"context"
	"testing"

	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestListPolicyRulesHandlerRequiresRulesService(t *testing.T) {
	handler := ListPolicyRulesHandler{}

	_, err := handler.Handle(context.Background(), ListPolicyRules{TenantID: "tenant-a"})
	if err == nil {
		t.Fatal("expected nil rules service to fail")
	}
}

func TestListPolicyRulesHandlerUsesAuthorizeAndAuditHooks(t *testing.T) {
	rules := &stubRuleReader{
		items: []policy.Rule{
			{
				TenantID:    "tenant-a",
				ID:          "rule-a",
				CommandName: "purchase.approve",
				ActorID:     "*",
				Decision:    policy.DecisionRequireApproval,
				Priority:    100,
				Active:      true,
			},
		},
	}

	authorizeCalled := false
	auditCalled := false
	handler := ListPolicyRulesHandler{
		Rules: rules,
		Authorize: func(_ context.Context, q ListPolicyRules) error {
			authorizeCalled = true
			if q.TenantID != "tenant-a" {
				t.Fatalf("unexpected tenant id: %s", q.TenantID)
			}
			return nil
		},
		Audit: func(_ context.Context, got []policy.Rule) error {
			auditCalled = true
			if len(got) != 1 {
				t.Fatalf("expected one rule, got %d", len(got))
			}
			return nil
		},
	}

	out, err := handler.Handle(context.Background(), ListPolicyRules{
		TenantID:    "tenant-a",
		CommandName: "purchase.approve",
		ActiveOnly:  true,
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
	if len(out) != 1 {
		t.Fatalf("expected one rule, got %d", len(out))
	}
}

type stubRuleReader struct {
	items []policy.Rule
}

func (s *stubRuleReader) List(_ context.Context, _ policy.RuleFilter) ([]policy.Rule, error) {
	return s.items, nil
}
