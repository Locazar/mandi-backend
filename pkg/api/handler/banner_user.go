package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type BannerUserHandler struct {
	bannerUseCase usecaseInterface.BannerUseCase
	cloudService  cloud.CloudService
}

func NewBannerUserHandler(bannerUseCase usecaseInterface.BannerUseCase, cloudService cloud.CloudService) *BannerUserHandler {
	return &BannerUserHandler{
		bannerUseCase: bannerUseCase,
		cloudService:  cloudService,
	}
}

// GetFilteredBanners godoc
//
//	@Summary		Get filtered banners by location and category
//	@Description	Returns active banners filtered by lat/lng with radius (Haversine),
//	@Description	pincode match, department_id, category_id, and app_type.
//	@Tags			User Banners
//	@Accept			json
//	@Produce		json
//	@Param			input	body		request.BannerFilterRequest	true	"Banner filter criteria"
//	@Router			/api/banners/filtered [post]
//	@Success		200	{object}	response.Response{}	"Successfully retrieved filtered banners"
//	@Failure		400	{object}	response.Response{}	"Invalid request parameters"
//	@Failure		500	{object}	response.Response{}	"Failed to retrieve banners"
func (h *BannerUserHandler) GetFilteredBanners(ctx *gin.Context) {
	var req request.BannerFilterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request format", err, nil)
		return
	}

	banners, err := h.bannerUseCase.GetFilteredBanners(ctx.Request.Context(), req)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch banners", err, nil)
		return
	}

	for i := range banners {
		banners[i].ImageURL = cloud.ResolveURL(h.cloudService, banners[i].ImageURL)
	}

	response.SuccessResponse(ctx, http.StatusOK, "Banners fetched successfully", map[string]interface{}{
		"count":   len(banners),
		"banners": banners,
	})
}
