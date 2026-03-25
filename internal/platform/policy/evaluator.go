package policy

import "context"

// Input contains the minimal command attributes needed for policy checks.
type Input struct {
	CommandName string
	TenantID    string
	ActorID     string
	Payload     any
}

// Evaluator decides whether a command can proceed.
type Evaluator interface {
	Evaluate(ctx context.Context, input Input) (Decision, error)
}
