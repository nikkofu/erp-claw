package capability

import (
	"errors"
	"sort"
	"strings"
)

var ErrAgentProfileIDRequired = errors.New("agent profile id is required")

type AgentCapabilityPolicy struct {
	TenantID             string
	AgentProfileID       string
	AllowedModelEntryIDs []string
	AllowedToolEntryIDs  []string
}

func NewAgentCapabilityPolicy(tenantID, agentProfileID string, allowedModelEntryIDs, allowedToolEntryIDs []string) (*AgentCapabilityPolicy, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, ErrTenantIDRequired
	}
	if strings.TrimSpace(agentProfileID) == "" {
		return nil, ErrAgentProfileIDRequired
	}

	return &AgentCapabilityPolicy{
		TenantID:             strings.TrimSpace(tenantID),
		AgentProfileID:       strings.TrimSpace(agentProfileID),
		AllowedModelEntryIDs: normalizeCapabilityEntryIDs(allowedModelEntryIDs),
		AllowedToolEntryIDs:  normalizeCapabilityEntryIDs(allowedToolEntryIDs),
	}, nil
}

func normalizeCapabilityEntryIDs(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}
