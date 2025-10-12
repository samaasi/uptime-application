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
