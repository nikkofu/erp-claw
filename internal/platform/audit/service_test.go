package audit

import (
	"context"
	"testing"
	"time"

	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestNewServiceRequiresStore(t *testing.T) {
	_, err := NewService(nil)
	if err == nil {
		t.Fatal("expected nil store to fail")
	}
}

func TestServiceRecordStoresEventAndListReturnsTenantScopedEvents(t *testing.T) {
	store := NewInMemoryStore()
	svc, err := NewService(store)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	err = svc.Record(context.Background(), Record{
		TenantID:    "tenant-a",
		CommandName: "purchase.approve",
		ActorID:     "actor-a",
		Decision:    policy.DecisionAllow,
		Outcome:     "succeeded",
	})
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	err = svc.Record(context.Background(), Record{
		TenantID:    "tenant-b",
		CommandName: "purchase.approve",
		ActorID:     "actor-b",
		Decision:    policy.DecisionAllow,
		Outcome:     "succeeded",
	})
	if err != nil {
		t.Fatalf("record second tenant: %v", err)
	}

	events, err := svc.List(context.Background(), Query{TenantID: "tenant-a"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].TenantID != "tenant-a" {
		t.Fatalf("expected tenant-a event, got %s", events[0].TenantID)
	}
}

func TestServiceListAppliesLimitAndFilters(t *testing.T) {
	store := NewInMemoryStore()
	svc, err := NewService(store)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	for i := 0; i < 3; i++ {
		err = svc.Record(context.Background(), Record{
			TenantID:    "tenant-a",
			CommandName: "purchase.approve",
			ActorID:     "actor-a",
			Decision:    policy.DecisionAllow,
			Outcome:     "succeeded",
		})
		if err != nil {
			t.Fatalf("record %d: %v", i, err)
		}
	}
	err = svc.Record(context.Background(), Record{
		TenantID:    "tenant-a",
		CommandName: "inventory.adjust",
		ActorID:     "actor-z",
		Decision:    policy.DecisionDeny,
		Outcome:     "rejected",
	})
	if err != nil {
		t.Fatalf("record filtered event: %v", err)
	}

	events, err := svc.List(context.Background(), Query{
		TenantID:    "tenant-a",
		CommandName: "purchase.approve",
		ActorID:     "actor-a",
		Limit:       2,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	for _, event := range events {
		if event.CommandName != "purchase.approve" || event.ActorID != "actor-a" {
			t.Fatalf("unexpected event returned: %+v", event)
		}
	}
}

func TestServiceListAppliesTimeWindow(t *testing.T) {
	store := NewInMemoryStore()
	svc, err := NewService(store)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	base := time.Date(2026, 3, 25, 8, 0, 0, 0, time.UTC)
	for i := 0; i < 4; i++ {
		err = svc.Record(context.Background(), Record{
			TenantID:    "tenant-a",
			CommandName: "purchase.approve",
			ActorID:     "actor-a",
			Decision:    policy.DecisionAllow,
			Outcome:     "succeeded",
			OccurredAt:  base.Add(time.Duration(i) * time.Minute),
		})
		if err != nil {
			t.Fatalf("record %d: %v", i, err)
		}
	}

	events, err := svc.List(context.Background(), Query{
		TenantID:       "tenant-a",
		OccurredAfter:  base.Add(1 * time.Minute),
		OccurredBefore: base.Add(3 * time.Minute),
	})
	if err != nil {
		t.Fatalf("list with time window: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events in time window, got %d", len(events))
	}
	for _, event := range events {
		if event.OccurredAt.Before(base.Add(1 * time.Minute)) {
			t.Fatalf("event occurred before lower bound: %s", event.OccurredAt)
		}
		if event.OccurredAt.After(base.Add(3 * time.Minute)) {
			t.Fatalf("event occurred after upper bound: %s", event.OccurredAt)
		}
	}
}
