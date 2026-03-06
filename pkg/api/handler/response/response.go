package response

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type Response struct {
	Status  bool        `json:"success"`
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
	Data    interface{} `json:"data"`
}

// ErrorInfo contains detailed error information for structured errors
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func SuccessResponse(ctx *gin.Context, statusCode int, message string, data ...interface{}) {

	log.Printf("\033[0;32m%s\033[0m\n", message)

	var out interface{}
	if len(data) == 0 {
		out = nil
	} else if len(data) == 1 {
		out = data[0]
	} else {
		out = data
	}

	response := Response{
		Status:  true,
		Message: message,
		Error:   nil,
		Data:    out,
	}
	ctx.JSON(statusCode, response)
}

// ErrorResponse handles error responses with old signature (for backward compatibility)
func ErrorResponse(ctx *gin.Context, statusCode int, message string, err error, data interface{}) {

	var errFields []string
	if err != nil {
		log.Printf("\033[0;31m%s\033[0m\n", err.Error())
		errFields = strings.Split(err.Error(), "\n")
	} else {
		log.Printf("\033[0;31m%s\033[0m\n", message)
		errFields = []string{message}
	}

	response := Response{
		Status:  false,
		Message: message,
		Error:   errFields,
		Data:    data,
	}

	ctx.JSON(statusCode, response)
}

// ErrorResponseAppError handles error responses using the new AppError structure
func ErrorResponseAppError(ctx *gin.Context, appErr *domain.AppError) {
	if appErr == nil {
		appErr = domain.InternalError("unknown error occurred", nil)
	}

	log.Printf("\033[0;31m%s\033[0m\n", appErr.Error())

	errorInfo := &ErrorInfo{
		Code:    string(appErr.Code),
		Message: appErr.Message,
		Details: appErr.Details,
	}

	response := Response{
		Status:  false,
		Message: appErr.Message,
		Error:   errorInfo,
		Data:    nil,
	}

	ctx.JSON(appErr.StatusCode, response)
}
