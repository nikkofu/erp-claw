package policy

import "context"

type staticEvaluator struct {
	decision Decision
}

// StaticEvaluator returns a deterministic evaluator useful for tests.
func StaticEvaluator(decision Decision) Evaluator {
	return staticEvaluator{decision: decision}
}

func (s staticEvaluator) Evaluate(_ context.Context, _ Input) (Decision, error) {
	return s.decision, nil
}
