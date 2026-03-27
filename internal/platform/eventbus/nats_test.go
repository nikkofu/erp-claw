package eventbus

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/nats-io/nats.go"
)

type fakeJetStreamPublisher struct {
	msg *nats.Msg
	err error
}

func (f *fakeJetStreamPublisher) PublishMsg(msg *nats.Msg, _ ...nats.PubOpt) (*nats.PubAck, error) {
	f.msg = msg
	if f.err != nil {
		return nil, f.err
	}
	return &nats.PubAck{}, nil
}

func TestNATSBusPublishSetsHeadersAndPayload(t *testing.T) {
	t.Parallel()

	js := &fakeJetStreamPublisher{}
	bus := &NATSBus{js: js}

	err := bus.Publish(context.Background(), Event{
		Topic:       "orders.created",
		TenantID:    "tenant-1",
		Correlation: "corr-1",
		MessageID:   "msg-1",
		Payload: map[string]any{
			"id": "order-1",
		},
	})
	if err != nil {
		t.Fatalf("Publish() error = %v, want nil", err)
	}

	if js.msg == nil {
		t.Fatal("expected published message, got nil")
	}
	if got := js.msg.Subject; got != "orders.created" {
		t.Fatalf("expected subject orders.created, got %s", got)
	}
	if got := js.msg.Header.Get("X-Tenant-ID"); got != "tenant-1" {
		t.Fatalf("expected tenant header tenant-1, got %s", got)
	}
	if got := js.msg.Header.Get("X-Correlation-ID"); got != "corr-1" {
		t.Fatalf("expected correlation header corr-1, got %s", got)
	}
	if got := js.msg.Header.Get(nats.MsgIdHdr); got != "msg-1" {
		t.Fatalf("expected msg id header msg-1, got %s", got)
	}

	var payload map[string]string
	if err := json.Unmarshal(js.msg.Data, &payload); err != nil {
		t.Fatalf("unmarshal payload error = %v", err)
	}
	if payload["id"] != "order-1" {
		t.Fatalf("expected payload id order-1, got %s", payload["id"])
	}
}

func TestNATSBusPublishPropagatesJetStreamError(t *testing.T) {
	t.Parallel()

	js := &fakeJetStreamPublisher{
		err: errors.New("jetstream down"),
	}
	bus := &NATSBus{js: js}

	err := bus.Publish(context.Background(), Event{
		Topic: "orders.created",
	})
	if err == nil {
		t.Fatal("expected publish error, got nil")
	}
}

func TestNATSBusPublishRejectsEmptyTopic(t *testing.T) {
	t.Parallel()

	js := &fakeJetStreamPublisher{}
	bus := &NATSBus{js: js}

	err := bus.Publish(context.Background(), Event{
		Topic: " ",
	})
	if err == nil {
		t.Fatal("expected empty topic error, got nil")
	}
}

func TestNATSBusPublishReturnsPayloadEncodeError(t *testing.T) {
	t.Parallel()

	js := &fakeJetStreamPublisher{}
	bus := &NATSBus{js: js}

	err := bus.Publish(context.Background(), Event{
		Topic:   "orders.created",
		Payload: map[string]any{"ch": make(chan int)},
	})
	if err == nil {
		t.Fatal("expected payload encode error, got nil")
	}
}
