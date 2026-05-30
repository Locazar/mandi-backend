package domain

import (
	"time"
)

// represent a model of product
type Product struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	Name         string    `json:"product_name" gorm:"not null" binding:"required,min=3,max=50"`
	Description  string    `json:"description" gorm:"not null" binding:"required,min=10,max=100"`
	CategoryID   string    `json:"category_id" gorm:"type:varchar(32)" binding:"omitempty"`
	DepartmentID string    `json:"department_id" gorm:"type:varchar(32)" binding:"omitempty"`
	Image        string    `json:"image" gorm:"not null"`
	ShopID       string    `json:"shop_id" gorm:"type:varchar(32);not null"`
	CreatedBy    string    `json:"created_by" gorm:"type:varchar(32)"`
	UpdatedBy    string    `json:"updated_by" gorm:"type:varchar(32)"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// this for a specific variant of product
type ProductItem struct {
	ID                string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	SubCategoryName   string    `json:"sub_category_name" gorm:"not null" binding:"required"`
	SubCategoryID     string    `json:"sub_category_id" gorm:"type:varchar(32)" binding:"omitempty"`
	CategoryID        string    `json:"category_id" gorm:"type:varchar(32)" binding:"omitempty"`
	DepartmentID      string    `json:"department_id" gorm:"type:varchar(32)" binding:"omitempty"`
	DynamicFields     string    `json:"dynamic_fields" gorm:"type:jsonb;not null"`
	AdminID           string    `json:"admin_id" gorm:"type:varchar(32);index"`
	ProductItemImages []string  `json:"product_item_images" gorm:"type:text[]"`
	ShopID            string    `json:"shop_id" gorm:"type:varchar(32)"`
	CreatedBy         string    `json:"created_by" gorm:"type:varchar(32)"`
	UpdatedBy         string    `json:"updated_by" gorm:"type:varchar(32)"`
	CreatedAt         time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Stock             bool      `json:"stock" gorm:"not null;default:true"`
}

type ProductItemImage struct {
	ID            string      `json:"id" gorm:"primaryKey;type:varchar(32)"`
	ProductItemID string      `json:"product_item_id" gorm:"type:varchar(32);not null" binding:"required"`
	ProductItem   ProductItem `json:"-"`
	ImageURL      []string    `json:"image_urls" gorm:"type:text[]"`
	AltText       string      `json:"alt_text" gorm:"size:255" binding:"omitempty"`
	SortOrder     int         `json:"sort_order" gorm:"not null;default:0"`
	IsActive      bool        `json:"is_active" gorm:"not null;default:true"`
	ShopID        string      `json:"shop_id" gorm:"type:varchar(32);not null" binding:"required"`
	CreatedAt     time.Time   `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt     time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

type Department struct {
	ID         string `json:"id" gorm:"primaryKey;type:varchar(32)"`
	Name       string `json:"department_name" gorm:"unique;not null" binding:"required,min=1,max=50"`
	SortOrder int    `json:"sort_order" gorm:"not null;default:0"`
	IsActive  bool   `json:"is_active" gorm:"not null;default:true"`
	ImageUrl   string `json:"image_url"`
	Icon       string `json:"icon" gorm:"size:255"`
}

// for a products category main and sub category as self joining
type Category struct {
	ID           string `json:"-" gorm:"primaryKey;type:varchar(32)"`
	DepartmentID string `json:"department_id" gorm:"type:varchar(32);not null" binding:"required"`
	Name         string `json:"category_name" gorm:"not null" binding:"required,min=1,max=30"`
	SortOrder   int    `json:"sort_order" gorm:"not null;default:0"`
	IsActive    bool   `json:"is_active" gorm:"not null;default:true"`
	ImageUrl     string `json:"image_url"`
	Icon         string `json:"icon" gorm:"size:255"`
}

type SubCategory struct {
	ID           string `json:"-" gorm:"primaryKey;type:varchar(32)"`
	DepartmentID string `json:"department_id" gorm:"type:varchar(32);not null" binding:"required"`
	CategoryID   string `json:"category_id" gorm:"type:varchar(32);not null" binding:"required"`
	Name         string `json:"sub_category_name" gorm:"not null" binding:"required,min=1,max=30"`
	SortOrder   int    `json:"sort_order" gorm:"not null;default:0"`
	IsActive    bool   `json:"is_active" gorm:"not null;default:true"`
	ImageUrl     string `json:"image_url"`
	Icon         string `json:"icon" gorm:"size:255"`
}

