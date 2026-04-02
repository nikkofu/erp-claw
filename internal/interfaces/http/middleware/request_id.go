package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/platform/runtime"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := fmt.Sprintf("req-%d", time.Now().UnixNano())
		rc := requestContext(c)
		rc.RequestID = reqID
		c.Set("request_id", reqID)
		c.Next()
	}
}

func requestContext(c *gin.Context) *runtime.RequestContext {
	if val, ok := c.Get(runtime.RequestContextKey); ok {
		if rc, ok := val.(*runtime.RequestContext); ok {
			return rc
		}
	}
	rc := &runtime.RequestContext{}
	c.Set(runtime.RequestContextKey, rc)
	c.Request = c.Request.WithContext(runtime.WithRequestContext(c.Request.Context(), rc))
	return rc
}
