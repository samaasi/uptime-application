package common

type ContextKey string

// OTPType defines the purpose of the OTP.
type OTPType string

type Error struct {
	Code    string
	Message string
	Details interface{}
}
