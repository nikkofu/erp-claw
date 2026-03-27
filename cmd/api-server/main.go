package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/router"
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

	bootstrap.StartRuntime(bootstrap.APIServerRole)
	shutdownTelemetry, err := bootstrap.SetupRuntimeTelemetry(cfg, bootstrap.APIServerRole)
	if err != nil {
		log.Fatalf("failed to setup telemetry: %v", err)
	}
	defer func() {
		if shutdownErr := shutdownTelemetry(context.Background()); shutdownErr != nil {
			log.Printf("failed to shutdown telemetry: %v", shutdownErr)
		}
	}()

	gin.SetMode(gin.ReleaseMode)
	container := bootstrap.NewContainer(cfg)
	engine := router.New(router.WithContainer(container))

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  time.Duration(cfg.HTTP.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.HTTP.WriteTimeoutSeconds) * time.Second,
	}

	log.Printf("api server listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("api server stopped: %v", err)
	}
}
