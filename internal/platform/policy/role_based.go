package policy

import (
	"context"
	"strings"
)

type RoleLookup func(ctx context.Context, tenantID, actorID string) ([]string, error)
type TenantRuleLookup func(ctx context.Context, tenantID string) ([]Rule, error)

// Rule maps a command prefix to role requirements.
type Rule struct {
	CommandPrefix string
	AnyOfRoles    []string
}

type roleEvaluator struct {
	lookup           RoleLookup
	rules            []Rule
	tenantRuleLookup TenantRuleLookup
}

func NewRoleEvaluator(lookup RoleLookup, rules []Rule) Evaluator {
	return roleEvaluator{
		lookup: lookup,
		rules:  append([]Rule(nil), rules...),
	}
}

func NewRoleEvaluatorWithTenantRules(
	lookup RoleLookup,
	defaultRules []Rule,
	tenantRuleLookup TenantRuleLookup,
) Evaluator {
	return roleEvaluator{
		lookup:           lookup,
		rules:            append([]Rule(nil), defaultRules...),
		tenantRuleLookup: tenantRuleLookup,
	}
}

func (e roleEvaluator) Evaluate(ctx context.Context, input Input) (Decision, error) {
	rules, err := e.effectiveRules(ctx, input.TenantID)
	if err != nil {
		return "", err
	}

	rule, matched := matchRule(rules, strings.TrimSpace(input.CommandName))
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

func (e roleEvaluator) effectiveRules(ctx context.Context, tenantID string) ([]Rule, error) {
	out := make([]Rule, 0, len(e.rules))
	for _, rule := range e.rules {
		out = append(out, cloneRule(rule))
	}
	if e.tenantRuleLookup == nil {
		return out, nil
	}

	tenantRules, err := e.tenantRuleLookup(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if len(tenantRules) == 0 {
		return out, nil
	}

	indexByPrefix := make(map[string]int, len(out))
	for i, rule := range out {
		indexByPrefix[rule.CommandPrefix] = i
	}
	for _, rawRule := range tenantRules {
		rule := cloneRule(rawRule)
		if idx, ok := indexByPrefix[rule.CommandPrefix]; ok {
			out[idx] = rule
			continue
		}
		indexByPrefix[rule.CommandPrefix] = len(out)
		out = append(out, rule)
	}
	return out, nil
}

func matchRule(rules []Rule, commandName string) (Rule, bool) {
	bestPrefixLen := -1
	var best Rule
	for _, rule := range rules {
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
