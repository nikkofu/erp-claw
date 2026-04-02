package policy

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
)

var ErrInvalidRule = errors.New("invalid policy rule")
var ErrRuleNotFound = errors.New("policy rule not found")

type RuleStore interface {
	Upsert(ctx context.Context, tenantID string, rule Rule) error
	List(ctx context.Context, tenantID string) ([]Rule, error)
	Delete(ctx context.Context, tenantID string, commandPrefix string) error
}

type InMemoryRuleStore struct {
	mu    sync.RWMutex
	rules map[string]map[string]Rule
}

func NewInMemoryRuleStore() *InMemoryRuleStore {
	return &InMemoryRuleStore{
		rules: make(map[string]map[string]Rule),
	}
}

func (s *InMemoryRuleStore) Upsert(_ context.Context, tenantID string, rule Rule) error {
	tenantID = strings.TrimSpace(tenantID)
	normalized, ok := normalizeRule(rule)
	if tenantID == "" || !ok {
		return ErrInvalidRule
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.rules[tenantID]; !ok {
		s.rules[tenantID] = map[string]Rule{}
	}
	s.rules[tenantID][normalized.CommandPrefix] = normalized
	return nil
}

func (s *InMemoryRuleStore) List(_ context.Context, tenantID string) ([]Rule, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, ErrInvalidRule
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	current := s.rules[tenantID]
	if len(current) == 0 {
		return []Rule{}, nil
	}

	out := make([]Rule, 0, len(current))
	for _, rule := range current {
		out = append(out, cloneRule(rule))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CommandPrefix < out[j].CommandPrefix
	})
	return out, nil
}

func (s *InMemoryRuleStore) Delete(_ context.Context, tenantID string, commandPrefix string) error {
	tenantID = strings.TrimSpace(tenantID)
	commandPrefix = strings.TrimSpace(commandPrefix)
	if tenantID == "" || commandPrefix == "" {
		return ErrInvalidRule
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tenantRules, ok := s.rules[tenantID]
	if !ok {
		return ErrRuleNotFound
	}
	if _, ok := tenantRules[commandPrefix]; !ok {
		return ErrRuleNotFound
	}
	delete(tenantRules, commandPrefix)
	if len(tenantRules) == 0 {
		delete(s.rules, tenantID)
	}
	return nil
}

func normalizeRule(rule Rule) (Rule, bool) {
	commandPrefix := strings.TrimSpace(rule.CommandPrefix)
	if commandPrefix == "" {
		return Rule{}, false
	}

	seen := map[string]struct{}{}
	roles := make([]string, 0, len(rule.AnyOfRoles))
	for _, role := range rule.AnyOfRoles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		roles = append(roles, role)
	}
	if len(roles) == 0 {
		return Rule{}, false
	}

	return Rule{
		CommandPrefix: commandPrefix,
		AnyOfRoles:    roles,
	}, true
}

func cloneRule(rule Rule) Rule {
	return Rule{
		CommandPrefix: rule.CommandPrefix,
		AnyOfRoles:    append([]string(nil), rule.AnyOfRoles...),
	}
}
