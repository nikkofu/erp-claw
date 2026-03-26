package command

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

var errSetPolicyRuleActiveHandlerRulesRequired = errors.New("set policy rule active handler requires rules service")

type RuleActivationManager interface {
	SetActive(ctx context.Context, tenantID, ruleID string, active bool) (policy.Rule, error)
}

type SetPolicyRuleActive struct {
	TenantID string
	RuleID   string
	Active   bool
}

type SetPolicyRuleActiveHandler struct {
	Rules     RuleActivationManager
	Authorize func(context.Context, SetPolicyRuleActive) error
	Audit     func(context.Context, policy.Rule) error
}

func (h SetPolicyRuleActiveHandler) Handle(ctx context.Context, cmd SetPolicyRuleActive) (policy.Rule, error) {
	if h.Rules == nil {
		return policy.Rule{}, errSetPolicyRuleActiveHandlerRulesRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, cmd); err != nil {
			return policy.Rule{}, err
		}
	}

	rule, err := h.Rules.SetActive(ctx, cmd.TenantID, cmd.RuleID, cmd.Active)
	if err != nil {
		return policy.Rule{}, err
	}

	if h.Audit != nil {
		if err := h.Audit(ctx, rule); err != nil {
			return policy.Rule{}, err
		}
	}

	return rule, nil
}
