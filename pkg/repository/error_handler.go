package repository

import (
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

// HandleDBErrorContext is a helper to convert database errors in repository layer
// It provides consistent error handling across all repository operations
func HandleDBErrorContext(dbErr error, operation, resource string) *domain.AppError {
	if dbErr == nil {
		return nil
	}
	return utils.ConvertDBError(dbErr, resource)
}

// ValidateRequired checks if a value is empty and returns validation error
func ValidateRequired(field, value string) *domain.AppError {
	if value == "" {
		return domain.ValidationError(field, "field is required")
	}
	return nil
}

// ValidateID checks if an ID is valid (non-zero for numeric IDs)
func ValidateID(fieldName string, id uint) *domain.AppError {
	if id == 0 {
		return domain.ValidationError(fieldName, "invalid or missing ID")
	}
	return nil
}

// ValidateEmail is a simple email validation helper
func ValidateEmail(email string) *domain.AppError {
	if email == "" {
		return domain.ValidationError("email", "email is required")
	}
	// Add regex validation if needed
	return nil
}
