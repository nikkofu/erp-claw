package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		rc := requestContext(c)
		log.Printf("req=%s tenant=%s method=%s path=%s status=%d duration=%s",
			rc.RequestID,
			rc.TenantID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
		)
	}
}
