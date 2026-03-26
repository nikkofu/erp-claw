package controlplane

import (
	"errors"
	"strings"
)

var (
	errAgentProfileTenantIDRequired = errors.New("agent profile tenant id is required")
	errAgentProfileNameRequired     = errors.New("agent profile name is required")
	errAgentProfileModelRequired    = errors.New("agent profile model is required")
)

// AgentProfile captures a governed AI agent profile in the catalog.
type AgentProfile struct {
	ID       string
	TenantID string
	Name     string
	Model    string
}

// NewAgentProfile validates required catalog fields.
func NewAgentProfile(tenantID, name, model string) (AgentProfile, error) {
	if strings.TrimSpace(tenantID) == "" {
		return AgentProfile{}, errAgentProfileTenantIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return AgentProfile{}, errAgentProfileNameRequired
	}
	if strings.TrimSpace(model) == "" {
		return AgentProfile{}, errAgentProfileModelRequired
	}

	return AgentProfile{
		TenantID: tenantID,
		Name:     name,
		Model:    model,
	}, nil
}
