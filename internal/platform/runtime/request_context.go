package runtime

import "context"

const RequestContextKey = "erp_claw_request_context"

type requestContextContextKey struct{}

// RequestContext carries request-scoped metadata through middleware and handlers.
type RequestContext struct {
	RequestID     string
	TenantID      string
	ActorID       string
	ActorProvided bool
	TraceID       string
}

func WithRequestContext(ctx context.Context, rc *RequestContext) context.Context {
	if rc == nil {
		return ctx
	}
	return context.WithValue(ctx, requestContextContextKey{}, rc)
}

func RequestContextFromContext(ctx context.Context) (*RequestContext, bool) {
	rc, ok := ctx.Value(requestContextContextKey{}).(*RequestContext)
	return rc, ok
}
