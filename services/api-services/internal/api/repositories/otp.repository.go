package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/samaasi/uptime-application/services/api-services/pkg/cache"
	"github.com/samaasi/uptime-application/services/api-services/pkg/otp"
)

// otpRepository implements otp.Repository using cache (same idea as before).
type otpRepository struct {
	cacheService *cache.Service
}

func NewOTPRepository(cacheService *cache.Service) otp.Repository {
	return &otpRepository{
		cacheService: cacheService,
	}
}

func (r *otpRepository) GenerateOTPKey(otpType string, identifier string) string {
	return fmt.Sprintf("otp:%s:%s", otpType, identifier)
}

func (r *otpRepository) SaveOTP(ctx context.Context, otp *otp.OTP, ttl time.Duration) error {
	key := r.GenerateOTPKey(otp.Type, otp.Identifier)

	otpData, err := json.Marshal(otp)
	if err != nil {
		return fmt.Errorf("failed to marshal OTP: %w", err)
	}

	return r.cacheService.Set(ctx, key, otpData, ttl)
}

func (r *otpRepository) GetOTP(ctx context.Context, otpType string, identifier string) (*otp.OTP, error) {
	key := r.GenerateOTPKey(otpType, identifier)

	var otpData []byte
	if err := r.cacheService.Get(ctx, key, &otpData); err != nil {
		return nil, err
	}

	var o otp.OTP
	if err := json.Unmarshal(otpData, &o); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OTP: %w", err)
	}
	return &o, nil
}

func (r *otpRepository) UpdateOTP(ctx context.Context, otp *otp.OTP) error {
	key := r.GenerateOTPKey(otp.Type, otp.Identifier)

	otpData, err := json.Marshal(otp)
	if err != nil {
		return fmt.Errorf("failed to marshal OTP: %w", err)
	}

	ttl := time.Until(otp.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("otp expired")
	}

	return r.cacheService.Set(ctx, key, otpData, ttl)
}

func (r *otpRepository) DeleteOTP(ctx context.Context, otpType string, identifier string) error {
	key := r.GenerateOTPKey(otpType, identifier)
	return r.cacheService.Delete(ctx, key)
}
