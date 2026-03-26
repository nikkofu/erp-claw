package policy

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"
)

// RuleRepository persists and queries policy rules.
type RuleRepository interface {
	UpsertRule(ctx context.Context, rule Rule) (Rule, error)
	ListRules(ctx context.Context, filter RuleFilter) ([]Rule, error)
	SetRuleActive(ctx context.Context, tenantID, ruleID string, active bool) (Rule, error)
}

// InMemoryRuleRepository stores rules in-memory for tests and bootstrap defaults.
type InMemoryRuleRepository struct {
	mu     sync.Mutex
	nextID int
	rules  map[string]Rule
}

func NewInMemoryRuleRepository() *InMemoryRuleRepository {
	return &InMemoryRuleRepository{
		rules: make(map[string]Rule),
	}
}

func (r *InMemoryRuleRepository) UpsertRule(_ context.Context, rule Rule) (Rule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rule.ID == "" {
		r.nextID++
		rule.ID = "rule-" + strconv.Itoa(r.nextID)
	}
	if rule.ActorID == "" {
		rule.ActorID = wildcard
	}
	if rule.CommandName == "" {
		rule.CommandName = wildcard
	}

	now := time.Now().UTC()
	key := fmt.Sprintf("%s::%s", rule.TenantID, rule.ID)
	if existing, ok := r.rules[key]; ok {
		rule.CreatedAt = existing.CreatedAt
		rule.UpdatedAt = now
	} else {
		rule.CreatedAt = now
		rule.UpdatedAt = now
	}

	r.rules[key] = rule
	return rule, nil
}

func (r *InMemoryRuleRepository) ListRules(_ context.Context, filter RuleFilter) ([]Rule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]Rule, 0)
	for _, rule := range r.rules {
		if filter.TenantID != "" && rule.TenantID != filter.TenantID {
			continue
		}
		if filter.ActiveOnly && !rule.Active {
			continue
		}
		if filter.CommandName != "" && rule.CommandName != filter.CommandName && rule.CommandName != wildcard {
			continue
		}
		if filter.ActorID != "" && rule.ActorID != filter.ActorID && rule.ActorID != wildcard {
			continue
		}
		out = append(out, rule)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Priority != out[j].Priority {
			return out[i].Priority > out[j].Priority
		}
		return out[i].ID < out[j].ID
	})

	if filter.Limit > 0 && len(out) > filter.Limit {
		return out[:filter.Limit], nil
	}

	return out, nil
}

func (r *InMemoryRuleRepository) SetRuleActive(_ context.Context, tenantID, ruleID string, active bool) (Rule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s::%s", tenantID, ruleID)
	rule, ok := r.rules[key]
	if !ok {
		return Rule{}, ErrRuleNotFound
	}

	rule.Active = active
	rule.UpdatedAt = time.Now().UTC()
	r.rules[key] = rule

	return rule, nil
}
