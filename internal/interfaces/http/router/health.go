package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
	"github.com/nikkofu/erp-claw/internal/interfaces/http/presenter"
)

func registerHealthRoutes(rg *gin.RouterGroup, container *bootstrap.Container) {
	if container == nil {
		panic("router: container must not be nil")
	}

	healthGroup := rg.Group("/health")
	healthGroup.GET("/livez", func(c *gin.Context) {
		if !container.Health.Liveness() {
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		presenter.OK(c, gin.H{"status": "live"})
	})

	healthGroup.GET("/readyz", func(c *gin.Context) {
		if !container.Health.Readiness() {
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		presenter.OK(c, gin.H{"status": "ready"})
	})
}
