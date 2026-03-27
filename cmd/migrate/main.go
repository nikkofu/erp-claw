package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/postgres"
)

func main() {
	configPath := os.Getenv("ERP_CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/local/app.yaml"
	}

	migrationsPath := os.Getenv("ERP_MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	cfg, err := bootstrap.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config (%s): %v", configPath, err)
	}

	bootstrap.StartRuntime(bootstrap.MigrateRole)
	shutdownTelemetry, err := bootstrap.SetupRuntimeTelemetry(cfg, bootstrap.MigrateRole)
	if err != nil {
		log.Fatalf("failed to setup telemetry: %v", err)
	}
	defer func() {
		if shutdownErr := shutdownTelemetry(context.Background()); shutdownErr != nil {
			log.Printf("failed to shutdown telemetry: %v", shutdownErr)
		}
	}()

	pgCfg := postgres.Config{
		DSN:          cfg.Database.DSN,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	}

	direction := strings.ToLower(strings.TrimSpace(os.Getenv("ERP_MIGRATIONS_DIRECTION")))
	if direction == "down" {
		migrator, err := postgres.NewMigrator(pgCfg, migrationsPath)
		if err != nil {
			log.Fatalf("failed to initialize migrator for %s: %v", migrationsPath, err)
		}
		defer func() {
			if closeErr := migrator.Close(); closeErr != nil {
				log.Printf("failed to close migrator: %v", closeErr)
			}
		}()

		if err := migrator.Down(context.Background()); err != nil {
			log.Fatalf("failed to roll back migrations from %s: %v", migrationsPath, err)
		}
		log.Printf("migrations rolled back from %s", migrationsPath)
		return
	}

	if err := postgres.ApplyUp(context.Background(), pgCfg, migrationsPath); err != nil {
		log.Fatalf("failed to apply migrations from %s: %v", migrationsPath, err)
	}

	log.Printf("migrations applied from %s", migrationsPath)
}
