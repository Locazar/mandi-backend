package request

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Limit  uint64
	Offset uint64
}

const (
	defaultLimit  = 25
	defaultOffset = 0
)

func GetPagination(ctx *gin.Context) Pagination {

	pagination := Pagination{
		Limit:  defaultLimit,
		Offset: defaultOffset,
	}

	num, err := strconv.ParseUint(ctx.Query("limit"), 10, 64)
	if err == nil {
		pagination.Limit = num
	}

	num, err = strconv.ParseUint(ctx.Query("offset"), 10, 64)
	if err == nil {
		pagination.Offset = num
	}
	return pagination
}
