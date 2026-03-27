package bootstrap

import (
	"context"
	"fmt"

	observabilityotel "github.com/nikkofu/erp-claw/internal/infrastructure/observability/otel"
)

type RuntimeRole string

const (
	APIServerRole    RuntimeRole = "api-server"
	AgentGatewayRole RuntimeRole = "agent-gateway"
	WorkerRole       RuntimeRole = "worker"
	SchedulerRole    RuntimeRole = "scheduler"
	MigrateRole      RuntimeRole = "migrate"
)

func (r RuntimeRole) String() string {
	return string(r)
}

// StartRuntime is a minimal stub used by the runtime command entry points.
func StartRuntime(role RuntimeRole) {
	fmt.Printf("starting runtime: %s\n", role)
}

func SetupRuntimeTelemetry(cfg Config, role RuntimeRole) (func(context.Context) error, error) {
	return observabilityotel.SetupTracing(observabilityotel.SetupConfig{
		ServiceName:     "erp-claw",
		Role:            role.String(),
		TracingEndpoint: cfg.Telemetry.TracingEndpoint,
	})
}
