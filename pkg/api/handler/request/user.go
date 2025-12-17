package request

import "mime/multipart"

type UserSignUp struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Age             uint   `json:"age"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

// for address add address
type Address struct {
	Name        string  `json:"name" binding:"required,min=2,max=50"`
	PhoneNumber string  `json:"phone_number" binding:"required,min=10,max=10"`
	House       string  `json:"house" binding:"required"`
	Area        string  `json:"area"`
	LandMark    string  `json:"land_mark" binding:"required"`
	City        string  `json:"city"`
	Pincode     uint    `json:"pincode" binding:"required"`
	CountryID   uint    `json:"country_id" binding:"required"`
	Latitude    float64 `json:"latitude" binding:"omitempty"`
	Longitude   float64 `json:"longitude" binding:"omitempty"`
	IsDefault   *bool   `json:"is_default"`
}
type EditAddress struct {
	ID          uint    `json:"address_id" binding:"required"`
	Name        string  `json:"name" binding:"required,min=2,max=50"`
	PhoneNumber string  `json:"phone_number" binding:"required,min=10,max=10"`
	House       string  `json:"house" binding:"required"`
	Area        string  `json:"area"`
	LandMark    string  `json:"land_mark" binding:"required"`
	City        string  `json:"city"`
	Pincode     uint    `json:"pincode" binding:"required"`
	CountryID   uint    `json:"country_id" binding:"required"`
	Latitude    float64 `json:"latitude" binding:"omitempty"`
	Longitude   float64 `json:"longitude" binding:"omitempty"`
	IsDefault   *bool   `json:"is_default"`
}

// user side
type Cart struct {
	UserID        uint `json:"-"`
	ProductItemID uint `json:"product_item_id" binding:"required"`
}

type UpdateCartItem struct {
	UserID        uint `json:"-"`
	ProductItemID uint `json:"product_item_id" binding:"required"`
	Count         uint `json:"count" binding:"omitempty,gte=1"`
}

type EditUser struct {
	FirstName       string `json:"first_name"  binding:"required,min=2,max=50"`
	LastName        string `json:"last_name"  binding:"required,min=1,max=50"`
	Age             uint   `json:"age"  binding:"required,numeric"`
	Email           string `json:"email" binding:"required,email"`
	Phone           string `json:"phone" binding:"required,min=10,max=10"`
	Password        string `json:"password"  binding:"omitempty,eqfield=ConfirmPassword"`
	ConfirmPassword string `json:"confirm_password" binding:"omitempty"`
}

type UploadImageRequest struct {
	UserID string                `form:"user_id" binding:"required"`
	Image  *multipart.FileHeader `form:"image" binding:"required"`
}

type SellerRadiusRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	RadiusKm  float64 `json:"radius_km" binding:"required"`
	Pagination
}
