package outbox

import (
	"context"
	"errors"
)

var errOutboxListMessagesReaderRequired = errors.New("outbox list messages handler requires reader")
var errOutboxRequeueFailedHandlerRecoveryRequired = errors.New("outbox requeue failed handler requires recovery service")

type ListMessages struct {
	TenantID string
	Status   string
	Limit    int
}

type ListMessagesHandler struct {
	Messages MessageReader
}

func (h ListMessagesHandler) Handle(ctx context.Context, q ListMessages) ([]Message, error) {
	if h.Messages == nil {
		return nil, errOutboxListMessagesReaderRequired
	}

	return h.Messages.ListMessages(ctx, q.TenantID, q.Status, q.Limit)
}

type FailureRequeuer interface {
	RequeueFailed(ctx context.Context, ids []int64) (int, error)
}

type RequeueFailedMessages struct {
	IDs []int64
}

type RequeueFailedMessagesHandler struct {
	Recovery FailureRequeuer
}

func (h RequeueFailedMessagesHandler) Handle(ctx context.Context, cmd RequeueFailedMessages) (int, error) {
	if h.Recovery == nil {
		return 0, errOutboxRequeueFailedHandlerRecoveryRequired
	}

	return h.Recovery.RequeueFailed(ctx, cmd.IDs)
}
