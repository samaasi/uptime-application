package services

import (
	"context"
	"fmt"

	"github.com/samaasi/uptime-application/services/api-services/internal/api/models"
	"github.com/samaasi/uptime-application/services/api-services/internal/api/repositories"
	"github.com/samaasi/uptime-application/services/api-services/internal/common"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"
	"github.com/samaasi/uptime-application/services/api-services/pkg/notifier/email"
	"github.com/samaasi/uptime-application/services/api-services/pkg/otp"
	"gorm.io/gorm"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepository repositories.UserRepository
	otpService     *otp.OTPService
	emailService   email.Service
}

func NewAuthService(
	userRepository repositories.UserRepository,
	otpService *otp.OTPService,
	emailService email.Service,
) *AuthService {
	return &AuthService{
		userRepository: userRepository,
		otpService:     otpService,
		emailService:   emailService,
	}
}

// SignUpByEmail handles user registration with email verification
func (s *AuthService) SignUpByEmail(ctx context.Context, req *dtos.SignUpRequest) (*models.User, error) {
	existingUser, err := s.userRepository.GetByEmail(ctx, req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Error("Failed to check existing user", logger.String("email", req.Email), logger.ErrorField(err))
		return nil, common.ErrInternalServer
	}

	if existingUser != nil {
		return nil, common.ErrEmailAlreadyRegistered
	}

	// Create new user
	user := &models.User{
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		HashedPassword: req.Password,
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		logger.Error("Failed to create user", logger.String("email", req.Email), logger.ErrorField(err))
		return nil, common.ErrInternalServer
	}

	// Generate OTP for email verification
	otpToken, err := s.otpService.GenerateAndSaveOTP(ctx, common.OTPTypeEmailVerification, user.ID)
	if err != nil {
		logger.Error("Failed to generate OTP", logger.String("email", req.Email), logger.ErrorField(err))
		return nil, common.ErrInternalServer
	}

	// Send verification email
	if err := s.emailService.SendEmail(ctx, req.Email, "Email Verification OTP", fmt.Sprintf("Your OTP for email verification is: %s", otpToken)); err != nil {
		logger.Error("Failed to send verification email", logger.String("email", req.Email), logger.ErrorField(err))
	}

	logger.Info("User registered successfully", logger.String("user_id", user.ID.String()), logger.String("email", user.Email))
	return user, nil
}
