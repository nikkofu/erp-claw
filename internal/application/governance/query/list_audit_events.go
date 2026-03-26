package query

import (
	"context"
	"errors"
	"time"

	"github.com/nikkofu/erp-claw/internal/platform/audit"
)

var errListAuditEventsHandlerAuditServiceRequired = errors.New("list audit events handler requires audit service")

type AuditEventLister interface {
	List(ctx context.Context, query audit.Query) ([]audit.Record, error)
}

type ListAuditEvents struct {
	TenantID       string
	CommandName    string
	ActorID        string
	OccurredAfter  time.Time
	OccurredBefore time.Time
	Limit          int
}

type ListAuditEventsHandler struct {
	AuditEvents AuditEventLister
	Authorize   func(context.Context, ListAuditEvents) error
	Audit       func(context.Context, []audit.Record) error
}

func (h ListAuditEventsHandler) Handle(ctx context.Context, q ListAuditEvents) ([]audit.Record, error) {
	if h.AuditEvents == nil {
		return nil, errListAuditEventsHandlerAuditServiceRequired
	}
	if h.Authorize != nil {
		if err := h.Authorize(ctx, q); err != nil {
			return nil, err
		}
	}

	events, err := h.AuditEvents.List(ctx, audit.Query{
		TenantID:       q.TenantID,
		CommandName:    q.CommandName,
		ActorID:        q.ActorID,
		OccurredAfter:  q.OccurredAfter,
		OccurredBefore: q.OccurredBefore,
		Limit:          q.Limit,
	})
	if err != nil {
		return nil, err
	}
	if h.Audit != nil {
		if err := h.Audit(ctx, events); err != nil {
			return nil, err
		}
	}

	return events, nil
}
