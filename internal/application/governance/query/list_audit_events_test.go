package query

import (
	"context"
	"testing"
	"time"

	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestListAuditEventsHandlerRequiresAuditService(t *testing.T) {
	handler := ListAuditEventsHandler{}

	_, err := handler.Handle(context.Background(), ListAuditEvents{TenantID: "tenant-a"})
	if err == nil {
		t.Fatal("expected nil audit service to fail")
	}
}

func TestListAuditEventsHandlerUsesAuthorizeAndAuditHooks(t *testing.T) {
	events := &stubAuditEventLister{
		items: []audit.Record{
			{
				ID:          "evt-1",
				TenantID:    "tenant-a",
				CommandName: "purchase.approve",
				ActorID:     "actor-a",
				Decision:    policy.DecisionAllow,
				Outcome:     "succeeded",
				OccurredAt:  time.Date(2026, 3, 25, 8, 0, 0, 0, time.UTC),
			},
		},
	}

	authorizeCalled := false
	auditCalled := false
	handler := ListAuditEventsHandler{
		AuditEvents: events,
		Authorize: func(_ context.Context, q ListAuditEvents) error {
			authorizeCalled = true
			if q.TenantID != "tenant-a" {
				t.Fatalf("unexpected tenant id: %s", q.TenantID)
			}
			return nil
		},
		Audit: func(_ context.Context, got []audit.Record) error {
			auditCalled = true
			if len(got) != 1 {
				t.Fatalf("expected one event, got %d", len(got))
			}
			return nil
		},
	}

	out, err := handler.Handle(context.Background(), ListAuditEvents{
		TenantID:    "tenant-a",
		CommandName: "purchase.approve",
		ActorID:     "actor-a",
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !authorizeCalled {
		t.Fatal("expected authorize hook to be called")
	}
	if !auditCalled {
		t.Fatal("expected audit hook to be called")
	}
	if len(out) != 1 {
		t.Fatalf("expected one event, got %d", len(out))
	}
}

type stubAuditEventLister struct {
	items []audit.Record
}

func (s *stubAuditEventLister) List(_ context.Context, _ audit.Query) ([]audit.Record, error) {
	return s.items, nil
}
