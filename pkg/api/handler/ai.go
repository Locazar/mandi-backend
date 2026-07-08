package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	service "github.com/rohit221990/mandi-backend/pkg/service/ai"
)

// AIHandler proxies AI operations to the ai-service microservice via the injected
// HTTP client (pkg/service/ai). It exposes the ai-service contract under /api/ai so
// the backend's own clients can reach Vision, object-detection, and embedding features
// without talking to the microservice directly. The target service URL is configured
// via AI_SERVICE_URL.
type AIHandler struct {
	aiClient *service.Client
}

// NewAIHandler creates a new AIHandler backed by the ai-service client.
func NewAIHandler(aiClient *service.Client) *AIHandler {
	return &AIHandler{aiClient: aiClient}
}

// HealthCheck godoc
//
//	@Summary		AI service health check
//	@Description	Proxies to the ai-service health endpoint to report whether it is reachable.
//	@Tags			AI Service
//	@ID				AIHealthCheck
//	@Produce		json
//	@Router			/ai/health [get]
//	@Success		200	{object}	response.Response{}	"AI service is healthy"
//	@Failure		503	{object}	response.Response{}	"AI service is unreachable"
func (h *AIHandler) HealthCheck(ctx *gin.Context) {
	if err := h.aiClient.HealthCheck(); err != nil {
		response.ErrorResponse(ctx, http.StatusServiceUnavailable, "AI service is unreachable", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "AI service is healthy", gin.H{"status": "ok"})
}

// ValidateProduct godoc
//
//	@Summary		Validate a product image against a category
//	@Description	Checks whether a product image (referenced by a server-readable path) matches the given category. Returns valid/confidence/reason.
//	@Tags			AI Service
//	@ID				AIValidateProduct
//	@Accept			json
//	@Produce		json
//	@Param			input	body		request.AIValidateProductRequest	true	"Image path and category"
//	@Router			/ai/validate-product [post]
//	@Success		200	{object}	response.Response{}	"Product validation completed"
//	@Failure		400	{object}	response.Response{}	"Invalid request body"
//	@Failure		502	{object}	response.Response{}	"AI service call failed"
func (h *AIHandler) ValidateProduct(ctx *gin.Context) {
	var req request.AIValidateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}

	result, err := h.aiClient.ValidateProduct(req.ImagePath, req.Category)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadGateway, "Failed to validate product image", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Product validation completed", result)
}

// CompareImages godoc
//
//	@Summary		Compare two product images by category
//	@Description	Determines whether two images (referenced by server-readable paths) belong to the same product category.
//	@Tags			AI Service
//	@ID				AICompareImages
//	@Accept			json
//	@Produce		json
//	@Param			input	body		request.AICompareImagesRequest	true	"Two image paths"
//	@Router			/ai/compare-images [post]
//	@Success		200	{object}	response.Response{}	"Image comparison completed"
//	@Failure		400	{object}	response.Response{}	"Invalid request body"
//	@Failure		502	{object}	response.Response{}	"AI service call failed"
func (h *AIHandler) CompareImages(ctx *gin.Context) {
	var req request.AICompareImagesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}

	result, err := h.aiClient.CompareImages(req.ImagePath1, req.ImagePath2)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadGateway, "Failed to compare images", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Image comparison completed", result)
}

// DetectObjects godoc
//
//	@Summary		Detect objects in an image
//	@Description	Detects objects (label/category/color/material/brand/confidence) in a base64-encoded JPEG. Set detect_all=false to return only the most prominent object.
//	@Tags			AI Service
//	@ID				AIDetectObjects
//	@Accept			json
//	@Produce		json
//	@Param			input	body		request.AIDetectObjectsRequest	true	"Base64 image and detect_all flag"
//	@Router			/ai/detect-objects [post]
//	@Success		200	{object}	response.Response{}	"Object detection completed"
//	@Failure		400	{object}	response.Response{}	"Invalid request body"
//	@Failure		502	{object}	response.Response{}	"AI service call failed"
func (h *AIHandler) DetectObjects(ctx *gin.Context) {
	var req request.AIDetectObjectsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}

	result, err := h.aiClient.DetectObjects(req.ImageBase64, req.DetectAll)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadGateway, "Failed to detect objects", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Object detection completed", result)
}

// VerifyImage godoc
//
//	@Summary		Verify an image against a prompt
//	@Description	Checks whether a base64-encoded JPEG satisfies a free-text condition prompt. Returns matches/confidence/reason.
//	@Tags			AI Service
//	@ID				AIVerifyImage
//	@Accept			json
//	@Produce		json
//	@Param			input	body		request.AIVerifyImageRequest	true	"Base64 image and condition prompt"
//	@Router			/ai/verify-image [post]
//	@Success		200	{object}	response.Response{}	"Image verification completed"
//	@Failure		400	{object}	response.Response{}	"Invalid request body"
//	@Failure		502	{object}	response.Response{}	"AI service call failed"
func (h *AIHandler) VerifyImage(ctx *gin.Context) {
	var req request.AIVerifyImageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}

	result, err := h.aiClient.VerifyImage(req.ImageBase64, req.Prompt)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadGateway, "Failed to verify image", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Image verification completed", result)
}

// GenerateEmbedding godoc
//
//	@Summary		Generate a text embedding
//	@Description	Generates an embedding vector for a single text for semantic search.
//	@Tags			AI Service
//	@ID				AIGenerateEmbedding
//	@Accept			json
//	@Produce		json
//	@Param			input	body		request.AIEmbedRequest	true	"Text to embed"
//	@Router			/ai/embed [post]
//	@Success		200	{object}	response.Response{}	"Embedding generated successfully"
//	@Failure		400	{object}	response.Response{}	"Invalid request body"
//	@Failure		502	{object}	response.Response{}	"AI service call failed"
func (h *AIHandler) GenerateEmbedding(ctx *gin.Context) {
	var req request.AIEmbedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}

	result, err := h.aiClient.GenerateEmbedding(req.Text)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadGateway, "Failed to generate embedding", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Embedding generated successfully", result)
}

// GenerateEmbeddings godoc
//
//	@Summary		Generate batch text embeddings
//	@Description	Generates embedding vectors for multiple texts in a single call.
//	@Tags			AI Service
//	@ID				AIGenerateEmbeddings
//	@Accept			json
//	@Produce		json
//	@Param			input	body		request.AIEmbedBatchRequest	true	"Texts to embed"
//	@Router			/ai/embed-batch [post]
//	@Success		200	{object}	response.Response{}	"Embeddings generated successfully"
//	@Failure		400	{object}	response.Response{}	"Invalid request body"
//	@Failure		502	{object}	response.Response{}	"AI service call failed"
func (h *AIHandler) GenerateEmbeddings(ctx *gin.Context) {
	var req request.AIEmbedBatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}

	result, err := h.aiClient.GenerateEmbeddings(req.Texts)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadGateway, "Failed to generate embeddings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Embeddings generated successfully", result)
}
