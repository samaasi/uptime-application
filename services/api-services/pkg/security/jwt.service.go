package security

import (
    "errors"
    "github.com/samaasi/uptime-application/services/api-services/pkg/logger"
    "time"
)

// JWTService provides methods to create and verify JWT tokens using a configured secret.
type JWTService struct {
    secret     string
    expiration time.Duration
}

// NewJWTService constructs a JWTService with the provided secret and default expiration.
func NewJWTService(secret string, expiration time.Duration) (*JWTService, error) {
    if secret == "" {
        logger.Error("jwt service requires non-empty secret")
        return nil, errors.New("invalid jwt secret: empty")
    }
    return &JWTService{secret: secret, expiration: expiration}, nil
}

// CreateToken signs the provided payload using the service secret.
func (s *JWTService) CreateToken(payload *Payload) (string, error) {
    return CreateToken(payload, s.secret)
}

// VerifyToken validates a token string using the service secret.
func (s *JWTService) VerifyToken(tokenStr string) (*Payload, error) {
    return VerifyToken(tokenStr, s.secret)
}

// Expiration returns the configured default expiration.
func (s *JWTService) Expiration() time.Duration { return s.expiration }