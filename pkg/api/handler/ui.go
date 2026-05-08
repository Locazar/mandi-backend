package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type UIHandler struct {
	adminUseCase usecaseInterface.AdminUseCase
}

func NewUIHandler(adminUsecase usecaseInterface.AdminUseCase) interfaces.UIHandler {
	return &UIHandler{
		adminUseCase: adminUsecase,
	}
}

// SellerUIEndpoint godoc
//
//	@Summary		Get seller UI template based on intent
//	@Security		BearerAuth
//	@Description	API for seller to get UI template for different actions
//	@Id				SellerUIEndpoint
//	@Tags			Seller UI
//	@Param			input	body	request.UISellerRequest{}	true	"Intent and context"
//	@Router			/ui/seller [post]
//	@Success		200	{object}	response.Response{}	"Successfully retrieved UI template"
//	@Failure		400	{object}	response.Response{}	"Invalid request or missing parameters"
//	@Failure		401	{object}	response.Response{}	"Unauthorized - Invalid token"
//	@Failure		500	{object}	response.Response{}	"Failed to retrieve UI template"
func (h *UIHandler) SellerUIEndpoint(ctx *gin.Context) {
	var req request.UISellerRequest

	// Parse request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request format", err, nil)
		return
	}

	// Get token from header and decode to get seller/admin ID
	tokenString := ctx.GetHeader("Authorization")
	if tokenString == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Missing authorization token", nil, nil)
		return
	}

	sellerID := h.adminUseCase.DecodeTokenData(tokenString)
	if sellerID == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid or expired token", nil, nil)
		return
	}

	// Extract shop_id from context
	shopIDInterface, ok := req.Context["shop_id"]
	if !ok {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Missing shop_id in context", nil, nil)
		return
	}

	shopIDStr, ok := shopIDInterface.(string)
	if !ok {
		response.ErrorResponse(ctx, http.StatusBadRequest, "shop_id must be a string", nil, nil)
		return
	}

	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop_id format", err, nil)
		return
	}

	// Get shop details to validate shop exists
	shop, err := h.adminUseCase.GetShopByID(ctx, uint(shopID))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch shop details", err, nil)
		return
	}

	if shop.ID == 0 {
		response.ErrorResponse(ctx, http.StatusNotFound, "Shop not found", nil, nil)
		return
	}

	// Forward request to external UI service on port 3002
	// Pass shop details to the forwarding function
	externalResponse, err := h.forwardToExternalUIService(ctx, req, sellerID, &shop)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get UI template from service", err, nil)
		return
	}

	// Return the external service response
	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved UI template", externalResponse)
}

// forwardToExternalUIService sends enriched request with seller and shop data to external UI service
func (h *UIHandler) forwardToExternalUIService(ctx *gin.Context, req request.UISellerRequest, sellerID string, shop *domain.ShopDetails) (interface{}, error) {
	// Convert seller ID string to uint
	sellerIDUint, err := strconv.ParseUint(sellerID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid seller ID format: %w", err)
	}

	// Get seller/admin details from database
	admin, err := h.adminUseCase.GetAdminByID(ctx, uint(sellerIDUint))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch admin details: %w", err)
	}

	// Get shop verification details
	_, shopVerification, err := h.adminUseCase.GetAdminWithShopVerificationByPhone(ctx, admin.Mobile)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shop verification details: %w", err)
	}

	// Determine pending items for seller
	sellerPendingItems := h.getSellerPendingItems(&admin, &shopVerification)

	// Determine pending items for shop
	shopPendingItems := h.getShopPendingItems(shop, &shopVerification)

	// Check if anything is missing/pending and create appropriate intents
	missingIntents := h.determineMissingIntents(&admin, shop, &shopVerification, sellerPendingItems, shopPendingItems)

	// If there are missing intents, forward those instead of the original intent
	intentsToForward := []string{}
	if len(missingIntents) > 0 {
		intentsToForward = missingIntents
	}

	// Forward simplified requests for each intent with only intent, shop_id, and seller_id
	var externalResponse interface{}
	for _, intent := range intentsToForward {
		resp, err := h.forwardSimplifiedRequest(ctx, intent, strconv.FormatUint(uint64(shop.ID), 10), sellerID)
		if err != nil {
			return nil, fmt.Errorf("failed to forward intent '%s': %w", intent, err)
		}
		externalResponse = resp
	}

	return externalResponse, nil
}

