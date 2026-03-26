package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/erp-claw/internal/platform/iam"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		actorID := strings.TrimSpace(c.GetHeader("X-Actor-ID"))
		if actorID == "" {
			actorID = iam.SystemActor.ID
		}
		rc := requestContext(c)
		rc.ActorID = actorID
		c.Set("actor_id", actorID)
		c.Header("X-Actor-ID", actorID)
		c.Next()
	}
}
