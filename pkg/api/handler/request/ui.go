package request

// UISellerRequest represents the UI seller endpoint request
type UISellerRequest struct {
	Context map[string]interface{} `json:"context" binding:"required"`
}