type Brand struct {
	ID         string `json:"id" gorm:"primaryKey;type:varchar(32)"`
	Name       string `json:"brand_name" gorm:"unique;not null"`
	SortOrder int    `json:"sort_order" gorm:"not null;default:0"`
	IsActive  bool   `json:"is_active" gorm:"not null;default:true"`
}

// variation means size color etc..
type Variation struct {
	ID            string   `json:"-" gorm:"primaryKey;type:varchar(32)"`
	SubCategoryID string   `json:"sub_category_id" gorm:"type:varchar(32);not null" binding:"required"`
	SubCategory   Category `json:"-"`
	Name          string   `json:"variation_name" gorm:"not null" binding:"required"`
	SortOrder    int      `json:"sort_order" gorm:"not null;default:0"`
	IsActive     bool     `json:"is_active" gorm:"not null;default:true"`
}

// variation option means values are like s,m,xl for size and blue,white,black for Color
type VariationOption struct {
	ID          string    `json:"-" gorm:"primaryKey;type:varchar(32)"`
	VariationID string    `json:"variation_id" gorm:"type:varchar(32);not null" binding:"required"` // a specific field of variation like color/size
	Variation   Variation `json:"-"`
	Value       string    `json:"variation_value" gorm:"not null" binding:"required"` // the variations value like blue/XL
	SortOrder  int       `json:"sort_order" gorm:"not null;default:0"`
	IsActive   bool      `json:"is_active" gorm:"not null;default:true"`
}

type ProductConfiguration struct {
	ProductItemID     string          `json:"product_item_id" gorm:"type:varchar(32);not null"`
	ProductItem       ProductItem     `json:"-"`
	VariationOptionID string          `json:"variation_option_id" gorm:"type:varchar(32);not null"`
	VariationOption   VariationOption `json:"-"`
	SortOrder        int             `json:"sort_order" gorm:"not null;default:0"`
	IsActive         bool            `json:"is_active" gorm:"not null;default:true"`
}

