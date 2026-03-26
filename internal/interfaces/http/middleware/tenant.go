package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/platform/tenant"
)

var errMissingTenantHeader = errors.New("missing tenant header")

func Tenant(resolver tenant.Resolver) gin.HandlerFunc {
	if resolver == nil {
		resolver = tenant.SimpleResolver{}
	}

	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			c.AbortWithError(400, errMissingTenantHeader)
			return
		}
		route, err := resolver.Resolve(tenantID)
		if err != nil {
			c.AbortWithError(404, err)
			return
		}
		rc := requestContext(c)
		rc.TenantID = route.TenantID
		c.Set("tenant_id", route.TenantID)
		c.Header("X-Tenant-ID", route.TenantID)
		if route.Isolation != "" {
			c.Header("X-Tenant-Isolation", route.Isolation)
		}
		c.Next()
	}
}
