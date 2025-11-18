package domain

import "time"

type ShopDetails struct {
	ID        uint   `json:"id" gorm:"primaryKey;not null"`
	OwnerID   uint   `json:"owner_id" gorm:"not null"`
	ShopID    string `json:"shop_id" gorm:"size:100;uniqueIndex;not null"`
	ShopName  string `json:"shop_name" gorm:"size:100;not null"`
	OwnerName string `json:"owner_name" gorm:"size:100;not null"`
	Email     string `json:"email" gorm:"size:100;uniqueIndex;not null"`
	Mobile    string `json:"mobile" gorm:"size:50;not null"`

	AddressLine1 string  `json:"address_line_1" gorm:"size:255" binding:"required"`
	AddressLine2 string  `json:"address_line_2" gorm:"size:255" binding:"omitempty"`
	City         string  `json:"city" gorm:"size:50" binding:"required"`
	State        string  `json:"state" gorm:"size:50" binding:"required"`
	Country      string  `json:"country" gorm:"size:50" binding:"required"`
	Pincode      string  `json:"pincode" gorm:"size:50" binding:"required"`
	Latitude     float64 `json:"latitude" gorm:"type:decimal(10,7);"`
	Longitude    float64 `json:"longitude" gorm:"type:decimal(10,7);"`

	ShopDescription        string `json:"shop_description" gorm:"type:text" binding:"omitempty"`
	ShopVerificationDocs   string `json:"shop_verification_docs" gorm:"type:text" binding:"omitempty"`
	GSTIN                  string `json:"gstin" gorm:"size:50" binding:"omitempty"`
	MSMERegistrationNumber string `json:"msme_registration_number" gorm:"size:50" binding:"omitempty"`
	ElectricityBill        string `json:"electricity_bill" gorm:"size:255" binding:"omitempty"`
	PanNumber              string `json:"pan_number" gorm:"size:20" binding:"omitempty"`
	ITRDocuments           string `json:"itr_documents" gorm:"type:text" binding:"omitempty"`

	ShopType   string `json:"shop_type" gorm:"size:50" binding:"omitempty"`
	ShopStatus string `json:"shop_status" gorm:"size:50" binding:"omitempty"`

	BankAccountNumber string `json:"bank_account_number" gorm:"size:50" binding:"omitempty"`
	BankIFSC          string `json:"bank_ifsc" gorm:"size:20" binding:"omitempty"`

	ShopVerificationStatus  string `json:"shop_verification_status" gorm:"size:50" binding:"omitempty"`
	ShopVerificationRemarks string `json:"shop_verification_remarks" gorm:"size:255" binding:"omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
