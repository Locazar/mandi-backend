package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// response for product
type Product struct {
	ID               uint           `json:"product_id"`
	CategoryID       uint           `json:"category_id"`
	Price            uint           `json:"price"`
	DiscountPrice    uint           `json:"discount_price"`
	Name             string         `json:"product_name"`
	Description      string         `json:"description" `
	CategoryName     string         `json:"category_name"`
	CategoryImageURL string         `json:"category_image_url"`
	MainCategoryName string         `json:"main_category_name"`
	BrandID          uint           `json:"brand_id"`
	BrandName        string         `json:"brand_name"`
	Image            string         `json:"image"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	LocationID       *uuid.UUID     `json:"location_id,omitempty"`
	Stock            int            `json:"stock"`
	ProductItems     []ProductItems `json:"product_items"`
}

// for a specific category representation
type Category struct {
	ID           uint   `json:"category_id"`
	Name         string `json:"category_name"`
	DepartmentID uint   `json:"department_id"`
	ImageUrl     string `json:"image_url"`
}

type SubCategory struct {
	ID           uint   `json:"sub_category_id"`
	Name         string `json:"category_name"`
	DepartmentID uint   `json:"department_id"`
	CategoryID   uint   `json:"parent_category_id"`
	ImageUrl     string `json:"image_url"`
}

type SubTypeAttribute struct {
	ID            uint   `json:"id"`
	SubCategoryID uint   `json:"sub_category_id"`
	FieldName     string `json:"field_name"`
	FieldType     string `json:"field_type"`
	IsRequired    bool   `json:"is_required"`
	SortOrder     int    `json:"sort_order"`
	ImageUrl      string `json:"image_url"`
}

type SubTypeAttributeOption struct {
	ID                 uint   `json:"id"`
	SubTypeAttributeID uint   `json:"sub_type_attribute_id"`
	OptionValue        string `json:"option_value"`
	SortOrder          int    `json:"sort_order"`
}

type CategoryImage struct {
	ID         uint   `json:"id"`
	CategoryID uint   `json:"category_id"`
	ImageURL   string `json:"image_url"`
	AltText    string `json:"alt_text"`
	SortOrder  int    `json:"sort_order"`
	IsActive   bool   `json:"is_active"`
}

// for a specific variation representation
type Variation struct {
	ID               uint              `json:"variation_id"`
	Name             string            `json:"variation_name"`
	VariationOptions []VariationOption `gorm:"-"`
}

// for a specific variation Value representation
type VariationOption struct {
	ID    uint   `json:"variation_option_id"`
	Value string `json:"variation_value"`
}

// for response a specific products all product items
type ProductItems struct {
	ID                  uint                   `json:"product_item_id"`
	Name                string                 `json:"product_name"`
	CategoryID          uint                   `json:"category_id"`
	DepartmentID        uint                   `json:"department_id"`
	SubCategoryID       uint                   `json:"sub_category_id"`
	CategoryName        string                 `json:"category_name"`
	MainCategoryName    string                 `json:"main_category_name"`
	SubCategoryImageURL string                 `json:"sub_category_image_url"`
	ProductItemImages   []string               `json:"product_item_images"`
	DynamicFields       map[string]interface{} `json:"dynamic_fields"`
	OfferProducts       []OfferProduct         `json:"offer	_products"`
	DiscountRate        *uint                  `json:"discount_rate,omitempty"`
	ShopID              uint                   `json:"shop_id"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	ViewCount           uint                   `json:"view_count"`
}

type ProductVariationValue struct {
	VariationID       uint   `json:"variation_id"`
	Name              string `json:"variation_name"`
	VariationOptionID uint   `json:"variation_option_id"`
	Value             string `json:"variation_value"`
}

// offer response
type OfferCategory struct {
	OfferCategoryID uint   `json:"offer_category_id"`
	CategoryID      uint   `json:"category_id"`
	CategoryName    string `json:"category_name"`
	DiscountRate    uint   `json:"discount_rate"`
	OfferID         uint   `json:"offer_id"`
	OfferName       string `json:"offer_name"`
}

type OfferProduct struct {
	OfferProductID    uint              `json:"offer_product_id"`
	ProductName       string            `json:"product_name"`
	OfferID           uint              `json:"offer_id"`
	OfferName         string            `json:"offer_name"`
	DiscountRate      uint              `json:"discount_rate"`
	Description       string            `json:"description"`
	StartDate         string            `json:"start_date"`
	EndDate           string            `json:"end_date"`
	Image             string            `json:"image"`
	Thumbnail         string            `json:"thumbnail"`
	PromotionCategory PromotionCategory `json:"promotion_category"`
	PromotionType     PromotionsType    `json:"promotion_type"`
}

