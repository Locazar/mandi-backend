package request

import "time"

type Notification struct {
	SenderType           string    `json:"sender_type" binding:"required,oneof=user seller admin"`
	ReceiverType         string    `json:"receiver_type" binding:"required,oneof=user seller admin"`
	Title                string    `json:"title" binding:"required,min=2,max=100"`
	Message              string    `json:"message" binding:"required,min=5,max=500"`
	Body                 string    `json:"body" binding:"required,min=5,max=500"`
	SenderID             uint      `json:"sender_id" binding:"required"`
	ReceiverID           uint      `json:"receiver_id" binding:"required"`
	CategoryID           uint      `json:"category_id" binding:"omitempty"`
	ProductID            uint      `json:"product_id" binding:"omitempty"`
	VariationID          uint      `json:"variation_id" binding:"omitempty"`
	ShopID               uint      `json:"shop_id" binding:"omitempty"`
	OrderID              uint      `json:"order_id" binding:"omitempty"`
	IsRead               bool      `json:"is_read" binding:"omitempty"`
	OfferID              uint      `json:"offer_id" binding:"omitempty"`
	NotificationMetaData string    `json:"notification_meta_data" binding:"omitempty"`
	Status               string    `json:"status" binding:"omitempty,min=2,max=50"`
	CreatedAt            time.Time `json:"created_at" binding:"omitempty"`
	UpdatedAt            time.Time `json:"updated_at" binding:"omitempty"`
}

type GetNotification struct {
	UserID    uint   `form:"user_id" binding:"omitempty"`
	AdminID   uint   `form:"admin_id" binding:"omitempty"`
	ShopID    uint   `form:"shop_id" binding:"omitempty"`
	Status    string `form:"status" binding:"omitempty"`
	ProductID uint   `form:"product_id" binding:"omitempty"`
	OrderID   uint   `form:"order_id" binding:"omitempty"`
	IsRead    *bool  `form:"is_read" binding:"omitempty"`
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

// NotificationDeviceToken is used by clients to register their FCM device token.
type NotificationDeviceToken struct {
	OwnerID   string `json:"owner_id" binding:"required"`
	OwnerType string `json:"owner_type" binding:"required,oneof=user seller"`
	Token     string `json:"token" binding:"required"`
	Platform  string `json:"platform" binding:"omitempty,oneof=android ios web"`
}

// UnregisterDeviceToken removes a device token on logout or token refresh.
type UnregisterDeviceToken struct {
	OwnerID   string `json:"owner_id" binding:"required"`
	OwnerType string `json:"owner_type" binding:"required,oneof=user seller"`
	Token     string `json:"token" binding:"required"`
}

// SendPushRequest triggers a direct FCM push from the backend.
type SendPushRequest struct {
	// OwnerID is the user or seller ID whose tokens to look up.
	OwnerID string `json:"owner_id" binding:"required"`
	// OwnerType is "user" or "seller".
	OwnerType string `json:"owner_type" binding:"required,oneof=user seller"`
	Title     string `json:"title" binding:"required,min=1,max=100"`
	Body      string `json:"body" binding:"required,min=1,max=500"`
	// Data is an optional map of key-value pairs delivered alongside the notification.
	Data map[string]string `json:"data" binding:"omitempty"`
	// EventType is a hint for the client app (e.g. "order_status_changed").
	EventType string `json:"event_type" binding:"omitempty"`
}

type NotificationRadiusRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	RadiusM   float64 `json:"radius_meters" binding:"required"`
	Title     string  `json:"title" binding:"required"`
	Body      string  `json:"body" binding:"required"`
}
