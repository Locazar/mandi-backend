package response

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  bool        `json:"success"`
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
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
