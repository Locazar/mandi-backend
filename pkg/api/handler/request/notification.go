package request

import "time"

type Notification struct {
	SenderType           string    `json:"sender_type" binding:"required,oneof=user seller"`
	ReceiverType         string    `json:"receiver_type" binding:"required,oneof=user seller"`
	Title                string    `json:"title" binding:"required,min=2,max=100"`
	Message              string    `json:"message" binding:"required,min=5,max=500"`
	Body                 string    `json:"body" binding:"required,min=5,max=500"`
	SenderID             uint      `json:"sender_id" binding:"required"`
	ReceiverID           uint      `json:"receiver_id" binding:"required"`
	CategoryID           uint      `json:"category_id" binding:"required"`
	ProductID            uint      `json:"product_id" binding:"required"`
	VariationID          uint      `json:"variation_id" binding:"required"`
	ShopID               uint      `json:"shop_id" binding:"required"`
	OrderID              uint      `json:"order_id" binding:"required"`
	IsRead               bool      `json:"is_read" binding:"omitempty"`
	OfferID              uint      `json:"offer_id" binding:"required"`
	NotificationMetaData string    `json:"notification_meta_data" binding:"omitempty"`
	Status               string    `json:"status" binding:"required,min=2,max=50"` // e.g. "accepted", "pending", "rejected"
	CreatedAt            time.Time `json:"created_at" binding:"omitempty"`
	UpdatedAt            time.Time `json:"updated_at" binding:"omitempty"`
}

type GetNotification struct {
	UserID    uint   `json:"user_id" binding:"required"`
	AdminID   uint   `json:"admin_id" binding:"required"`
	ShopID    uint   `json:"shop_id" binding:"required"`
	Status    string `json:"status" binding:"omitempty"`
	ProductID uint   `json:"product_id" binding:"omitempty"`
	OrderID   uint   `json:"order_id" binding:"omitempty"`
	IsRead    *bool  `json:"is_read" binding:"omitempty"`
}

type DeviceToken struct {
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	OwnerID   string     `gorm:"type:varchar(100);not null"`
	OwnerType string     `gorm:"type:varchar(10);not null;check:owner_type IN ('user','seller')"`
	Token     string     `gorm:"type:varchar(255);unique;not null"`
	Platform  string     `gorm:"type:varchar(50)"`
	IsActive  bool       `gorm:"default:true"`
	CreatedAt time.Time  `gorm:"not null;autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
}

type NotificationDeviceToken struct {
	OwnerID   string `json:"owner_id" binding:"required"`
	OwnerType string `json:"owner_type" binding:"required,oneof=user seller"`
	Token     string `json:"token" binding:"required"`
	Platform  string `json:"platform" binding:"omitempty"`
}
type NotificationRadiusRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	RadiusM   float64 `json:"radius_meters" binding:"required"`
	Title     string  `json:"title" binding:"required"`
	Body      string  `json:"body" binding:"required"`
}
