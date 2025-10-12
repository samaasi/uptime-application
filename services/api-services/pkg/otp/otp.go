package otp

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/samaasi/uptime-application/services/api-services/internal/common"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"
)

// Config for OTP generation/validation
type OTPConfig struct {
	Length                  int
	ExpirationPasswordReset time.Duration
	ExpirationEmailVerify   time.Duration
	ExpirationPhoneVerify   time.Duration
	MaxAttempts             int
}

func DefaultOTPConfig() OTPConfig {
	return OTPConfig{
		Length:                  6,
		ExpirationPasswordReset: 15 * time.Minute,
		ExpirationEmailVerify:   30 * time.Minute,
		ExpirationPhoneVerify:   15 * time.Minute,
		MaxAttempts:             3,
	}
}

// OTPService contains only domain logic: create OTP model and validate rules.
type OTPService struct {
	config OTPConfig
}

func NewOTPService(cfg OTPConfig) *OTPService {
	return &OTPService{config: cfg}
}

// Generate creates a new models.OTP and returns it along with the ttl to be used by the repo.
func (s *OTPService) Generate(identifier string, otpType common.OTPType) (*OTP, time.Duration, error) {
	code, err := generateNumericOTP(s.config.Length)
	if err != nil {
		logger.Error("security: failed to generate OTP code", logger.ErrorField(err))
		return nil, 0, fmt.Errorf("failed to generate OTP code: %w", err)
	}

	var ttl time.Duration
	switch otpType {
	case common.OTPTypePasswordReset:
		ttl = s.config.ExpirationPasswordReset
	case common.OTPTypeEmailVerification:
		ttl = s.config.ExpirationEmailVerify
	case common.OTPTypePhoneVerification:
		ttl = s.config.ExpirationPhoneVerify
	default:
		return nil, 0, fmt.Errorf("unsupported otp type: %s", otpType)
	}

	now := time.Now()
	otp := &OTP{
		Code:       code,
		Identifier: identifier,
		Type:       string(otpType),
		ExpiresAt:  now.Add(ttl),
		Used:       false,
		CreatedAt:  now,
		Attempts:   0,
	}

	return otp, ttl, nil
}

// Validate applies OTP rules to the provided models.OTP and the provided code.
// It mutates the otp (increments Attempts and sets Used when appropriate) and returns an error describing the rule violation (if any).
func (s *OTPService) Validate(otp *OTP, code string) error {
	now := time.Now()
	if now.After(otp.ExpiresAt) {
		return common.ErrOTPExpired
	}
	if otp.Used {
		return common.ErrOTPAlreadyUsed
	}
	if otp.Attempts >= s.config.MaxAttempts {
		return common.ErrTooManyAttempts
	}

	if otp.Code != code {
		otp.Attempts++
		if otp.Attempts >= s.config.MaxAttempts {
			return common.ErrTooManyAttempts
		}
		return common.ErrInvalidOTP
	}

	otp.Attempts++
	otp.Used = true
	return nil
}

func generateNumericOTP(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("OTP length must be positive")
	}
	const digits = "0123456789"
	otp := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random digit: %w", err)
		}
		otp[i] = digits[num.Int64()]
	}
	return string(otp), nil
}
