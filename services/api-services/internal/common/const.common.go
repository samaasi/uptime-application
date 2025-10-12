package common

const (
	RequestIDContextKey ContextKey = "requestID"
	RequestStartTimeKey ContextKey = "requestStartTime"

	UserIDContextKey               ContextKey = "userID"
	AuthorizationPayloadContextKey ContextKey = "authorizationPayload"

	OTPCacheKeyPrefix                = "otp:"
	OTPTypePasswordReset     OTPType = "password_reset"
	OTPTypeEmailVerification OTPType = "email_verification"
	OTPTypePhoneVerification OTPType = "phone_verification"

	OTPTypePasswordResetPhone OTPType = "password_reset_phone"
	OTPTypePasswordResetEmail OTPType = "password_reset_email"
)
