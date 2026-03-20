package middleware

import (
	"time"

	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/constants"
	"github.com/AmithSAI007/prj-apex-transcode-submitter/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger returns a Gin middleware that emits a structured log entry
// for every completed HTTP request. The log includes method, path, status
// code, latency, client IP, and the request correlation ID. This data is
// intended for the BigQuery audit log sink.
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		c.Next()

		elapsed := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		ctx := c.Request.Context()

		fields := append(utils.LogFieldsFromContext(ctx),
			zap.String(constants.LogKeyLayer, "middleware"),
			zap.String(constants.LogKeyMethod, c.Request.Method),
			zap.String("path", path),
			zap.Int("http_status", status),
			zap.Int64(constants.LogKeyDuration, elapsed),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("body_size", c.Writer.Size()),
		)

		switch {
		case status >= 500:
			logger.Error("request completed", fields...)
		case status >= 400:
			logger.Warn("request completed", fields...)
		default:
			logger.Info("request completed", fields...)
		}
	}
}
