package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikkofu/erp-claw/internal/bootstrap"
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

	bootstrap.StartRuntime(bootstrap.AgentGatewayRole)

	container := bootstrap.NewContainer(cfg)

	log.Printf(
		"agent gateway started (env=%s, config=%s, health_service=%t, workspace_channels=%d)",
		cfg.Env,
		configPath,
		container.Health != nil,
		container.WorkspaceGateway.ChannelCount(),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Print("agent gateway shutdown requested")
}
