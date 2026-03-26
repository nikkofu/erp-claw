package shared

import (
	"context"
	"errors"
	"fmt"
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

func TestPipelineCapabilityDenialSkipsHandlerAndTransaction(t *testing.T) {
	t.Parallel()

	txManager := &transactionManagerStub{}
	capabilities := &capabilityAuthorizerStub{
		err: fmt.Errorf("%w: tool catalog entry %q is not effectively allowed", ErrCapabilityDenied, "tool-2"),
	}
	pipeline := NewPipeline(PipelineDeps{
		Policy:       policy.StaticEvaluator(policy.DecisionAllow),
		Transactions: txManager,
		Audit:        audit.NoopRecorder(),
		Capabilities: capabilities,
	})

	handlerCalls := 0
	err := pipeline.Execute(context.Background(), Command{
		Name:     "agent.run",
		TenantID: "tenant-a",
		ActorID:  "user-a",
		Payload: map[string]any{
			"agent_profile_id": "profile-1",
			"model_entry_id":   "model-1",
			"tool_entry_ids":   []string{"tool-2"},
		},
	}, func(context.Context, Command) error {
		handlerCalls++
		return nil
	})
	if !errors.Is(err, ErrCapabilityDenied) {
		t.Fatalf("expected capability denied error, got %v", err)
	}
	if capabilities.calls != 1 {
		t.Fatalf("expected capability authorizer to be called once, got %d", capabilities.calls)
	}
	if txManager.calls != 0 {
		t.Fatalf("expected no transaction when capability check fails, got %d", txManager.calls)
	}
	if handlerCalls != 0 {
		t.Fatalf("expected handler not to run, got %d calls", handlerCalls)
	}
}

func TestPipelineCapabilityAuthorizerAllowsHandlerExecution(t *testing.T) {
	t.Parallel()

	txManager := &transactionManagerStub{}
	capabilities := &capabilityAuthorizerStub{}
	pipeline := NewPipeline(PipelineDeps{
		Policy:       policy.StaticEvaluator(policy.DecisionAllow),
		Transactions: txManager,
		Audit:        audit.NoopRecorder(),
		Capabilities: capabilities,
	})

	handlerCalls := 0
	err := pipeline.Execute(context.Background(), Command{
		Name:     "agent.run",
		TenantID: "tenant-a",
		ActorID:  "user-a",
		Payload: map[string]any{
			"agent_profile_id": "profile-1",
			"model_entry_id":   "model-1",
			"tool_entry_ids":   []string{"tool-1"},
		},
	}, func(context.Context, Command) error {
		handlerCalls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if capabilities.calls != 1 {
		t.Fatalf("expected capability authorizer to be called once, got %d", capabilities.calls)
	}
	if txManager.calls != 1 {
		t.Fatalf("expected one transaction call, got %d", txManager.calls)
	}
	if handlerCalls != 1 {
		t.Fatalf("expected handler to run once, got %d", handlerCalls)
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

type capabilityAuthorizerStub struct {
	calls int
	err   error
}

func (s *capabilityAuthorizerStub) AuthorizeCommandCapabilities(context.Context, Command) error {
	s.calls++
	return s.err
}
