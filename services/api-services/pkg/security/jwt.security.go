package security

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"
)

// Payload represents the JWT claims structure.
type Payload struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// NewPayload creates a new JWT payload with the given user details and expiration duration.
func NewPayload(userID uuid.UUID, duration time.Duration) *Payload {
	return &Payload{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}
}

// CreateToken generates a signed JWT token from the payload using the provided secret.
func CreateToken(payload *Payload, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		logger.Error("failed to sign JWT token", logger.ErrorField(err))
		return "", err
	}
	return signedToken, nil
}

// VerifyToken parses and validates the JWT token using the provided secret, returning the payload if valid.
func VerifyToken(tokenStr string, secret string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Payload{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			logger.Warn("JWT token expired", logger.String("token", tokenStr))
		} else {
			logger.Error("failed to parse JWT token", logger.ErrorField(err), logger.String("token", tokenStr))
		}
		return nil, err
	}

	if !token.Valid {
		logger.Warn("invalid JWT token", logger.String("token", tokenStr))
		return nil, jwt.ErrSignatureInvalid
	}

	payload, ok := token.Claims.(*Payload)
	if !ok {
		logger.Error("failed to assert JWT claims as Payload")
		return nil, jwt.ErrInvalidKeyType
	}

	return payload, nil
}

// ExtractTokenFromHeader extracts the JWT token from the Authorization header.
func ExtractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}
