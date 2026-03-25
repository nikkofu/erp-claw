package bootstrap

import "github.com/nikkofu/erp-claw/internal/platform/health"

type Container struct {
	Config Config
	Health *health.Service
}

func NewContainer(cfg Config) *Container {
	return &Container{
		Config: cfg,
		Health: health.NewService(),
	}
}
