package presenter

import "github.com/gin-gonic/gin"

func OK(c *gin.Context, data any) {
	c.JSON(200, gin.H{
		"data": data,
		"meta": gin.H{"request_id": c.GetString("request_id")},
	})
}
