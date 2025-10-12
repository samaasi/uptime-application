package services

import (
	"context"
	"fmt"

	"github.com/samaasi/uptime-application/services/api-services/internal/common"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"
	"github.com/samaasi/uptime-application/services/api-services/pkg/otp"
)

// UserOTPManagerService orchestrates application flow: uses security (domain rules) + repo + logging.
type UserOTPManagerService struct {
	repo   otp.Repository
	secSvc *otp.OTPService
}

func NewUserOTPManagerService(repo otp.Repository, secSvc *otp.OTPService) *UserOTPManagerService {
	return &UserOTPManagerService{
		repo:   repo,
		secSvc: secSvc,
	}
}

// GenerateAndSaveOTP: generate domain OTP via security service and persist via repo.
func (s *UserOTPManagerService) GenerateAndSaveOTP(ctx context.Context, otpType common.OTPType, identifier string) (string, error) {
	otpObj, ttl, err := s.secSvc.Generate(identifier, otpType)
	if err != nil {
		logger.Error("service: failed to generate OTP",
			logger.String("identifier", identifier),
			logger.String("otp_type", string(otpType)),
			logger.ErrorField(err))
		return "", fmt.Errorf("failed to generate otp: %w", err)
	}

	if err := s.repo.SaveOTP(ctx, otpObj, ttl); err != nil {
		logger.Error("service: failed to save OTP",
			logger.String("identifier", identifier),
			logger.String("otp_type", string(otpType)),
			logger.ErrorField(err))
		return "", fmt.Errorf("failed to persist otp: %w", err)
	}

	logger.Info("service: otp generated and persisted",
		logger.String("identifier", identifier),
		logger.String("otp_type", string(otpType)),
	)

	return otpObj.Code, nil
}

// VerifyOTP: orchestrates retrieval, calls secSvc.Validate (domain rules), persists changes & cleans up as required.
func (s *UserOTPManagerService) VerifyOTP(ctx context.Context, otpType common.OTPType, identifier string, code string) (bool, error) {
	storedOTP, err := s.repo.GetOTP(ctx, string(otpType), identifier)
	if err != nil {
		logger.Warn("service: otp not found or repo error",
			logger.String("identifier", identifier),
			logger.String("otp_type", string(otpType)),
			logger.ErrorField(err))
		return false, common.ErrOTPNotFound
	}

	// Validate (mutates storedOTP: attempts/used)
	if err := s.secSvc.Validate(storedOTP, code); err != nil {
		switch err {
		case common.ErrInvalidOTP:
			if updateErr := s.repo.UpdateOTP(ctx, storedOTP); updateErr != nil {
				logger.Error("service: failed to update OTP attempts",
					logger.String("identifier", identifier),
					logger.String("otp_type", string(otpType)),
					logger.ErrorField(updateErr))
			}
			logger.Warn("service: invalid otp provided",
				logger.String("identifier", identifier),
				logger.String("otp_type", string(otpType)))
			return false, common.ErrInvalidOTP
		case common.ErrTooManyAttempts:
			// attempts reached: remove OTP
			_ = s.repo.DeleteOTP(ctx, string(otpType), identifier)
			logger.Warn("service: too many attempts - otp deleted",
				logger.String("identifier", identifier),
				logger.String("otp_type", string(otpType)))
			return false, common.ErrTooManyAttempts
		case common.ErrOTPExpired:
			_ = s.repo.DeleteOTP(ctx, string(otpType), identifier)
			logger.Warn("service: otp expired and removed",
				logger.String("identifier", identifier),
				logger.String("otp_type", string(otpType)))
			return false, common.ErrOTPExpired
		case common.ErrOTPAlreadyUsed:
			return false, common.ErrOTPAlreadyUsed
		default:
			logger.Error("service: validation returned unexpected error",
				logger.String("identifier", identifier),
				logger.String("otp_type", string(otpType)),
				logger.ErrorField(err))
			return false, fmt.Errorf("validation error: %w", err)
		}
	}

	// success: OTP was marked Used inside Validate; persist or delete as desired.
	if dErr := s.repo.DeleteOTP(ctx, string(otpType), identifier); dErr != nil {
		logger.Error("service: failed to delete OTP after successful verification",
			logger.String("identifier", identifier),
			logger.String("otp_type", string(otpType)),
			logger.ErrorField(dErr),
		)
	}

	logger.Info("service: otp verified successfully",
		logger.String("identifier", identifier),
		logger.String("otp_type", string(otpType)),
	)
	return true, nil
}
