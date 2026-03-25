package eventbus

import (
	"context"
	"testing"
)

func TestMemoryBusPublish(t *testing.T) {
	bus := NewMemory()
	err := bus.Publish(context.Background(), Event{
		Topic:       "orders.created",
		TenantID:    "tenant-1",
		Correlation: "corr-1",
		Payload: map[string]any{
			"id": "order-1",
		},
	})
	if err != nil {
		t.Fatalf("Publish() error = %v, want nil", err)
	}
}
