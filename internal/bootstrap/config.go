package bootstrap

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env           string              `yaml:"env"`
	HTTP          HTTPConfig          `yaml:"http"`
	Database      DatabaseConfig      `yaml:"database"`
	Redis         RedisConfig         `yaml:"redis"`
	NATS          NATSConfig          `yaml:"nats"`
	ObjectStorage ObjectStorageConfig `yaml:"objectStorage"`
	Telemetry     TelemetryConfig     `yaml:"telemetry"`
}

type HTTPConfig struct {
	Host                string `yaml:"host"`
	Port                int    `yaml:"port"`
	ReadTimeoutSeconds  int    `yaml:"readTimeoutSeconds"`
	WriteTimeoutSeconds int    `yaml:"writeTimeoutSeconds"`
}

type DatabaseConfig struct {
	DSN          string `yaml:"dsn"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
}

type NATSConfig struct {
	Servers []string `yaml:"servers"`
	Cluster string   `yaml:"cluster"`
}

type ObjectStorageConfig struct {
	Provider string `yaml:"provider"`
	Bucket   string `yaml:"bucket"`
	Region   string `yaml:"region"`
}

type TelemetryConfig struct {
	MetricsPath     string `yaml:"metricsPath"`
	TracingEndpoint string `yaml:"tracingEndpoint"`
}

func LoadConfig(path string) (Config, error) {
	cfg := defaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := applyEnvOverrides(&cfg); err != nil {
				return Config{}, err
			}
			return cfg, nil
		}
		return Config{}, err
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}

	if err := applyEnvOverrides(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		Env: "local",
		HTTP: HTTPConfig{
			Host:                "0.0.0.0",
			Port:                8080,
			ReadTimeoutSeconds:  5,
			WriteTimeoutSeconds: 5,
		},
		Database: DatabaseConfig{
			DSN:          "postgres://erp:erp@localhost:5432/erp_claw?sslmode=disable",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
		},
		NATS: NATSConfig{
			Servers: []string{"nats://localhost:4222"},
			Cluster: "local",
		},
		ObjectStorage: ObjectStorageConfig{
			Provider: "minio",
			Bucket:   "erp-claw-local",
			Region:   "us-east-1",
		},
		Telemetry: TelemetryConfig{
			MetricsPath:     "/metrics",
			TracingEndpoint: "http://localhost:4317",
		},
	}
}

// DefaultConfig returns the baseline configuration.
func DefaultConfig() Config {
	return defaultConfig()
}

func applyEnvOverrides(cfg *Config) error {
	if v := strings.TrimSpace(os.Getenv("ERP_ENV")); v != "" {
		cfg.Env = v
	}
	if v := strings.TrimSpace(os.Getenv("ERP_HTTP_HOST")); v != "" {
		cfg.HTTP.Host = v
	}
	if v := strings.TrimSpace(os.Getenv("ERP_HTTP_PORT")); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid ERP_HTTP_PORT %q: %w", v, err)
		}
		cfg.HTTP.Port = port
	}
	if v := strings.TrimSpace(os.Getenv("ERP_STORE_DATABASE_DSN")); v != "" {
		cfg.Database.DSN = v
	}
	if v := strings.TrimSpace(os.Getenv("ERP_STORE_REDIS_ADDR")); v != "" {
		cfg.Redis.Addr = v
	}
	if v := strings.TrimSpace(os.Getenv("ERP_TELEMETRY_METRICS_PATH")); v != "" {
		cfg.Telemetry.MetricsPath = v
	}
	if v := strings.TrimSpace(os.Getenv("ERP_TELEMETRY_TRACING_ENDPOINT")); v != "" {
		cfg.Telemetry.TracingEndpoint = v
	}
	return nil
}