// forwardSimplifiedRequest sends a simplified request with only intent, shop_id, and seller_id
func (h *UIHandler) forwardSimplifiedRequest(ctx *gin.Context, intent string, shopID string, sellerID string) (interface{}, error) {
	// Build simplified payload with only intent, shop_id, and seller_id
	simplifiedPayload := map[string]interface{}{
		"intent":    intent,
		"shop_id":   shopID,
		"seller_id": sellerID,
	}

	// Marshal simplified payload
	payload, err := json.Marshal(simplifiedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal simplified request: %w", err)
	}

	// Create HTTP request
	externalURL := "http://localhost:3002/api/ui/seller"
	httpReq, err := http.NewRequest("POST", externalURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sellerID))

	// Make HTTP call
	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call external UI service: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if external service returned an error
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external service returned status %d: %s", httpResp.StatusCode, string(respBody))
	}

	// Parse response
	var externalResp interface{}
	if err := json.Unmarshal(respBody, &externalResp); err != nil {
		return nil, fmt.Errorf("failed to parse external response: %w", err)
	}

	return externalResp, nil
}

// determineMissingIntents checks for missing/false/null/empty data and returns appropriate intents
func (h *UIHandler) determineMissingIntents(admin *domain.Admin, shop *domain.ShopDetails, shopVerification *domain.ShopVerification, sellerPendingItems []string, shopPendingItems []string) []string {
	var missingIntents []string

	if admin == nil || shop == nil {
		return missingIntents
	}

	// Map pending items to intents
	for _, pendingItem := range sellerPendingItems {
		switch pendingItem {
		case "email_verification":
			missingIntents = append(missingIntents, "seller_email_verification")
		case "document_verification":
			missingIntents = append(missingIntents, "seller_document_verification")
		case "shop_setup":
			missingIntents = append(missingIntents, "shop_setup")
		case "address_verification":
			missingIntents = append(missingIntents, "seller_address_verification")
		case "shop_image_upload":
			missingIntents = append(missingIntents, "shop_image_upload")
		}
	}

	// Map shop pending items to intents
	for _, pendingItem := range shopPendingItems {
		switch pendingItem {
		case "shop_image":
			// Only add if not already added
			if !h.contains(missingIntents, "shop_image_upload") {
				missingIntents = append(missingIntents, "shop_image_upload")
			}
		case "shop_address":
			if !h.contains(missingIntents, "shop_address_update") {
				missingIntents = append(missingIntents, "shop_address_update")
			}
		case "shop_pincode":
			if !h.contains(missingIntents, "shop_address_update") {
				missingIntents = append(missingIntents, "shop_address_update")
			}
		case "address_not_verified":
			if !h.contains(missingIntents, "shop_address_verification") {
				missingIntents = append(missingIntents, "shop_address_verification")
			}
		case "shop_not_verified":
			if !h.contains(missingIntents, "shop_verification") {
				missingIntents = append(missingIntents, "shop_verification")
			}
		}
	}

	return missingIntents
}

