package middleware

import (
	"time"

	"github.com/samaasi/uptime-application/services/api-services/internal/common"
	"github.com/samaasi/uptime-application/services/api-services/internal/utils"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware logs the start and completion time of each HTTP request, along with the estimated duration of the request.
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := utils.GetRequestID(c)
		clientIP := utils.GetClientIP(c)

		logger.Info("Request started",
			logger.String("request_id", requestID),
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("client_ip", clientIP),
		)

		c.Set(string(common.RequestStartTimeKey), time.Now())

		c.Next()

		duration := time.Duration(0)
		if startTimeVal, ok := c.Get(string(common.RequestStartTimeKey)); ok {
			if startTime, isTime := startTimeVal.(time.Time); isTime {
				duration = time.Since(startTime)
			}
		}

		logger.Info("Request completed",
			logger.String("request_id", requestID),
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.Int("status", c.Writer.Status()),
			logger.Duration("duration", duration),
			logger.String("client_ip", clientIP),
		)
	}
}
