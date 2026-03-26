package policy

import (
	"errors"
	"time"
)

const wildcard = "*"

var ErrRuleNotFound = errors.New("policy rule not found")

// Rule represents a tenant-scoped governance rule used by the policy evaluator.
type Rule struct {
	TenantID    string
	ID          string
	CommandName string
	ActorID     string
	Decision    Decision
	Priority    int
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RuleFilter defines a minimal query contract for policy rules.
type RuleFilter struct {
	TenantID    string
	CommandName string
	ActorID     string
	ActiveOnly  bool
	Limit       int
}

func ruleMatchesInput(rule Rule, input Input) bool {
	if rule.TenantID != input.TenantID {
		return false
	}

	if rule.CommandName != wildcard && rule.CommandName != input.CommandName {
		return false
	}

	if rule.ActorID != wildcard && rule.ActorID != input.ActorID {
		return false
	}

	return true
}
