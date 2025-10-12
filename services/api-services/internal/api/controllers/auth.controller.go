package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samaasi/uptime-application/services/api-services/internal/api/dtos"
	"github.com/samaasi/uptime-application/services/api-services/internal/api/services"
	"github.com/samaasi/uptime-application/services/api-services/internal/common"
	"github.com/samaasi/uptime-application/services/api-services/internal/utils"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"
)

// AuthController handles authentication-related HTTP requests
type AuthController struct {
	authService *services.AuthService
}

// NewAuthController creates a new auth controller instance
func NewAuthController(
	authService *services.AuthService,
) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// SignUp handles POST /auth/signup - Register a new user with OTP verification
func (ac *AuthController) SignUp(c *gin.Context) {
	var req dtos.SignUpRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request payload", logger.ErrorField(err))
		utils.SendError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	response, err := ac.authService.SignUpByEmail(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case common.ErrEmailAlreadyRegistered:
			utils.SendConflict(c, "Email already registered")
		case common.ErrInvalidOTP:
			utils.SendBadRequest(c, "Invalid or expired OTP")
		default:
			logger.Error("Failed to sign up user", logger.ErrorField(err))
			utils.SendError(c, http.StatusInternalServerError, "SIGNUP_FAILED", "Failed to sign up user")
		}
		return
	}

	logger.Info("User signed up successfully")
	utils.SendCreated(c, response, "User signed up successfully")
}

// SignIn handles POST /auth/signin - Login user with OTP verification
func (ac *AuthController) SignIn(c *gin.Context) {
	var req dtos.SignInRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request payload", logger.ErrorField(err))
		utils.SendError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	response, err := ac.authService.SignIn(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case common.ErrInvalidCredentials:
			utils.SendUnauthorizedWithDetail(c, "INVALID_CREDENTIALS", "Invalid credentials")
		case common.ErrInvalidOTP:
			utils.SendBadRequest(c, "Invalid or expired OTP")
		case common.ErrEmailNotVerified:
			utils.SendUnauthorizedWithDetail(c, "EMAIL_NOT_VERIFIED", "Email not verified")
		default:
			logger.Error("Failed to sign in user", logger.ErrorField(err))
			utils.SendError(c, http.StatusInternalServerError, "SIGNIN_FAILED", "Failed to sign in user")
		}
		return
	}

	logger.Info("User signed in successfully")
	utils.SendSuccess(c, response, "User signed in successfully")
}
