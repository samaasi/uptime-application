package dtos

import (
    "time"

    "github.com/google/uuid"
)

type SignInRequestDto struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

type SignInResponseDto struct {
    Token     string    `json:"token"`
    UserID    uuid.UUID `json:"user_id"`
    ExpiresAt time.Time `json:"expires_at"`
}

type SignUpRequestDto struct {
    FirstName string `json:"first_name" validate:"required"`
    LastName  string `json:"last_name" validate:"required"`
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=8"`
}

type SignUpResponseDto struct{}

type ForgotPasswordRequest struct {
    Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
    Email       string `json:"email" validate:"required,email"`
    OTP         string `json:"otp" validate:"required"`
    NewPassword string `json:"new_password" validate:"required,min=8"`
}

type VerifyEmailRequest struct {
    Email string `json:"email" validate:"required,email"`
    OTP   string `json:"otp" validate:"required"`
}
