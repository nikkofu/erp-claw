package bootstrap

import "fmt"

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
