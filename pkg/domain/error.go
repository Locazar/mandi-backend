package domain

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorCode represents standard error codes across the application
type ErrorCode string

const (
	// Validation errors (4000-4099)
	ErrCodeValidation       ErrorCode = "VALIDATION_ERROR"
	ErrCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrCodeMissingField     ErrorCode = "MISSING_FIELD"
	ErrCodeInvalidFormat    ErrorCode = "INVALID_FORMAT"

	// Authentication errors (4010-4099)
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeTokenExpired     ErrorCode = "TOKEN_EXPIRED"
	ErrCodeInvalidToken     ErrorCode = "INVALID_TOKEN"

	// Resource errors (4040-4099)
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists    ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict         ErrorCode = "CONFLICT"

	// Business logic errors (4220-4299)
	ErrCodeBusinessLogic    ErrorCode = "BUSINESS_LOGIC_ERROR"
	ErrCodeImageMismatch    ErrorCode = "IMAGE_MISMATCH"
	ErrCodeInsufficientStock ErrorCode = "INSUFFICIENT_STOCK"
	ErrCodePaymentFailed    ErrorCode = "PAYMENT_FAILED"
	ErrCodeInvalidOperation ErrorCode = "INVALID_OPERATION"

	// Server errors (5000-5099)
	ErrCodeInternal         ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeDatabase         ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalService  ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrCodeTimeout          ErrorCode = "TIMEOUT"
	ErrCodeFileOperation    ErrorCode = "FILE_OPERATION_ERROR"
)

// AppError represents a standardized error throughout the application
type AppError struct {
	Code       ErrorCode   `json:"code"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
	Details    string      `json:"details,omitempty"`
	Err        error       `json:"-"` // Internal error for logging
	Timestamp  int64       `json:"timestamp"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAppError creates a new AppError with status code mapping
func NewAppError(code ErrorCode, message, details string, err error) *AppError {
	appErr := &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		Err:        err,
		Timestamp:  time.Now().Unix(),
	}
	appErr.StatusCode = mapErrorCodeToStatus(code)
	return appErr
}

// mapErrorCodeToStatus maps error codes to HTTP status codes
func mapErrorCodeToStatus(code ErrorCode) int {
	switch code {
	case ErrCodeValidation, ErrCodeInvalidInput, ErrCodeMissingField, ErrCodeInvalidFormat:
		return http.StatusBadRequest
	case ErrCodeUnauthorized, ErrCodeTokenExpired, ErrCodeInvalidToken:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeAlreadyExists, ErrCodeConflict, ErrCodeImageMismatch, ErrCodeInvalidOperation:
		return http.StatusConflict
	case ErrCodeBusinessLogic, ErrCodeInsufficientStock, ErrCodePaymentFailed:
		return http.StatusUnprocessableEntity
	case ErrCodeInternal, ErrCodeDatabase, ErrCodeExternalService, ErrCodeTimeout, ErrCodeFileOperation:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Predefined error constructors for common scenarios

// ValidationError creates a validation error
func ValidationError(field, message string) *AppError {
	return NewAppError(ErrCodeValidation, "Validation failed", fmt.Sprintf("Field: %s, Reason: %s", field, message), nil)
}

// NotFoundError creates a not found error
func NotFoundError(resource string) *AppError {
	return NewAppError(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), "", nil)
}

// AlreadyExistsError creates an already exists error
func AlreadyExistsError(resource string) *AppError {
	return NewAppError(ErrCodeAlreadyExists, fmt.Sprintf("%s already exists", resource), "", nil)
}

// UnauthorizedError creates an unauthorized error
func UnauthorizedError(message string) *AppError {
	return NewAppError(ErrCodeUnauthorized, message, "", nil)
}

// ForbiddenError creates a forbidden error
func ForbiddenError(message string) *AppError {
	return NewAppError(ErrCodeForbidden, message, "", nil)
}

// InternalError creates an internal server error with wrapped error
func InternalError(message string, err error) *AppError {
	return NewAppError(ErrCodeInternal, message, "", err)
}

// DatabaseError creates a database error with wrapped error
func DatabaseError(operation string, err error) *AppError {
	return NewAppError(ErrCodeDatabase, fmt.Sprintf("Database operation failed: %s", operation), "", err)
}

// ExternalServiceError creates an external service error
func ExternalServiceError(service string, message string, err error) *AppError {
	return NewAppError(ErrCodeExternalService, fmt.Sprintf("External service failed: %s", service), message, err)
}

// ImageMismatchError creates an image mismatch error
func ImageMismatchError(reason string) *AppError {
	return NewAppError(ErrCodeImageMismatch, "Image does not match the product category", reason, nil)
}

// BusinessLogicError creates a business logic error
func BusinessLogicError(message string, details string) *AppError {
	return NewAppError(ErrCodeBusinessLogic, message, details, nil)
}

// TimeoutError creates a timeout error
func TimeoutError(operation string) *AppError {
	return NewAppError(ErrCodeTimeout, fmt.Sprintf("Operation timed out: %s", operation), "", nil)
}

// ConflictError creates a conflict error
func ConflictError(message string, details string) *AppError {
	return NewAppError(ErrCodeConflict, message, details, nil)
}

// FileOperationError creates a file operation error
func FileOperationError(operation string, err error) *AppError {
	return NewAppError(ErrCodeFileOperation, fmt.Sprintf("File operation failed: %s", operation), "", err)
}

// InvalidOperationError creates an invalid operation error
func InvalidOperationError(message string, details string) *AppError {
	return NewAppError(ErrCodeInvalidOperation, message, details, nil)
}

// InsufficientStockError creates an insufficient stock error
func InsufficientStockError(productName string, requested, available int) *AppError {
	details := fmt.Sprintf("Product: %s, Requested: %d, Available: %d", productName, requested, available)
	return NewAppError(ErrCodeInsufficientStock, "Insufficient stock available", details, nil)
}

// PaymentFailedError creates a payment failed error
func PaymentFailedError(reason string) *AppError {
	return NewAppError(ErrCodePaymentFailed, "Payment processing failed", reason, nil)
}