// contains checks if a string slice contains a specific string
func (h *UIHandler) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// validateDataForIntent checks shop and admin data based on the request intent
func (h *UIHandler) validateDataForIntent(intent string, admin *domain.Admin, shop *domain.ShopDetails, shopVerification *domain.ShopVerification) error {
	switch intent {
	case "shop_image":
		// Shop image requires shop to exist and have basic info
		if shop == nil || shop.ID == 0 {
			return fmt.Errorf("invalid shop data for shop_image intent")
		}

	case "shop_address":
		// Shop address requires shop to exist
		if shop == nil || shop.ID == 0 {
			return fmt.Errorf("invalid shop data for shop_address intent")
		}

	case "shop_verification":
		// Shop verification requires admin and shop to exist
		if admin == nil || admin.ID == 0 {
			return fmt.Errorf("invalid admin data for shop_verification intent")
		}
		if shop == nil || shop.ID == 0 {
			return fmt.Errorf("invalid shop data for shop_verification intent")
		}

	case "add_product":
		// Add product requires seller to be verified and shop to exist
		if admin == nil || admin.ID == 0 {
			return fmt.Errorf("invalid admin data for add_product intent")
		}
		if shop == nil || shop.ID == 0 {
			return fmt.Errorf("invalid shop data for add_product intent")
		}
		if shopVerification == nil || !shopVerification.VerificationStatus {
			return fmt.Errorf("seller verification required for add_product intent")
		}

	case "product_image":
		// Product image requires admin and shop to exist
		if admin == nil || admin.ID == 0 {
			return fmt.Errorf("invalid admin data for product_image intent")
		}
		if shop == nil || shop.ID == 0 {
			return fmt.Errorf("invalid shop data for product_image intent")
		}

	case "product_details":
		// Product details requires admin and shop to exist
		if admin == nil || admin.ID == 0 {
			return fmt.Errorf("invalid admin data for product_details intent")
		}
		if shop == nil || shop.ID == 0 {
			return fmt.Errorf("invalid shop data for product_details intent")
		}

	default:
		// For unknown intents, ensure basic data exists
		if admin == nil || admin.ID == 0 {
			return fmt.Errorf("invalid admin data for intent: %s", intent)
		}
		if shop == nil || shop.ID == 0 {
			return fmt.Errorf("invalid shop data for intent: %s", intent)
		}
	}

	return nil
}

func (h *UIHandler) getSellerPendingItems(admin *domain.Admin, shopVerification *domain.ShopVerification) []string {
	var pendingItems []string

	if admin == nil {
		return pendingItems
	}

	// Check seller basic info
	if admin.Email == "" || admin.Email == "null" {
		pendingItems = append(pendingItems, "email_verification")
	}

	// Check if seller is verified
	if !admin.VerifiedSeller {
		pendingItems = append(pendingItems, "document_verification")
	}

	// Check if seller has set up shop
	if shopVerification == nil {
		pendingItems = append(pendingItems, "shop_setup")
	}

	// Check shop verification status
	if shopVerification != nil && !shopVerification.VerificationStatus {
		pendingItems = append(pendingItems, "address_verification")
		pendingItems = append(pendingItems, "shop_image_upload")
	}

	return pendingItems
}

// getShopPendingItems determines what items are pending for the shop
func (h *UIHandler) getShopPendingItems(shop *domain.ShopDetails, shopVerification *domain.ShopVerification) []string {
	var pendingItems []string

	if shop == nil {
		return pendingItems
	}

	// Check shop image
	if shop.Shop_Image_URL == "" || shop.Shop_Image_URL == "null" {
		pendingItems = append(pendingItems, "shop_image")
	}

	// Check shop address
	if shop.AddressLine1 == "" || shop.AddressLine1 == "null" {
		pendingItems = append(pendingItems, "shop_address")
	}

	// Check shop pincode
	if shop.Pincode == "" || shop.Pincode == "null" {
		pendingItems = append(pendingItems, "shop_pincode")
	}

	// Check shop city and state
	if shop.City == "" || shop.State == "" {
		pendingItems = append(pendingItems, "shop_address")
	}

	// Check shop verification status
	if shopVerification != nil {
		if !shopVerification.VerificationStatus {
			pendingItems = append(pendingItems, "address_not_verified")
			pendingItems = append(pendingItems, "shop_not_verified")
		}
	}

	return pendingItems
}
