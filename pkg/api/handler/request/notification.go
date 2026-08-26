package request

import "time"

type Notification struct {
	SenderType           string    `json:"sender_type" binding:"required,oneof=user seller admin"`
	ReceiverType         string    `json:"receiver_type" binding:"required,oneof=user seller admin"`
	Title                string    `json:"title" binding:"required,min=2,max=100"`
	Message              string    `json:"message" binding:"required,min=5,max=500"`
	Body                 string    `json:"body" binding:"required,min=5,max=500"`
	SenderID             string    `json:"sender_id" binding:"required"`
	ReceiverID           string    `json:"receiver_id" binding:"required"`
	CategoryID           string    `json:"category_id" binding:"omitempty"`
	ProductID            string    `json:"product_id" binding:"omitempty"`
	VariationID          string    `json:"variation_id" binding:"omitempty"`
	ShopID               string    `json:"shop_id" binding:"omitempty"`
	OrderID              string    `json:"order_id" binding:"omitempty"`
	IsRead               bool      `json:"is_read" binding:"omitempty"`
	OfferID              string    `json:"offer_id" binding:"omitempty"`
	NotificationMetaData string    `json:"notification_meta_data" binding:"omitempty"`
	Status               string    `json:"status" binding:"omitempty,min=2,max=50"`
	CreatedAt            time.Time `json:"created_at" binding:"omitempty"`
	UpdatedAt            time.Time `json:"updated_at" binding:"omitempty"`
}

type GetNotification struct {
	UserID    string `form:"user_id" binding:"omitempty"`
	AdminID   string `form:"admin_id" binding:"omitempty"`
	ShopID    string `form:"shop_id" binding:"omitempty"`
	Status    string `form:"status" binding:"omitempty"`
	ProductID string `form:"product_id" binding:"omitempty"`
	OrderID   string `form:"order_id" binding:"omitempty"`
	IsRead    *bool  `form:"is_read" binding:"omitempty"`
}

type DeviceToken struct {
	ID        string     `gorm:"primaryKey;type:varchar(32)"`
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

// SendBroadcastRequest triggers a topic broadcast to a whole audience (e.g. all
// customer-app devices) in a single FCM call — no per-user token lookup.
type SendBroadcastRequest struct {
	Title string `json:"title" binding:"required,min=1,max=100"`
	Body  string `json:"body" binding:"required,min=1,max=500"`
	// Audience selects the topic. Optional — defaults to "all_users" (all
	// customer-app devices). Allowed: "all_users" | "all_sellers".
	Audience string `json:"audience" binding:"omitempty,oneof=all_users all_sellers"`
	// Data is an optional map of key-value pairs delivered alongside the notification.
	Data map[string]string `json:"data" binding:"omitempty"`
}

// RadiusAnnouncement pushes to every customer with a saved address within
// RadiusKm of a point the admin supplies. The shop-centred version is
// ShopLaunchAnnouncement below, which derives the point from the shop row.
type RadiusAnnouncement struct {
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
	RadiusKm  float64 `json:"radius_km" binding:"required,min=1,max=50"`
	Title     string  `json:"title" binding:"required,min=1,max=100"`
	Body      string  `json:"body" binding:"required,min=1,max=500"`
	// ImageURL is optional; a stored object key or uploads/ path is absolutised
	// server-side, since FCM silently drops an image it cannot fetch.
	ImageURL string `json:"image_url" binding:"omitempty"`
}

// ShopLaunchAnnouncement announces a newly opened shop to every customer with a
// saved address within RadiusKm of the shop's pinned location.
//
// Unlike NotificationRadiusRequest (raw coordinates), the centre point is taken
// from the shop row itself, so the admin never re-types a lat/long that could
// drift from the shop's real position.
type ShopLaunchAnnouncement struct {
	ShopID string `json:"shop_id" binding:"required"`
	// RadiusKm is capped at 10: past that the "shop near you" premise stops
	// being true and the push reads as spam.
	RadiusKm float64 `json:"radius_km" binding:"required,min=1,max=10"`
	Title    string  `json:"title" binding:"required,min=1,max=100"`
	Body     string  `json:"body" binding:"required,min=1,max=500"`
	// ImageURL overrides the shop's own photo. Empty means "use the shop image",
	// which is the normal case.
	ImageURL string `json:"image_url" binding:"omitempty"`
}
