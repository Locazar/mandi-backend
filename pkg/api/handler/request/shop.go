package request

import "time"

type ShopVerification struct {
	Photo_Shop_Verification    bool      `json:"photo_shop_verification" gorm:"type:text;" binding:"omitempty"`
	Business_Doc_Verification  bool      `json:"business_doc_verification" gorm:"type:text;" binding:"omitempty"`
	Identity_Doc_Verification  bool      `json:"identity_doc_verification" gorm:"type:text;" binding:"omitempty"`
	Address_Proof_Verification bool      `json:"address_proof_verification" gorm:"type:text;" binding:"omitempty"`
	UpdatedAt                  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type SetShopTimeRequest struct {
	Status    bool   `json:"status" binding:"required"`
	OpenTime  string `json:"open_time" binding:"required"`
	CloseTime string `json:"close_time" binding:"required"`
}
