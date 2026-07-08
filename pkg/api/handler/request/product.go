package request

import "mime/multipart"

// for a new product
type Product struct {
	Name            string `json:"product_name" binding:"required,min=3,max=50"`
	Description     string `json:"description" binding:"required,min=10,max=100"`
	CategoryID      string `json:"category_id" binding:"required"`
	Department      string `json:"department" binding:"required"`
	DepartmentID    string `json:"department_id" binding:"required"`
	ImageFileHeader *multipart.FileHeader
}
type UpdateProduct struct {
	ID          string `json:"product_id" binding:"required"`
	Name        string `json:"product_name" binding:"required,min=3,max=50"`
	Description string `json:"description" binding:"required,min=10,max=100"`
	CategoryID  string `json:"category_id" binding:"required"`
	BrandID     string `json:"brand_id" binding:"required"`
	Price       uint   `json:"price" binding:"required,numeric"`
	Image       string `json:"image"`
}

// for a new productItem
type ProductItem struct {
	SubCategoryID     string                 `json:"sub_category_id" binding:"required"`
	SubCategoryName   string                 `json:"sub_category_name" binding:"required"`
	DynamicFields     map[string]interface{} `json:"dynamic_fields" binding:"required"`
	ProductItemImages []string               `json:"product_item_images " binding:"omitempty,dive,required"`
	DepartmentID      string                 `json:"department_id" binding:"required"`
	CategoryID        string                 `json:"category_id" binding:"required"`
	Description       string                 `json:"description"`
	Highlights        []string               `json:"highlights"`
	RetainedImages    []string               `json:"retained_images"`
}

// UpdateProductItemStock toggles the in-stock / out-of-stock status of a product item.
type UpdateProductItemStock struct {
	Stock *bool `json:"stock" binding:"required"`
}

type Variation struct {
	Names []string `json:"variation_names" binding:"required,dive,min=1"`
}

type VariationOption struct {
	Values []string `json:"variation_value" binding:"required,dive,min=1"`
}

type Category struct {
	Name         string `json:"category_name" binding:"required"`
	DepartmentId string `json:"department_id" binding:"required"`
	SortOrder    int    `json:"sort_order"`
	IsActive     bool   `json:"is_active"`
	ImageURL     string `json:"image_url" binding:"required"`
	Icon         string `json:"icon"`
}

type SubCategory struct {
	Name         string `json:"sub_category_name" binding:"required"`
	DepartmentID string `json:"department_id" binding:"required"`
	CategoryID   string `json:"category_id" binding:"required"`
	SortOrder    int    `json:"sort_order"`
	IsActive     bool   `json:"is_active"`
	ImageURL     string `json:"image_url" binding:"required"`
}

type Brand struct {
	Name string `json:"brand_name" binding:"required,min=3,max=25"`
}

type Department struct {
	Name      string `json:"department_name" binding:"required,min=3,max=25"`
	Slug      string `json:"slug" binding:"required,min=3,max=25"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
	ImageURL  string `json:"image_url" binding:"required"`
	Icon      string `json:"icon"`
}

type SubTypeAttribute struct {
	FieldName  string `json:"field_name" binding:"required,min=2,max=50"`
	FieldType  string `json:"field_type" binding:"required,oneof=dropdown number text"`
	IsRequired bool   `json:"is_required"`
	SortOrder  int    `json:"sort_order"`
}

type SubTypeAttributeOption struct {
	OptionValue string `json:"option_value" binding:"required,min=1,max=50"`
	SortOrder   int    `json:"sort_order"`
}

type CategoryImage struct {
	ImageURL  string `json:"image_url" binding:"required"`
	AltText   string `json:"alt_text" binding:"omitempty"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}
