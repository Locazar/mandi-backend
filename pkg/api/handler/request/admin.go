package request

import "mime/multipart"

type AdminUploadImageRequest struct {
	Image *multipart.FileHeader `form:"image" binding:"required"`
}

type DocumentRequest struct {
	DocumentValue string `json:"document_value" binding:"required"`
	DocumentType  string `json:"document_type" binding:"required"`
}

type AddressRequest struct {
	ShopName     string `json:"shop_name" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	AddressLine1 string `json:"address_line_1" binding:"required"`
	AddressLine2 string `json:"address_line_2" binding:"omitempty"`
	City         string `json:"city" binding:"required"`
	State        string `json:"state" binding:"required"`
	Pincode      string `json:"pincode"`
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type VerifyShopDocumentRequest struct {
	OTP           string `json:"otp" binding:"required"`
	DocumentValue string `json:"document_value" binding:"required"`
	DocumentType  string `json:"document_type" binding:"required"`
}
