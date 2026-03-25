package integration

import (
	"context"
	"testing"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/platform/policy"
)

func TestCommandPipelineRejectsDeniedPolicy(t *testing.T) {
	p := shared.NewPipeline(shared.PipelineDeps{
		Policy: policy.StaticEvaluator(policy.DecisionDeny),
	})

	err := p.Execute(context.Background(), shared.Command{
		Name:     "customers.create",
		TenantID: "tenant-a",
		ActorID:  "actor-a",
		Payload:  map[string]any{"name": "Acme"},
	})

	if err == nil {
		t.Fatalf("expected denied policy to reject command")
	}
}
