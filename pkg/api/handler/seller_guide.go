package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
)

// SellerGuideHandler serves static seller onboarding guide data.
type SellerGuideHandler struct{}

// NewSellerGuideHandler creates a new SellerGuideHandler.
func NewSellerGuideHandler() *SellerGuideHandler {
	return &SellerGuideHandler{}
}

// GetCategories handles GET /api/seller-guide/categories
// Returns a static list of seller guide categories for the Khangaro-seller app.
func (h *SellerGuideHandler) GetCategories(ctx *gin.Context) {
	categories := []map[string]interface{}{
		{"id": "1", "name": "Getting Started", "description": "Learn the basics of selling on Localzar"},
		{"id": "2", "name": "Product Management", "description": "How to add and manage your products"},
		{"id": "3", "name": "Orders & Fulfillment", "description": "Managing orders and delivery"},
		{"id": "4", "name": "Promotions & Offers", "description": "Running promotions and creating offers"},
		{"id": "5", "name": "Account & Settings", "description": "Account management and seller settings"},
	}
	response.SuccessResponse(ctx, http.StatusOK, "Categories retrieved", map[string]interface{}{
		"categories": categories,
	})
}
