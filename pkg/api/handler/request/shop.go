package request

import "time"

type ShopVerification struct {
	ShopId                     string    `json:"shop_id" binding:"required"`
	Photo_Shop_Verification    bool      `json:"photo_shop_verification" gorm:"type:text;" binding:"omitempty"`
	Business_Doc_Verification  bool      `json:"business_doc_verification" gorm:"type:text;" binding:"omitempty"`
	Identity_Doc_Verification  bool      `json:"identity_doc_verification" gorm:"type:text;" binding:"omitempty"`
	Address_Proof_Verification bool      `json:"address_proof_verification" gorm:"type:text;" binding:"omitempty"`
	UpdatedAt                  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ShopSearch filters for the admin shop search endpoint.
// All fields are optional; provided filters are combined with AND.
type ShopSearch struct {
	Phone     string  `form:"phone"`
	Pincode   string  `form:"pincode"`
	City      string  `form:"city"`
	Search    string  `form:"search"` // matches shop name or owner name
	Latitude  float64 `form:"lat"`
	Longitude float64 `form:"lng"`
	RadiusKm  float64 `form:"radius_km"`
	Limit     int     `form:"limit"`
}

type SetShopTimeRequest struct {
	Status    bool   `json:"status"`
	OpenTime  string `json:"open_time" binding:"required"`
	CloseTime string `json:"close_time" binding:"required"`
}
