package domain

import "time"

type ShopDetails struct {
	ID        uint   `json:"id" gorm:"primaryKey;"`
	AdminID   uint   `json:"admin_id" gorm:";uniqueIndex"`
	ShopName  string `json:"shop_name" gorm:"size:100;"`
	OwnerName string `json:"owner_name" gorm:"size:100;"`
	Email     string `json:"email" gorm:"size:100;uniqueIndex;"`
	Phone     string `json:"phone" gorm:"size:50;"`

	AddressLine1 string  `json:"address_line1" gorm:"size:255"`
	AddressLine2 string  `json:"address_line2" gorm:"size:255" binding:"omitempty"`
	City         string  `json:"city" gorm:"size:50" binding:"required"`
	State        string  `json:"state" gorm:"size:50" binding:"required"`
	Country      string  `json:"country" gorm:"size:50" binding:"required"`
	Pincode      string  `json:"pincode" gorm:"size:50" binding:"required"`
	Latitude     float64 `json:"latitude" gorm:"type:decimal(10,7);"`
	Longitude    float64 `json:"longitude" gorm:"type:decimal(10,7);"`

	ShopDescription      string `json:"shop_description" gorm:"type:text" binding:"omitempty"`
	ShopVerificationDocs string `json:"shop_verification_docs" gorm:"type:text;" binding:"omitempty"`
	Document_Type        string `json:"document_type" gorm:"size:50" binding:"omitempty"`
	Document_Value       string `json:"document_value" gorm:"type:text" binding:"omitempty"`
	PanNumber            string `json:"pan_number" gorm:"size:20" binding:"omitempty"`
	ITRDocuments         string `json:"itr_documents" gorm:"type:text" binding:"omitempty"`

	ShopType   string `json:"shop_type" gorm:"size:50" binding:"omitempty"`
	ShopStatus string `json:"shop_status" gorm:"size:50" binding:"omitempty"`

	BankAccountNumber string `json:"bank_account_number" gorm:"size:50" binding:"omitempty"`
	BankIFSC          string `json:"bank_ifsc" gorm:"size:20" binding:"omitempty"`
	Shop_Image_URL    string `json:"shop_image_url" gorm:"size:255" binding:"omitempty"`

	ShopVerificationStatus     bool   `json:"shop_verification_status" gorm:"not null;default:false" binding:"omitempty"`
	ShopVerificationRemarks    string `json:"shop_verification_remarks" gorm:"not null;default:false" binding:"omitempty"`
	Photo_Shop_Verification    bool   `json:"photo_shop_verification" gorm:"not null;default:false" binding:"omitempty"`
	Business_Doc_Verification  bool   `json:"business_doc_verification" gorm:"not null;default:false" binding:"omitempty"`
	Identity_Doc_Verification  bool   `json:"identity_doc_verification" gorm:"not null;default:false" binding:"omitempty"`
	Address_Proof_Verification bool   `json:"address_proof_verification" gorm:"not null;default:false" binding:"omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:";autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
