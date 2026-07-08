package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler"
)

// AIRoutes registers the AI Service proxy endpoints under /api/ai. These forward to the
// ai-service microservice via the injected client (AI_SERVICE_URL).
func AIRoutes(api *gin.RouterGroup, aiHandler *handler.AIHandler) {
	ai := api.Group("/ai")
	{
		ai.GET("/health", aiHandler.HealthCheck)
		ai.POST("/validate-product", aiHandler.ValidateProduct)
		ai.POST("/compare-images", aiHandler.CompareImages)
		ai.POST("/detect-objects", aiHandler.DetectObjects)
		ai.POST("/verify-image", aiHandler.VerifyImage)
		ai.POST("/embed", aiHandler.GenerateEmbedding)
		ai.POST("/embed-batch", aiHandler.GenerateEmbeddings)
	}
}
