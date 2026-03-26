package command

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

var errUpsertPolicyRuleHandlerRulesRequired = errors.New("upsert policy rule handler requires rules service")

type RuleManager interface {
	Upsert(ctx context.Context, rule policy.Rule) (policy.Rule, error)
}

type UpsertPolicyRule struct {
	TenantID    string
	ID          string
	CommandName string
	ActorID     string
	Decision    policy.Decision
	Priority    int
	Active      bool
}

type UpsertPolicyRuleHandler struct {
	Rules     RuleManager
	Authorize func(context.Context, UpsertPolicyRule) error
	Audit     func(context.Context, policy.Rule) error
}

func (h UpsertPolicyRuleHandler) Handle(ctx context.Context, cmd UpsertPolicyRule) (policy.Rule, error) {
	if h.Rules == nil {
		return policy.Rule{}, errUpsertPolicyRuleHandlerRulesRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, cmd); err != nil {
			return policy.Rule{}, err
		}
	}

	rule, err := h.Rules.Upsert(ctx, policy.Rule{
		TenantID:    cmd.TenantID,
		ID:          cmd.ID,
		CommandName: cmd.CommandName,
		ActorID:     cmd.ActorID,
		Decision:    cmd.Decision,
		Priority:    cmd.Priority,
		Active:      cmd.Active,
	})
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
