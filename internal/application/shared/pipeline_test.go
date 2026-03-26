package shared

import (
	"context"
	"errors"
	"testing"

	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestPipelineStartsApprovalInsideTransaction(t *testing.T) {
	t.Parallel()

	txManager := &transactionManagerStub{}
	approvals := &approvalStarterStub{}
	pipeline := NewPipeline(PipelineDeps{
		Policy: policy.StaticEvaluator(policy.DecisionRequireApproval),
		Transactions: txManager,
		Audit: audit.NoopRecorder(),
		Approvals: approvals,
	})

	err := pipeline.Execute(context.Background(), Command{
		Name: "purchase.submit",
		TenantID: "tenant-a",
		ActorID: "user-a",
		Payload: map[string]any{
			"approval_definition_id": "def-a",
			"resource_id": "po-1",
		},
	})
	if !errors.Is(err, ErrApprovalRequired) {
		t.Fatalf("expected approval required error, got %v", err)
	}
	if txManager.calls != 1 {
		t.Fatalf("expected approval path to use one transaction, got %d", txManager.calls)
	}
	if approvals.calls != 1 {
		t.Fatalf("expected approval starter to be called once, got %d", approvals.calls)
	}
}

type transactionManagerStub struct {
	calls int
}

func (s *transactionManagerStub) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	s.calls++
	if fn == nil {
		return nil
	}
	return fn(ctx)
}

type approvalStarterStub struct {
	calls int
}

func (s *approvalStarterStub) StartApprovalForCommand(context.Context, Command) error {
	s.calls++
	return nil
}
