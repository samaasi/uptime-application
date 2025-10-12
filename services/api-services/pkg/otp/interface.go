package otp

import (
	"context"
	"time"
)

type OTP struct {
	Code       string    `json:"code"`
	Identifier string    `json:"identifier"`
	Type       string    `json:"type"`
	ExpiresAt  time.Time `json:"expires_at"`
	Used       bool      `json:"used"`
	CreatedAt  time.Time `json:"created_at"`
	Attempts   int       `json:"attempts"`
}

// Repository defines persistence operations for OTPs.
type Repository interface {
	SaveOTP(ctx context.Context, otp *OTP, ttl time.Duration) error
	GetOTP(ctx context.Context, otpType string, identifier string) (*OTP, error)
	UpdateOTP(ctx context.Context, otp *OTP) error
	DeleteOTP(ctx context.Context, otpType string, identifier string) error
	GenerateOTPKey(otpType string, identifier string) string
}