// to store a url of productItem Id along a unique url
// so we can ote multiple images url for a ProductItem
// one to many connection
type ProductImage struct {
	ID            string      `json:"id" gorm:"primaryKey;type:varchar(32)"`
	ProductItemID string      `json:"product_item_id" gorm:"type:varchar(32);not null"`
	ProductItem   ProductItem `json:"-"`
	ImageURL      []string    `json:"image_url" gorm:"type:text[]"`
	ShopID        string      `json:"shop_id" gorm:"type:varchar(32);not null"`
	ProductID     string      `json:"product_id" gorm:"type:varchar(32);not null"`
	AltText       string      `json:"alt_text" gorm:"size:255" binding:"omitempty"`
	SortOrder     int         `json:"sort_order" gorm:"not null;default:0"`
	IsActive      bool        `json:"is_active" gorm:"not null;default:true"`
	CreatedAt     time.Time   `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt     time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// offer
type Offer struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(32)" swaggerignore:"true"`
	Name         string    `json:"offer_name" gorm:"not null" binding:"required"`
	Description  string    `json:"description" gorm:"not null" binding:"required,min=6,max=50"`
	DiscountRate uint      `json:"discount_rate" gorm:"not null" binding:"required,numeric,min=1,max=100"`
	OfferType    OfferType `json:"offer_type" gorm:"not null" binding:"required"` // percentage,fixed
	StartDate    time.Time `json:"start_date" gorm:"not null" binding:"required"`
	EndDate      time.Time `json:"end_date" gorm:"not null" binding:"required"`
	Image        string    `json:"image_url" gorm:"not null" binding:"required"`
	Thumbnail    string    `json:"thumbnail_url" binding:"required"`
	CreatedBy   string    `json:"created_by" gorm:"type:varchar(32)"`
	UpdatedBy   string    `json:"updated_by" gorm:"type:varchar(32)"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	SortOrder   int       `json:"sort_order" gorm:"not null;default:0"`
	IsActive    bool      `json:"is_active" gorm:"not null;default:true"`
}

type OfferCategory struct {
	ID         string   `json:"id" gorm:"primaryKey;type:varchar(32)"`
	OfferID    string   `json:"offer_id" gorm:"type:varchar(32);not null"`
	Offer      Offer    `json:"-"`
	CategoryID string   `json:"category_id" gorm:"type:varchar(32);not null"`
	Category   Category `json:"-"`
	SortOrder int      `json:"sort_order" gorm:"not null;default:0"`
	IsActive  bool     `json:"is_active" gorm:"not null;default:true"`
}

type OfferProduct struct {
	ID            string `json:"id" gorm:"primaryKey;type:varchar(32)"`
	OfferID       string `json:"offer_id" gorm:"type:varchar(32);not null"`
	ProductItemID string `json:"product_item_id" gorm:"type:varchar(32);not null"`
	SortOrder    int  `json:"sort_order" gorm:"not null;default:0"`
	IsActive     bool `json:"is_active" gorm:"not null;default:true"`
}

type ProductItemView struct {
	ID            string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	ProductItemID string    `json:"product_item_id" gorm:"type:varchar(32);not null"`
	ShopID        string    `json:"shop_id" gorm:"type:varchar(32);not null"`
	AdminID       string    `json:"admin_id" gorm:"type:varchar(32);index"`
	ViewedAt      time.Time `json:"viewed_at" gorm:"not null;autoCreateTime"`
	ViewCount     int       `json:"view_count" gorm:"not null;default:1"`
	CreatedAt     time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ProductItemFilterType struct {
	ID         string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	FilterName string    `json:"filter_name" gorm:"not null"`
	ShopID     string    `json:"shop_id" gorm:"type:varchar(32)"`
	CreatedAt  time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type PromotionsType struct {
	ID                    string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	Name                  string    `json:"name"`
	IsActive             bool      `json:"is_active"`
	ShopID                string    `json:"shop_id" gorm:"type:varchar(32)"`
	PromotionCategoryID   string    `json:"promotion_category_id" gorm:"type:varchar(32)"`
	PromotionOfferDetails string    `json:"offer_details" gorm:"type:jsonb;not null"`
	IconPath             string    `json:"icon_path"`
	CreatedAt             time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt             time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type PromotionCategory struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	Name      string    `json:"name" gorm:"not null"`
	ShopID    string    `json:"shop_id" gorm:"type:varchar(32)"`
	IsActive bool      `json:"is_active" gorm:"not null;default:true"`
	IconPath string    `json:"icon_path"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type PromotionOfferDetails struct {
	OfferName              string    `json:"offer_name"`
	Description            string    `json:"description"`
	DiscountRate           float64   `json:"discount_rate"`
	StartDate              time.Time `json:"start_date"`
	EndDate                time.Time `json:"end_date"`
	MinimumPurchaseAmount  *float64  `json:"minimum_purchase_amount,omitempty"`
	TierQuantity           *int      `json:"tier_quantity,omitempty"`
	BogoGetQuantity        *int      `json:"bogo_get_quantity,omitempty"`
	BogoBuyQuantity        *int      `json:"bogo_buy_quantity,omitempty"`
	BogoCombinationEnabled *bool     `json:"bogo_combination_enabled,omitempty"`
	GiftDescription        *string   `json:"gift_description,omitempty"`
}

type Promotion struct {
	ID                     string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	PromotionCategoryID    string    `json:"promotion_category_id" gorm:"type:varchar(32)"`
	PromotionTypeID        string    `json:"promotion_type_id" gorm:"type:varchar(32)"`
	OfferName              string    `json:"offer_name"`
	Description            string    `json:"description"`
	DiscountRate           float64   `json:"discount_rate"`
	StartDate              time.Time `json:"start_date"`
	EndDate                time.Time `json:"end_date"`
	MinimumPurchaseAmount  *float64  `json:"minimum_purchase_amount,omitempty"`
	TierQuantity           *int      `json:"tier_quantity,omitempty"`
	BogoGetQuantity        *int      `json:"bogo_get_quantity,omitempty"`
	BogoBuyQuantity        *int      `json:"bogo_buy_quantity,omitempty"`
	BogoCombinationEnabled *bool     `json:"bogo_combination_enabled,omitempty"`
	GiftDescription        *string   `json:"gift_description,omitempty"`
	ShopID                 string    `json:"shop_id" gorm:"type:varchar(32)"`
	IsActive               bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt              time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt              time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	PromotionCategory PromotionCategory `json:"promotion_category" gorm:"foreignKey:PromotionCategoryID"`
	PromotionType     PromotionsType    `json:"promotion_type" gorm:"foreignKey:PromotionTypeID"`
}
