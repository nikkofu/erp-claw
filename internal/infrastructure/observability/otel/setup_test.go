package otel

import (
	"context"
	"testing"
)

func TestSetupTracingAllowsEmptyEndpoint(t *testing.T) {
	shutdown, err := SetupTracing(SetupConfig{
		ServiceName:     "erp-claw",
		Role:            "api-server",
		TracingEndpoint: "",
	})
	if err != nil {
		t.Fatalf("expected no error for empty endpoint, got %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown function for empty endpoint")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("expected no-op shutdown error nil, got %v", err)
	}
}

func TestSetupTracingRejectsInvalidEndpoint(t *testing.T) {
	if _, err := SetupTracing(SetupConfig{
		ServiceName:     "erp-claw",
		Role:            "worker",
		TracingEndpoint: "localhost:4317",
	}); err == nil {
		t.Fatal("expected error for invalid endpoint without scheme")
	}
}

func TestSetupTracingAcceptsValidEndpoint(t *testing.T) {
	shutdown, err := SetupTracing(SetupConfig{
		ServiceName:     "erp-claw",
		Role:            "scheduler",
		TracingEndpoint: "http://localhost:4317",
	})
	if err != nil {
		t.Fatalf("expected valid endpoint, got error %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown function for valid endpoint")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("expected shutdown to succeed, got %v", err)
	}
}
