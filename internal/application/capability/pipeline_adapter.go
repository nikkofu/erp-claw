package capability

import (
	"context"
	"errors"
	"fmt"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	domaincap "github.com/nikkofu/erp-claw/internal/domain/capability"
)

var errCapabilityAuthorizerResolverRequired = errors.New("capability authorizer resolver is required")

type effectiveAgentCapabilityPolicyResolver interface {
	Handle(ctx context.Context, tenantID, agentProfileID string) (*domaincap.EffectiveAgentCapabilityPolicy, error)
}

type SharedCommandCapabilityAuthorizer struct {
	Resolver effectiveAgentCapabilityPolicyResolver
}

func (s SharedCommandCapabilityAuthorizer) AuthorizeCommandCapabilities(ctx context.Context, cmd shared.Command) error {
	payload, ok := cmd.Payload.(map[string]any)
	if !ok {
		return nil
	}

	modelEntryID := capabilityStringValue(payload["model_entry_id"])
	toolEntryIDs := capabilityStringList(payload["tool_entry_ids"])
	if modelEntryID == "" && len(toolEntryIDs) == 0 {
		return nil
	}

	if s.Resolver == nil {
		return errCapabilityAuthorizerResolverRequired
	}

	agentProfileID := capabilityStringValue(payload["agent_profile_id"])
	if agentProfileID == "" {
		return fmt.Errorf("%w: agent_profile_id is required", shared.ErrCapabilityDenied)
	}

	policy, err := s.Resolver.Handle(ctx, cmd.TenantID, agentProfileID)
	if err != nil {
		return err
	}

	if modelEntryID != "" {
		if err := ensureEffectiveEntry("model catalog entry", modelEntryID, agentProfileID, policy.EffectiveModelEntryIDs, policy.StaleModelEntryIDs); err != nil {
			return err
		}
	}

	for _, toolEntryID := range toolEntryIDs {
		if err := ensureEffectiveEntry("tool catalog entry", toolEntryID, agentProfileID, policy.EffectiveToolEntryIDs, policy.StaleToolEntryIDs); err != nil {
			return err
		}
	}

	return nil
}

func ensureEffectiveEntry(kind, entryID, agentProfileID string, effectiveIDs, staleIDs []string) error {
	if containsCapabilityEntryID(effectiveIDs, entryID) {
		return nil
	}
	if containsCapabilityEntryID(staleIDs, entryID) {
		return fmt.Errorf("%w: %s %q is stale for agent profile %q", shared.ErrCapabilityDenied, kind, entryID, agentProfileID)
	}
	return fmt.Errorf("%w: %s %q is not effectively allowed for agent profile %q", shared.ErrCapabilityDenied, kind, entryID, agentProfileID)
}

func containsCapabilityEntryID(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func capabilityStringValue(raw any) string {
	switch v := raw.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return ""
	}
}

func capabilityStringList(raw any) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)

	appendValue := func(value string) {
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}

	switch v := raw.(type) {
	case []string:
		for _, value := range v {
			appendValue(value)
		}
	case []any:
		for _, value := range v {
			appendValue(capabilityStringValue(value))
		}
	case string:
		appendValue(v)
	}

	return out
}
