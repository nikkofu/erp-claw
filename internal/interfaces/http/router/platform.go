package router

import (
	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/bootstrap"
)

func registerPlatformRoutes(rg *gin.RouterGroup, container *bootstrap.Container) {
	registerHealthRoutes(rg, container)
	registerControlPlaneRoutes(rg, container)
}
