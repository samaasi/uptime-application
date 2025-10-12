package utils

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/samaasi/uptime-application/services/api-services/internal/common"
	"github.com/samaasi/uptime-application/services/api-services/internal/config"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"
)

var (
	appConfig     *config.Config
	configOnce    sync.Once
	configInitErr error
	isDevMode     bool
)

const (
	ErrCodeValidation               = "VALIDATION_ERROR"
	ErrCodeNotFound                 = "NOT_FOUND"
	ErrCodeUnauthorized             = "UNAUTHORIZED"
	ErrCodeForbidden                = "FORBIDDEN"
	ErrCodeBadRequest               = "BAD_REQUEST"
	ErrCodeInternalError            = "INTERNAL_SERVER_ERROR"
	ErrCodeConflict                 = "CONFLICT_ERROR"
	DefaultSuccessMessage           = "Request processed successfully"
	DefaultValidationErrMsg         = "Validation failed: Please check the provided data."
	DefaultTopLevelValidationErrMsg = "Request failed due to validation errors."
)

// ErrorDetails provides a structured format for error information
type ErrorDetails struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    any    `json:"details,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// Meta contains metadata about the API response.
type Meta struct {
	RequestID  string      `json:"request_id"`
	Timestamp  string      `json:"timestamp"`
	Version    string      `json:"version"`
	DurationMs int64       `json:"duration_ms,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	UserID     string      `json:"user_id,omitempty"`
}

// GenericResponse is a standardized API response format with typed data.
type GenericResponse[T any] struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Data    T             `json:"data,omitempty"`
	Error   *ErrorDetails `json:"error,omitempty"`
	Meta    Meta          `json:"meta"`
}

// ResponseBuilder provides a fluent API for constructing and sending API responses.
type ResponseBuilder[T any] struct {
	c          *gin.Context
	appVersion string
	statusCode int
	headers    map[string]string
	data       T
	message    string
	errDetails *ErrorDetails
	pagination *Pagination
	startTime  time.Time
}

// InitResponseUtil initializes the response utilities with configuration.
func InitResponseUtil(cfg *config.Config, devMode bool) error {
	configOnce.Do(func() {
		if cfg == nil {
			configInitErr = errors.New("nil config provided")
			return
		}
		appConfig = cfg
		isDevMode = devMode
	})
	return configInitErr
}

// NewResponse creates a new ResponseBuilder instance.
func NewResponse[T any](c *gin.Context) (*ResponseBuilder[T], error) {
	if appConfig == nil {
		errMsg := "response utilities not initialized. Call InitResponseUtil first"
		logger.Fatal(errMsg)

		return nil, errors.New(errMsg)
	}

	startTime := time.Now()
	if ctxStartTime, ok := c.Request.Context().Value(common.RequestStartTimeKey).(time.Time); ok {
		startTime = ctxStartTime
	}

	return &ResponseBuilder[T]{
		c:          c,
		appVersion: appConfig.App.Version,
		statusCode: http.StatusOK,
		startTime:  startTime,
	}, nil
}

// Status sets the HTTP status code for the response.
func (r *ResponseBuilder[T]) Status(status int) *ResponseBuilder[T] {
	r.statusCode = status
	return r
}

// WithData sets the data payload for the response.
func (r *ResponseBuilder[T]) WithData(data T) *ResponseBuilder[T] {
	r.data = data
	return r
}

// WithMessage sets a descriptive message for the response.
func (r *ResponseBuilder[T]) WithMessage(message string) *ResponseBuilder[T] {
	r.message = message
	return r
}

// WithError sets the error details for the response.
func (r *ResponseBuilder[T]) WithError(code, message string, details ...any) *ResponseBuilder[T] {
	var d any
	if len(details) > 0 {
		d = details[0]
	}

	errDetails := &ErrorDetails{
		Code:    code,
		Message: message,
		Details: d,
	}

	if isDevMode {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		errDetails.StackTrace = string(buf[:n])
	}

	r.errDetails = errDetails
	return r
}

// WithPagination adds pagination metadata to the response.
func (r *ResponseBuilder[T]) WithPagination(p *Pagination) *ResponseBuilder[T] {
	r.pagination = p
	return r
}

// WithHeader adds a custom header to the response.
func (r *ResponseBuilder[T]) WithHeader(key, value string) *ResponseBuilder[T] {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[key] = value
	return r
}

// WithCommonError handles common error types consistently.
func (r *ResponseBuilder[T]) WithCommonError(err common.Error) *ResponseBuilder[T] {
	return r.WithError(err.Code, err.Message, err.Details)
}

// WithCacheControl adds cache control headers.
func (r *ResponseBuilder[T]) WithCacheControl(maxAge time.Duration) *ResponseBuilder[T] {
	return r.WithHeader("Cache-Control", fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds())))
}

