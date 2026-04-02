package otel

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
)

type SetupConfig struct {
	ServiceName     string
	Role            string
	TracingEndpoint string
}

type ShutdownFunc func(context.Context) error

func SetupTracing(cfg SetupConfig) (ShutdownFunc, error) {
	serviceName := strings.TrimSpace(cfg.ServiceName)
	if serviceName == "" {
		serviceName = "erp-claw"
	}

	role := strings.TrimSpace(cfg.Role)
	if role == "" {
		role = "unknown"
	}

	endpoint := strings.TrimSpace(cfg.TracingEndpoint)
	if endpoint == "" {
		log.Printf("otel tracing disabled (service=%s role=%s, empty endpoint)", serviceName, role)
		return noOpShutdown, nil
	}

	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid tracing endpoint %q", endpoint)
	}

	log.Printf(
		"otel tracing configured (service=%s role=%s endpoint=%s)",
		serviceName,
		role,
		endpoint,
	)

	// Phase 1 seam: provider/exporter bootstrap is intentionally deferred.
	return noOpShutdown, nil
}

func noOpShutdown(context.Context) error {
	return nil
}
