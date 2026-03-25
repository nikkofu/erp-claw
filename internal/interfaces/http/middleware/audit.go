package middleware

import "github.com/gin-gonic/gin"

func Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		rc := requestContext(c)
		c.Header("X-Audit-Request-ID", rc.RequestID)
		c.Next()
	}
}
