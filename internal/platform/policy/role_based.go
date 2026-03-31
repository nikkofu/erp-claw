package policy

import (
	"context"
	"strings"
)

type RoleLookup func(ctx context.Context, tenantID, actorID string) ([]string, error)

// Rule maps a command prefix to role requirements.
type Rule struct {
	CommandPrefix string
	AnyOfRoles    []string
}

type roleEvaluator struct {
	lookup RoleLookup
	rules  []Rule
}

func NewRoleEvaluator(lookup RoleLookup, rules []Rule) Evaluator {
	return roleEvaluator{
		lookup: lookup,
		rules:  append([]Rule(nil), rules...),
	}
}

func (e roleEvaluator) Evaluate(ctx context.Context, input Input) (Decision, error) {
	rule, matched := e.matchRule(strings.TrimSpace(input.CommandName))
	if !matched {
		return DecisionAllow, nil
	}
	if len(rule.AnyOfRoles) == 0 {
		return DecisionAllow, nil
	}
	if e.lookup == nil {
		return DecisionDeny, nil
	}

	roles, err := e.lookup(ctx, input.TenantID, input.ActorID)
	if err != nil {
		return "", err
	}
	if hasAnyRole(roles, rule.AnyOfRoles) {
		return DecisionAllow, nil
	}
	return DecisionDeny, nil
}

func (e roleEvaluator) matchRule(commandName string) (Rule, bool) {
	bestPrefixLen := -1
	var best Rule
	for _, rule := range e.rules {
		prefix := strings.TrimSpace(rule.CommandPrefix)
		if prefix == "" {
			continue
		}
		if matchesCommand(prefix, commandName) && len(prefix) > bestPrefixLen {
			bestPrefixLen = len(prefix)
			best = rule
		}
	}
	if bestPrefixLen < 0 {
		return Rule{}, false
	}
	return best, true
}

func matchesCommand(prefix, commandName string) bool {
	if strings.HasSuffix(prefix, ".") {
		return strings.HasPrefix(commandName, prefix)
	}
	return commandName == prefix
}

func hasAnyRole(assigned []string, required []string) bool {
	set := make(map[string]struct{}, len(assigned))
	for _, role := range assigned {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		set[role] = struct{}{}
	}
	for _, role := range required {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, ok := set[role]; ok {
			return true
		}
	}
	return false
}
