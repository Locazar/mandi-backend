package alert_engine

import (
	"context"
	"fmt"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// MissingShopPhotoRule triggers when shop has no photo
type MissingShopPhotoRule struct{}

func (r MissingShopPhotoRule) Key() string {
	return "missing_shop_photo"
}

func (r MissingShopPhotoRule) Evaluate(ctx context.Context, sellerID string, data domain.AggregatedData) (*domain.Alert, error) {
	if data.HasShopPhoto {
		return nil, nil // Rule does not apply
	}

	now := time.Now()
	alert := &domain.Alert{
		SellerID:    data.AdminID,
		Key:         r.Key(),
		Title:       "Add Shop Photo",
		Content:     "Your shop photo is missing. Add a professional photo to attract more customers.",
		Description: "Shop photo helps customers recognize and trust your business. Upload a clear photo of your shop.",
		Type:        "warning",
		Priority:    2,
		IsActive:    true,
		ValidFrom:   &now,
		Frequency:   "daily",
		Actions: []domain.AlertAction{
			{
				Label:      "Upload Photo",
				ActionType: "navigate",
				ActionURL:  "/seller/shop/edit",
			},
			{
				Label:      "Dismiss",
				ActionType: "dismiss",
			},
		},
	}
	return alert, nil
}

// NoProductsRule triggers when seller has not added any products
type NoProductsRule struct{}

func (r NoProductsRule) Key() string {
	return "no_products_added"
}

func (r NoProductsRule) Evaluate(ctx context.Context, sellerID string, data domain.AggregatedData) (*domain.Alert, error) {
	if data.ProductCount > 0 {
		return nil, nil // Rule does not apply
	}

	now := time.Now()
	alert := &domain.Alert{
		SellerID:    data.AdminID,
		Key:         r.Key(),
		Title:       "Start Selling",
		Content:     "You haven't added any products yet. Add your first product to start selling.",
		Description: "Products are the foundation of your business. Upload your first product to reach customers.",
		Type:        "critical",
		Priority:    3,
		IsActive:    true,
		ValidFrom:   &now,
		Frequency:   "weekly",
		Actions: []domain.AlertAction{
			{
				Label:      "Add Product",
				ActionType: "navigate",
				ActionURL:  "/seller/products/add",
			},
			{
				Label:      "Learn More",
				ActionType: "navigate",
				ActionURL:  "/help/add-product",
			},
			{
				Label:      "Dismiss",
				ActionType: "dismiss",
			},
		},
	}
	return alert, nil
}

// ShopNotVerifiedRule triggers when shop is not verified
type ShopNotVerifiedRule struct{}

func (r ShopNotVerifiedRule) Key() string {
	return "shop_not_verified"
}

func (r ShopNotVerifiedRule) Evaluate(ctx context.Context, sellerID string, data domain.AggregatedData) (*domain.Alert, error) {
	if data.IsVerified {
		return nil, nil // Rule does not apply
	}

	now := time.Now()
	alert := &domain.Alert{
		SellerID:    data.AdminID,
		Key:         r.Key(),
		Title:       "Complete Shop Verification",
		Content:     "Your shop is not verified yet. Complete the verification process to unlock all features.",
		Description: "Shop verification builds customer trust and unlocks premium features like promotions and higher limits.",
		Type:        "warning",
		Priority:    2,
		IsActive:    true,
		ValidFrom:   &now,
		Frequency:   "weekly",
		Actions: []domain.AlertAction{
			{
				Label:      "Start Verification",
				ActionType: "navigate",
				ActionURL:  "/seller/verification",
			},
			{
				Label:      "View Status",
				ActionType: "navigate",
				ActionURL:  "/seller/verification/status",
			},
			{
				Label:      "Dismiss",
				ActionType: "dismiss",
			},
		},
		Metadata: map[string]interface{}{
			"verification_steps": []string{
				"Photo verification",
				"Business document verification",
				"Identity document verification",
				"Address proof verification",
			},
		},
	}
	return alert, nil
}

// LowReviewScoreRule triggers when shop has low rating (example for extensibility)
type LowReviewScoreRule struct {
	MinimumScore float64
}

func NewLowReviewScoreRule(minScore float64) *LowReviewScoreRule {
	return &LowReviewScoreRule{
		MinimumScore: minScore,
	}
}

func (r *LowReviewScoreRule) Key() string {
	return "low_review_score"
}

func (r *LowReviewScoreRule) Evaluate(ctx context.Context, sellerID string, data domain.AggregatedData) (*domain.Alert, error) {
	// This is a template - in real implementation, fetch shop rating from repository
	// For now, we skip this as we need the actual rating from DB
	return nil, nil
}

// ProductQualityRule checks if products meet quality standards (example for extensibility)
type ProductQualityRule struct{}

func (r ProductQualityRule) Key() string {
	return "product_quality_check"
}

func (r ProductQualityRule) Evaluate(ctx context.Context, sellerID string, data domain.AggregatedData) (*domain.Alert, error) {
	// This is a template - in real implementation, check product images, descriptions, etc.
	return nil, nil
}

// SessionInactivityRule alerts seller if inactive for too long (example for extensibility)
type SessionInactivityRule struct {
	InactiveDays int
}

func NewSessionInactivityRule(days int) *SessionInactivityRule {
	return &SessionInactivityRule{
		InactiveDays: days,
	}
}

func (r *SessionInactivityRule) Key() string {
	return fmt.Sprintf("session_inactivity_%d_days", r.InactiveDays)
}

func (r *SessionInactivityRule) Evaluate(ctx context.Context, sellerID string, data domain.AggregatedData) (*domain.Alert, error) {
	// Template - requires LastLoginAt from shop details
	return nil, nil
}
