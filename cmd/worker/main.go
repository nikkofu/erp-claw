package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nikkofu/erp-claw/internal/bootstrap"
	messagingnats "github.com/nikkofu/erp-claw/internal/infrastructure/messaging/nats"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/postgres"
	"github.com/nikkofu/erp-claw/internal/platform/eventbus"
)

func main() {
	configPath := os.Getenv("ERP_CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/local/app.yaml"
	}

	cfg, err := bootstrap.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config (%s): %v", configPath, err)
	}

	bootstrap.StartRuntime(bootstrap.WorkerRole)
	shutdownTelemetry, err := bootstrap.SetupRuntimeTelemetry(cfg, bootstrap.WorkerRole)
	if err != nil {
		log.Fatalf("failed to setup telemetry: %v", err)
	}
	defer func() {
		if shutdownErr := shutdownTelemetry(context.Background()); shutdownErr != nil {
			log.Printf("failed to shutdown telemetry: %v", shutdownErr)
		}
	}()

	db, err := postgres.New(postgres.Config{
		DSN:          cfg.Database.DSN,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	})
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close postgres connection: %v", closeErr)
		}
	}()

	nc, err := messagingnats.New(messagingnats.Config{
		Servers: cfg.NATS.Servers,
		Cluster: cfg.NATS.Cluster,
	})
	if err != nil {
		log.Fatalf("failed to connect nats: %v", err)
	}
	defer nc.Close()

	bus, err := eventbus.NewNATS(nc)
	if err != nil {
		log.Fatalf("failed to initialize event bus: %v", err)
	}

	log.Printf("worker started (env=%s, config=%s, nats_servers=%v)", cfg.Env, configPath, cfg.NATS.Servers)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runOutboxPoller(ctx, db, bus, 2*time.Second); err != nil {
		log.Fatalf("worker stopped with error: %v", err)
	}
}

func runOutboxPoller(ctx context.Context, db *sql.DB, bus eventbus.Bus, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Print("worker shutdown requested")
			return nil
		case <-ticker.C:
			if err := pollOutboxBatch(ctx, db, bus); err != nil {
				log.Printf("outbox poll iteration failed: %v", err)
			}
		}
	}
}

func pollOutboxBatch(_ context.Context, _ *sql.DB, _ eventbus.Bus) error {
	return nil
}
