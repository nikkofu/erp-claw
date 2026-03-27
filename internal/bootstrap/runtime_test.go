package bootstrap

import "testing"

func TestRuntimeRoleString(t *testing.T) {
	if APIServerRole.String() != "api-server" {
		t.Fatalf("expected api-server, got %q", APIServerRole.String())
	}
}

func TestNewContainerProvidesSupplyChainService(t *testing.T) {
	container := NewContainer(DefaultConfig())
	if container.SupplyChain == nil {
		t.Fatal("expected supply-chain service to be wired")
	}
}

func TestSetupRuntimeTelemetryWithEmptyEndpoint(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Telemetry.TracingEndpoint = ""

	shutdown, err := SetupRuntimeTelemetry(cfg, APIServerRole)
	if err != nil {
		t.Fatalf("expected no telemetry setup error, got %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected telemetry shutdown function")
	}
}

func TestSetupRuntimeTelemetryRejectsInvalidEndpoint(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Telemetry.TracingEndpoint = "localhost:4317"

	if _, err := SetupRuntimeTelemetry(cfg, WorkerRole); err == nil {
		t.Fatal("expected invalid tracing endpoint error")
	}
}
