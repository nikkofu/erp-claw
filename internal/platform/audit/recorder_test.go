package audit

import (
	"context"
	"testing"

	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestInMemoryRecorderListsRecordsByTenantAndLimit(t *testing.T) {
	recorder := NewInMemoryRecorder()
	for _, record := range []Record{
		{
			CommandName: "procurement.purchase_orders.create",
			TenantID:    "tenant-a",
			ActorID:     "actor-a",
			Decision:    policy.DecisionAllow,
			Outcome:     "succeeded",
		},
		{
			CommandName: "procurement.purchase_orders.submit",
			TenantID:    "tenant-a",
			ActorID:     "actor-a",
			Decision:    policy.DecisionAllow,
			Outcome:     "succeeded",
		},
		{
			CommandName: "procurement.purchase_orders.create",
			TenantID:    "tenant-b",
			ActorID:     "actor-b",
			Decision:    policy.DecisionAllow,
			Outcome:     "succeeded",
		},
	} {
		if err := recorder.Record(context.Background(), record); err != nil {
			t.Fatalf("record audit: %v", err)
		}
	}

	records, err := recorder.List(context.Background(), Query{
		TenantID: "tenant-a",
		Limit:    1,
	})
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].CommandName != "procurement.purchase_orders.submit" {
		t.Fatalf("expected latest record submit, got %s", records[0].CommandName)
	}
}

func TestInMemoryRecorderListsByCommandName(t *testing.T) {
	recorder := NewInMemoryRecorder()
	for _, record := range []Record{
		{
			CommandName: "procurement.purchase_orders.create",
			TenantID:    "tenant-a",
			ActorID:     "actor-a",
			Decision:    policy.DecisionAllow,
			Outcome:     "succeeded",
		},
		{
			CommandName: "procurement.purchase_orders.submit",
			TenantID:    "tenant-a",
			ActorID:     "actor-a",
			Decision:    policy.DecisionAllow,
			Outcome:     "succeeded",
		},
	} {
		if err := recorder.Record(context.Background(), record); err != nil {
			t.Fatalf("record audit: %v", err)
		}
	}

	records, err := recorder.List(context.Background(), Query{
		TenantID:    "tenant-a",
		CommandName: "procurement.purchase_orders.create",
	})
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].CommandName != "procurement.purchase_orders.create" {
		t.Fatalf("expected create command, got %s", records[0].CommandName)
	}
}
