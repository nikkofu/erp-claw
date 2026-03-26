package query

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

var errListPolicyRulesHandlerRulesRequired = errors.New("list policy rules handler requires rules service")

type RuleReader interface {
	List(ctx context.Context, filter policy.RuleFilter) ([]policy.Rule, error)
}

type ListPolicyRules struct {
	TenantID    string
	CommandName string
	ActorID     string
	ActiveOnly  bool
	Limit       int
}

type ListPolicyRulesHandler struct {
	Rules     RuleReader
	Authorize func(context.Context, ListPolicyRules) error
	Audit     func(context.Context, []policy.Rule) error
}

func (h ListPolicyRulesHandler) Handle(ctx context.Context, q ListPolicyRules) ([]policy.Rule, error) {
	if h.Rules == nil {
		return nil, errListPolicyRulesHandlerRulesRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, q); err != nil {
			return nil, err
		}
	}

	rules, err := h.Rules.List(ctx, policy.RuleFilter{
		TenantID:    q.TenantID,
		CommandName: q.CommandName,
		ActorID:     q.ActorID,
		ActiveOnly:  q.ActiveOnly,
		Limit:       q.Limit,
	})
	if err != nil {
		return nil, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, rules); err != nil {
			return nil, err
		}
	}

	return rules, nil
}
