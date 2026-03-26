package utils

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// To append the message to the error
func AppendMessageToError(err error, message string) error {
	if err == nil {
		return errors.New(message)
	}
	return fmt.Errorf("%w \n%s", err, message)
}

// To prepend the message to the error
func PrependMessageToError(err error, message string) error {
	if err == nil {
		return errors.New(message)
	}
	return fmt.Errorf("%s \n%w", message, err)
}

// ErrorResponse represents the structure of error responses
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"` // Only include in development
}

// SendError sends a standardized error response
// In production, it hides internal errors and logs them
func SendError(c *gin.Context, statusCode int, message string, err error) {
	response := ErrorResponse{
		Success: false,
		Message: message,
	}

	// Log the actual error for debugging
	if err != nil {
		log.Printf("Error: %v\nStack: %s", err, debug.Stack())
	}

	// In production, don't expose internal errors
	// You can check environment variable or config
	// For now, assuming production mode, omit the error field
	// If you want to include in dev, uncomment:
	// if gin.Mode() != gin.ReleaseMode && err != nil {
	//     response.Error = err.Error()
	// }

	c.JSON(statusCode, response)
	c.Abort()
}

// Common error handlers
func SendBadRequest(c *gin.Context, message string, err error) {
	SendError(c, http.StatusBadRequest, message, err)
}

func SendUnauthorized(c *gin.Context, message string, err error) {
	SendError(c, http.StatusUnauthorized, message, err)
}

func SendForbidden(c *gin.Context, message string, err error) {
	SendError(c, http.StatusForbidden, message, err)
}

func SendNotFound(c *gin.Context, message string, err error) {
	SendError(c, http.StatusNotFound, message, err)
}

func SendInternalServerError(c *gin.Context, message string, err error) {
	SendError(c, http.StatusInternalServerError, message, err)
}

func SendConflict(c *gin.Context, message string, err error) {
	SendError(c, http.StatusConflict, message, err)
}

// Recovery middleware to catch panics
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("Panic recovered: %v\nStack: %s", recovered, debug.Stack())
		SendInternalServerError(c, "Internal server error", nil)
	})
}