// WithSecurityHeaders adds common security headers.
func (r *ResponseBuilder[T]) WithSecurityHeaders() *ResponseBuilder[T] {
	return r.
		WithHeader("X-Content-Type-Options", "nosniff").
		WithHeader("X-Frame-Options", "DENY").
		WithHeader("X-XSS-Protection", "1; mode=block")
}

// Send finalizes and sends the JSON response.
func (r *ResponseBuilder[T]) Send() {
	totalDuration := time.Since(r.startTime)

	for key, value := range r.headers {
		r.c.Writer.Header().Set(key, value)
	}

	if r.errDetails == nil && r.message == "" {
		r.message = DefaultSuccessMessage
	}

	resp := GenericResponse[T]{
		Success: r.errDetails == nil,
		Message: r.message,
		Data:    r.data,
		Error:   r.errDetails,
		Meta: Meta{
			RequestID:  GetRequestID(r.c),
			UserID:     GetUserIDFromContext(r.c),
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			Version:    r.appVersion,
			DurationMs: totalDuration.Milliseconds(),
			Pagination: r.pagination,
		},
	}

	fields := []logger.Field{
		logger.Int("status", r.statusCode),
		logger.String("path", r.c.Request.URL.Path),
		logger.String("method", r.c.Request.Method),
		logger.Duration("duration", totalDuration),
		logger.String("request_id", GetRequestID(r.c)),
		logger.String("client_ip", GetClientIP(r.c)),
	}

	if r.errDetails != nil {
		logger.Error("Request failed",
			append(fields,
				logger.String("error_code", r.errDetails.Code),
				logger.String("error_message", r.errDetails.Message),
				logger.Any("error_details", r.errDetails.Details),
			)...,
		)
	} else {
		logger.Info("Request completed", fields...)
	}

	r.c.JSON(r.statusCode, resp)
}

// --- Common Response Helpers ---
// SendSuccess sends a 200 OK success response.
func SendSuccess[T any](c *gin.Context, data T, message string) {
	builder, err := NewResponse[T](c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response builder"})
		return
	}
	builder.
		Status(http.StatusOK).
		WithData(data).
		WithMessage(message).
		Send()
}

// SendCreated sends a 201 Created response for successful resource creation.
func SendCreated[T any](c *gin.Context, data T, message string) {
	builder, err := NewResponse[T](c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response builder"})
		return
	}
	builder.
		Status(http.StatusCreated).
		WithData(data).
		WithMessage(message).
		Send()
}

// SendAccepted sends a 202 Accepted response, indicating the request has been accepted for processing.
func SendAccepted[T any](c *gin.Context, data T, message string) {
	builder, err := NewResponse[T](c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response builder"})
		return
	}
	builder.
		Status(http.StatusAccepted).
		WithData(data).
		WithMessage(message).
		Send()
}

// SendNoContent sends a 204 No Content response for successful requests with no content to return.
func SendNoContent(c *gin.Context, message string) {
	builder, err := NewResponse[any](c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response builder"})
		return
	}
	builder.
		Status(http.StatusNoContent).
		WithMessage(message).
		Send()
}

// SendBadRequest sends a 400 Bad Request error response.
func SendBadRequest(c *gin.Context, message string, details ...any) {
	builder, err := NewResponse[any](c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response builder"})
		return
	}
	builder.
		Status(http.StatusBadRequest).
		WithError(ErrCodeBadRequest, message, details...).
		Send()
}

// SendValidationError sends a 400 Bad Request error with validation details.
// It now expects a `validator.ValidationErrors` type for `err` and the `targetStruct` for JSON tag extraction.
func SendValidationError(c *gin.Context, err error, targetStruct any) {
	validationErrors := FormatValidationErrors(err, targetStruct)

	builder, buildErr := NewResponse[any](c)
	if buildErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response builder"})
		return
	}
	builder.
		Status(http.StatusBadRequest).
		WithError(ErrCodeValidation, DefaultTopLevelValidationErrMsg, validationErrors).
		Send()
}

// SendNotFound sends a 404 Not Found error response.
func SendNotFound(c *gin.Context, message string, details ...any) {
	builder, err := NewResponse[any](c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response builder"})
		return
	}
	builder.
		Status(http.StatusNotFound).
		WithError(ErrCodeNotFound, message, details...).
		Send()
}

// SendUnauthorized sends a 401 Unauthorized error response.
func SendUnauthorized(c *gin.Context) {
	builder, err := NewResponse[any](c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response builder"})
		return
	}
	builder.
		Status(http.StatusUnauthorized).
		WithError(ErrCodeUnauthorized, "Authentication required.").
		Send()
}

// SendUnauthorizedWithDetail sends a 401 Unauthorized error with specific details.
func SendUnauthorizedWithDetail(c *gin.Context, code, message string, details ...any) {
	builder, err := NewResponse[any](c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response builder"})
		return
	}
	builder.
		Status(http.StatusUnauthorized).
		WithError(code, message, details...).
		Send()
}

