package policy

import (
	"context"
	"errors"
)

var errRuleServiceRepositoryRequired = errors.New("policy rule service requires repository")

// RuleService exposes an application seam for governance rule management.
type RuleService struct {
	repository RuleRepository
}

func NewRuleService(repository RuleRepository) (*RuleService, error) {
	if repository == nil {
		return nil, errRuleServiceRepositoryRequired
	}
	return &RuleService{repository: repository}, nil
}

func (s *RuleService) Upsert(ctx context.Context, rule Rule) (Rule, error) {
	return s.repository.UpsertRule(ctx, rule)
}

func (s *RuleService) List(ctx context.Context, filter RuleFilter) ([]Rule, error) {
	return s.repository.ListRules(ctx, filter)
}

func (s *RuleService) SetActive(ctx context.Context, tenantID, ruleID string, active bool) (Rule, error) {
	return s.repository.SetRuleActive(ctx, tenantID, ruleID, active)
}

func (s *RuleService) Activate(ctx context.Context, tenantID, ruleID string) (Rule, error) {
	return s.SetActive(ctx, tenantID, ruleID, true)
}

func (s *RuleService) Deactivate(ctx context.Context, tenantID, ruleID string) (Rule, error) {
	return s.SetActive(ctx, tenantID, ruleID, false)
}

type ruleEvaluator struct {
	repository RuleRepository
	fallback   Decision
}

// NewRuleEvaluator returns an evaluator backed by persisted policy rules.
func NewRuleEvaluator(repository RuleRepository, fallback Decision) Evaluator {
	if fallback == "" {
		fallback = DecisionDeny
	}
	return ruleEvaluator{
		repository: repository,
		fallback:   fallback,
	}
}

func (e ruleEvaluator) Evaluate(ctx context.Context, input Input) (Decision, error) {
	if e.repository == nil {
		return e.fallback, nil
	}

	rules, err := e.repository.ListRules(ctx, RuleFilter{
		TenantID:    input.TenantID,
		CommandName: input.CommandName,
		ActorID:     input.ActorID,
		ActiveOnly:  true,
	})
	if err != nil {
		return "", err
	}

	if len(rules) == 0 {
		return e.fallback, nil
	}

	best := rules[0]
	bestScore := ruleSpecificityScore(best, input)
	for _, candidate := range rules[1:] {
		score := ruleSpecificityScore(candidate, input)
		if score > bestScore {
			best = candidate
			bestScore = score
		}
	}

	return best.Decision, nil
}

func ruleSpecificityScore(rule Rule, input Input) int {
	score := rule.Priority * 10
	if rule.CommandName == input.CommandName {
		score += 2
	}
	if rule.ActorID == input.ActorID {
		score += 1
	}
	return score
}
