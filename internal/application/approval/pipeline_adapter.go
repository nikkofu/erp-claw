package approval

import (
	"context"
	"errors"
	"fmt"

	"github.com/nikkofu/erp-claw/internal/application/shared"
)

var (
	errApprovalPayloadMustBeMap     = errors.New("approval starter requires map payload")
	errApprovalDefinitionIDRequired = errors.New("approval starter requires approval_definition_id")
	errApprovalResourceIDRequired   = errors.New("approval starter requires resource_id")
)

type SharedCommandApprovalStarter struct {
	Handler StartApprovalHandler
}

func (s SharedCommandApprovalStarter) StartApprovalForCommand(ctx context.Context, cmd shared.Command) error {
	payload, ok := cmd.Payload.(map[string]any)
	if !ok {
		return errApprovalPayloadMustBeMap
	}

	definitionID := stringValue(payload["approval_definition_id"])
	if definitionID == "" {
		return errApprovalDefinitionIDRequired
	}
	resourceID := stringValue(payload["resource_id"])
	if resourceID == "" {
		return errApprovalResourceIDRequired
	}
	resourceType := stringValue(payload["resource_type"])
	if resourceType == "" {
		resourceType = cmd.Name
	}
	requestedBy := stringValue(payload["requested_by"])
	if requestedBy == "" {
		requestedBy = cmd.ActorID
	}

	_, err := s.Handler.Handle(ctx, StartApproval{
		TenantID:     cmd.TenantID,
		DefinitionID: definitionID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		RequestedBy:  requestedBy,
	})
	return err
}

func stringValue(raw any) string {
	switch v := raw.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return ""
	}
}
