package capability

import "context"

type AgentCapabilityPolicyRepository interface {
	SaveAgentCapabilityPolicy(ctx context.Context, policy *AgentCapabilityPolicy) error
	GetAgentCapabilityPolicy(ctx context.Context, tenantID, agentProfileID string) (*AgentCapabilityPolicy, error)
}
