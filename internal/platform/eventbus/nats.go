package eventbus

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/nats-io/nats.go"
)

type NATSBus struct {
	js jetStreamPublisher
}

type jetStreamPublisher interface {
	PublishMsg(msg *nats.Msg, opts ...nats.PubOpt) (*nats.PubAck, error)
}

func NewNATS(conn *nats.Conn) (*NATSBus, error) {
	if conn == nil {
		return nil, errors.New("nats connection is required")
	}

	js, err := conn.JetStream()
	if err != nil {
		return nil, err
	}

	return &NATSBus{js: js}, nil
}

func (b *NATSBus) Publish(ctx context.Context, evt Event) error {
	if strings.TrimSpace(evt.Topic) == "" {
		return errors.New("event topic is required")
	}

	data, err := encodePayload(evt.Payload)
	if err != nil {
		return err
	}

	msg := &nats.Msg{
		Subject: evt.Topic,
		Data:    data,
		Header:  nats.Header{},
	}
	if evt.TenantID != "" {
		msg.Header.Set("X-Tenant-ID", evt.TenantID)
	}
	if evt.Correlation != "" {
		msg.Header.Set("X-Correlation-ID", evt.Correlation)
	}
	if strings.TrimSpace(evt.MessageID) != "" {
		msg.Header.Set(nats.MsgIdHdr, evt.MessageID)
	}

	_, err = b.js.PublishMsg(msg, nats.Context(ctx))
	return err
}

func encodePayload(payload any) ([]byte, error) {
	switch value := payload.(type) {
	case nil:
		return nil, nil
	case []byte:
		return value, nil
	default:
		return json.Marshal(value)
	}
}
