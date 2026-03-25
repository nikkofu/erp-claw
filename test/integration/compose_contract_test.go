package integration

import (
	"os"
	"strings"
	"testing"
)

func TestDockerComposeIncludesRequiredServices(t *testing.T) {
	data, err := os.ReadFile("../../docker-compose.yml")
	if err != nil {
		t.Fatalf("read compose: %v", err)
	}

	required := []string{"postgres", "redis", "nats", "minio", "otel-collector", "prometheus", "grafana"}
	content := string(data)
	for _, service := range required {
		if !strings.Contains(content, service+":") {
			t.Fatalf("expected service %q in compose file", service)
		}
	}
	if !strings.Contains(content, "configs/local/docker.env") {
		t.Fatalf("expected compose file to reference configs/local/docker.env so local env overrides work")
	}
}
