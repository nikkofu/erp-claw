package runtime

// WorkspaceEvent is the typed event envelope emitted by the agent gateway runtime.
type WorkspaceEvent struct {
	Type      string
	TenantID  string
	SessionID string
	TaskID    string
	Payload   any
}

type WorkspaceEventSink interface {
	Broadcast(evt WorkspaceEvent) error
}
