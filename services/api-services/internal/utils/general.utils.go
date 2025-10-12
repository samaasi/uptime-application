package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"unicode"

	"github.com/samaasi/uptime-application/services/api-services/internal/common"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"
	"github.com/samaasi/uptime-application/services/api-services/pkg/security"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9_ -]+`)

// GetRequestID extracts the request ID from the Gin context.
func GetRequestID(c *gin.Context) string {
	if requestID, ok := c.Get(string(common.RequestIDContextKey)); ok {
		return requestID.(string)
	}

	newID := uuid.New().String()
	logger.Debug("Generated new request ID", logger.String("requestID", newID))
	return newID
}

// GetUserIDFromContext extracts the user ID from the Gin context.
func GetUserIDFromContext(c *gin.Context) string {
	if userID, ok := c.Get(string(common.UserIDContextKey)); ok {
		return userID.(string)
	}
	logger.Debug("User ID not found in context")
	return ""
}

// GetAuthUser retrieves the authenticated userId from the Gin context.
func GetAuthUser(c *gin.Context) (userID uuid.UUID, err error) {
	payload, exists := c.Get(string(common.AuthorizationPayloadContextKey))
	if !exists {
		err = errors.New("no payload found in context")
		SendError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "Unauthorized: No payload found in context", err)
		return uuid.Nil, err
	}

	authPayload, ok := payload.(*security.Payload)
	if !ok {
		err = errors.New("failed to cast payload")
		SendError(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to cast payload", err)
		return uuid.Nil, err
	}

	return authPayload.UserID, nil
}

// GetClientIP extracts the client's IP address from the Gin context.
func GetClientIP(c *gin.Context) string {
	return c.ClientIP()
}

func GetUserAgent(c *gin.Context) string {
	return c.Request.Header.Get("User-Agent")
}

func IsDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}

	// Common PostgreSQL unique constraint violation error message part
	if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return true
	}

	if strings.Contains(err.Error(), "Duplicate entry") {
		return true
	}

	return false
}

// StringPtr returns a pointer to the given string.
// Returns nil if the string is empty.
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Must is a helper to simplify error checking for functions that return (T, error).
// It panics if an error occurs, useful for compile-time constants or early initialization.
// This function should be used sparingly and only for truly unrecoverable errors
// e.g: during application startup or configuration loading.
func Must[T any](x T, err error) T {
	if err != nil {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		stackTrace := string(buf[:n])

		logger.Fatal("Fatal error during initialization/must operation",
			logger.ErrorField(err),
			logger.String("stack", stackTrace),
			logger.String("function", GetFunctionName(Must[T])),
		)
	}
	return x
}

// CheckError is a simple helper to log errors that are not critical enough to panic.
// It uses the custom logger for consistent error reporting.
func CheckError(err error) {
	if err != nil {
		if runtime.GOOS == "windows" && strings.Contains(err.Error(), "sync /dev/stdout: The handle is invalid") {
			return
		}

		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		stackTrace := string(buf[:n])

		logger.Error("Unexpected error",
			logger.ErrorField(err),
			logger.String("stack", stackTrace),
		)
	}
}

// CheckErrorWithMsg is a simple helper to log errors with a custom message.
// It uses the custom logger for consistent error reporting with additional context.
func CheckErrorWithMsg(err error, msg string) {
	if err != nil {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		stackTrace := string(buf[:n])

		logger.Error(msg,
			logger.ErrorField(err),
			logger.String("stack", stackTrace),
		)
	}
}

// GetStringValue safely dereferences a string pointer, returning an empty string if nil.
func GetStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// GetFunctionName returns the name of the function passed to it.
// It handles anonymous functions and methods correctly.
func GetFunctionName(i interface{}) string {
	if i == nil {
		return "nil"
	}

	name := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()

	if strings.Contains(name, ".func") {
		parts := strings.Split(name, ".")
		return parts[len(parts)-1]
	}

	return name
}

// ToPointer converts a value to a pointer.
// This is a generic helper that can be placed here or in a more general utils package.
func ToPointer[T any](v T) *T {
	return &v
}

// IsNumeric checks if a string consists only of digits.
func IsNumeric(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// IsEmail Check if it is an email
func IsEmail(identifier string) bool {
	return strings.Contains(identifier, "@") && strings.Contains(identifier, ".")
}

func GenerateSlug(name string) (string, error) {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	slug = slugRegex.ReplaceAllString(slug, "")
	slug = strings.Trim(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")

	maxRandomNum := big.NewInt(900000)
	minRandomNum := big.NewInt(100000)

	randomInt, err := rand.Int(rand.Reader, maxRandomNum)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number for slug: %w", err)
	}
	randomNumber := randomInt.Int64() + minRandomNum.Int64()

	finalSlug := fmt.Sprintf("%s-%d", slug, randomNumber)

	return finalSlug, nil
}

// GenerateRandomString generates a random string of the specified length using crypto/rand.
func GenerateRandomString(length int) (string, error) {
	if length%2 != 0 {
		return "", fmt.Errorf("length must be even")
	}
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetContentType returns the MIME type for a given file extension.
func GetContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".mp4":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}
