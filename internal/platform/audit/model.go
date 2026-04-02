package audit

import (
	"time"

	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

// Record captures command pipeline audit metadata.
type Record struct {
	CommandName   string
	TenantID      string
	ActorID       string
	Decision      policy.Decision
	Outcome       string
	Error         string
	CorrelationID string
	ResourceType  string
	ResourceID    string
	OccurredAt    time.Time
}
