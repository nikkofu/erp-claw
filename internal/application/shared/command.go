package shared

// Command is the canonical shape entering application handlers.
type Command struct {
	Name     string
	TenantID string
	ActorID  string
	Payload  any
}
