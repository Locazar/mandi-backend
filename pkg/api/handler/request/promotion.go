package request

import "time"

type CreatePromotionRequest struct {
	ShopID                 string    `json:"shop_id" binding:"required"`
	PromotionCategoryID    string    `json:"promotion_category_id" binding:"required"`
	PromotionTypeID        string    `json:"promotion_type_id" binding:"required"`
	OfferName              string    `json:"offer_name" binding:"required"`
	Description            string    `json:"description" binding:"required"`
	DiscountRate           float64   `json:"discount_rate"`
	StartDate              time.Time `json:"start_date" binding:"required"`
	EndDate                time.Time `json:"end_date" binding:"required"`
	IsActive               bool      `json:"is_active"`
	MinimumPurchaseAmount  *float64  `json:"minimum_purchase_amount,omitempty"`
	TierQuantity           *int      `json:"tier_quantity,omitempty"`
	BogoGetQuantity        *int      `json:"bogo_get_quantity,omitempty"`
	BogoBuyQuantity        *int      `json:"bogo_buy_quantity,omitempty"`
	BogoCombinationEnabled *bool     `json:"bogo_combination_enabled,omitempty"`
	GiftDescription        *string   `json:"gift_description,omitempty"`
}

// UpdatePromotionRequest holds the fields a seller may update on an existing promotion.
type UpdatePromotionRequest struct {
	OfferName              string    `json:"offer_name"`
	Description            string    `json:"description"`
	DiscountRate           *float64  `json:"discount_rate,omitempty"`
	StartDate              *time.Time `json:"start_date,omitempty"`
	EndDate                *time.Time `json:"end_date,omitempty"`
	IsActive               *bool     `json:"is_active,omitempty"`
	MinimumPurchaseAmount  *float64  `json:"minimum_purchase_amount,omitempty"`
	TierQuantity           *int      `json:"tier_quantity,omitempty"`
	BogoGetQuantity        *int      `json:"bogo_get_quantity,omitempty"`
	BogoBuyQuantity        *int      `json:"bogo_buy_quantity,omitempty"`
	BogoCombinationEnabled *bool     `json:"bogo_combination_enabled,omitempty"`
	GiftDescription        *string   `json:"gift_description,omitempty"`
}
