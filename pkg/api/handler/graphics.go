package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/service/graphics"
)

type GraphicsHandler struct {
	graphicsService *graphics.GraphicsService
}

func NewGraphicsHandler() *GraphicsHandler {
	graphicsService := graphics.NewGraphicsService("./uploads/offers")
	return &GraphicsHandler{
		graphicsService: graphicsService.(*graphics.GraphicsService),
	}
}

// GenerateOfferImagePreview godoc
//
//	@Summary		Generate Offer Image Preview (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to generate offer image preview for testing
//	@Id				GenerateOfferImagePreview
//	@Tags			Admin Graphics
//	@Param			input	body	request.Offer{}	true	"Offer details for image generation"
//	@Router			/admin/graphics/offer-preview [post]
//	@Success		200	{object}	response.Response{}	"Successfully generated offer image"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Failed to generate image"
func (g *GraphicsHandler) GenerateOfferImagePreview(ctx *gin.Context) {
	var body request.Offer

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid input", err, nil)
		return
	}

	// Generate unique filename for preview
	imageUUID := uuid.New().String()
	mainImageFilename := fmt.Sprintf("preview_offer_%s.png", imageUUID)
	thumbnailFilename := fmt.Sprintf("preview_offer_thumb_%s.png", imageUUID)

	// Generate main offer image
	mainImagePath, err := g.graphicsService.GenerateOfferImage(body, mainImageFilename)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to generate offer image", err, nil)
		return
	}

	// Generate thumbnail
	thumbnailPath, err := g.graphicsService.GenerateThumbnail(body, thumbnailFilename)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to generate thumbnail", err, nil)
		return
	}

	// Return the file paths for frontend to display
	ctx.JSON(http.StatusOK, response.Response{
		Status:  true,
		Message: "Successfully generated offer images",
		Data: map[string]interface{}{
			"main_image":     fmt.Sprintf("/uploads/offers/%s", mainImageFilename),
			"thumbnail":      fmt.Sprintf("/uploads/offers/thumbnail/%s", thumbnailFilename),
			"image_path":     mainImagePath,
			"thumbnail_path": thumbnailPath,
		},
	})
}
