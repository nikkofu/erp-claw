package router

import (
	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/middleware"
	"github.com/nikkofu/erp-claw/internal/platform/tenant"
)

type Option func(*options)

type options struct {
	container *bootstrap.Container
}

func WithContainer(container *bootstrap.Container) Option {
	return func(o *options) {
		if container != nil {
			o.container = container
		}
	}
}

func New(opts ...Option) *gin.Engine {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.container == nil {
		cfg.container = bootstrap.NewContainer(bootstrap.DefaultConfig())
	}
	tenantResolver := tenant.ChainResolver{
		Primary:  tenant.CatalogResolver{Catalog: cfg.container.TenantCatalog},
		Fallback: tenant.SimpleResolver{},
	}

	ginEngine := gin.New()
	ginEngine.Use(gin.Recovery())
	ginEngine.Use(middleware.RequestID())
	ginEngine.Use(middleware.Logging())
	ginEngine.Use(middleware.Tenant(tenantResolver))
	ginEngine.Use(middleware.Auth())
	ginEngine.Use(middleware.Audit())

	registerAdminRoutes(ginEngine.Group("/api/admin/v1"), cfg.container)
	registerPlatformRoutes(ginEngine.Group("/api/platform/v1"), cfg.container)
	registerWorkspaceRoutes(ginEngine.Group("/api/workspace/v1"), cfg.container)
	registerIntegrationRoutes(ginEngine.Group("/api/integration/v1"), cfg.container)

	return ginEngine
}

func defaultOptions() *options {
	return &options{}
}
