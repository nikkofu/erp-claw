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

func TestInMemoryRecorderListsByActorDecisionOutcomeAndOffset(t *testing.T) {
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
			Outcome:     "failed",
		},
		{
			CommandName: "procurement.purchase_orders.approve",
			TenantID:    "tenant-a",
			ActorID:     "actor-a",
			Decision:    policy.DecisionAllow,
			Outcome:     "succeeded",
		},
		{
			CommandName: "procurement.purchase_orders.create",
			TenantID:    "tenant-a",
			ActorID:     "actor-b",
			Decision:    policy.DecisionDeny,
			Outcome:     "rejected",
		},
	} {
		if err := recorder.Record(context.Background(), record); err != nil {
			t.Fatalf("record audit: %v", err)
		}
	}

	records, err := recorder.List(context.Background(), Query{
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Decision: policy.DecisionAllow,
		Outcome:  "succeeded",
		Offset:   1,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record after offset, got %d", len(records))
	}
	if records[0].CommandName != "procurement.purchase_orders.create" {
		t.Fatalf("expected offset record to be create, got %s", records[0].CommandName)
	}
}

func TestInMemoryRecorderListsByCommandPrefix(t *testing.T) {
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
			CommandName: "controlplane.actors.upsert",
			TenantID:    "tenant-a",
			ActorID:     "actor-admin",
			Decision:    policy.DecisionAllow,
			Outcome:     "succeeded",
		},
	} {
		if err := recorder.Record(context.Background(), record); err != nil {
			t.Fatalf("record audit: %v", err)
		}
	}

	records, err := recorder.List(context.Background(), Query{
		TenantID:      "tenant-a",
		CommandPrefix: "procurement.purchase_orders.",
	})
	if err != nil {
		t.Fatalf("list audit by command prefix: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	for _, record := range records {
		if record.CommandName != "procurement.purchase_orders.create" && record.CommandName != "procurement.purchase_orders.submit" {
			t.Fatalf("unexpected command in prefix filter result: %s", record.CommandName)
		}
	}
}
