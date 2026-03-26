package request

import "github.com/gin-gonic/gin"

type ProductItemSearchRequest struct {
	Latitude   float64 `form:"latitude" binding:"omitempty,numeric"`
	Longitude  float64 `form:"longitude" binding:"omitempty,numeric"`
	RadiusKm   float64 `form:"radius_km" binding:"omitempty,numeric"`
	Pincode    string  `form:"pincode" binding:"omitempty"`
	Pagination Pagination
}

func (p *ProductItemSearchRequest) GetPagination(ctx *gin.Context) {
	p.Pagination = GetPagination(ctx)
}
