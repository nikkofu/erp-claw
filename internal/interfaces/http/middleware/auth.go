package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/platform/iam"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		rc := requestContext(c)
		rc.ActorID = iam.SystemActor.ID
		c.Header("X-Actor-ID", iam.SystemActor.ID)
		c.Next()
	}
}
