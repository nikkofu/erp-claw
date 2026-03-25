package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nikkofu/erp-claw/internal/bootstrap"
	messagingnats "github.com/nikkofu/erp-claw/internal/infrastructure/messaging/nats"
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

	bootstrap.StartRuntime(bootstrap.SchedulerRole)

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

	log.Printf("scheduler started (env=%s, config=%s, nats_servers=%v)", cfg.Env, configPath, cfg.NATS.Servers)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runScheduleLoop(ctx, bus, time.Minute); err != nil {
		log.Fatalf("scheduler stopped with error: %v", err)
	}
}

func runScheduleLoop(ctx context.Context, bus eventbus.Bus, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Print("scheduler shutdown requested")
			return nil
		case tickAt := <-ticker.C:
			if err := emitScheduleTick(ctx, bus, tickAt); err != nil {
				log.Printf("scheduler tick emit failed: %v", err)
			}
		}
	}
}

func emitScheduleTick(ctx context.Context, bus eventbus.Bus, tickAt time.Time) error {
	return bus.Publish(ctx, eventbus.Event{
		Topic: "platform.scheduler.tick",
		Payload: map[string]any{
			"emitted_at": tickAt.UTC().Format(time.RFC3339),
		},
	})
}
