package domain

import "time"

// ShopUpdate is a curated per-shop "What's New" advertisement: it features a
// registered shop (shop_details) plus a few of that shop's catalog products the
// seller wants customers to see. Shown to customers filtered by proximity.
type ShopUpdate struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	ShopID      string    `json:"shop_id" gorm:"type:varchar(32);not null;index"`
	ImageURL    string    `json:"image_url" gorm:"type:text"` // optional per-entry override; else shop logo
	ActionLabel string    `json:"action_label" gorm:"size:50;not null;default:'Visit'"`
	IsActive    bool      `json:"is_active" gorm:"not null;default:true"`
	SortOrder   int       `json:"sort_order" gorm:"not null;default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Products is populated for admin reads (not persisted on this table).
	Products []ShopUpdateProduct `json:"products,omitempty" gorm:"-"`
}

// ShopUpdateProduct is one product showcased under a ShopUpdate. It links a real
// catalog product (ProductItemID) and may override the display title/image; the
// customer read derives missing title/image from product_items.
type ShopUpdateProduct struct {
	ID            string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	ShopUpdateID  string    `json:"shop_update_id" gorm:"type:varchar(32);not null;index"`
	ProductItemID string    `json:"product_item_id" gorm:"type:varchar(32)"`
	Title         string    `json:"title" gorm:"size:150"`
	ImageURL      string    `json:"image_url" gorm:"type:text"`
	Attribute     string    `json:"attribute" gorm:"size:100"`
	SortOrder     int       `json:"sort_order" gorm:"not null;default:0"`
	IsActive      bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
