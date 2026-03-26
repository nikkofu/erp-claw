package outbox

import (
	"context"
	"errors"

	"github.com/nikkofu/erp-claw/internal/platform/eventbus"
)

var errNilEventBus = errors.New("outbox publisher requires non-nil event bus")

type EventBusPublisher struct {
	bus eventbus.Bus
}

func NewEventBusPublisher(bus eventbus.Bus) (*EventBusPublisher, error) {
	if bus == nil {
		return nil, errNilEventBus
	}

	return &EventBusPublisher{bus: bus}, nil
}

func (p *EventBusPublisher) Publish(ctx context.Context, message Message) error {
	return p.bus.Publish(ctx, eventbus.Event{
		Topic:    message.Topic,
		TenantID: message.TenantID,
		Payload:  message.Payload,
	})
}