type Offer struct {
	ID           uint      `json:"offer_id"`
	Name         string    `json:"offer_name"`
	Description  string    `json:"description"`
	DiscountRate uint      `json:"discount_rate"`
	OfferType    string    `json:"offer_type"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Image        string    `json:"image_url"`
	Thumbnail    string    `json:"thumbnail_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProductFilters struct {
	Brands     []BrandFilter    `json:"brands"`
	Locations  []LocationFilter `json:"locations"`
	Prices     PriceFilter      `json:"prices"`
	Categories []CategoryFilter `json:"categories"`
}

type BrandFilter struct {
	BrandID   uuid.UUID `json:"brand_id"` // UUID type
	BrandName string    `json:"brand_name"`
	Count     int       `json:"count"`
}

type LocationFilter struct {
	LocationID   uint   `json:"location_id"`
	LocationName string `json:"location_name"`
	Count        uint   `json:"count"`
}

type PriceFilter struct {
	MinPrice uint `json:"min_price"`
	MaxPrice uint `json:"max_price"`
}

type Location struct {
	LocationID uuid.UUID `json:"location_id"`
	Name       string    `json:"name"`
	Country    string    `json:"country"`
	State      string    `json:"state"`
	City       string    `json:"city"`
	ZipCode    string    `json:"zip_code"`
	Area       string    `json:"area"`
	Pincode    string    `json:"pincode"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
}

type Brand struct {
	BrandID     uuid.UUID `json:"brand_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
}

type CategoryFilter struct {
	CategoryID   uint   `json:"category_id"`
	CategoryName string `json:"category_name"`
	Count        uint   `json:"count"`
}

type City struct {
	ID   uint   `json:"city_id"`
	Name string `json:"city_name"`
}

type State struct {
	ID   uint   `json:"state_id"`
	Name string `json:"state_name"`
}

type Country struct {
	ID   uint   `json:"country_id"`
	Name string `json:"country_name"`
}

type Pincode struct {
	ID      uint   `json:"pincode_id"`
	Pincode string `json:"pincode"`
}

type Area struct {
	ID   uint   `json:"area_id"`
	Name string `json:"area_name"`
}

type Department struct {
	ID       uint   `json:"department_id"`
	Name     string `json:"department_name"`
	ImageUrl string `json:"image_url"`
}

type PromotionCategory struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	ShopID    uint      `json:"shop_id"`
	IsActive  bool      `json:"is_active"`
	IconPath  string    `json:"icon_path"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PromotionsType struct {
	ID                    uint      `json:"id"`
	Name                  string    `json:"name"`
	IsActive              bool      `json:"is_active"`
	ShopID                string    `json:"shop_id"`
	PromotionCategoryID   uint      `json:"promotion_category_id"`
	CategoryName          string    `json:"category_name"`
	PromotionOfferDetails string    `json:"offer_details"`
	Type                  string    `json:"type"`
	IconPath              string    `json:"icon_path"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type Promotion struct {
	ID                     uint              `json:"id"`
	PromotionCategoryID    uint              `json:"promotion_category_id"`
	PromotionTypeID        uint              `json:"promotion_type_id"`
	OfferName              string            `json:"offer_name"`
	Description            string            `json:"description"`
	DiscountRate           float64           `json:"discount_rate"`
	StartDate              string            `json:"start_date"`
	EndDate                string            `json:"end_date"`
	MinimumPurchaseAmount  *float64          `json:"minimum_purchase_amount,omitempty"`
	TierQuantity           *int              `json:"tier_quantity,omitempty"`
	BogoGetQuantity        *int              `json:"bogo_get_quantity,omitempty"`
	BogoBuyQuantity        *int              `json:"bogo_buy_quantity,omitempty"`
	BogoCombinationEnabled *bool             `json:"bogo_combination_enabled,omitempty"`
	GiftDescription        *string           `json:"gift_description,omitempty"`
	ShopID                 uint              `json:"shop_id"`
	IsActive               bool              `json:"is_active"`
	CreatedAt              time.Time         `json:"created_at"`
	UpdatedAt              time.Time         `json:"updated_at"`
	PromotionCategory      PromotionCategory `json:"promotion_category" gorm:"foreignKey:PromotionCategoryID"`
	PromotionType          PromotionsType    `json:"type" gorm:"foreignKey:PromotionTypeID"`
	IconPath               string            `json:"icon_path"`
	Thumbnail              string            `json:"thumbnail"`
}

type OffersAndPromotions struct {
	Offers     []domain.Offer `json:"offers"`
	Promotions []Promotion    `json:"promotions"`
}
