package common

import (
	"errors"
)

var (
	ErrNotFound        = errors.New("record not found")
	ErrDuplicateEntry  = errors.New("duplicate entry")
	ErrOTPNotFound     = errors.New("OTP not found")
	ErrOTPExpired      = errors.New("OTP expired")
	ErrOTPAlreadyUsed  = errors.New("OTP already used")
	ErrInvalidOTP      = errors.New("invalid OTP code")
	ErrTooManyAttempts = errors.New("too many OTP attempts")

	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrAccountLocked          = errors.New("account locked")
	ErrEmailNotVerified       = errors.New("email not verified")
	ErrPhoneNotVerified       = errors.New("phone not verified")
	ErrEmailAlreadyRegistered = errors.New("email address already registered")
	ErrPhoneAlreadyRegistered = errors.New("phone number already registered")
	ErrUserNotFound           = errors.New("user not found")
	ErrPasswordMismatch       = errors.New("password mismatch")
	ErrOTPAlreadySent         = errors.New("OTP already sent, please wait before retrying")
	ErrOldPasswordMismatch    = errors.New("old password does not match")
	ErrNoIdentifierProvided   = errors.New("email or phone number must be provided")
	ErrInvalidRefreshToken    = errors.New("invalid refresh token")
	ErrTokenMissing           = errors.New("token missing")
	ErrUnauthorized           = errors.New("unauthorized")

	ErrInvalidTokenDuration = errors.New("invalid token duration")
	ErrTokenExpired         = errors.New("token has expired")
	ErrInvalidKey           = errors.New("invalid key provided")
	ErrTokenBlacklisted     = errors.New("token has been blacklisted")
	ErrInvalidToken         = errors.New("invalid token")
	ErrCodeUnauthorized     = "unauthorized"
	ErrSessionBlocked       = errors.New("session is blocked")
	ErrSessionExpired       = errors.New("session has expired")
	ErrSessionNotFound      = errors.New("session not found")
	ErrBadRequest           = errors.New("bad request")
	ErrInternalServer       = errors.New("internal server error")
)
