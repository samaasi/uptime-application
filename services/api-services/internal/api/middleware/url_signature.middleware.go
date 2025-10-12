package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"
	"github.com/samaasi/uptime-application/services/api-services/pkg/urlsigner"
)

// URLSignatureMiddleware is a Gin middleware to validate the signed URL.
// It reconstructs the URL from the request (path + query) and validates it.
func URLSignatureMiddleware(signer *urlsigner.Signer) gin.HandlerFunc {
	return func(c *gin.Context) {
		fullURL := c.Request.URL.String()

		valid, err := signer.Validate(fullURL)
		if err != nil || !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired URL signature"})

			if err != nil {
				logger.Error("URL validation error", logger.ErrorField(err))
			}
			return
		}

		c.Next()
	}
}
