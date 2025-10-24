package middleware

import (
	"net/http"

	"github.com/LiusCraft/x-mindflow/mock-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Recovery 恢复 panic
func Recovery(log logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		log.Error("panic recovered",
			"request_id", c.GetString("request_id"),
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
	})
}
