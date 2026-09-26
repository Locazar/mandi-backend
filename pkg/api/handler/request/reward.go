package request

type SubmitShopPurchaseRequest struct {
	ShopID          string   `json:"shop_id" binding:"required,max=32"`
	BillAmountPaise int64    `json:"bill_amount_paise" binding:"required,gt=0"`
	Latitude        *float64 `json:"latitude" binding:"required,gte=-90,lte=90"`
	Longitude       *float64 `json:"longitude" binding:"required,gte=-180,lte=180"`
	ClientRequestID string   `json:"client_request_id" binding:"required,max=64"`
}

type UpdateShopRewardSettingsRequest struct {
	DiscountEnabled *bool `json:"discount_enabled" binding:"required"`
	DiscountPercent int   `json:"discount_percent" binding:"gte=0,lte=100"`
}

type RejectShopPurchaseRequest struct {
	Reason string `json:"reason" binding:"max=200"`
}

type AdjustRewardAccountRequest struct {
	DeltaPoints     int64  `json:"delta_points" binding:"required,gte=-100000,lte=100000"`
	Reason          string `json:"reason" binding:"required,min=5,max=300"`
	ClientRequestID string `json:"client_request_id" binding:"omitempty,max=40"`
}