// SendForbidden sends a 403 Forbidden error response.
func SendForbidden(c *gin.Context, message string, details ...any) {
	builder, err := NewResponse[any](c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response builder"})
		return
	}
	builder.
		Status(http.StatusForbidden).
		WithError(ErrCodeForbidden, message, details...).
		Send()
}

// SendConflict sends a 409-Conflict error response.
func SendConflict(c *gin.Context, message string, details ...any) {
	builder, err := NewResponse[any](c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response builder"})
		return
	}
	builder.
		Status(http.StatusConflict).
		WithError(ErrCodeConflict, message, details...).
		Send()
}

// SendError sends a structured error JSON response with the specified HTTP status, code, message, and details.
func SendError(c *gin.Context, httpStatus int, code string, message string, details ...any) {
	builder, err := NewResponse[any](c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response builder"})
		return
	}
	builder.
		Status(httpStatus).
		WithError(code, message, details...).
		Send()
}

// SendInternalServerError sends a 500 Internal Server Error response.
func SendInternalServerError(c *gin.Context, details ...any) {
	message := "An unexpected error occurred."
	code := ErrCodeInternalError
	if isDevMode && len(details) > 0 {
		message = fmt.Sprintf("Internal Server Error: %v", details[0])
	}

	builder, err := NewResponse[any](c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response builder"})
		return
	}
	builder.
		Status(http.StatusInternalServerError).
		WithError(code, message, details...).
		Send()
}

// --- Validation Helpers ---
// FormatValidationErrors processes a validator.ValidationErrors into a map for API response.
// It requires the targetStruct to correctly extract JSON tags.
func FormatValidationErrors(err error, targetStruct any) map[string]string {
	formattedErrors := make(map[string]string)

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		structType := reflect.TypeOf(targetStruct)
		if structType.Kind() == reflect.Ptr {
			structType = structType.Elem()
		}

		for _, e := range validationErrors {
			jsonTag := extractJSONTag(e, structType)
			formattedErrors[jsonTag] = formatErrorMessage(e)
		}
	} else {
		formattedErrors["general"] = err.Error()
	}

	return formattedErrors
}

// formatErrorMessage generates a user-friendly error message for a validation field error.
func formatErrorMessage(e validator.FieldError) string {
	field := e.Field()
	param := e.Param()

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("The %s field is required.", field)
	case "email":
		return fmt.Sprintf("The %s field must be a valid email address.", field)
	case "url":
		return fmt.Sprintf("The %s field must be a valid URL.", field)
	case "uuid":
		return fmt.Sprintf("The %s field must be a valid UUID.", field)
	case "ip":
		return fmt.Sprintf("The %s field must be a valid IP address.", field)
	case "ipv4":
		return fmt.Sprintf("The %s field must be a valid IPv4 address.", field)
	case "ipv6":
		return fmt.Sprintf("The %s field must be a valid IPv6 address.", field)
	case "len":
		return fmt.Sprintf("The %s field must be exactly %s characters long.", field, param)
	case "min":
		return fmt.Sprintf("The %s field must be at least %s characters long.", field, param)
	case "max":
		return fmt.Sprintf("The %s field must not exceed %s characters.", field, param)
	case "eq":
		return fmt.Sprintf("The %s field must be equal to %s.", field, param)
	case "ne":
		return fmt.Sprintf("The %s field must not be equal to %s.", field, param)
	case "lt":
		return fmt.Sprintf("The %s field must be less than %s.", field, param)
	case "lte":
		return fmt.Sprintf("The %s field must be less than or equal to %s.", field, param)
	case "gt":
		return fmt.Sprintf("The %s field must be greater than %s.", field, param)
	case "gte":
		return fmt.Sprintf("The %s field must be greater than or equal to %s.", field, param)
	case "alphanum":
		return fmt.Sprintf("The %s field must be alphanumeric.", field)
	case "contains":
		return fmt.Sprintf("The %s field must contain '%s'.", field, param)
	case "startswith":
		return fmt.Sprintf("The %s field must start with '%s'.", field, param)
	case "endswith":
		return fmt.Sprintf("The %s field must end with '%s'.", field, param)
	case "oneof":
		return fmt.Sprintf("The %s field must be one of [%s].", field, param)
	case "datetime":
		return fmt.Sprintf("The %s field must be a valid datetime in format %s.", field, param)
	case "phone_number":
		return fmt.Sprintf("The %s field must be a valid phone number.", field)
	default:
		return fmt.Sprintf("The %s field is invalid.", field)
	}
}

// extractJSONTag extracts the JSON tag name for a given field error.
// It tries to get the struct type from the context if not provided.
func extractJSONTag(fe validator.FieldError, structType reflect.Type) string {
	if structType != nil {
		if structType.Kind() == reflect.Ptr {
			structType = structType.Elem()
		}

		fieldName := fe.StructField()
		if field, found := structType.FieldByName(fieldName); found {
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" && jsonTag != "-" {
				return strings.Split(jsonTag, ",")[0]
			}
		}
	}
	return fe.Field()
}
