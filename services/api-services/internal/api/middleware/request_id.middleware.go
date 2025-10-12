package middleware

import (
	"context"
	"time"

	"github.com/samaasi/uptime-application/services/api-services/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware is a Gin middleware that extracts or generates a request ID from the "X-Request-ID" header.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")

		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Writer.Header().Set("X-Request-ID", requestID)

		startTime := time.Now()

		c.Set(string(common.RequestIDContextKey), requestID)
		c.Set(string(common.RequestStartTimeKey), startTime)

		newCtx := context.WithValue(c.Request.Context(), common.RequestIDContextKey, requestID)
		newCtx = context.WithValue(newCtx, common.RequestStartTimeKey, startTime)
		c.Request = c.Request.WithContext(newCtx)

		c.Next()
	}
}
