package domain

import (
	"time"
)

// represent a model of product
type Product struct {
	ID           uint      `json:"id" gorm:"primaryKey;not null"`
	Name         string    `json:"product_name" gorm:"not null" binding:"required,min=3,max=50"`
	Description  string    `json:"description" gorm:"not null" binding:"required,min=10,max=100"`
	CategoryID   uint      `json:"category_id" binding:"omitempty,numeric"`
	DepartmentID uint      `json:"department_id" binding:"omitempty,numeric"`
	Image        string    `json:"image" gorm:"not null"`
	ShopID       uint      `json:"shop_id" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// this for a specific variant of product
type ProductItem struct {
	ID                uint      `json:"id" gorm:"primaryKey;not null"`
	SubCategoryName   string    `json:"sub_category_name" gorm:"not null" binding:"required"`
	SubCategoryID     uint      `json:"sub_category_id" binding:"omitempty,numeric"`
	CategoryID        uint      `json:"category_id" binding:"omitempty,numeric"`
	DepartmentID      uint      `json:"department_id" binding:"omitempty,numeric"`
	DynamicFields     string    `json:"dynamic_fields" gorm:"type:jsonb;not null"`
	AdminID           string    `json:"admin_id" gorm:"type:jsonb;not null"` // stored as JSONB in DB
	ProductItemImages []string  `json:"product_item_images" gorm:"type:text[]"`
	CreatedAt         time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ProductItemImage struct {
	ID            uint        `json:"id" gorm:"primaryKey;not null"`
	ProductItemID uint        `json:"product_item_id" binding:"required,numeric" gorm:"not null"`
	ProductItem   ProductItem `json:"-"`
	ImageURL      []string    `json:"image_urls" gorm:"type:text[]"`
	AltText       string      `json:"alt_text" gorm:"size:255" binding:"omitempty"`
	SortOrder     int         `json:"sort_order" gorm:"not null;default:0"`
	IsActive      bool        `json:"is_active" gorm:"not null;default:true"`
	ShopID        uint        `json:"shop_id" gorm:"not null" binding:"required,numeric"`
	CreatedAt     time.Time   `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt     time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

type Department struct {
	ID         uint   `json:"id" gorm:"primaryKey;not null"`
	Name       string `json:"department_name" gorm:"unique;not null" binding:"required,min=1,max=50"`
	Sort_Order int    `json:"sort_order" gorm:"not null;default:0"`
	Is_Active  bool   `json:"is_active" gorm:"not null;default:true"`
	ImageUrl   string `json:"image_url"`
}

// for a products category main and sub category as self joining
type Category struct {
	ID           uint   `json:"-" gorm:"primaryKey;not null"`
	DepartmentID uint   `json:"department_id" gorm:"not null" binding:"required,numeric"`
	Name         string `json:"category_name" gorm:"not null" binding:"required,min=1,max=30"`
	Sort_Order   int    `json:"sort_order" gorm:"not null;default:0"`
	Is_Active    bool   `json:"is_active" gorm:"not null;default:true"`
	ImageUrl     string `json:"image_url"`
}

type SubCategory struct {
	ID           uint   `json:"-" gorm:"primaryKey;not null"`
	DepartmentID uint   `json:"department_id" gorm:"not null" binding:"required,numeric"`
	CategoryID   uint   `json:"category_id" gorm:"not null" binding:"required,numeric"`
	Name         string `json:"sub_category_name" gorm:"not null" binding:"required,min=1,max=30"`
	Sort_Order   int    `json:"sort_order" gorm:"not null;default:0"`
	Is_Active    bool   `json:"is_active" gorm:"not null;default:true"`
	ImageUrl     string `json:"image_url"`
}

type Brand struct {
	ID         uint   `json:"id" gorm:"primaryKey;not null"`
	Name       string `json:"brand_name" gorm:"unique;not null"`
	Sort_Order int    `json:"sort_order" gorm:"not null;default:0"`
	Is_Active  bool   `json:"is_active" gorm:"not null;default:true"`
}

// variation means size color etc..
type Variation struct {
	ID            uint     `json:"-" gorm:"primaryKey;not null"`
	SubCategoryID uint     `json:"sub_category_id" gorm:"not null" binding:"required,numeric"`
	SubCategory   Category `json:"-"`
	Name          string   `json:"variation_name" gorm:"not null" binding:"required"`
	Sort_Order    int      `json:"sort_order" gorm:"not null;default:0"`
	Is_Active     bool     `json:"is_active" gorm:"not null;default:true"`
}

// variation option means values are like s,m,xl for size and blue,white,black for Color
type VariationOption struct {
	ID          uint      `json:"-" gorm:"primaryKey;not null"`
	VariationID uint      `json:"variation_id" gorm:"not null" binding:"required,numeric"` // a specific field of variation like color/size
	Variation   Variation `json:"-"`
	Value       string    `json:"variation_value" gorm:"not null" binding:"required"` // the variations value like blue/XL
	Sort_Order  int       `json:"sort_order" gorm:"not null;default:0"`
	Is_Active   bool      `json:"is_active" gorm:"not null;default:true"`
}

type ProductConfiguration struct {
	ProductItemID     uint            `json:"product_item_id" gorm:"not null"`
	ProductItem       ProductItem     `json:"-"`
	VariationOptionID uint            `json:"variation_option_id" gorm:"not null"`
	VariationOption   VariationOption `json:"-"`
	Sort_Order        int             `json:"sort_order" gorm:"not null;default:0"`
	Is_Active         bool            `json:"is_active" gorm:"not null;default:true"`
}

// to store a url of productItem Id along a unique url
// so we can ote multiple images url for a ProductItem
// one to many connection
type ProductImage struct {
	ID            uint        `json:"id" gorm:"primaryKey;not null"`
	ProductItemID uint        `json:"product_item_id" gorm:"not null"`
	ProductItem   ProductItem `json:"-"`
	ImageURL      []string    `json:"image_url" gorm:"type:text[]"`
	ShopID        uint        `json:"shop_id" gorm:"not null"`
	ProductID     uint        `json:"product_id" gorm:"not null"`
	AltText       string      `json:"alt_text" gorm:"size:255" binding:"omitempty"`
	SortOrder     int         `json:"sort_order" gorm:"not null;default:0"`
	IsActive      bool        `json:"is_active" gorm:"not null;default:true"`
	CreatedAt     time.Time   `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt     time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// offer
type Offer struct {
	ID           uint      `json:"id" gorm:"primaryKey;not null" swaggerignore:"true"`
	Name         string    `json:"offer_name" gorm:"not null;unique" binding:"required"`
	Description  string    `json:"description" gorm:"not null" binding:"required,min=6,max=50"`
	DiscountRate uint      `json:"discount_rate" gorm:"not null" binding:"required,numeric,min=1,max=100"`
	OfferType    string    `json:"offer_type" gorm:"not null" binding:"required"` // percentage,fixed
	StartDate    time.Time `json:"start_date" gorm:"not null" binding:"required"`
	EndDate      time.Time `json:"end_date" gorm:"not null" binding:"required"`
	Image        string    `json:"image_url" gorm:"not null" binding:"required"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Sort_Order   int       `json:"sort_order" gorm:"not null;default:0"`
	Is_Active    bool      `json:"is_active" gorm:"not null;default:true"`
}

type OfferCategory struct {
	ID         uint     `json:"id" gorm:"primaryKey;not null"`
	OfferID    uint     `json:"offer_id" gorm:"not null"`
	Offer      Offer    `json:"-"`
	CategoryID uint     `json:"category_id" gorm:"not null"`
	Category   Category `json:"-"`
	Sort_Order int      `json:"sort_order" gorm:"not null;default:0"`
	Is_Active  bool     `json:"is_active" gorm:"not null;default:true"`
}

type OfferProduct struct {
	ID            uint `json:"id" gorm:"primaryKey;not null"`
	OfferID       uint `json:"offer_id" gorm:"not null"`
	Offer         Offer
	ProductItemID uint `json:"product_item_id" gorm:"not null"`
	Sort_Order    int  `json:"sort_order" gorm:"not null;default:0"`
	Is_Active     bool `json:"is_active" gorm:"not null;default:true"`
}

type SubCategoryDetails struct {
	ID                  uint   `json:"id"`
	SubCategoryID       uint   `json:"sub_category_id"`
	SubCategoryImageUrl string `json:"sub_category_image_url"`
}
