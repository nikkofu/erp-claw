package runtime

import "time"

const (
	WorkspaceEventTypeTaskStatusChanged    = "agentruntime.task.status.changed"
	WorkspaceEventTypeSessionStatusChanged = "agentruntime.session.status.changed"
)

// WorkspaceEvent is the typed event envelope emitted by the agent gateway runtime.
type WorkspaceEvent struct {
	Type       string
	TenantID   string
	SessionID  string
	TaskID     string
	Payload    any
	OccurredAt time.Time
}
