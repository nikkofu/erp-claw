package runtime

const RequestContextKey = "erp_claw_request_context"

// RequestContext carries request-scoped metadata through middleware and handlers.
type RequestContext struct {
	RequestID string
	TenantID  string
	ActorID   string
	TraceID   string
}
