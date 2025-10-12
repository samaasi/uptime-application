package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/samaasi/uptime-application/services/api-services/internal/common"
	"github.com/samaasi/uptime-application/services/api-services/internal/utils"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"
	"github.com/samaasi/uptime-application/services/api-services/pkg/security"
)

// AuthMiddleware is a Gin middleware that verifies JWT authentication.
func AuthMiddleware(appKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := security.ExtractTokenFromHeader(c)
		if tokenStr == "" {
			utils.SendUnauthorizedWithDetail(c, "Missing or invalid Authorization header", "Authorization header is required")
			c.Abort()
			return
		}

		payload, err := security.VerifyToken(tokenStr, appKey)
		if err != nil {
			logger.Warn("Invalid JWT token", logger.ErrorField(err), logger.String("request_id", utils.GetRequestID(c)))
			utils.SendUnauthorizedWithDetail(c, "Invalid or expired token", "Token is either invalid or expired")
			c.Abort()
			return
		}

		c.Set(string(common.AuthorizationPayloadContextKey), payload)
		c.Set(string(common.UserIDContextKey), payload.UserID.String())

		c.Next()
	}
}

// OptionalAuthMiddleware is a Gin middleware that verifies JWT if present, but allows unauthenticated requests.
func OptionalAuthMiddleware(appKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := security.ExtractTokenFromHeader(c)
		if tokenStr == "" {
			c.Next()
			return
		}

		payload, err := security.VerifyToken(tokenStr, appKey)
		if err != nil {
			logger.Warn("Invalid JWT token in optional auth", logger.ErrorField(err), logger.String("request_id", utils.GetRequestID(c)))
			c.Next()
			return
		}

		c.Set(string(common.AuthorizationPayloadContextKey), payload)
		c.Set(string(common.UserIDContextKey), payload.UserID.String())

		c.Next()
	}
}
