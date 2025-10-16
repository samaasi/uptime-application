package services

import (
	"context"
	"fmt"
	"time"

	"github.com/samaasi/uptime-application/services/api-services/internal/api/models"
	"github.com/samaasi/uptime-application/services/api-services/internal/api/repositories"
	"github.com/samaasi/uptime-application/services/api-services/internal/common"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"
	"github.com/samaasi/uptime-application/services/api-services/pkg/notifier/email"
	"github.com/samaasi/uptime-application/services/api-services/pkg/security"

	"github.com/samaasi/uptime-application/services/api-services/internal/api/dtos"
	"gorm.io/gorm"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepository repositories.UserRepository
	otpService     *UserOTPManagerService
	emailService   email.Service
	jwtService     *security.JWTService
}

func NewAuthService(
	userRepository repositories.UserRepository,
	otpService *UserOTPManagerService,
	emailService email.Service,
	jwtService *security.JWTService,
) *AuthService {
	return &AuthService{
		userRepository: userRepository,
		otpService:     otpService,
		emailService:   emailService,
		jwtService:     jwtService,
	}
}

// SignUpByEmail handles user registration with email verification
func (s *AuthService) SignUpByEmail(ctx context.Context, req *dtos.SignUpRequestDto) (*models.User, error) {
	existingUser, err := s.userRepository.GetByEmail(ctx, req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Error("Failed to check existing user", logger.String("email", req.Email), logger.ErrorField(err))
		return nil, common.ErrInternalServer
	}

	if existingUser != nil {
		return nil, common.ErrEmailAlreadyRegistered
	}

	user := &models.User{
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          &req.Email,
		HashedPassword: req.Password,
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		logger.Error("Failed to create user", logger.String("email", req.Email), logger.ErrorField(err))
		return nil, common.ErrInternalServer
	}

	// Generate OTP for email verification
	otpToken, err := s.otpService.GenerateAndSaveOTP(ctx, common.OTPTypeEmailVerification, req.Email)
	if err != nil {
		logger.Error("Failed to generate OTP", logger.String("email", req.Email), logger.ErrorField(err))
		return nil, common.ErrInternalServer
	}

	// Send verification email
	if err := s.emailService.SendEmail(ctx, req.Email, "Email Verification OTP", fmt.Sprintf("Your OTP for email verification is: %s", otpToken)); err != nil {
		logger.Error("Failed to send verification email", logger.String("email", req.Email), logger.ErrorField(err))
	}

	logger.Info("User registered successfully", logger.String("user_id", user.ID.String()), logger.String("email", req.Email))
	return user, nil
}

// SignIn handles user authentication
func (s *AuthService) SignIn(ctx context.Context, req *dtos.SignInRequestDto) (*dtos.SignInResponseDto, error) {
	user, err := s.userRepository.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrInvalidCredentials
		}
		logger.Error("Failed to get user", logger.String("email", req.Email), logger.ErrorField(err))
		return nil, common.ErrInternalServer
	}

	if !security.VerifyPassword(user.HashedPassword, req.Password) {
		return nil, common.ErrInvalidCredentials
	}

	if !user.EmailVerified() {
		return nil, common.ErrEmailNotVerified
	}

	// Generate JWT access token
	payload := security.NewPayload(user.ID, time.Hour*24)

	accessToken, err := s.jwtService.CreateToken(payload)
	if err != nil {
		logger.Error("Failed to sign JWT token", logger.String("user_id", user.ID.String()), logger.ErrorField(err))
		return nil, common.ErrInternalServer
	}

	response := &dtos.SignInResponseDto{
		Token:     accessToken,
		UserID:    user.ID,
		ExpiresAt: payload.ExpiresAt.Time,
	}

	// Safe email logging
	emailVal := ""
	if user.Email != nil {
		emailVal = *user.Email
	}
	logger.Info("User signed in successfully", logger.String("user_id", user.ID.String()), logger.String("email", emailVal))
	return response, nil
}

// ForgotPassword initiates password reset process
func (s *AuthService) ForgotPassword(ctx context.Context, req *dtos.ForgotPasswordRequest) error {
	// Check if user exists
	_, err := s.userRepository.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Don't reveal if user exists or not
			return nil
		}
		logger.Error("Failed to get user", logger.String("email", req.Email), logger.ErrorField(err))
		return common.ErrInternalServer
	}

	// Generate OTP for password reset
	otp, err := s.otpService.GenerateAndSaveOTP(ctx, common.OTPTypePasswordReset, req.Email)
	if err != nil {
		logger.Error("Failed to generate OTP", logger.String("email", req.Email), logger.ErrorField(err))
		return common.ErrInternalServer
	}

	// Send password reset email
	if err := s.emailService.SendEmail(ctx, req.Email, "Password Reset OTP", fmt.Sprintf("Your OTP for password reset is: %s", otp)); err != nil {
		logger.Error("Failed to send password reset email", logger.String("email", req.Email), logger.ErrorField(err))
		return common.ErrInternalServer
	}

	logger.Info("Password reset initiated", logger.String("email", req.Email))
	return nil
}

