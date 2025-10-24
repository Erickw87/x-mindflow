package middleware

import (
	"time"

	"github.com/LiusCraft/x-mindflow/mock-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Logger 结构化日志中间件
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Info("request completed",
			"request_id", c.GetString("request_id"),
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", status,
			"latency", latency.String(),
			"client_ip", c.ClientIP())
	}
}
