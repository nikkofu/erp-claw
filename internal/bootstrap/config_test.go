package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := LoadConfig("testdata/does-not-need-to-exist-yet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTP.Port != 8080 {
		t.Fatalf("expected HTTP port 8080, got %d", cfg.HTTP.Port)
	}
}

func TestLoadConfigWithYAMLOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
http:
  port: 5050
database:
  dsn: postgres://override
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cfg.HTTP.Port; got != 5050 {
		t.Fatalf("expected overridden HTTP port, got %d", got)
	}
	if got := cfg.Database.DSN; !strings.Contains(got, "override") {
		t.Fatalf("expected database DSN override, got %q", got)
	}
	if cfg.Redis.Addr != "localhost:6379" {
		t.Fatalf("expected default redis addr, got %q", cfg.Redis.Addr)
	}
}

func TestLoadConfigInvalidHTTPPortEnv(t *testing.T) {
	t.Setenv("ERP_HTTP_PORT", "bad-port")
	if _, err := LoadConfig("testdata/does-not-need-to-exist-yet"); err == nil {
		t.Fatalf("expected error for invalid ERP_HTTP_PORT")
	} else if !strings.Contains(err.Error(), "invalid ERP_HTTP_PORT") {
		t.Fatalf("unexpected error for bad HTTP port: %v", err)
	}
}

func TestLoadConfigRejectsUnknownFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "unknown.yaml")
	content := `
http:
  port: 9090
unknownSection:
  value: yes
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := LoadConfig(path); err == nil {
		t.Fatalf("expected error for unknown fields")
	} else if !strings.Contains(err.Error(), "unknown field") && !strings.Contains(err.Error(), "field unknownSection not found") {
		t.Fatalf("unexpected error for unknown field: %v", err)
	}
}

func TestEnvOverrideWinsOverYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "override.yaml")
	content := `
http:
  port: 5051
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("ERP_HTTP_PORT", "6060")

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTP.Port != 6060 {
		t.Fatalf("expected env override to win (6060), got %d", cfg.HTTP.Port)
	}
}
