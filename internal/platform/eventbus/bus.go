package eventbus

import "context"

type Event struct {
	Topic       string
	TenantID    string
	Correlation string
	MessageID   string
	Payload     any
}

type Bus interface {
	Publish(ctx context.Context, evt Event) error
}
