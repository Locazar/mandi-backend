package request

// Request DTOs for the AI Service proxy endpoints (pkg/api/handler/ai.go).
// These mirror the contract exposed by the ai-service microservice and carry
// gin binding tags for input validation.

// AIValidateProductRequest is the body for POST /api/ai/validate-product.
type AIValidateProductRequest struct {
	ImagePath string `json:"image_path" binding:"required"`
	Category  string `json:"category" binding:"required"`
}

// AICompareImagesRequest is the body for POST /api/ai/compare-images.
type AICompareImagesRequest struct {
	ImagePath1 string `json:"image_path1" binding:"required"`
	ImagePath2 string `json:"image_path2" binding:"required"`
}

// AIDetectObjectsRequest is the body for POST /api/ai/detect-objects.
// The image is sent inline as a base64-encoded JPEG.
type AIDetectObjectsRequest struct {
	ImageBase64 string `json:"image_base64" binding:"required"`
	DetectAll   bool   `json:"detect_all"`
}

// AIVerifyImageRequest is the body for POST /api/ai/verify-image.
// The image is sent inline as a base64-encoded JPEG.
type AIVerifyImageRequest struct {
	ImageBase64 string `json:"image_base64" binding:"required"`
	Prompt      string `json:"prompt" binding:"required"`
}

// AIEmbedRequest is the body for POST /api/ai/embed.
type AIEmbedRequest struct {
	Text string `json:"text" binding:"required"`
}

// AIEmbedBatchRequest is the body for POST /api/ai/embed-batch.
type AIEmbedBatchRequest struct {
	Texts []string `json:"texts" binding:"required,min=1"`
}
