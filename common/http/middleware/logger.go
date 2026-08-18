package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Drathveloper/go-web-skeleton/common/log"
)

func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		requestID := uuid.New().String()
		c.Header("X-Request-Id", requestID)
		contextLogger := logger.With("request_id", requestID)
		ctxWithLogger := log.WithLogger(c.Request.Context(), contextLogger)
		c.Request = c.Request.WithContext(ctxWithLogger)
		contextLogger.Debug("incoming request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path)
		c.Next()
		contextLogger.Debug("outgoing response", "status", c.Writer.Status())
	}
}
