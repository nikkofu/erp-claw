package shared

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nikkofu/erp-claw/internal/platform/audit"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

var (
	ErrPolicyDenied      = errors.New("policy denied command")
	ErrApprovalRequired  = errors.New("policy requires approval")
)

// Handler executes a command inside the transaction boundary.
type Handler func(context.Context, Command) error

type PipelineDeps struct {
	Policy       policy.Evaluator
	Transactions TransactionManager
	Audit        audit.Recorder
}

type Pipeline struct {
	policy       policy.Evaluator
	transactions TransactionManager
	audit        audit.Recorder
}

func NewPipeline(deps PipelineDeps) *Pipeline {
	if deps.Policy == nil {
		deps.Policy = policy.StaticEvaluator(policy.DecisionAllow)
	}
	if deps.Transactions == nil {
		deps.Transactions = NoopTransactionManager()
	}
	if deps.Audit == nil {
		deps.Audit = audit.NoopRecorder()
	}

	return &Pipeline{
		policy:       deps.Policy,
		transactions: deps.Transactions,
		audit:        deps.Audit,
	}
}

func (p *Pipeline) Execute(ctx context.Context, cmd Command, handlers ...Handler) error {
	decision, err := p.policy.Evaluate(ctx, policy.Input{
		CommandName: cmd.Name,
		TenantID:    cmd.TenantID,
		ActorID:     cmd.ActorID,
		Payload:     cmd.Payload,
	})
	if err != nil {
		return err
	}

	switch decision {
	case policy.DecisionDeny:
		err = ErrPolicyDenied
		return p.recordAndReturn(ctx, cmd, decision, "rejected", err)
	case policy.DecisionRequireApproval:
		err = ErrApprovalRequired
		return p.recordAndReturn(ctx, cmd, decision, "pending_approval", err)
	}

	handler := noOpHandler
	if len(handlers) > 0 && handlers[0] != nil {
		handler = handlers[0]
	}

	err = p.transactions.WithinTransaction(ctx, func(txCtx context.Context) error {
		return handler(txCtx, cmd)
	})
	if err != nil {
		return p.recordAndReturn(ctx, cmd, decision, "failed", err)
	}

	return p.recordAndReturn(ctx, cmd, decision, "succeeded", nil)
}

func (p *Pipeline) recordAndReturn(ctx context.Context, cmd Command, decision policy.Decision, outcome string, execErr error) error {
	record := audit.Record{
		CommandName:   cmd.Name,
		TenantID:      cmd.TenantID,
		ActorID:       cmd.ActorID,
		Decision:      decision,
		Outcome:       outcome,
		CorrelationID: auditPayloadString(cmd.Payload, "correlation_id"),
		ResourceType:  auditPayloadString(cmd.Payload, "resource_type"),
		ResourceID:    auditPayloadString(cmd.Payload, "resource_id"),
		OccurredAt:    time.Now().UTC(),
	}
	if execErr != nil {
		record.Error = execErr.Error()
	}

	if err := p.audit.Record(ctx, record); err != nil {
		if execErr != nil {
			return fmt.Errorf("%w (audit: %v)", execErr, err)
		}
		return err
	}

	return execErr
}

func noOpHandler(context.Context, Command) error {
	return nil
}

func auditPayloadString(payload any, key string) string {
	asMap, ok := payload.(map[string]any)
	if !ok {
		return ""
	}
	value, ok := asMap[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}
