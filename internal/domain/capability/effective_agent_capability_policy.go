package capability

type EffectiveAgentCapabilityPolicy struct {
	TenantID               string
	AgentProfileID         string
	EffectiveModelEntryIDs []string
	EffectiveToolEntryIDs  []string
	StaleModelEntryIDs     []string
	StaleToolEntryIDs      []string
}

func NewEffectiveAgentCapabilityPolicy(
	tenantID,
	agentProfileID string,
	effectiveModelEntryIDs,
	effectiveToolEntryIDs,
	staleModelEntryIDs,
	staleToolEntryIDs []string,
) (*EffectiveAgentCapabilityPolicy, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}
	if agentProfileID == "" {
		return nil, ErrAgentProfileIDRequired
	}

	return &EffectiveAgentCapabilityPolicy{
		TenantID:               tenantID,
		AgentProfileID:         agentProfileID,
		EffectiveModelEntryIDs: normalizeCapabilityEntryIDs(effectiveModelEntryIDs),
		EffectiveToolEntryIDs:  normalizeCapabilityEntryIDs(effectiveToolEntryIDs),
		StaleModelEntryIDs:     normalizeCapabilityEntryIDs(staleModelEntryIDs),
		StaleToolEntryIDs:      normalizeCapabilityEntryIDs(staleToolEntryIDs),
	}, nil
}
