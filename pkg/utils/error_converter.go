package utils

import (
	"database/sql"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"strings"
)

// ConvertDBError converts database errors to AppError
func ConvertDBError(err error, context string) *domain.AppError {
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return domain.NotFoundError(context)
	}

	// Handle PostgreSQL specific errors
	errStr := err.Error()
	if strings.Contains(errStr, "unique constraint") {
		return domain.AlreadyExistsError(context)
	}
	if strings.Contains(errStr, "foreign key constraint") {
		return domain.BusinessLogicError("Invalid reference", context)
	}
	if strings.Contains(errStr, "connection refused") {
		return domain.DatabaseError("connection failed", err)
	}

	return domain.DatabaseError(context, err)
}

// ConvertValidationError converts validation errors to AppError
func ConvertValidationError(field, reason string) *domain.AppError {
	return domain.ValidationError(field, reason)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*domain.AppError)
	return ok
}

// GetAppError extracts AppError from error chain
func GetAppError(err error) *domain.AppError {
	if appErr, ok := err.(*domain.AppError); ok {
		return appErr
	}
	return nil
}

// WrapError wraps a regular error with context as AppError
func WrapError(err error, code domain.ErrorCode, message string) *domain.AppError {
	if appErr, ok := err.(*domain.AppError); ok {
		return appErr
	}
	return domain.NewAppError(code, message, "", err)
}