// ResetPassword completes password reset process
func (s *AuthService) ResetPassword(ctx context.Context, req *dtos.ResetPasswordRequest) error {
	// Verify OTP
	verified, err := s.otpService.VerifyOTP(ctx, common.OTPTypePasswordReset, req.Email, req.OTP)
	if err != nil || !verified {
		logger.Error("Invalid OTP for password reset", logger.String("email", req.Email), logger.ErrorField(err))
		return common.ErrInvalidOTP
	}

	// Get user
	user, err := s.userRepository.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.ErrUserNotFound
		}
		logger.Error("Failed to get user", logger.String("email", req.Email), logger.ErrorField(err))
		return common.ErrInternalServer
	}

	// Hash new password
	hashedPassword, err := security.HashPassword(req.NewPassword, nil)
	if err != nil {
		logger.Error("Failed to hash password", logger.String("email", req.Email), logger.ErrorField(err))
		return common.ErrInternalServer
	}

	// Update user password
	user.HashedPassword = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.userRepository.Update(ctx, user); err != nil {
		logger.Error("Failed to update user password", logger.String("email", req.Email), logger.ErrorField(err))
		return common.ErrInternalServer
	}

	// OTP is automatically deleted by the VerifyOTP method
	// No need to manually delete it

	logger.Info("Password reset successfully", logger.String("email", req.Email))
	return nil
}

// VerifyEmail handles email verification
func (s *AuthService) VerifyEmail(ctx context.Context, req *dtos.VerifyEmailRequest) error {
	// Verify OTP
	verified, err := s.otpService.VerifyOTP(ctx, common.OTPTypeEmailVerification, req.Email, req.OTP)
	if err != nil || !verified {
		logger.Error("Invalid OTP for email verification", logger.String("email", req.Email), logger.ErrorField(err))
		return common.ErrInvalidOTP
	}

	// Get user
	user, err := s.userRepository.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.ErrUserNotFound
		}
		logger.Error("Failed to get user", logger.String("email", req.Email), logger.ErrorField(err))
		return common.ErrInternalServer
	}

	// Update email verification status
	now := time.Now()
	user.EmailVerifiedAt = &now
	user.UpdatedAt = now

	if err := s.userRepository.Update(ctx, user); err != nil {
		logger.Error("Failed to update user email verification", logger.String("email", req.Email), logger.ErrorField(err))
		return common.ErrInternalServer
	}

	// OTP is automatically deleted by the VerifyOTP method
	// No need to manually delete it

	logger.Info("Email verified successfully", logger.String("email", req.Email))
	return nil
}

// ResendOTP resends OTP for various operations
func (s *AuthService) ResendOTP(ctx context.Context, otpType common.OTPType, email string) error {
	// Check if user exists
	_, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.ErrUserNotFound
		}
		logger.Error("Failed to get user", logger.String("email", email), logger.ErrorField(err))
		return common.ErrInternalServer
	}

	// Generate new OTP
	otp, err := s.otpService.GenerateAndSaveOTP(ctx, otpType, email)
	if err != nil {
		logger.Error("Failed to generate OTP", logger.String("email", email), logger.String("type", string(otpType)), logger.ErrorField(err))
		return common.ErrInternalServer
	}

	// Determine email subject and message based on OTP type
	var subject, message string
	switch otpType {
	case common.OTPTypeEmailVerification:
		subject = "Email Verification"
		message = fmt.Sprintf("Your verification code is: %s", otp)
	case common.OTPTypePasswordReset:
		subject = "Password Reset"
		message = fmt.Sprintf("Your password reset code is: %s", otp)
	default:
		subject = "OTP Code"
		message = fmt.Sprintf("Your OTP code is: %s", otp)
	}

	// Send email
	if err := s.emailService.SendEmail(ctx, email, subject, message); err != nil {
		logger.Error("Failed to send OTP email", logger.String("email", email), logger.String("type", string(otpType)), logger.ErrorField(err))
		return common.ErrInternalServer
	}

	logger.Info("OTP resent successfully", logger.String("email", email), logger.String("type", string(otpType)))
	return nil
}
