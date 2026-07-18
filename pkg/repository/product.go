package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/elasticsearch"
)

type productDatabase struct {
	DB            *gorm.DB
	ElasticClient *elasticsearch.ElasticService
}

// GetProductItemsByOfferID implements [interfaces.ProductRepository].
func (c *productDatabase) GetProductItemsByOfferID(ctx context.Context, offerID string, categoryID int, departmentID int, subCategoryID int, latStr string, lngStr string, pincode string, radiusKm float64, limit int, offset int) ([]response.ProductItems, error) {
	return GetProductItemsByOfferID(ctx, c.DB, offerID, categoryID, departmentID, subCategoryID, latStr, lngStr, pincode, radiusKm, limit, offset)
}

// DeleteProductItem deletes a product item and all its related data.
func (c *productDatabase) DeleteProductItem(ctx context.Context, productItemID string) error {
	// Delete all related records in cascade order to avoid foreign key constraints

	// 1. Delete from offer_products (offers/promotions linked to this product)
	if err := c.DB.Exec(`DELETE FROM offer_products WHERE product_item_id = $1`, productItemID).Error; err != nil {
		return fmt.Errorf("failed to delete offer_products: %w", err)
	}

	// 2. Delete from product_configurations (variation configurations)
	if err := c.DB.Exec(`DELETE FROM product_configurations WHERE product_item_id = $1`, productItemID).Error; err != nil {
		return fmt.Errorf("failed to delete product_configurations: %w", err)
	}

	// 3. Delete from product_item_views (view count records)
	if err := c.DB.Exec(`DELETE FROM product_item_views WHERE product_item_id = $1`, productItemID).Error; err != nil {
		return fmt.Errorf("failed to delete product_item_views: %w", err)
	}

	// 4. Delete from product_images (product images)
	if err := c.DB.Exec(`DELETE FROM product_images WHERE product_item_id = $1`, productItemID).Error; err != nil {
		return fmt.Errorf("failed to delete product_images: %w", err)
	}

	if err := c.DB.Exec(`DELETE FROM product_items WHERE id = $1`, productItemID).Error; err != nil {
		return fmt.Errorf("failed to delete product_item: %w", err)
	}

	// 5. Finally, delete the product_item itself
	return nil
}

func NewProductRepository(db *gorm.DB, elasticClient *elasticsearch.ElasticService) interfaces.ProductRepository {
	return &productDatabase{
		DB:            db,
		ElasticClient: elasticClient,
	}
}

func (c *productDatabase) Transactions(ctx context.Context, trxFn func(repo interfaces.ProductRepository) error) error {
	trx := c.DB.Begin()

	repo := NewProductRepository(trx, c.ElasticClient)

	if err := trxFn(repo); err != nil {
		trx.Rollback()
		return err
	}

	if err := trx.Commit().Error; err != nil {
		trx.Rollback()
		return err
	}
	return nil
}

// To check the category name exist
func (c *productDatabase) IsCategoryNameExist(ctx context.Context, name string, departmentId string) (exist bool, err error) {
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1 AND department_id = $2)`
	err = c.DB.Raw(query, name, departmentId).Scan(&exist).Error

	return
}

// Save Category
func (c *productDatabase) SaveCategory(ctx context.Context, category request.Category, departmentId string) (err error) {
	query := `INSERT INTO categories (name, department_id, sort_order, is_active, image_url, icon) VALUES ($1, $2, $3, $4, $5, $6)`
	err = c.DB.Exec(query, category.Name, departmentId, category.SortOrder, category.IsActive, category.ImageURL, category.Icon).Error

	return err
}

// To check the sub category name already exist for the category
func (c *productDatabase) IsSubCategoryNameExist(ctx context.Context, name string, departmentId string) (exist bool, err error) {
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1 AND department_id = $2)`
	err = c.DB.Raw(query, name, departmentId).Scan(&exist).Error

	return
}

// Save Category as sub category
func (c *productDatabase) SaveSubCategory(ctx context.Context, body request.SubCategory, brandID string, categoryID string) (err error) {
	print("department id in repo", brandID, "category id in repo", categoryID)
	query := `INSERT INTO sub_categories (department_id, category_id, name, sort_order,  is_active, image_url) VALUES ($1, $2, $3, $4, $5, $6)`
	err = c.DB.Exec(query, brandID, categoryID, body.Name, body.SortOrder, body.IsActive, body.ImageURL).Error

	return err
}

// Find all main category(its not have a category_id)
func (c *productDatabase) FindAllMainCategories(ctx context.Context,
	pagination request.Pagination,
) (categories []response.Category, err error) {
	limit := pagination.Limit
	offset := pagination.Offset

	query := `SELECT id, name, image_url, icon FROM categories 
	LIMIT $1 OFFSET $2`
	err = c.DB.Raw(query, limit, offset).Scan(&categories).Error

	return
}

// Find all sub categories of a category
func (c *productDatabase) FindAllSubCategories(ctx context.Context,
	categoryID string,
) (subCategories []response.SubCategory, err error) {
	query := `SELECT id, name FROM sub_categories WHERE category_id = $1`
	err = c.DB.Raw(query, categoryID).Scan(&subCategories).Error

	return
}

// Find all variations which related to given category id
func (c *productDatabase) FindAllVariationsByCategoryID(ctx context.Context,
	categoryID string,
) (variations []response.Variation, err error) {
	query := `SELECT id, name FROM variations WHERE category_id = $1`
	err = c.DB.Raw(query, categoryID).Scan(&variations).Error

	return
}

// Find all variation options which related to given variation id
func (c productDatabase) FindAllVariationOptionsByVariationID(ctx context.Context,
	variationID string,
) (variationOptions []response.VariationOption, err error) {
	query := `SELECT id, value FROM variation_options WHERE variation_id = $1`
	err = c.DB.Raw(query, variationID).Scan(&variationOptions).Error

	return
}

// To check a variation exist for the given category
func (c *productDatabase) IsVariationNameExistForCategory(ctx context.Context,
	name string, categoryID string,
) (exist bool, err error) {
	query := `SELECT EXISTS(SELECT 1 FROM variations WHERE name = $1 AND category_id = $2)`
	err = c.DB.Raw(query, name, categoryID).Scan(&exist).Error

	return
}

// To check a variation value exist for the given variation
func (c *productDatabase) IsVariationValueExistForVariation(ctx context.Context,
	value string, variationID string,
) (exist bool, err error) {
	query := `SELECT EXISTS(SELECT 1 FROM variation_options WHERE value = $1 AND variation_id = $2)`
	err = c.DB.Raw(query, value, variationID).Scan(&exist).Error

	return
}

// Save Variation for category
func (c *productDatabase) SaveVariation(ctx context.Context, categoryID string, variationName string) error {
	query := `INSERT INTO variations (category_id, name) VALUES($1, $2)`
	err := c.DB.Exec(query, categoryID, variationName).Error

	return err
}

// add variation option
func (c *productDatabase) SaveVariationOption(ctx context.Context, variationID string, variationValue string) error {
	query := `INSERT INTO variation_options (variation_id, value) VALUES($1, $2)`
	err := c.DB.Exec(query, variationID, variationValue).Error

	return err
}

// find product by id
func (c *productDatabase) FindProductByID(ctx context.Context, productID string) (product domain.Product, err error) {
	query := `SELECT * FROM products WHERE id = $1`
	err = c.DB.Raw(query, productID).Scan(&product).Error

	return
}

func (c *productDatabase) IsProductNameExistForOtherProduct(ctx context.Context,
	name string, productID string,
) (exist bool, err error) {
	query := `SELECT EXISTS(SELECT id FROM products WHERE name = $1 AND id != $2)`
	err = c.DB.Raw(query, name, productID).Scan(&exist).Error

	return
}

func (c *productDatabase) IsProductNameExist(ctx context.Context, productName string) (exist bool, err error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE name = $1)`
	err = c.DB.Raw(query, productName).Scan(&exist).Error

	return
}

// Check if product with same name exists for a specific shop
func (c *productDatabase) IsProductNameExistForShop(ctx context.Context, productName string, shopID *string) (exist bool, err error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE name = $1 AND shop_id = $2)`
	err = c.DB.Raw(query, productName, shopID).Scan(&exist).Error

	return
}

// to add a new product in database
func (c *productDatabase) SaveProduct(ctx context.Context, product domain.Product, adminID string) (productID string, err error) {
	// Get the shop Id and Shop name using adminID

	var shopDetails struct {
		ShopID   string `gorm:"column:id"`
		ShopName string `gorm:"column:shop_name"`
	}

	query := `SELECT id, shop_name FROM shop_details WHERE admin_id = $1`
	err = c.DB.Raw(query, adminID).Scan(&shopDetails).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("failed to fetch shop details for admin %s: %v", adminID, err)
	}

	// Check if product with shop_id and category_id already exists
	checkQuery := `SELECT id FROM products WHERE shop_id = $1 AND category_id = $2 LIMIT 1`
	err = c.DB.Raw(checkQuery, shopDetails.ShopID, product.CategoryID).Scan(&productID).Error

	if err == nil && productID != "" {
		return productID, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("failed to check if product exists: %v", err)
	}

	// Validate department_id exists
	if product.DepartmentID != "" {
		var deptExists bool
		deptQuery := `SELECT EXISTS(SELECT 1 FROM departments WHERE id = $1)`
		err = c.DB.Raw(deptQuery, product.DepartmentID).Scan(&deptExists).Error
		if err != nil {
			return "", fmt.Errorf("failed to check if department exists: %v", err)
		}
		if !deptExists {
			return "", fmt.Errorf("department with id %s does not exist", product.DepartmentID)
		}
	}

	// Insert new product
	productID = domain.NewID(domain.PrefixProduct)
	query = `INSERT INTO products (id, name, description, category_id, image, department_id, shop_id, created_at)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8)`

	createdAt := time.Now()
	err = c.DB.Exec(query, productID, product.Name, product.Description, product.CategoryID, product.Image, product.DepartmentID, shopDetails.ShopID, createdAt).Error
	if err != nil {
		return "", fmt.Errorf("failed to insert product: %v", err)
	}

	return productID, nil
}

// update product
func (c *productDatabase) UpdateProduct(ctx context.Context, product domain.Product) error {
	query := `UPDATE products SET name = $1, description = $2, category_id = $3, image = $4, updated_at = $5 
	WHERE id = $6`

	updatedAt := time.Now()

	err := c.DB.Exec(query, product.Name, product.Description, product.CategoryID,
		product.Image, updatedAt, product.ID).Error

	return err
}

// get all products from database
func (c *productDatabase) FindAllProducts(ctx context.Context, pagination request.Pagination, search string) (products []response.Product, err error) {
	limit := pagination.Limit
	offset := pagination.Offset

	query := `SELECT p.*, c.name AS category_name, c.image_url AS category_image_url
	FROM products p
	LEFT JOIN categories c ON p.category_id = c.id
	ORDER BY p.created_at DESC LIMIT $1 OFFSET $2`

	// Temporary struct for scanning (without ProductItems slice)
	type productDB struct {
		ID               string     `gorm:"column:id"`
		CategoryID       string     `gorm:"column:category_id"`
		Price            uint       `gorm:"column:price"`
		DiscountPrice    uint       `gorm:"column:discount_price"`
		Name             string     `gorm:"column:name"`
		Description      string     `gorm:"column:description"`
		CategoryName     string     `gorm:"column:category_name"`
		CategoryImageURL string     `gorm:"column:category_image_url"`
		MainCategoryName string     `gorm:"column:main_category_name"`
		BrandID          string     `gorm:"column:brand_id"`
		BrandName        string     `gorm:"column:brand_name"`
		Image            string     `gorm:"column:image"`
		CreatedAt        time.Time  `gorm:"column:created_at"`
		UpdatedAt        time.Time  `gorm:"column:updated_at"`
		LocationID       *uuid.UUID `gorm:"column:location_id"`
		Stock            int        `gorm:"column:stock"`
	}

	var dbProducts []productDB
	err = c.DB.Raw(query, limit, offset).Scan(&dbProducts).Error
	if err != nil {
		return nil, err
	}

	// Map to response.Product
	products = make([]response.Product, len(dbProducts))
	for i, dbProd := range dbProducts {
		products[i] = response.Product{
			ID:               dbProd.ID,
			CategoryID:       dbProd.CategoryID,
			Price:            dbProd.Price,
			DiscountPrice:    dbProd.DiscountPrice,
			Name:             dbProd.Name,
			Description:      dbProd.Description,
			CategoryName:     dbProd.CategoryName,
			CategoryImageURL: dbProd.CategoryImageURL,
			MainCategoryName: dbProd.MainCategoryName,
			BrandID:          dbProd.BrandID,
			BrandName:        dbProd.BrandName,
			Image:            dbProd.Image,
			CreatedAt:        dbProd.CreatedAt,
			UpdatedAt:        dbProd.UpdatedAt,
			LocationID:       dbProd.LocationID,
			Stock:            dbProd.Stock,
			ProductItems:     []response.ProductItems{}, // Initialize empty, will be populated below
		}
	}

	// Fetch product items for each product
	for i := range products {
		productItems, itemErr := c.findProductItemsByProductID(ctx, products[i].ID)
		if itemErr != nil {
			products[i].ProductItems = []response.ProductItems{}
		} else {
			products[i].ProductItems = productItems
		}
	}

	return
}

// helper method to get product items by product ID (internal use)
func (c *productDatabase) findProductItemsByProductID(ctx context.Context, productID string) (productItems []response.ProductItems, err error) {
	query := `SELECT pi.sub_category_name, pi.id, pi.category_id, 
		   sc.name AS category_name, mc.name AS main_category_name
	       FROM product_items pi 
	       LEFT JOIN categories sc ON pi.category_id = sc.id 
	       LEFT JOIN categories mc ON pi.category_id = mc.id 
	       WHERE pi.id = $1`

	type productItemDB struct {
		Name             string `gorm:"column:sub_category_name"`
		ID               string `gorm:"column:id"`
		CategoryID       string `gorm:"column:category_id"`
		CategoryName     string `gorm:"column:category_name"`
		MainCategoryName string `gorm:"column:main_category_name"`
	}
	var dbItems []productItemDB
	err = c.DB.Raw(query, productID).Scan(&dbItems).Error
	if err != nil {
		return
	}

	for _, dbItem := range dbItems {
		item := response.ProductItems{
			ID:               dbItem.ID,
			Name:             dbItem.Name,
			CategoryName:     dbItem.CategoryName,
			MainCategoryName: dbItem.MainCategoryName,
		}
		images, imgErr := c.FindAllProductItemImages(ctx, dbItem.ID)
		if imgErr != nil {
			item.ProductItemImages = []string{}
		} else {
			item.ProductItemImages = images
		}
		productItems = append(productItems, item)
	}
	return
}

// to get productItem id
func (c *productDatabase) FindProductItemByID(ctx context.Context, productItemID string) (productItem domain.ProductItem, err error) {
	type tempProductItem struct {
		ID                string    `gorm:"column:id"`
		SubCategoryName   string    `gorm:"column:sub_category_name"`
		SubCategoryID     string    `gorm:"column:sub_category_id"`
		CategoryID        string    `gorm:"column:category_id"`
		DepartmentID      string    `gorm:"column:department_id"`
		DynamicFields     string    `gorm:"column:dynamic_fields"`
		AdminID           string    `gorm:"column:admin_id"`
		ProductItemImages string    `gorm:"column:product_item_images"`
		Description       string    `gorm:"column:description"`
		Highlights        string    `gorm:"column:highlights"`
		ShopID            string    `gorm:"column:shop_id"`
		CreatedAt         time.Time `gorm:"column:created_at"`
		UpdatedAt         time.Time `gorm:"column:updated_at"`
	}

	var temp tempProductItem
	err = c.DB.WithContext(ctx).Table("product_items").Where("id = ?", productItemID).First(&temp).Error
	if err != nil {
		return productItem, err
	}

	productItem.ID = temp.ID
	productItem.SubCategoryName = temp.SubCategoryName
	productItem.SubCategoryID = temp.SubCategoryID
	productItem.CategoryID = temp.CategoryID
	productItem.DepartmentID = temp.DepartmentID
	productItem.DynamicFields = temp.DynamicFields
	productItem.AdminID = temp.AdminID
	productItem.ShopID = temp.ShopID
	productItem.Description = temp.Description
	productItem.CreatedAt = temp.CreatedAt
	productItem.UpdatedAt = temp.UpdatedAt

	// Parse product_item_images from PostgreSQL array format
	if temp.ProductItemImages != "" {
		imageStr := temp.ProductItemImages
		if len(imageStr) > 2 && imageStr[0] == '{' && imageStr[len(imageStr)-1] == '}' {
			imageStr = imageStr[1 : len(imageStr)-1]
			if imageStr != "" {
				productItem.ProductItemImages = strings.Split(imageStr, ",")
			}
		}
	}

	// Parse highlights PostgreSQL array format
	if temp.Highlights != "" {
		h := temp.Highlights
		if len(h) > 2 && h[0] == '{' && h[len(h)-1] == '}' {
			h = h[1 : len(h)-1]
		}
		if h != "" {
			for _, part := range strings.Split(h, ",") {
				productItem.Highlights = append(productItem.Highlights, strings.Trim(strings.TrimSpace(part), `"`))
			}
		}
	}

	return productItem, nil
}

// to get how many variations are available for a product
func (c *productDatabase) FindVariationCountForProduct(ctx context.Context, productID string) (variationCount uint, err error) {
	query := `SELECT COUNT(v.id) FROM variations v
	INNER JOIN categories c ON c.id = v.category_id 
	INNER JOIN products p ON p.category_id = v.category_id 
	WHERE p.id = $1`

	err = c.DB.Raw(query, productID).Scan(&variationCount).Error

	return
}

// To find all product item ids which related to the given product id and variation option id
func (c *productDatabase) FindAllProductItemIDsByProductIDAndVariationOptionID(ctx context.Context, productID,
	variationOptionID string,
) (productItemIDs []string, err error) {
	query := `SELECT id FROM product_items pi 
		INNER JOIN product_configurations pc ON pi.id = pc.product_item_id 
		WHERE pi.product_id = $1 AND variation_option_id = $2`
	err = c.DB.Raw(query, productID, variationOptionID).Scan(&productItemIDs).Error

	return
}

func (c *productDatabase) SaveProductConfiguration(ctx context.Context, productItemID, variationOptionID string) error {
	query := `INSERT INTO product_configurations (product_item_id, variation_option_id) VALUES ($1, $2)`
	err := c.DB.Exec(query, productItemID, variationOptionID).Error

	return err
}

// UpdateShopDepartments adds a department to the shop_departments table if not already present
// Uses atomic PostgreSQL INSERT with ON CONFLICT to avoid race conditions and duplicates
// Creates an entry in shop_departments table for tracking shop-department relationships
func (c *productDatabase) UpdateShopDepartments(ctx context.Context, shopID string, departmentID string, adminID string, categoryID string, subCategoryId string) error {
	if shopID == "" || departmentID == "" || categoryID == "" || subCategoryId == "" {
		return fmt.Errorf("shopID, departmentID, categoryID, and subCategoryId must not be empty")
	}

	// adminID and shopID are already string IDs

	// Insert into shop_departments with ON CONFLICT DO NOTHING to handle duplicates
	query := `INSERT INTO shop_departments (id, admin_id, shop_id, department_id, category_id, sub_category_id, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) ON CONFLICT (shop_id, department_id) DO NOTHING`

	result := c.DB.WithContext(ctx).Exec(query, domain.NewID(domain.PrefixShopDepartment), adminID, shopID, departmentID, categoryID, subCategoryId)
	if result.Error != nil {
		log.Printf("Failed to update shop departments: shopID=%s, departmentID=%s, adminID=%s, categoryID=%s, error=%v", shopID, departmentID, adminID, categoryID, result.Error)
		return fmt.Errorf("failed to update shop departments: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Printf("Department already exists for shop: shopID=%s, departmentID=%s, categoryID=%s", shopID, departmentID, categoryID)
	} else {
		log.Printf("Shop department record created: shopID=%s, departmentID=%s, adminID=%s, categoryID=%s", shopID, departmentID, adminID, categoryID)
	}

	return nil
}

// CountProductItemsByShopID returns how many product items already exist for a shop,
// used to enforce the seller's admins.product_limit quota before a new upload.
func (c *productDatabase) CountProductItemsByShopID(ctx context.Context, shopID string) (count int64, err error) {
	err = c.DB.WithContext(ctx).Model(&domain.ProductItem{}).Where("shop_id = ?", shopID).Count(&count).Error
	return
}

func (c *productDatabase) SaveProductItem(ctx context.Context, productItem request.ProductItem, adminID string, shopID string) (productItemID string, err error) {
	query := `INSERT INTO product_items (id, admin_id, sub_category_name, dynamic_fields, product_item_images, category_id, department_id, sub_category_id, shop_id, description, highlights, created_at, updated_at)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`

	createdAt := time.Now()
	newID := domain.NewID(domain.PrefixProductItem)

	// Marshal DynamicFields to JSON for JSONB column
	dynamicFieldsJSON, err := json.Marshal(productItem.DynamicFields)
	if err != nil {
		return "", err
	}

	var productItemImagesStr string
	if len(productItem.ProductItemImages) > 0 {
		productItemImagesStr = "{" + strings.Join(productItem.ProductItemImages, ",") + "}"
	} else {
		productItemImagesStr = "{}"
	}

	var highlightsStr string
	if len(productItem.Highlights) > 0 {
		escaped := make([]string, len(productItem.Highlights))
		for i, h := range productItem.Highlights {
			escaped[i] = `"` + strings.ReplaceAll(h, `"`, `\"`) + `"`
		}
		highlightsStr = "{" + strings.Join(escaped, ",") + "}"
	} else {
		highlightsStr = "{}"
	}

	err = c.DB.Exec(query, newID, adminID, productItem.SubCategoryName, dynamicFieldsJSON, productItemImagesStr, productItem.CategoryID, productItem.DepartmentID, productItem.SubCategoryID, shopID, productItem.Description, highlightsStr, createdAt, createdAt).Error
	if err == nil {
		productItemID = newID
	}

	if err == nil && c.ElasticClient != nil {
		domainItem := domain.ProductItem{
			ID:                productItemID,
			SubCategoryName:   productItem.SubCategoryName,
			CategoryID:        productItem.CategoryID,
			DepartmentID:      productItem.DepartmentID,
			SubCategoryID:     productItem.SubCategoryID,
			AdminID:           adminID,
			DynamicFields:     string(dynamicFieldsJSON),
			ProductItemImages: productItem.ProductItemImages,
			ShopID:            shopID,
		}
		go func() {
			if indexErr := c.ElasticClient.IndexProductItem(ctx, domainItem); indexErr != nil {
				log.Printf("Failed to index product item in Elasticsearch: %v", indexErr)
			}
		}()
	}

	// Update shop's departments array after successful product item save
	if err == nil && productItem.DepartmentID != "" {
		if updateDeptErr := c.UpdateShopDepartments(ctx, shopID, productItem.DepartmentID, adminID, productItem.CategoryID, productItem.SubCategoryID); updateDeptErr != nil {
			log.Printf("Warning: Failed to update shop departments during SaveProductItem: %v", updateDeptErr)
			// Don't return error here - product was saved successfully, department update is a side effect
		}
	}

	return productItemID, err
}

func (c *productDatabase) UpdateProductItem(ctx context.Context, productItemID string, productItem request.ProductItem) error {
	existing, err := c.FindProductItemByID(ctx, productItemID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing product item: %w", err)
	}

	// Merge dynamic_fields: existing as base, new values overwrite
	var existingDynamicFields map[string]interface{}
	if existing.DynamicFields != "" {
		if err := json.Unmarshal([]byte(existing.DynamicFields), &existingDynamicFields); err != nil {
			return fmt.Errorf("failed to unmarshal existing dynamic fields: %w", err)
		}
	}
	if existingDynamicFields == nil {
		existingDynamicFields = make(map[string]interface{})
	}
	if productItem.DynamicFields != nil {
		for k, v := range productItem.DynamicFields {
			existingDynamicFields[k] = v
		}
	}
	dynamicFieldsJSON, err := json.Marshal(existingDynamicFields)
	if err != nil {
		return err
	}

	// Scalar fields: fall back to existing when not provided
	subCategoryName := productItem.SubCategoryName
	if subCategoryName == "" {
		subCategoryName = existing.SubCategoryName
	}
	categoryID := productItem.CategoryID
	if categoryID == "" {
		categoryID = existing.CategoryID
	}
	departmentID := productItem.DepartmentID
	if departmentID == "" {
		departmentID = existing.DepartmentID
	}
	subCategoryID := productItem.SubCategoryID
	if subCategoryID == "" {
		subCategoryID = existing.SubCategoryID
	}
	description := productItem.Description
	if description == "" {
		description = existing.Description
	}
	highlights := productItem.Highlights
	if len(highlights) == 0 {
		highlights = existing.Highlights
	}

	// Image set: retained (client's kept keys) + newly uploaded.
	// If client sends no retained_images at all, preserve existing (back-compat).
	var finalImages []string
	if len(productItem.RetainedImages) > 0 {
		finalImages = append(finalImages, productItem.RetainedImages...)
	} else if len(productItem.ProductItemImages) == 0 {
		finalImages = existing.ProductItemImages
	}
	finalImages = append(finalImages, productItem.ProductItemImages...)

	var productItemImagesStr string
	if len(finalImages) > 0 {
		productItemImagesStr = "{" + strings.Join(finalImages, ",") + "}"
	} else {
		productItemImagesStr = "{}"
	}

	updatedAt := time.Now()
	query := `UPDATE product_items
		SET sub_category_name = $1, dynamic_fields = $2, product_item_images = $3,
		    category_id = $4, department_id = $5, sub_category_id = $6, updated_at = $7
		WHERE id = $8`

	err = c.DB.Exec(query, subCategoryName, dynamicFieldsJSON, productItemImagesStr, categoryID, departmentID, subCategoryID, updatedAt, productItemID).Error
	if err != nil {
		return err
	}

	if c.ElasticClient != nil {
		domainItem := domain.ProductItem{
			ID:                productItemID,
			SubCategoryName:   subCategoryName,
			CategoryID:        categoryID,
			DepartmentID:      departmentID,
			SubCategoryID:     subCategoryID,
			DynamicFields:     string(dynamicFieldsJSON),
			ProductItemImages: finalImages,
			Description:       description,
			Highlights:        highlights,
		}
		go func() {
			if updateErr := c.ElasticClient.UpdateProductItem(ctx, domainItem); updateErr != nil {
				log.Printf("Failed to update product item in Elasticsearch: %v", updateErr)
			}
		}()
	}

	return nil
}

// UpdateProductItemStock flips the in-stock / out-of-stock boolean for a single product item.
func (c *productDatabase) UpdateProductItemStock(ctx context.Context, productItemID string, inStock bool) error {
	query := `UPDATE product_items SET stock = $1, updated_at = $2 WHERE id = $3`
	return c.DB.Exec(query, inStock, time.Now(), productItemID).Error
}

// for get all products items for a product filtered by admin_id and additional filters
func (c *productDatabase) FindAllProductItems(ctx context.Context, adminID string, keyword string, categoryID *string, brandID *string, locationID *string, offer string, sortby string, pagination *request.Pagination, filterByShopID string, customerView bool) (productItems []response.ProductItems, err error) {
	var ids []string
	if keyword != "" && c.ElasticClient != nil {
		limit := 100
		offset := 0
		if pagination != nil {
			limit = int(pagination.Limit)
			offset = int(pagination.Offset)
		}
		var shopIDPtr *string
		if filterByShopID != "" {
			shopIDPtr = &filterByShopID
		}
		var err error
		ids, err = c.ElasticClient.SearchProductItems(ctx, keyword, categoryID, nil, brandID, shopIDPtr, limit, offset)
		if err != nil {
			log.Printf("ES search failed, falling back to PG: %v", err)
		} else if len(ids) == 0 {
			return []response.ProductItems{}, nil
		}
	}

	// Default to true if offer param is not explicitly "false"
	includeOffers := offer != "false"
	log.Printf("includeOffers: %v", includeOffers)

	// Define offerSubquery conditionally
	var offerSubquery string
	if !includeOffers {
		offerSubquery = `(SELECT '[]'::json) AS offer_products`
	} else {
		offerSubquery = `(SELECT COALESCE(json_agg(json_build_object(
			'offer_product_id', op2.id,
			'product_name', pi2.sub_category_name,
			'offer_id', p2.id,
			'offer_name', p2.offer_name,
			'discount_rate', p2.discount_rate,
			'description', p2.description,
			'start_date', p2.start_date,
			'end_date', p2.end_date,
			'promotion_category', json_build_object(
				'id', pc2.id,
				'name', pc2.name,
				'shop_id', pc2.shop_id,
				'is_active', pc2.is_active,
				'icon_path', pc2.icon_path,
				'created_at', pc2.created_at,
				'updated_at', pc2.updated_at
			),
			'promotion_type', json_build_object(
				'id', pt2.id,
				'name', pt2.name,
				'is_active', pt2.is_active,
				'shop_id', pt2.shop_id,
				'promotion_category_id', pt2.promotion_category_id,
				'icon_path', pt2.icon_path,
				'created_at', pt2.created_at,
				'updated_at', pt2.updated_at
			)
		) ORDER BY p2.created_at DESC), '[]')
		FROM offer_products op2
		LEFT JOIN product_items pi2 ON pi2.id = op2.product_item_id
		INNER JOIN promotions p2 ON p2.id = op2.offer_id
		LEFT JOIN promotion_categories pc2 ON p2.promotion_category_id = pc2.id
		LEFT JOIN promotions_types pt2 ON p2.promotion_type_id = pt2.id
		WHERE op2.product_item_id = pi.id
		AND p2.is_active = true) AS offer_products`
	}
	log.Printf("Offer parameter: '%s', includeOffers: %v", offer, includeOffers)
	log.Printf("offerSubquery: %s", offerSubquery)

	// Relevance ranking: active only for a text query resolved via the Postgres
	// fallback (no Elasticsearch id set). Mirrors SearchProducts so shop-detail /
	// category listing search ranks the same way as global search.
	rankingActive := searchRankingEnabled && keyword != "" && len(ids) == 0 && len(tokenizeQuery(keyword)) > 0
	scoreColumn := ""
	if rankingActive {
		// Layered additive weighted score on the FULL query phrase, plus a
		// per-token name-contains bonus. Uses GORM named params (reused below).
		scoreColumn = fmt.Sprintf(`(
			CASE WHEN pi.sub_category_name ILIKE @phraseExact THEN %d ELSE 0 END
			+ CASE WHEN pi.sub_category_name ILIKE @phrasePrefix THEN %d ELSE 0 END
			+ CASE WHEN pi.sub_category_name ILIKE @phraseContains THEN %d ELSE 0 END
			+ CASE WHEN (c.name::text ILIKE @phraseContains OR sc.name::text ILIKE @phraseContains) THEN %d ELSE 0 END
			+ CASE WHEN d.name::text ILIKE @phraseContains THEN %d ELSE 0 END
			+ CASE WHEN (pi.description ILIKE @phraseContains OR array_to_string(pi.highlights, ' ') ILIKE @phraseContains OR pi.dynamic_fields::text ILIKE @phraseContains) THEN %d ELSE 0 END
		) AS relevance_score,`,
			scoreExactName, scoreNamePrefix, scoreNameContains,
			scoreCategoryContains, scoreDepartmentContains, scoreLongTextContains)
	}

	query := `SELECT pi.sub_category_name, pi.id, pi.category_id, pi.department_id, pi.sub_category_id,
			pi.product_item_images, pi.dynamic_fields, pi.stock, pi.created_at, pi.updated_at,
				c.name AS category_name, d.name AS department_name, sc.name AS sub_category_name_ref,
			(SELECT MAX(p3.discount_rate) FROM offer_products op3 INNER JOIN promotions p3 ON p3.id = op3.offer_id WHERE op3.product_item_id = pi.id AND p3.is_active = true AND (p3.start_date)::timestamp <= CURRENT_TIMESTAMP AND (p3.end_date)::timestamp >= CURRENT_TIMESTAMP) AS discount_rate,
				sc.image_url AS sub_category_image_url,
				(SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = pi.id) AS view_count,
				` + scoreColumn + `
				` + offerSubquery + `
			FROM product_items pi
			LEFT JOIN categories c ON pi.category_id = c.id
			LEFT JOIN departments d ON pi.department_id = d.id
			LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
			LEFT JOIN shop_details sd ON sd.id = pi.shop_id
			WHERE 1=1`

	// Subscription gate: on customer-facing calls, hide products whose owning shop
	// has no active subscription. Skipped for seller/admin views (customerView=false)
	// so a lapsed seller still sees their own catalogue.
	if customerView && subscriptionGateEnabled {
		query += " AND " + SubscribedShopPredicate
	}

	// Add offer filter - this ensures different data sets based on offer parameter
	log.Printf("DEBUG: offer=%s, len=%d", offer, len(offer))
	log.Printf("DEBUG: offer == 'true': %v, offer == 'false': %v", offer == "true", offer == "false")

	if offer == "true" {
		// Return ONLY products WITH active, valid offers
		log.Printf("DEBUG: Applying offer=true filter (products WITH offers)")
		query += " AND EXISTS (SELECT 1 FROM offer_products op INNER JOIN promotions p ON p.id = op.offer_id WHERE op.product_item_id = pi.id AND p.is_active = true AND (p.end_date)::timestamp >= CURRENT_TIMESTAMP)"
	} else if offer == "false" {
		// Return ONLY products WITHOUT ANY offers at all (regardless of date/active status)
		log.Printf("DEBUG: Applying offer=false filter (products WITHOUT any offers)")
		query += " AND NOT EXISTS (SELECT 1 FROM offer_products WHERE product_item_id = pi.id)"
	} else {
		log.Printf("DEBUG: No offer filter applied (returning all products)")
	}
	// If offer is neither "true" nor "false" (empty or other values), return all products without filtering
	// Add filters dynamically
	params := map[string]interface{}{}
	if adminID != "" {
		params["adminID"] = adminID
	}
	if filterByShopID != "" {
		query += " AND pi.shop_id = @shopID"
		params["shopID"] = filterByShopID
	}
	// Removed shop_id filter as it may not be set correctly in DB
	if len(ids) > 0 {
		params["ids"] = ids
		query += " AND pi.id = ANY(@ids)"
	}
	if rankingActive {
		// AND-of-tokens candidacy across every searched field (unified with
		// SearchProducts): each token must hit at least one field.
		params["phraseExact"] = keyword
		params["phrasePrefix"] = keyword + "%"
		params["phraseContains"] = "%" + keyword + "%"
		tokens := tokenizeQuery(keyword)
		tokenClauses := make([]string, 0, len(tokens))
		tokenNameBonuses := make([]string, 0, len(tokens))
		for i, tok := range tokens {
			key := fmt.Sprintf("tok%d", i)
			params[key] = "%" + tok + "%"
			tokenClauses = append(tokenClauses, fmt.Sprintf(
				"(pi.sub_category_name ILIKE @%s OR c.name::text ILIKE @%s OR sc.name::text ILIKE @%s OR d.name::text ILIKE @%s OR pi.description ILIKE @%s OR array_to_string(pi.highlights, ' ') ILIKE @%s OR pi.dynamic_fields::text ILIKE @%s)",
				key, key, key, key, key, key, key))
			tokenNameBonuses = append(tokenNameBonuses, fmt.Sprintf(
				"CASE WHEN pi.sub_category_name ILIKE @%s THEN %d ELSE 0 END", key, scorePerTokenNameBonus))
		}
		query += " AND (" + strings.Join(tokenClauses, " AND ") + ")"
		// Fold the per-token name bonus into the score expression built above.
		if len(tokenNameBonuses) > 0 {
			query = strings.Replace(query,
				"\n\t\t) AS relevance_score,",
				"\n\t\t\t+ "+strings.Join(tokenNameBonuses, " + ")+"\n\t\t) AS relevance_score,", 1)
		}
	} else if keyword != "" && len(ids) == 0 {
		query += " AND (pi.sub_category_name ILIKE @keyword OR c.name ILIKE @keyword OR sc.name ILIKE @keyword)"
		params["keyword"] = "%" + keyword + "%"
	}
	if categoryID != nil && *categoryID != "" && len(ids) == 0 {
		query += " AND pi.category_id = @categoryID"
		params["categoryID"] = *categoryID
	}
	if brandID != nil && *brandID != "" {
		query += " AND pi.brand_id = @brandID"
		params["brandID"] = *brandID
	}
	if locationID != nil && *locationID != "" {
		query += " AND pi.location_id = @locationID"
		params["locationID"] = *locationID
	}
	orderBy := "pi.created_at"
	if sortby != "" {
		switch sortby {
		case "created_at":
			orderBy = "pi.created_at"
		case "updated_at":
			orderBy = "pi.updated_at"
		case "name":
			orderBy = "pi.sub_category_name"
		case "views":
			orderBy = "view_count"
		}
	}
	// When relevance ranking is active, sort by score first, keep the existing
	// column as the tie-breaker, and add pi.id for deterministic pagination.
	orderByClause := orderBy + " DESC"
	if rankingActive {
		orderByClause = "relevance_score DESC, " + orderBy + " DESC, pi.id"
	}
	if pagination != nil {
		query += " ORDER BY " + orderByClause + " LIMIT @limit OFFSET @offset"
		params["limit"] = pagination.Limit
		params["offset"] = pagination.Offset
	} else {
		query += " ORDER BY " + orderByClause
	}

	// Internal struct for scanning DB result
	type productItemDB struct {
		Name                string        `gorm:"column:sub_category_name"`
		ID                  string        `gorm:"column:id"`
		CategoryID          string        `gorm:"column:category_id"`
		DepartmentID        string        `gorm:"column:department_id"`
		SubCategoryID       string        `gorm:"column:sub_category_id"`
		CategoryName        string        `gorm:"column:category_name"`
		DepartmentName      string        `gorm:"column:department_name"`
		SubCategoryNameRef  string        `gorm:"column:sub_category_name_ref"`
		SubCategoryImageURL string        `gorm:"column:sub_category_image_url"`
		ProductItemImages   string        `gorm:"column:product_item_images"` // Store as string
		DynamicFields       []byte        `gorm:"column:dynamic_fields"`
		OfferProducts       []byte        `gorm:"column:offer_products"`
		CreatedAt           time.Time     `gorm:"column:created_at"`
		UpdatedAt           time.Time     `gorm:"column:updated_at"`
		DiscountRate        sql.NullInt64 `gorm:"column:discount_rate"`
		ViewCount           uint          `gorm:"column:view_count"`
		Stock               bool          `gorm:"column:stock"`
	}
	var dbItems []productItemDB
	log.Printf("Query: %s, Params: %v", query, params)
	err = c.DB.Raw(query, params).Scan(&dbItems).Error
	if err != nil {
		return
	}

	log.Printf("DEBUG: dbItems count=%d", len(dbItems))
	if len(dbItems) > 0 {
		log.Printf("First dbItem: %+v", dbItems[0])
	}

	// Map to response.ProductItems
	for _, dbItem := range dbItems {
		// Parse product_item_images from PostgreSQL array format
		var images []string
		if dbItem.ProductItemImages != "" {
			// Remove curly braces and parse comma-separated values
			imageStr := dbItem.ProductItemImages
			if len(imageStr) > 2 && imageStr[0] == '{' && imageStr[len(imageStr)-1] == '}' {
				imageStr = imageStr[1 : len(imageStr)-1]
				if imageStr != "" {
					images = []string{}
					for _, img := range parsePostgresArray(imageStr) {
						images = append(images, img)
					}
				}
			}
		}
		if images == nil {
			images = []string{}
		}

		item := response.ProductItems{
			ID:                  dbItem.ID,
			Name:                dbItem.Name,
			CategoryName:        dbItem.CategoryName,
			MainCategoryName:    dbItem.DepartmentName,
			SubCategoryImageURL: dbItem.SubCategoryImageURL,
			CategoryID:          dbItem.CategoryID,
			DepartmentID:        dbItem.DepartmentID,
			SubCategoryID:       dbItem.SubCategoryID,
			ProductItemImages:   images,
			CreatedAt:           dbItem.CreatedAt,
			UpdatedAt:           dbItem.UpdatedAt,
			ViewCount:           dbItem.ViewCount,
			Stock:               dbItem.Stock,
		}

		// Unmarshal offer_products if present
		if len(dbItem.OfferProducts) > 0 {
			var offerProducts []response.OfferProduct
			log.Printf("Attempting to unmarshal offer_products: %s", string(dbItem.OfferProducts))
			if err := json.Unmarshal(dbItem.OfferProducts, &offerProducts); err != nil {
				log.Printf("Failed to unmarshal offer_products: %v", err)
			} else {
				item.OfferProducts = offerProducts
				log.Printf("Loaded %d offer_products for dbItem.ID=%s", len(offerProducts), dbItem.ID)
			}
		} else {
			log.Printf("No offer_products data for product %s", dbItem.ID)
		}

		// Unmarshal DynamicFields if present
		if len(dbItem.DynamicFields) > 0 {
			var dynamicFields map[string]interface{}
			if err := json.Unmarshal(dbItem.DynamicFields, &dynamicFields); err == nil {
				item.DynamicFields = dynamicFields
			}
		}

		productItems = append(productItems, item)
	}
	return
}

// Helper function to parse PostgreSQL array format
func parsePostgresArray(s string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	var current string
	inQuotes := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuotes = !inQuotes
		} else if c == ',' && !inQuotes {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// FindLowViewProductItems finds products with less than 5 views in the last 10 days for a specific shop
// Used to identify underperforming products that need promotion
func (c *productDatabase) FindLowViewProductItems(ctx context.Context,
	adminID string, keyword string, categoryID *string, brandID *string, locationID *string, sortby string, pagination *request.Pagination, filterByShopID *string,
) (productItems []response.ProductItems, err error) {
	log.Printf("FindLowViewProductItems called with shopID: %v", filterByShopID)

	var ids []string
	if keyword != "" && c.ElasticClient != nil {
		limit := 100
		offset := 0
		if pagination != nil {
			limit = int(pagination.Limit)
			offset = int(pagination.Offset)
		}
		var err error
		ids, err = c.ElasticClient.SearchProductItems(ctx, keyword, categoryID, nil, brandID, filterByShopID, limit, offset)
		if err != nil {
			log.Printf("ES search failed, falling back to PG: %v", err)
		} else if len(ids) == 0 {
			return []response.ProductItems{}, nil
		}
	}

	// Subquery to get offer products
	offerSubquery := `(SELECT COALESCE(json_agg(json_build_object(
		'offer_product_id', op2.id,
		'product_name', pi2.sub_category_name,
		'offer_id', p2.id,
		'offer_name', p2.offer_name,
		'discount_rate', p2.discount_rate,
		'description', p2.description,
		'start_date', p2.start_date,
		'end_date', p2.end_date,
		'promotion_category', json_build_object(
			'id', pc2.id,
			'name', pc2.name,
			'shop_id', pc2.shop_id,
			'is_active', pc2.is_active,
			'icon_path', pc2.icon_path,
			'created_at', pc2.created_at,
			'updated_at', pc2.updated_at
		),
		'promotion_type', json_build_object(
			'id', pt2.id,
			'name', pt2.name,
			'is_active', pt2.is_active,
			'shop_id', pt2.shop_id,
			'promotion_category_id', pt2.promotion_category_id,
			'icon_path', pt2.icon_path,
			'created_at', pt2.created_at,
			'updated_at', pt2.updated_at
		)
	) ORDER BY p2.created_at DESC), '[]')
	FROM offer_products op2
	LEFT JOIN product_items pi2 ON pi2.id = op2.product_item_id
	INNER JOIN promotions p2 ON p2.id = op2.offer_id
	LEFT JOIN promotion_categories pc2 ON p2.promotion_category_id = pc2.id
	LEFT JOIN promotions_types pt2 ON p2.promotion_type_id = pt2.id
	WHERE op2.product_item_id = pi.id
	AND p2.is_active = true) AS offer_products`

	query := `SELECT pi.sub_category_name, pi.id, pi.category_id, pi.department_id, pi.sub_category_id,
				pi.product_item_images, pi.dynamic_fields, pi.created_at, pi.updated_at,
				c.name AS category_name, d.name AS department_name, sc.name AS sub_category_name_ref,
			(SELECT MAX(p3.discount_rate) FROM offer_products op3 INNER JOIN promotions p3 ON p3.id = op3.offer_id WHERE op3.product_item_id = pi.id AND p3.is_active = true AND (p3.start_date)::timestamp <= CURRENT_TIMESTAMP AND (p3.end_date)::timestamp >= CURRENT_TIMESTAMP) AS discount_rate,
				sc.image_url AS sub_category_image_url,
				(SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = pi.id AND viewed_at >= (pi.created_at + INTERVAL '10 days') AND viewed_at < (pi.created_at + INTERVAL '20 days')) AS view_count_10days,
				` + offerSubquery + `
			FROM product_items pi 
			LEFT JOIN categories c ON pi.category_id = c.id 
			LEFT JOIN departments d ON pi.department_id = d.id
			LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
			WHERE (pi.admin_id ->> 'id' = @adminID OR pi.admin_id #>> '{}' = @adminID)
			AND (SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = pi.id AND viewed_at >= (pi.created_at + INTERVAL '10 days') AND viewed_at < (pi.created_at + INTERVAL '20 days')) < 5`

	// Add filters dynamically
	params := map[string]interface{}{
		"adminID": adminID,
	}
	if filterByShopID != nil && *filterByShopID != "" {
		shopIDUint, err := strconv.ParseUint(*filterByShopID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid shop_id: %w", err)
		}
		query += " AND pi.shop_id = @shopID"
		params["shopID"] = uint(shopIDUint)
	}
	// Removed shop_id filter

	if len(ids) > 0 {
		params["ids"] = ids
		query += " AND pi.id = ANY(@ids)"
	}

	if keyword != "" && len(ids) == 0 {
		query += " AND (pi.sub_category_name ILIKE @keyword OR c.name ILIKE @keyword OR sc.name ILIKE @keyword)"
		params["keyword"] = "%" + keyword + "%"
	}

	if categoryID != nil && *categoryID != "" && len(ids) == 0 {
		query += " AND pi.category_id = @categoryID"
		params["categoryID"] = *categoryID
	}

	if brandID != nil && *brandID != "" {
		query += " AND pi.brand_id = @brandID"
		params["brandID"] = *brandID
	}

	if locationID != nil && *locationID != "" {
		query += " AND pi.location_id = @locationID"
		params["locationID"] = *locationID
	}

	orderBy := "view_count_10days"
	if sortby != "" {
		switch sortby {
		case "created_at":
			orderBy = "pi.created_at"
		case "updated_at":
			orderBy = "pi.updated_at"
		case "name":
			orderBy = "pi.sub_category_name"
		case "views":
			orderBy = "view_count_10days"
		}
	}

	if pagination != nil {
		query += " ORDER BY " + orderBy + " ASC LIMIT @limit OFFSET @offset"
		params["limit"] = pagination.Limit
		params["offset"] = pagination.Offset
	} else {
		query += " ORDER BY " + orderBy + " ASC"
	}

	// Internal struct for scanning DB result
	type productItemDB struct {
		Name                string        `gorm:"column:sub_category_name"`
		ID                  string        `gorm:"column:id"`
		CategoryID          string        `gorm:"column:category_id"`
		DepartmentID        string        `gorm:"column:department_id"`
		SubCategoryID       string        `gorm:"column:sub_category_id"`
		CategoryName        string        `gorm:"column:category_name"`
		DepartmentName      string        `gorm:"column:department_name"`
		SubCategoryNameRef  string        `gorm:"column:sub_category_name_ref"`
		SubCategoryImageURL string        `gorm:"column:sub_category_image_url"`
		ProductItemImages   string        `gorm:"column:product_item_images"`
		DynamicFields       []byte        `gorm:"column:dynamic_fields"`
		OfferProducts       []byte        `gorm:"column:offer_products"`
		CreatedAt           time.Time     `gorm:"column:created_at"`
		UpdatedAt           time.Time     `gorm:"column:updated_at"`
		DiscountRate        sql.NullInt64 `gorm:"column:discount_rate"`
		ViewCount10Days     uint          `gorm:"column:view_count_10days"`
	}

	var dbItems []productItemDB
	log.Printf("Query: %s, Params: %v", query, params)
	err = c.DB.Raw(query, params).Scan(&dbItems).Error
	if err != nil {
		return
	}

	log.Printf("DEBUG: dbItems count=%d", len(dbItems))

	// Map to response.ProductItems
	for _, dbItem := range dbItems {
		// Parse product_item_images from PostgreSQL array format
		var images []string
		if dbItem.ProductItemImages != "" {
			imageStr := dbItem.ProductItemImages
			if len(imageStr) > 2 && imageStr[0] == '{' && imageStr[len(imageStr)-1] == '}' {
				imageStr = imageStr[1 : len(imageStr)-1]
				if imageStr != "" {
					images = []string{}
					for _, img := range parsePostgresArray(imageStr) {
						images = append(images, img)
					}
				}
			}
		}
		if images == nil {
			images = []string{}
		}

		item := response.ProductItems{
			ID:                  dbItem.ID,
			Name:                dbItem.Name,
			CategoryName:        dbItem.CategoryName,
			MainCategoryName:    dbItem.DepartmentName,
			SubCategoryImageURL: dbItem.SubCategoryImageURL,
			CategoryID:          dbItem.CategoryID,
			DepartmentID:        dbItem.DepartmentID,
			SubCategoryID:       dbItem.SubCategoryID,
			ProductItemImages:   images,
			CreatedAt:           dbItem.CreatedAt,
			UpdatedAt:           dbItem.UpdatedAt,
			ViewCount:           dbItem.ViewCount10Days,
		}

		// Unmarshal offer_products if present
		if len(dbItem.OfferProducts) > 0 {
			var offerProducts []response.OfferProduct
			log.Printf("Attempting to unmarshal offer_products for low-view product: %s", string(dbItem.OfferProducts))
			if err := json.Unmarshal(dbItem.OfferProducts, &offerProducts); err != nil {
				log.Printf("Failed to unmarshal offer_products: %v", err)
			} else {
				item.OfferProducts = offerProducts
				log.Printf("Loaded %d offer_products for dbItem.ID=%s", len(offerProducts), dbItem.ID)
			}
		}

		// Unmarshal DynamicFields if present
		if len(dbItem.DynamicFields) > 0 {
			var dynamicFields map[string]interface{}
			if err := json.Unmarshal(dbItem.DynamicFields, &dynamicFields); err == nil {
				item.DynamicFields = dynamicFields
			}
		}

		productItems = append(productItems, item)
	}

	log.Printf("DEBUG: productItems count=%d", len(productItems))
	return
}

// Find all variation and value of a product item
func (c *productDatabase) FindAllVariationValuesOfProductItem(ctx context.Context,
	productItemID string,
) (productVariationsValues []response.ProductVariationValue, err error) {
	query := `SELECT v.id AS variation_id, v.name, vo.id AS variation_option_id, vo.value 
	FROM  product_configurations pc 
	INNER JOIN variation_options vo ON vo.id = pc.variation_option_id 
	INNER JOIN variations v ON v.id = vo.variation_id 
	WHERE pc.product_item_id = $1`
	err = c.DB.Raw(query, productItemID).Scan(&productVariationsValues).Error

	return
}

// To save image for product item
func (c *productDatabase) SaveProductItemImage(ctx context.Context, productItemID string, image domain.ProductItemImage) error {
	query := `INSERT INTO product_images (product_item_id, image) VALUES ($1, $2)`
	err := c.DB.Exec(query, productItemID, image).Error

	return err
}

// To find all images of a product item
func (c *productDatabase) FindAllProductItemImages(ctx context.Context, productItemID string) (images []string, err error) {
	query := `SELECT COALESCE(product_item_images, ARRAY[]::text[])::text AS product_item_images FROM product_items WHERE id = $1`

	var rawImages string
	err = c.DB.Raw(query, productItemID).Scan(&rawImages).Error
	if err != nil {
		return nil, err
	}

	if rawImages == "" || rawImages == "{}" {
		return []string{}, nil
	}

	if strings.HasPrefix(rawImages, "{") && strings.HasSuffix(rawImages, "}") {
		rawImages = rawImages[1 : len(rawImages)-1]
	}

	if strings.TrimSpace(rawImages) == "" {
		return []string{}, nil
	}

	parts := strings.Split(rawImages, ",")
	images = make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.Trim(strings.TrimSpace(part), `"`)
		if trimmed != "" {
			images = append(images, trimmed)
		}
	}

	return
}

func buildGeoDistanceQuery(lat, lng, radius float64, startParam int, locationColumnOrLatCol, longitudeCol string, defaultOrderBy string) (distanceExpr string, geoFilter string, orderBy string, params []interface{}, nextParam int) {
	distanceExpr = "NULL::double precision AS distance_km"
	orderBy = defaultOrderBy
	nextParam = startParam

	if lat != 0 && lng != 0 {
		// Check if we're using latitude/longitude columns (both parameters provided) or PostGIS geometry
		if longitudeCol != "" {
			// Using latitude/longitude numeric columns with haversine formula
			latIdx := nextParam
			lonIdx := nextParam + 1
			radiusIdx := nextParam + 2

			distanceExpr = fmt.Sprintf(
				`(6371 * acos(
					LEAST(1, GREATEST(-1,
						cos(radians($%d)) * cos(radians(%s::double precision)) *
						cos(radians(%s::double precision) - radians($%d)) +
						sin(radians($%d)) * sin(radians(%s::double precision))
					))
				)) AS distance_km`,
				latIdx,
				locationColumnOrLatCol,
				longitudeCol,
				lonIdx,
				latIdx,
				locationColumnOrLatCol,
			)

			whereFilterTemplate := ` AND %s IS NOT NULL AND %s IS NOT NULL AND
			  (6371 * acos(
				LEAST(1, GREATEST(-1,
					cos(radians($%d)) * cos(radians(%s::double precision)) *
					cos(radians(%s::double precision) - radians($%d)) +
					sin(radians($%d)) * sin(radians(%s::double precision))
				))
			  )) <= $%d`

			if radius > 0 {
				geoFilter = fmt.Sprintf(whereFilterTemplate,
					locationColumnOrLatCol,
					longitudeCol,
					latIdx,
					locationColumnOrLatCol,
					longitudeCol,
					lonIdx,
					latIdx,
					locationColumnOrLatCol,
					radiusIdx,
				)
				params = append(params, lat, lng, radius)
				nextParam += 3
			} else {
				// No radius filter, just calculate distance
				geoFilter = fmt.Sprintf(" AND %s IS NOT NULL AND %s IS NOT NULL", locationColumnOrLatCol, longitudeCol)
				params = append(params, lat, lng)
				nextParam += 2
			}

			orderBy = fmt.Sprintf(" ORDER BY distance_km ASC NULLS LAST, pi.created_at DESC")
		} else {
			// Using PostGIS geometry column (old behavior)
			lonIdx := nextParam
			latIdx := nextParam + 1

			distanceExpr = fmt.Sprintf(
				`ST_Distance(%s, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography) / 1000.0 AS distance_km`,
				locationColumnOrLatCol,
				lonIdx,
				latIdx,
			)

			params = append(params, lng, lat)
			nextParam += 2

			if radius > 0 {
				radiusIdx := nextParam
				geoFilter = fmt.Sprintf(
					` AND %s IS NOT NULL AND ST_DWithin(%s, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography, $%d)`,
					locationColumnOrLatCol,
					locationColumnOrLatCol,
					lonIdx,
					latIdx,
					radiusIdx,
				)
				params = append(params, radius*1000)
				nextParam++
			}

			orderBy = fmt.Sprintf(" ORDER BY distance_km ASC NULLS LAST%s", defaultOrderBy)
		}
	}

	return
}

// SearchProducts implements interfaces.ProductRepository.
func (c *productDatabase) SearchProducts(ctx context.Context, keyword string, categoryID, departmentID, brandID, locationID, shopID *string, latitude, longitude, radius float64, pincode *uint, pagination request.Pagination) (products []response.ProductItems, err error) {
	limit := int(pagination.Limit)
	offset := int(pagination.Offset)

	// Initialize products as empty slice to avoid nil in JSON response
	products = []response.ProductItems{}

	var ids []string
	var useElasticsearchResults bool
	if keyword != "" && c.ElasticClient != nil {
		// Use Elasticsearch for search with all filters applied with AND logic
		ids, err = c.ElasticClient.SearchProductItems(ctx, keyword, categoryID, departmentID, brandID, shopID, limit, offset)
		if err != nil {
			log.Printf("ES search failed, falling back to PG: %v", err)
		} else if len(ids) > 0 {
			useElasticsearchResults = true
		}
	}

	// Build the base query
	baseQueryTemplate := `SELECT pi.sub_category_name, pi.id, pi.category_id, pi.department_id, pi.sub_category_id,
			pi.product_item_images, pi.dynamic_fields, pi.created_at, pi.updated_at,
			c.name AS category_name, d.name AS department_name, sc.name AS sub_category_name_ref,
			sc.image_url AS sub_category_image_url,
			sd.id AS shop_id, sd.shop_name,
			%s,
			%s
			(SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = pi.id) AS view_count,
			(
				SELECT COALESCE(json_agg(json_build_object(
					'offer_product_id', op2.id,
					'product_name', pi2.sub_category_name,
					'offer_id', p2.id,
					'offer_name', p2.offer_name,
					'discount_rate', p2.discount_rate,
					'description', p2.description,
					'start_date', p2.start_date,
					'end_date', p2.end_date
				)), '[]')
				FROM offer_products op2
				LEFT JOIN product_items pi2 ON pi2.id = op2.product_item_id
				INNER JOIN promotions p2 ON p2.id = op2.offer_id
				WHERE op2.product_item_id = pi.id
				AND p2.is_active = true
				AND (p2.start_date)::timestamp <= CURRENT_TIMESTAMP
				AND (p2.end_date)::timestamp >= CURRENT_TIMESTAMP
			) AS offer_products
		FROM product_items pi
		LEFT JOIN categories c ON pi.category_id = c.id
		LEFT JOIN departments d ON pi.department_id = d.id
		LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
		LEFT JOIN shop_details sd ON sd.id = pi.shop_id`

	params := []interface{}{}
	paramIndex := 1
	whereClause := " WHERE 1=1"

	// scoreExpr is the relevance_score SELECT column; empty unless ranking is on.
	// rankingOrderBy is the ORDER BY prefix that sorts by relevance_score first;
	// empty unless ranking is on (falls back to the geo/created_at ordering).
	scoreExpr := ""
	rankingOrderBy := ""
	rankingActive := searchRankingEnabled && keyword != "" && !useElasticsearchResults && len(tokenizeQuery(keyword)) > 0

	// Subscription gate: hide products whose owning shop has no active subscription.
	// Applied in the Postgres WHERE (never in Elasticsearch), so it filters
	// ES-derived candidates during hydration too. Uses the shop_details sd alias.
	if subscriptionGateEnabled {
		whereClause += " AND " + SubscribedShopPredicate
	}

	// If we have IDs from Elasticsearch, filter by them
	if useElasticsearchResults && len(ids) > 0 {
		placeholders := make([]string, len(ids))
		for i, id := range ids {
			placeholders[i] = fmt.Sprintf("$%d", paramIndex)
			params = append(params, id)
			paramIndex++
		}
		whereClause += " AND pi.id IN (" + strings.Join(placeholders, ",") + ")"
	} else if rankingActive {
		// Relevance ranking ON: AND-of-tokens candidacy across every searched
		// field (each token must hit at least one field), plus a computed
		// relevance_score that ranks contiguous full-phrase matches highest.
		tokens := tokenizeQuery(keyword)

		// Full-phrase param (used by the score tiers and the phrase-contains match).
		phraseExact := paramIndex // ILIKE <phrase> (no wildcards) -> exact
		params = append(params, keyword)
		paramIndex++
		phrasePrefix := paramIndex // ILIKE <phrase>%
		params = append(params, keyword+"%")
		paramIndex++
		phraseContains := paramIndex // ILIKE %<phrase>%
		params = append(params, "%"+keyword+"%")
		paramIndex++

		// Per-token candidacy: AND over tokens, each token ORs across all fields.
		tokenClauses := make([]string, 0, len(tokens))
		tokenNameBonuses := make([]string, 0, len(tokens))
		for _, tok := range tokens {
			idx := paramIndex
			params = append(params, "%"+tok+"%")
			paramIndex++
			tokenClauses = append(tokenClauses, fmt.Sprintf(
				"(pi.sub_category_name ILIKE $%d OR c.name::text ILIKE $%d OR sc.name::text ILIKE $%d OR d.name::text ILIKE $%d OR pi.description ILIKE $%d OR array_to_string(pi.highlights, ' ') ILIKE $%d OR pi.dynamic_fields::text ILIKE $%d)",
				idx, idx, idx, idx, idx, idx, idx))
			tokenNameBonuses = append(tokenNameBonuses, fmt.Sprintf(
				"CASE WHEN pi.sub_category_name ILIKE $%d THEN %d ELSE 0 END", idx, scorePerTokenNameBonus))
		}
		whereClause += " AND (" + strings.Join(tokenClauses, " AND ") + ")"

		// Layered, additive weighted score on the FULL query phrase, plus the
		// per-token name bonus so multi-word name matches beat category-only ones.
		scoreExpr = fmt.Sprintf(`(
			CASE WHEN pi.sub_category_name ILIKE $%d THEN %d ELSE 0 END
			+ CASE WHEN pi.sub_category_name ILIKE $%d THEN %d ELSE 0 END
			+ CASE WHEN pi.sub_category_name ILIKE $%d THEN %d ELSE 0 END
			+ CASE WHEN (c.name::text ILIKE $%d OR sc.name::text ILIKE $%d) THEN %d ELSE 0 END
			+ CASE WHEN d.name::text ILIKE $%d THEN %d ELSE 0 END
			+ CASE WHEN (pi.description ILIKE $%d OR array_to_string(pi.highlights, ' ') ILIKE $%d OR pi.dynamic_fields::text ILIKE $%d) THEN %d ELSE 0 END
			+ %s
		) AS relevance_score,`,
			phraseExact, scoreExactName,
			phrasePrefix, scoreNamePrefix,
			phraseContains, scoreNameContains,
			phraseContains, phraseContains, scoreCategoryContains,
			phraseContains, scoreDepartmentContains,
			phraseContains, phraseContains, phraseContains, scoreLongTextContains,
			strings.Join(tokenNameBonuses, " + "))
	} else if keyword != "" {
		// Fallback to keyword search if no Elasticsearch or ES failed (ranking off).
		whereClause += fmt.Sprintf(" AND (pi.sub_category_name ILIKE $%d OR pi.dynamic_fields::text ILIKE $%d OR c.name::text ILIKE $%d OR sc.name::text ILIKE $%d OR d.name::text ILIKE $%d)", paramIndex, paramIndex, paramIndex, paramIndex, paramIndex)
		params = append(params, "%"+keyword+"%")
		paramIndex++
	}
	// Note: If keyword is empty, we continue without keyword filter but apply other filters (category, department, geo, etc.)

	if categoryID != nil {
		if cid, err := strconv.ParseUint(*categoryID, 10, 64); err == nil {
			whereClause += fmt.Sprintf(" AND pi.category_id = $%d", paramIndex)
			params = append(params, cid)
			paramIndex++
		}
	}

	// department_id is varchar(32) (e.g. "dept_00001"), not an integer.
	if departmentID != nil && *departmentID != "" {
		whereClause += fmt.Sprintf(" AND pi.department_id = $%d", paramIndex)
		params = append(params, *departmentID)
		paramIndex++
	}

	if shopID != nil {
		if sid, err := strconv.ParseUint(*shopID, 10, 64); err == nil {
			whereClause += fmt.Sprintf(" AND pi.shop_id = $%d", paramIndex)
			params = append(params, sid)
			paramIndex++
		}
	}

	if brandID != nil {
		if bid, err := strconv.ParseUint(*brandID, 10, 64); err == nil {
			whereClause += fmt.Sprintf(" AND pi.brand_id = $%d", paramIndex)
			params = append(params, bid)
			paramIndex++
		}
	}

	if locationID != nil {
		if lid, err := strconv.ParseUint(*locationID, 10, 64); err == nil {
			whereClause += fmt.Sprintf(" AND pi.location_id = $%d", paramIndex)
			params = append(params, lid)
			paramIndex++
		}
	}

	distanceExpr, geoFilter, orderBy, geoArgs, paramIndex := buildGeoDistanceQuery(latitude, longitude, radius, paramIndex, "sd.latitude", "sd.longitude", " ORDER BY pi.created_at DESC")
	if geoFilter != "" {
		whereClause += geoFilter
	}
	params = append(params, geoArgs...)

	// When relevance ranking is active, order by score first, then the existing
	// distance/recency tie-breakers, with a final deterministic tiebreaker on
	// pi.id so pagination is stable. buildGeoDistanceQuery has already set orderBy
	// to distance-then-recency (geo) or created_at DESC (no geo); we splice the
	// score in front and reuse those existing tie-break columns.
	if rankingActive {
		useGeo := latitude != 0 && longitude != 0
		if useGeo {
			rankingOrderBy = " ORDER BY relevance_score DESC, distance_km ASC NULLS LAST, pi.created_at DESC, pi.id"
		} else {
			rankingOrderBy = " ORDER BY relevance_score DESC, pi.created_at DESC, pi.id"
		}
		orderBy = rankingOrderBy
	}

	useGeoDistance := latitude != 0 && longitude != 0

	if !useGeoDistance && pincode != nil {
		// Use pincode filter only if geolocation is not provided
		whereClause += fmt.Sprintf(" AND sd.pincode = $%d", paramIndex)
		params = append(params, fmt.Sprintf("%d", *pincode))
		paramIndex++
	}

	baseQuery := fmt.Sprintf(baseQueryTemplate, distanceExpr, scoreExpr)
	baseQuery += whereClause + orderBy + " LIMIT $" + fmt.Sprint(paramIndex) + " OFFSET $" + fmt.Sprint(paramIndex+1)
	params = append(params, limit, offset)

	// Log the final SQL and parameters for debugging

	// Scan into internal DB struct to correctly parse JSONB and array columns
	type productItemDB struct {
		Name                string    `gorm:"column:sub_category_name"`
		ID                  string    `gorm:"column:id"`
		CategoryID          string    `gorm:"column:category_id"`
		DepartmentID        string    `gorm:"column:department_id"`
		SubCategoryID       string    `gorm:"column:sub_category_id"`
		CategoryName        string    `gorm:"column:category_name"`
		DepartmentName      string    `gorm:"column:department_name"`
		SubCategoryNameRef  string    `gorm:"column:sub_category_name_ref"`
		SubCategoryImageURL string    `gorm:"column:sub_category_image_url"`
		ShopID              string    `gorm:"column:shop_id"`
		ShopName            string    `gorm:"column:shop_name"`
		ProductItemImages   string    `gorm:"column:product_item_images"`
		DynamicFields       []byte    `gorm:"column:dynamic_fields"`
		OfferProducts       []byte    `gorm:"column:offer_products"`
		DistanceKM          *float64  `gorm:"column:distance_km"`
		CreatedAt           time.Time `gorm:"column:created_at"`
		UpdatedAt           time.Time `gorm:"column:updated_at"`
		ViewCount           uint      `gorm:"column:view_count"`
	}

	var dbItems []productItemDB
	err = c.DB.Raw(baseQuery, params...).Scan(&dbItems).Error
	if err != nil {
		return
	}

	// Map to response.ProductItems
	for _, dbItem := range dbItems {
		// Parse product_item_images (Postgres array format stored as string)
		var images []string
		if dbItem.ProductItemImages != "" {
			imageStr := dbItem.ProductItemImages
			if len(imageStr) > 2 && imageStr[0] == '{' && imageStr[len(imageStr)-1] == '}' {
				imageStr = imageStr[1 : len(imageStr)-1]
				if imageStr != "" {
					images = []string{}
					for _, img := range parsePostgresArray(imageStr) {
						images = append(images, img)
					}
				}
			}
		}
		if images == nil {
			images = []string{}
		}

		item := response.ProductItems{
			ID:                  dbItem.ID,
			Name:                dbItem.Name,
			CategoryName:        dbItem.CategoryName,
			MainCategoryName:    dbItem.DepartmentName,
			SubCategoryImageURL: dbItem.SubCategoryImageURL,
			CategoryID:          dbItem.CategoryID,
			DepartmentID:        dbItem.DepartmentID,
			SubCategoryID:       dbItem.SubCategoryID,
			ShopID:              dbItem.ShopID,
			ShopName:            dbItem.ShopName,
			ProductItemImages:   images,
			CreatedAt:           dbItem.CreatedAt,
			UpdatedAt:           dbItem.UpdatedAt,
			ViewCount:           dbItem.ViewCount,
			DistanceKM:          dbItem.DistanceKM,
		}

		// Unmarshal offer_products if present. Handle both raw JSON bytes and JSON-as-string.
		if len(dbItem.OfferProducts) > 0 {
			var offerProducts []response.OfferProduct
			if err := json.Unmarshal(dbItem.OfferProducts, &offerProducts); err != nil {
				// try interpreting as string containing JSON
				rawStr := string(dbItem.OfferProducts)
				if rawStr != "" {
					if err2 := json.Unmarshal([]byte(rawStr), &offerProducts); err2 != nil {
					} else {
						item.OfferProducts = offerProducts
					}
				}
			} else {
				item.OfferProducts = offerProducts
			}
		}

		// Unmarshal DynamicFields if present
		if len(dbItem.DynamicFields) > 0 {
			var dynamicFields map[string]interface{}
			if err := json.Unmarshal(dbItem.DynamicFields, &dynamicFields); err == nil {
				item.DynamicFields = dynamicFields
			}
		} else {
			item.DynamicFields = make(map[string]interface{})
		}

		products = append(products, item)
	}

	return
}

func (c *productDatabase) SaveDepartment(ctx context.Context, department request.Department) error {
	query := `INSERT INTO departments (name, slug, sort_order, is_active, image_url, icon) VALUES ($1, $2, $3, $4, $5, $6)`
	err := c.DB.Exec(query, department.Name, department.Slug, department.SortOrder, department.IsActive, department.ImageURL, department.Icon).Error
	return err
}

func (c *productDatabase) GetAllDepartments(ctx context.Context) (departments []response.Department, err error) {
	query := `SELECT id, name, image_url, icon FROM departments where is_active = true ORDER BY sort_order ASC`
	err = c.DB.Raw(query).Scan(&departments).Error
	return
}

func (c *productDatabase) GetDepartmentByID(ctx context.Context, brandID string) (department response.Department, err error) {
	query := `SELECT id, name, image_url, icon FROM departments WHERE id = $1`
	err = c.DB.Raw(query, brandID).Scan(&department).Error

	return
}

func (c *productDatabase) GetAllSubCategories(ctx context.Context) (subCategories []response.SubCategory, err error) {
	query := `SELECT * FROM sub_categories ORDER BY sort_order ASC`
	err = c.DB.Raw(query).Scan(&subCategories).Error

	return
}

func (c *productDatabase) GetAllCategoriesByDepartmentID(ctx context.Context, brandID string) (categories []response.Category, err error) {
	query := `SELECT id, name, image_url, icon FROM categories WHERE department_id = $1 ORDER BY sort_order ASC`
	err = c.DB.Raw(query, brandID).Scan(&categories).Error

	return
}

func (c *productDatabase) GetAllSubCategoriesByCategoryID(ctx context.Context, categoryID string) (subCategories []response.SubCategory, err error) {
	query := `SELECT id, name, image_url FROM sub_categories WHERE category_id = $1 ORDER BY sort_order ASC`
	err = c.DB.Raw(query, categoryID).Scan(&subCategories).Error

	return
}

// SaveSubTypeAttribute saves a new sub type attribute for a subcategory
func (c *productDatabase) SaveSubTypeAttribute(ctx context.Context, locationID string, attribute domain.SubTypeAttributes) error {
	attribute.SubCategoryID = locationID
	return c.DB.Create(&attribute).Error
}

// GetAllSubTypeAttributes retrieves all sub type attributes for a subcategory
func (c *productDatabase) GetAllSubTypeAttributes(ctx context.Context, locationID string) (attributes []response.SubTypeAttribute, err error) {
	query := `SELECT id, sub_category_id, field_name, field_type, is_required, sort_order 
	          FROM sub_type_attributes 
	          WHERE sub_category_id = $1 
	          ORDER BY sort_order ASC`
	err = c.DB.Raw(query, locationID).Scan(&attributes).Error
	return
}

// GetSubTypeAttributeByID retrieves a single sub type attribute by ID
func (c *productDatabase) GetSubTypeAttributeByID(ctx context.Context, attributeID string) (attribute response.SubTypeAttribute, err error) {
	query := `SELECT id, sub_category_id, field_name, field_type, is_required, sort_order 
	          FROM sub_type_attributes 
	          WHERE id = $1`
	err = c.DB.Raw(query, attributeID).Scan(&attribute).Error
	return
}

// SaveSubTypeAttributeOption saves a new option for a sub type attribute
func (c *productDatabase) SaveSubTypeAttributeOption(ctx context.Context, attributeID string, option domain.SubTypeAttributeOptions) error {
	option.SubTypeAttributeID = attributeID
	return c.DB.Create(&option).Error
}

// GetAllSubTypeAttributeOptions retrieves all options for a sub type attribute
func (c *productDatabase) GetAllSubTypeAttributeOptions(ctx context.Context, attributeID string) (options []response.SubTypeAttributeOption, err error) {
	query := `SELECT id, sub_type_attribute_id, option_value, sort_order 
	          FROM sub_type_attribute_options 
	          WHERE sub_type_attribute_id = $1 
	          ORDER BY sort_order ASC`
	err = c.DB.Raw(query, attributeID).Scan(&options).Error
	return
}

// GetSubTypeAttributeOptionByID retrieves a single option by ID
func (c *productDatabase) GetSubTypeAttributeOptionByID(ctx context.Context, optionID string) (option response.SubTypeAttributeOption, err error) {
	query := `SELECT id, sub_type_attribute_id, option_value, sort_order
	          FROM sub_type_attribute_options
	          WHERE id = $1`
	err = c.DB.Raw(query, optionID).Scan(&option).Error
	return
}

// UpdateSubTypeAttribute updates field_name, field_type, is_required, sort_order for a sub type attribute
func (c *productDatabase) UpdateSubTypeAttribute(ctx context.Context, attributeID string, attribute domain.SubTypeAttributes) error {
	query := `UPDATE sub_type_attributes SET field_name = $1, field_type = $2, is_required = $3, sort_order = $4 WHERE id = $5`
	return c.DB.Exec(query, attribute.FieldName, attribute.FieldType, attribute.IsRequired, attribute.SortOrder, attributeID).Error
}

// DeleteSubTypeAttribute deletes a sub type attribute and its options
func (c *productDatabase) DeleteSubTypeAttribute(ctx context.Context, attributeID string) error {
	if err := c.DB.Exec(`DELETE FROM sub_type_attribute_options WHERE sub_type_attribute_id = $1`, attributeID).Error; err != nil {
		return err
	}
	return c.DB.Exec(`DELETE FROM sub_type_attributes WHERE id = $1`, attributeID).Error
}

// UpdateSubTypeAttributeOption updates option_value and sort_order for a sub type attribute option
func (c *productDatabase) UpdateSubTypeAttributeOption(ctx context.Context, optionID string, option domain.SubTypeAttributeOptions) error {
	query := `UPDATE sub_type_attribute_options SET option_value = $1, sort_order = $2 WHERE id = $3`
	return c.DB.Exec(query, option.OptionValue, option.SortOrder, optionID).Error
}

// DeleteSubTypeAttributeOption deletes a single sub type attribute option
func (c *productDatabase) DeleteSubTypeAttributeOption(ctx context.Context, optionID string) error {
	return c.DB.Exec(`DELETE FROM sub_type_attribute_options WHERE id = $1`, optionID).Error
}

// SaveCategoryImage saves a new category image
func (c *productDatabase) SaveCategoryImage(ctx context.Context, categoryID string, image domain.CategoryImage) error {
	image.CategoryID = categoryID
	query := `INSERT INTO category_images (id, category_id, image_url, alt_text, sort_order, is_active)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	return c.DB.Exec(query, domain.NewID(domain.PrefixCategoryImage), image.CategoryID, image.ImageURL, image.AltText, image.SortOrder, image.IsActive).Error
}

// GetAllCategoryImages retrieves all images for a category
func (c *productDatabase) GetAllCategoryImages(ctx context.Context, categoryID string) (images []response.CategoryImage, err error) {
	query := `SELECT id, category_id, image_url, alt_text, sort_order, is_active 
	          FROM category_images 
	          WHERE category_id = $1 AND is_active = true 
	          ORDER BY sort_order ASC, id ASC`
	err = c.DB.Raw(query, categoryID).Scan(&images).Error
	return
}

// GetCategoryImageByID retrieves a single category image by ID
func (c *productDatabase) GetCategoryImageByID(ctx context.Context, imageID string) (image response.CategoryImage, err error) {
	query := `SELECT id, category_id, image_url, alt_text, sort_order, is_active 
	          FROM category_images 
	          WHERE id = $1`
	err = c.DB.Raw(query, imageID).Scan(&image).Error
	return
}

// UpdateCategoryImage updates an existing category image
func (c *productDatabase) UpdateCategoryImage(ctx context.Context, image domain.CategoryImage) error {
	query := `UPDATE category_images 
	          SET image_url = $1, alt_text = $2, sort_order = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $5`
	return c.DB.Exec(query, image.ImageURL, image.AltText, image.SortOrder, image.IsActive, image.ID).Error
}

// DeleteCategoryImage soft deletes a category image
func (c *productDatabase) DeleteCategoryImage(ctx context.Context, imageID string) error {
	query := `UPDATE category_images SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	return c.DB.Exec(query, imageID).Error
}

func (c *productDatabase) GetProductItemByID(ctx context.Context, productItemID string, customerView bool) (productItem response.ProductItems, err error) {
	query := `SELECT pi.id, pi.sub_category_name, pi.category_id, pi.department_id, pi.sub_category_id, pi.stock,
	           sc.name AS category_name, mc.name AS main_category_name,
	           pi.dynamic_fields, pi.created_at, pi.updated_at, pi.shop_id,
	           (SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = pi.id) AS view_count,
	           ` + IsSubscribedColumn + `
	       FROM product_items pi
	       LEFT JOIN categories sc ON pi.category_id = sc.id
	       LEFT JOIN categories mc ON pi.category_id = mc.id
	       LEFT JOIN shop_details sd ON sd.id = pi.shop_id
	       WHERE pi.id = $1;`

	// Item detail is intentionally NOT hidden when the seller is unsubscribed: a
	// customer must still be able to open a PAST enquiry (which loads the product).
	// Discovery lists (search / shop-products) hide the product; here we only expose
	// is_subscribed so the client can disable engage actions (Ask for Price, negotiate).
	_ = customerView

	var dbItem struct {
		ID               string
		SubCategoryName  string
		CategoryID       string
		DepartmentID     string
		SubCategoryID    string
		CategoryName     string
		MainCategoryName string
		Stock            bool
		DynamicFields    []byte
		Description      string
		Highlights       string
		CreatedAt        time.Time
		UpdatedAt        time.Time
		ViewCount        uint
		ShopId           string
		IsSubscribed     bool
	}

	err = c.DB.Raw(query, productItemID).Scan(&dbItem).Error
	if err != nil {
		return
	}
	// Raw().Scan() yields no error for zero rows; treat an empty result as
	// not-found so the handler can return 404 (also the gated-out case above).
	if dbItem.ID == "" {
		return productItem, gorm.ErrRecordNotFound
	}

	productItem.ID = dbItem.ID
	productItem.Name = dbItem.SubCategoryName
	productItem.CategoryID = dbItem.CategoryID
	productItem.DepartmentID = dbItem.DepartmentID
	productItem.SubCategoryID = dbItem.SubCategoryID
	productItem.CategoryName = dbItem.CategoryName
	productItem.MainCategoryName = dbItem.MainCategoryName
	productItem.Stock = dbItem.Stock
	productItem.Description = dbItem.Description
	productItem.CreatedAt = dbItem.CreatedAt
	productItem.UpdatedAt = dbItem.UpdatedAt
	productItem.ViewCount = dbItem.ViewCount
	productItem.ShopID = dbItem.ShopId
	productItem.IsSubscribed = dbItem.IsSubscribed

	// Parse highlights PostgreSQL array format
	if dbItem.Highlights != "" {
		h := dbItem.Highlights
		if len(h) > 2 && h[0] == '{' && h[len(h)-1] == '}' {
			h = h[1 : len(h)-1]
		}
		if h != "" {
			for _, part := range strings.Split(h, ",") {
				productItem.Highlights = append(productItem.Highlights, strings.Trim(strings.TrimSpace(part), `"`))
			}
		}
	}

	// Fetch raw image keys from product_items.product_item_images
	images, imgErr := c.FindAllProductItemImages(ctx, dbItem.ID)
	if imgErr != nil {
		productItem.ProductItemImages = []string{}
	} else {
		productItem.ProductItemImages = images
	}

	// Unmarshal DynamicFields JSONB to map
	var dynamicFields map[string]interface{}
	if len(dbItem.DynamicFields) > 0 {
		err = json.Unmarshal(dbItem.DynamicFields, &dynamicFields)
		if err != nil {
			return
		}
		productItem.DynamicFields = dynamicFields
	} else {
		productItem.DynamicFields = make(map[string]interface{})
	}

	// Fetch offers from promotions table for this product item
	offerQuery := `SELECT op.id as offer_product_id, pi.sub_category_name as product_name,
	                  p.id as offer_id, p.offer_name, p.discount_rate, p.description,
	                  p.start_date, p.end_date,
	                  pc.id as promotion_category_id, pc.name as promotion_category_name, pc.shop_id as promotion_category_shop_id, pc.is_active as promotion_category_is_active, pc.icon_path as promotion_category_icon_path, pc.created_at as promotion_category_created_at, pc.updated_at as promotion_category_updated_at,
	                  pt.id as promotion_type_id, pt.name as promotion_type_name, pt.is_active as promotion_type_is_active, pt.shop_id as promotion_type_shop_id, pt.promotion_category_id as promotion_type_promotion_category_id, pt.icon_path as promotion_type_icon_path, pt.created_at as promotion_type_created_at, pt.updated_at as promotion_type_updated_at
	               FROM offer_products op
	               INNER JOIN promotions p ON p.id = op.offer_id
	               LEFT JOIN product_items pi ON pi.id = op.product_item_id
	               LEFT JOIN promotion_categories pc ON p.promotion_category_id = pc.id
	               LEFT JOIN promotions_types pt ON p.promotion_type_id = pt.id
	               WHERE op.product_item_id = $1 AND p.is_active = true`

	var offerRows []struct {
		OfferProductID                   string    `gorm:"column:offer_product_id"`
		ProductName                      string    `gorm:"column:product_name"`
		OfferID                          string    `gorm:"column:offer_id"`
		OfferName                        string    `gorm:"column:offer_name"`
		DiscountRate                     uint      `gorm:"column:discount_rate"`
		Description                      string    `gorm:"column:description"`
		StartDate                        string    `gorm:"column:start_date"`
		EndDate                          string    `gorm:"column:end_date"`
		PromotionCategoryID              string    `gorm:"column:promotion_category_id"`
		PromotionCategoryName            string    `gorm:"column:promotion_category_name"`
		PromotionCategoryShopID          string    `gorm:"column:promotion_category_shop_id"`
		PromotionCategoryIsActive        bool      `gorm:"column:promotion_category_is_active"`
		PromotionCategoryIconPath        string    `gorm:"column:promotion_category_icon_path"`
		PromotionCategoryCreatedAt       time.Time `gorm:"column:promotion_category_created_at"`
		PromotionCategoryUpdatedAt       time.Time `gorm:"column:promotion_category_updated_at"`
		PromotionTypeID                  string    `gorm:"column:promotion_type_id"`
		PromotionTypeName                string    `gorm:"column:promotion_type_name"`
		PromotionTypeIsActive            bool      `gorm:"column:promotion_type_is_active"`
		PromotionTypeShopID              string    `gorm:"column:promotion_type_shop_id"`
		PromotionTypePromotionCategoryID string    `gorm:"column:promotion_type_promotion_category_id"`
		PromotionTypeType                string    `gorm:"column:promotion_type_type"`
		PromotionTypeIconPath            string    `gorm:"column:promotion_type_icon_path"`
		PromotionTypeCreatedAt           time.Time `gorm:"column:promotion_type_created_at"`
		PromotionTypeUpdatedAt           time.Time `gorm:"column:promotion_type_updated_at"`
	}

	err = c.DB.Raw(offerQuery, productItemID).Scan(&offerRows).Error
	if err != nil {
		// If there's an error fetching offers, continue without them
		productItem.OfferProducts = []response.OfferProduct{}
	} else {
		// Convert to response format
		offerProducts := make([]response.OfferProduct, len(offerRows))
		for i, row := range offerRows {
			offerProducts[i] = response.OfferProduct{
				OfferProductID: row.OfferProductID,
				ProductName:    row.ProductName,
				OfferID:        row.OfferID,
				OfferName:      row.OfferName,
				DiscountRate:   row.DiscountRate,
				Description:    row.Description,
				StartDate:      row.StartDate,
				EndDate:        row.EndDate,
				PromotionCategory: response.PromotionCategory{
					ID:        row.PromotionCategoryID,
					Name:      row.PromotionCategoryName,
					ShopID:    row.PromotionCategoryShopID,
					IsActive:  row.PromotionCategoryIsActive,
					IconPath:  row.PromotionCategoryIconPath,
					CreatedAt: row.PromotionCategoryCreatedAt,
					UpdatedAt: row.PromotionCategoryUpdatedAt,
				},
				PromotionType: response.PromotionsType{
					ID:                  row.PromotionTypeID,
					Name:                row.PromotionTypeName,
					IsActive:            row.PromotionTypeIsActive,
					ShopID:              row.PromotionTypeShopID,
					PromotionCategoryID: row.PromotionTypePromotionCategoryID,
					Type:                row.PromotionTypeType,
					IconPath:            row.PromotionTypeIconPath,
					CreatedAt:           row.PromotionTypeCreatedAt,
					UpdatedAt:           row.PromotionTypeUpdatedAt,
				},
			}
		}
		productItem.OfferProducts = offerProducts
	}

	return
}

func (c *productDatabase) IncrementProductItemViewCount(ctx context.Context, productItemID string, adminID string) error {
	// Resolve the owning shop from the product item itself.
	// This keeps view rows tied to the product owner while deduping by viewer.
	var shopID string
	shopQuery := `SELECT shop_id FROM product_items WHERE id = $1`
	err := c.DB.Raw(shopQuery, productItemID).Scan(&shopID).Error
	if err != nil {
		return err
	}
	if shopID == "" {
		return fmt.Errorf("product item %s has no shop_id", productItemID)
	}

	// Per-user, per-product dedupe: if a row already exists for this viewer,
	// do not increment view_count again.
	updateQuery := `UPDATE product_item_views
		SET viewed_at = CURRENT_TIMESTAMP
		WHERE product_item_id = $1
		  AND shop_id = $2
		  AND admin_id = $3`
	result := c.DB.Exec(updateQuery, productItemID, shopID, adminID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	insertQuery := `INSERT INTO product_item_views (id, product_item_id, shop_id, admin_id, view_count, created_at, viewed_at)
		VALUES ($1, $2, $3, $4, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	return c.DB.Exec(insertQuery, domain.NewID(domain.PrefixProductItemView), productItemID, shopID, adminID).Error
}

func (c *productDatabase) GetProductItemViewCount(ctx context.Context, productItemID string, adminID string) (viewCount uint, err error) {
	// Get shop ID using admin ID
	var shopID string
	shopQuery := `SELECT id FROM shop_details WHERE admin_id = $1`
	err = c.DB.Raw(shopQuery, adminID).Scan(&shopID).Error
	if err != nil {
		return 0, err
	}
	if shopID == "" {
		return 0, nil
	}

	query := `SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = $1 AND shop_id = $2`
	err = c.DB.Raw(query, productItemID, shopID).Scan(&viewCount).Error
	return
}

func (c *productDatabase) FindProductItemsByDocument(ctx context.Context, documentID uint) (productItems []response.ProductItems, err error) {
	query := `SELECT pi.sub_category_name, pi.id, pi.category_id, 
	           sc.name AS category_name, mc.name AS main_category_name,
	           pi.product_item_images, pi.dynamic_fields, pi.created_at, pi.updated_at,
	           (SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = pi.id) AS view_count
	       FROM product_items pi 
	       LEFT JOIN categories sc ON pi.category_id = sc.id 
	       LEFT JOIN categories mc ON pi.category_id = mc.id 
	       INNER JOIN document_product_items dpi ON pi.id = dpi.product_item_id
	       WHERE dpi.document_id = $1;`

	type productItemDB struct {
		Name              string    `gorm:"column:sub_category_name"`
		ID                string    `gorm:"column:id"`
		CategoryID        string    `gorm:"column:category_id"`
		CategoryName      string    `gorm:"column:category_name"`
		MainCategoryName  string    `gorm:"column:main_category_name"`
		ProductItemImages string    `gorm:"column:product_item_images"`
		DynamicFields     []byte    `gorm:"column:dynamic_fields"`
		CreatedAt         time.Time `gorm:"column:created_at"`
		UpdatedAt         time.Time `gorm:"column:updated_at"`
		ViewCount         uint      `gorm:"column:view_count"`
	}

	var dbItems []productItemDB
	err = c.DB.Raw(query, documentID).Scan(&dbItems).Error
	if err != nil {
		return
	}

	// Map to response.ProductItems
	for _, dbItem := range dbItems {
		// Parse product_item_images from PostgreSQL array format
		var images []string
		if dbItem.ProductItemImages != "" {
			imageStr := dbItem.ProductItemImages
			if len(imageStr) > 2 && imageStr[0] == '{' && imageStr[len(imageStr)-1] == '}' {
				imageStr = imageStr[1 : len(imageStr)-1]
				if imageStr != "" {
					images = []string{}
					for _, img := range parsePostgresArray(imageStr) {
						images = append(images, img)
					}
				}
			}
		}
		if images == nil {
			images = []string{}
		}

		item := response.ProductItems{
			ID:                dbItem.ID,
			Name:              dbItem.Name,
			CategoryName:      dbItem.CategoryName,
			MainCategoryName:  dbItem.MainCategoryName,
			ProductItemImages: images,
			CreatedAt:         dbItem.CreatedAt,
			UpdatedAt:         dbItem.UpdatedAt,
			ViewCount:         dbItem.ViewCount,
		}

		// Unmarshal DynamicFields if present
		if len(dbItem.DynamicFields) > 0 {
			var dynamicFields map[string]interface{}
			if err := json.Unmarshal(dbItem.DynamicFields, &dynamicFields); err == nil {
				item.DynamicFields = dynamicFields
			}
		} else {
			item.DynamicFields = make(map[string]interface{})
		}

		productItems = append(productItems, item)
	}

	return
}

// GetProductItemsByDepartment returns product items for the department id provided.
// It joins shop_details to map the document/shop id to the admin_id stored on product_items.
func (c *productDatabase) GetProductItemsByDepartment(ctx context.Context, brandID string) (productItems []response.ProductItems, err error) {
	query := `SELECT pi.sub_category_name, pi.id, pi.category_id, pi.department_id, pi.sub_category_id,
				pi.product_item_images, pi.dynamic_fields, pi.created_at, pi.updated_at,
				c.name AS category_name, d.name AS department_name, sc.name AS sub_category_name_ref,
				sc.image_url AS sub_category_image_url,
				(SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = pi.id) AS view_count,
				(
					SELECT COALESCE(json_agg(json_build_object(
						'offer_product_id', op2.id,
						'product_name', pi2.sub_category_name,
						'offer_id', o2.id,
						'offer_name', o2.name,
						'discount_rate', o2.discount_rate,
						'description', o2.description,
						'start_date', o2.start_date,
						'end_date', o2.end_date,
						'image', o2.image,
						'thumbnail', o2.thumbnail
					)), '[]')
					FROM offer_products op2
					LEFT JOIN product_items pi2 ON pi2.id = op2.product_item_id
					INNER JOIN offers o2 ON o2.id = op2.offer_id

					WHERE op2.product_item_id = pi.id
				) AS offer_products
			FROM product_items pi
			LEFT JOIN categories c ON pi.category_id = c.id
			LEFT JOIN departments d ON pi.department_id = d.id
			LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
			WHERE pi.department_id = $1
			ORDER BY pi.created_at DESC`

	// Internal struct mirrors FindAllProductItems scanning for reuse
	type productItemDB struct {
		Name                string        `gorm:"column:sub_category_name"`
		ID                  string        `gorm:"column:id"`
		CategoryID          string        `gorm:"column:category_id"`
		DepartmentID        string        `gorm:"column:department_id"`
		SubCategoryID       string        `gorm:"column:sub_category_id"`
		CategoryName        string        `gorm:"column:category_name"`
		DepartmentName      string        `gorm:"column:department_name"`
		SubCategoryNameRef  string        `gorm:"column:sub_category_name_ref"`
		SubCategoryImageURL string        `gorm:"column:sub_category_image_url"`
		ProductItemImages   string        `gorm:"column:product_item_images"`
		DynamicFields       []byte        `gorm:"column:dynamic_fields"`
		OfferProducts       []byte        `gorm:"column:offer_products"`
		DiscountRate        sql.NullInt64 `gorm:"column:discount_rate"`
		CreatedAt           time.Time     `gorm:"column:created_at"`
		UpdatedAt           time.Time     `gorm:"column:updated_at"`
		ViewCount           uint          `gorm:"column:view_count"`
	}

	var dbItems []productItemDB
	err = c.DB.Raw(query, brandID).Scan(&dbItems).Error
	if err != nil {
		return
	}

	for _, dbItem := range dbItems {
		// reuse the same parsing/mapping logic as FindAllProductItems
		var images []string
		if dbItem.ProductItemImages != "" {
			imageStr := dbItem.ProductItemImages
			if len(imageStr) > 2 && imageStr[0] == '{' && imageStr[len(imageStr)-1] == '}' {
				imageStr = imageStr[1 : len(imageStr)-1]
				if imageStr != "" {
					images = []string{}
					for _, img := range parsePostgresArray(imageStr) {
						images = append(images, img)
					}
				}
			}
		}
		if images == nil {
			images = []string{}
		}

		var offerProducts []response.OfferProduct
		if len(dbItem.OfferProducts) > 0 {
			if err := json.Unmarshal(dbItem.OfferProducts, &offerProducts); err != nil {
				rawStr := string(dbItem.OfferProducts)
				if rawStr != "" {
					_ = json.Unmarshal([]byte(rawStr), &offerProducts)
				}
			}
		}

		var discountRatePtr *uint
		if dbItem.DiscountRate.Valid {
			val := uint(dbItem.DiscountRate.Int64)
			discountRatePtr = &val
		} else {
			discountRatePtr = nil
		}
		item := response.ProductItems{
			ID:                  dbItem.ID,
			Name:                dbItem.Name,
			CategoryName:        dbItem.CategoryName,
			MainCategoryName:    dbItem.DepartmentName,
			SubCategoryImageURL: dbItem.SubCategoryImageURL,
			CategoryID:          dbItem.CategoryID,
			DepartmentID:        dbItem.DepartmentID,
			SubCategoryID:       dbItem.SubCategoryID,
			ProductItemImages:   images,
			OfferProducts:       offerProducts,
			DiscountRate:        discountRatePtr,
			CreatedAt:           dbItem.CreatedAt,
			UpdatedAt:           dbItem.UpdatedAt,
			ViewCount:           dbItem.ViewCount,
		}

		if len(dbItem.DynamicFields) > 0 {
			var dynamicFields map[string]interface{}
			if err := json.Unmarshal(dbItem.DynamicFields, &dynamicFields); err == nil {
				item.DynamicFields = dynamicFields
			}
		} else {
			item.DynamicFields = make(map[string]interface{})
		}

		productItems = append(productItems, item)
	}

	return
}

// GetProductItemsByCategory returns product items for the category id provided.
func (c *productDatabase) GetProductItemsByCategory(ctx context.Context, categoryID string) (productItems []response.ProductItems, err error) {
	query := `SELECT pi.sub_category_name, pi.id, pi.category_id, pi.department_id, pi.sub_category_id,
				pi.product_item_images, pi.dynamic_fields, pi.created_at, pi.updated_at,
				c.name AS category_name, d.name AS department_name, sc.name AS sub_category_name_ref,
				sc.image_url AS sub_category_image_url,
				(SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = pi.id) AS view_count,
				(
					SELECT COALESCE(json_agg(json_build_object(
						'offer_product_id', op2.id,
						'product_name', pi2.sub_category_name,
						'offer_id', o2.id,
						'offer_name', o2.name,
						'discount_rate', o2.discount_rate,
						'description', o2.description,
						'start_date', o2.start_date,
						'end_date', o2.end_date,
						'image', o2.image,
						'thumbnail', o2.thumbnail
					)), '[]')
					FROM offer_products op2
					LEFT JOIN product_items pi2 ON pi2.id = op2.product_item_id
					INNER JOIN offers o2 ON o2.id = op2.offer_id

					WHERE op2.product_item_id = pi.id
				) AS offer_products
			FROM product_items pi
			LEFT JOIN categories c ON pi.category_id = c.id
			LEFT JOIN departments d ON pi.department_id = d.id
			LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
			WHERE pi.category_id = $1
			ORDER BY pi.created_at DESC`

	type productItemDB struct {
		Name                string    `gorm:"column:sub_category_name"`
		ID                  string    `gorm:"column:id"`
		CategoryID          string    `gorm:"column:category_id"`
		DepartmentID        string    `gorm:"column:department_id"`
		SubCategoryID       string    `gorm:"column:sub_category_id"`
		CategoryName        string    `gorm:"column:category_name"`
		DepartmentName      string    `gorm:"column:department_name"`
		SubCategoryNameRef  string    `gorm:"column:sub_category_name_ref"`
		SubCategoryImageURL string    `gorm:"column:sub_category_image_url"`
		ProductItemImages   string    `gorm:"column:product_item_images"`
		DynamicFields       []byte    `gorm:"column:dynamic_fields"`
		OfferProducts       []byte    `gorm:"column:offer_products"`
		CreatedAt           time.Time `gorm:"column:created_at"`
		UpdatedAt           time.Time `gorm:"column:updated_at"`
		ViewCount           uint      `gorm:"column:view_count"`
	}

	var dbItems []productItemDB
	err = c.DB.Raw(query, categoryID).Scan(&dbItems).Error
	if err != nil {
		return
	}

	for _, dbItem := range dbItems {
		var images []string
		if dbItem.ProductItemImages != "" {
			imageStr := dbItem.ProductItemImages
			if len(imageStr) > 2 && imageStr[0] == '{' && imageStr[len(imageStr)-1] == '}' {
				imageStr = imageStr[1 : len(imageStr)-1]
				if imageStr != "" {
					images = []string{}
					for _, img := range parsePostgresArray(imageStr) {
						images = append(images, img)
					}
				}
			}
		}
		if images == nil {
			images = []string{}
		}

		item := response.ProductItems{
			ID:                  dbItem.ID,
			Name:                dbItem.Name,
			CategoryName:        dbItem.CategoryName,
			MainCategoryName:    dbItem.DepartmentName,
			SubCategoryImageURL: dbItem.SubCategoryImageURL,
			CategoryID:          dbItem.CategoryID,
			DepartmentID:        dbItem.DepartmentID,
			SubCategoryID:       dbItem.SubCategoryID,
			ProductItemImages:   images,
			CreatedAt:           dbItem.CreatedAt,
			UpdatedAt:           dbItem.UpdatedAt,
		}

		if len(dbItem.OfferProducts) > 0 {
			var offerProducts []response.OfferProduct
			if err := json.Unmarshal(dbItem.OfferProducts, &offerProducts); err != nil {
				rawStr := string(dbItem.OfferProducts)
				if rawStr != "" {
					if err2 := json.Unmarshal([]byte(rawStr), &offerProducts); err2 == nil {
						item.OfferProducts = offerProducts
					}
				}
			} else {
				item.OfferProducts = offerProducts
			}
		}

		if len(dbItem.DynamicFields) > 0 {
			var dynamicFields map[string]interface{}
			if err := json.Unmarshal(dbItem.DynamicFields, &dynamicFields); err == nil {
				item.DynamicFields = dynamicFields
			}
		} else {
			item.DynamicFields = make(map[string]interface{})
		}

		productItems = append(productItems, item)
	}

	return
}

// GetProductItemsBySubCategory returns product items for the sub-category id provided.
func (c *productDatabase) GetProductItemsBySubCategory(ctx context.Context, locationID string) (productItems []response.ProductItems, err error) {
	query := `SELECT pi.sub_category_name, pi.id, pi.category_id, pi.department_id, pi.sub_category_id,
				pi.product_item_images, pi.dynamic_fields, pi.created_at, pi.updated_at,
				c.name AS category_name, d.name AS department_name, sc.name AS sub_category_name_ref,
				sc.image_url AS sub_category_image_url,
				(SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = pi.id) AS view_count,
				(
					SELECT COALESCE(json_agg(json_build_object(
						'offer_product_id', op2.id,
						'product_name', pi2.sub_category_name,
						'offer_id', o2.id,
						'offer_name', o2.name,
						'discount_rate', o2.discount_rate,
						'description', o2.description,
						'start_date', o2.start_date,
						'end_date', o2.end_date,
						'image', o2.image,
						'thumbnail', o2.thumbnail
					)), '[]')
					FROM offer_products op2
					LEFT JOIN product_items pi2 ON pi2.id = op2.product_item_id
					INNER JOIN offers o2 ON o2.id = op2.offer_id

					WHERE op2.product_item_id = pi.id
				) AS offer_products
			FROM product_items pi
			LEFT JOIN categories c ON pi.category_id = c.id
			LEFT JOIN departments d ON pi.department_id = d.id
			LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
			WHERE pi.sub_category_id = $1
			ORDER BY pi.created_at DESC`

	type productItemDB struct {
		Name                string    `gorm:"column:sub_category_name"`
		ID                  string    `gorm:"column:id"`
		CategoryID          string    `gorm:"column:category_id"`
		DepartmentID        string    `gorm:"column:department_id"`
		SubCategoryID       string    `gorm:"column:sub_category_id"`
		CategoryName        string    `gorm:"column:category_name"`
		DepartmentName      string    `gorm:"column:department_name"`
		SubCategoryNameRef  string    `gorm:"column:sub_category_name_ref"`
		SubCategoryImageURL string    `gorm:"column:sub_category_image_url"`
		ProductItemImages   string    `gorm:"column:product_item_images"`
		DynamicFields       []byte    `gorm:"column:dynamic_fields"`
		OfferProducts       []byte    `gorm:"column:offer_products"`
		CreatedAt           time.Time `gorm:"column:created_at"`
		UpdatedAt           time.Time `gorm:"column:updated_at"`
		ViewCount           uint      `gorm:"column:view_count"`
	}

	var dbItems []productItemDB
	err = c.DB.Raw(query, locationID).Scan(&dbItems).Error
	if err != nil {
		return
	}

	for _, dbItem := range dbItems {
		var images []string
		if dbItem.ProductItemImages != "" {
			imageStr := dbItem.ProductItemImages
			if len(imageStr) > 2 && imageStr[0] == '{' && imageStr[len(imageStr)-1] == '}' {
				imageStr = imageStr[1 : len(imageStr)-1]
				if imageStr != "" {
					images = []string{}
					for _, img := range parsePostgresArray(imageStr) {
						images = append(images, img)
					}
				}
			}
		}
		if images == nil {
			images = []string{}
		}

		item := response.ProductItems{
			ID:                  dbItem.ID,
			Name:                dbItem.Name,
			CategoryName:        dbItem.CategoryName,
			MainCategoryName:    dbItem.DepartmentName,
			SubCategoryImageURL: dbItem.SubCategoryImageURL,
			CategoryID:          dbItem.CategoryID,
			DepartmentID:        dbItem.DepartmentID,
			SubCategoryID:       dbItem.SubCategoryID,
			ProductItemImages:   images,
			CreatedAt:           dbItem.CreatedAt,
			UpdatedAt:           dbItem.UpdatedAt,
		}

		if len(dbItem.OfferProducts) > 0 {
			var offerProducts []response.OfferProduct
			if err := json.Unmarshal(dbItem.OfferProducts, &offerProducts); err != nil {
				rawStr := string(dbItem.OfferProducts)
				if rawStr != "" {
					if err2 := json.Unmarshal([]byte(rawStr), &offerProducts); err2 == nil {
						item.OfferProducts = offerProducts
					}
				}
			} else {
				item.OfferProducts = offerProducts
			}
		}

		if len(dbItem.DynamicFields) > 0 {
			var dynamicFields map[string]interface{}
			if err := json.Unmarshal(dbItem.DynamicFields, &dynamicFields); err == nil {
				item.DynamicFields = dynamicFields
			}
		} else {
			item.DynamicFields = make(map[string]interface{})
		}

		productItems = append(productItems, item)
	}

	return
}

// GetProductItemsByShop returns product items for the shop owned by the provided admin id.
// It joins shop_details to find the shop id for the admin and matches product_items.admin_id to shop id.
func (c *productDatabase) GetProductItemsByShop(ctx context.Context, adminID string) (productItems []response.ProductItems, err error) {
	query := `SELECT pi.sub_category_name, pi.id, pi.category_id, pi.department_id, pi.sub_category_id,
				pi.product_item_images, pi.dynamic_fields, pi.created_at, pi.updated_at,
				c.name AS category_name, d.name AS department_name, sc.name AS sub_category_name_ref,
				sc.image_url AS sub_category_image_url,
				(SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = pi.id) AS view_count,
				(
					SELECT COALESCE(json_agg(json_build_object(
						'offer_product_id', op2.id,
						'product_name', pi2.sub_category_name,
						'offer_id', o2.id,
						'offer_name', o2.name,
						'discount_rate', o2.discount_rate,
						'description', o2.description,
						'start_date', o2.start_date,
						'end_date', o2.end_date,
						'image', o2.image,
						'thumbnail', o2.thumbnail
					)), '[]')
					FROM offer_products op2
					LEFT JOIN product_items pi2 ON pi2.id = op2.product_item_id
					INNER JOIN offers o2 ON o2.id = op2.offer_id

					WHERE op2.product_item_id = pi.id AND o2.start_date <= CURRENT_TIMESTAMP AND o2.end_date >= CURRENT_TIMESTAMP
				) AS offer_products
			FROM product_items pi
			LEFT JOIN categories c ON pi.category_id = c.id
			LEFT JOIN departments d ON pi.department_id = d.id
			LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
			-- Match product_items.admin_id robustly against provided admin id.
			-- product_items.admin_id is jsonb and may store either an admin id or a shop id.
			WHERE (
				-- direct jsonb equality: admin_id = to_jsonb($1)
				pi.admin_id = to_jsonb($1)
				-- or textual match of jsonb value
				OR pi.admin_id::text = to_jsonb($1)::text
				-- or the admin owns a shop whose id matches the product_items.admin_id
				OR EXISTS (
					SELECT 1 FROM shop_details sd WHERE sd.admin_id = $1 AND sd.id::text = pi.admin_id::text
				)
			)
			ORDER BY pi.created_at DESC`

	type productItemDB struct {
		Name                string    `gorm:"column:sub_category_name"`
		ID                  string    `gorm:"column:id"`
		CategoryID          string    `gorm:"column:category_id"`
		DepartmentID        string    `gorm:"column:department_id"`
		SubCategoryID       string    `gorm:"column:sub_category_id"`
		CategoryName        string    `gorm:"column:category_name"`
		DepartmentName      string    `gorm:"column:department_name"`
		SubCategoryNameRef  string    `gorm:"column:sub_category_name_ref"`
		SubCategoryImageURL string    `gorm:"column:sub_category_image_url"`
		ProductItemImages   string    `gorm:"column:product_item_images"`
		DynamicFields       []byte    `gorm:"column:dynamic_fields"`
		OfferProducts       []byte    `gorm:"column:offer_products"`
		CreatedAt           time.Time `gorm:"column:created_at"`
		UpdatedAt           time.Time `gorm:"column:updated_at"`
		ViewCount           uint      `gorm:"column:view_count"`
	}

	// Log SQL and parameter to aid debugging when API returns no rows
	log.Printf("GetProductItemsByShop SQL: %s", query)
	log.Printf("GetProductItemsByShop adminID param: %s", adminID)

	var dbItems []productItemDB
	err = c.DB.Raw(query, adminID).Scan(&dbItems).Error
	if err != nil {
		return
	}

	log.Printf("DEBUG: dbItems count=%d", len(dbItems))

	for _, dbItem := range dbItems {
		var images []string
		if dbItem.ProductItemImages != "" {
			imageStr := dbItem.ProductItemImages
			if len(imageStr) > 2 && imageStr[0] == '{' && imageStr[len(imageStr)-1] == '}' {
				imageStr = imageStr[1 : len(imageStr)-1]
				if imageStr != "" {
					images = []string{}
					for _, img := range parsePostgresArray(imageStr) {
						images = append(images, img)
					}
				}
			}
		}
		if images == nil {
			images = []string{}
		}

		item := response.ProductItems{
			ID:                  dbItem.ID,
			Name:                dbItem.Name,
			CategoryName:        dbItem.CategoryName,
			MainCategoryName:    dbItem.DepartmentName,
			SubCategoryImageURL: dbItem.SubCategoryImageURL,
			CategoryID:          dbItem.CategoryID,
			DepartmentID:        dbItem.DepartmentID,
			SubCategoryID:       dbItem.SubCategoryID,
			ProductItemImages:   images,
			CreatedAt:           dbItem.CreatedAt,
			UpdatedAt:           dbItem.UpdatedAt,
			ViewCount:           dbItem.ViewCount,
		}

		if len(dbItem.OfferProducts) > 0 {
			var offerProducts []response.OfferProduct
			if err := json.Unmarshal(dbItem.OfferProducts, &offerProducts); err != nil {
				rawStr := string(dbItem.OfferProducts)
				if rawStr != "" {
					if err2 := json.Unmarshal([]byte(rawStr), &offerProducts); err2 == nil {
						item.OfferProducts = offerProducts
					}
				}
			} else {
				item.OfferProducts = offerProducts
			}
		}

		if len(dbItem.DynamicFields) > 0 {
			var dynamicFields map[string]interface{}
			if err := json.Unmarshal(dbItem.DynamicFields, &dynamicFields); err == nil {
				item.DynamicFields = dynamicFields
			}
		} else {
			item.DynamicFields = make(map[string]interface{})
		}

		productItems = append(productItems, item)
	}

	return
}

func (c *productDatabase) FindProductItemFilters(ctx context.Context, adminID string, shopID string) ([]domain.ProductItemFilterType, error) {
	shopQuery := `SELECT id FROM shop_details WHERE id = $1`
	err := c.DB.Raw(shopQuery, shopID).Scan(&shopID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shop ID for admin %s: %v", adminID, err)
	}

	var filters []domain.ProductItemFilterType
	query := `SELECT id, filter_name, shop_id FROM product_item_filter_types where shop_id = $1 ORDER BY id ASC`
	err = c.DB.Raw(query, shopID).Scan(&filters).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch product item filters: %v", err)
	}

	// Get products for the admin
	var products []struct {
		CategoryID string `json:"category_id"`
	}
	productQuery := `SELECT category_id FROM product_items WHERE shop_id = $1`
	err = c.DB.Raw(productQuery, shopID).Scan(&products).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %v", err)
	}

	// Get unique category IDs
	categoryIDMap := make(map[string]bool)
	for _, p := range products {
		categoryIDMap[p.CategoryID] = true
	}

	// Get category names and ids
	var categories []struct {
		ID   string
		Name string
	}
	if len(categoryIDMap) > 0 {
		categoryIDs := make([]string, 0, len(categoryIDMap))
		for id := range categoryIDMap {
			categoryIDs = append(categoryIDs, id)
		}
		categoryQuery := `SELECT id, name FROM categories WHERE id IN (?)`
		err = c.DB.Raw(categoryQuery, categoryIDs).Scan(&categories).Error
		if err != nil {
			return nil, fmt.Errorf("failed to fetch categories: %v", err)
		}
	}

	// Append categories as filter types
	for _, cat := range categories {
		filters = append(filters, domain.ProductItemFilterType{
			ID:         cat.ID,
			FilterName: cat.Name,
			ShopID:     shopID,
		})
	}

	return filters, nil
}

func GetProductItemsByOfferID(ctx context.Context, db *gorm.DB, offerID string, categoryID int, departmentID int, subCategoryID int, latStr string, lngStr string, pincode string, radiusKm float64, limit int, offset int) (productItems []response.ProductItems, err error) {
	offerQuery := `SELECT pi.sub_category_name, pi.id, pi.category_id, pi.department_id, pi.sub_category_id,
				pi.product_item_images, pi.dynamic_fields, pi.created_at, pi.updated_at,
				c.name AS category_name, d.name AS department_name, sc.name AS sub_category_name_ref,
				sc.image_url AS sub_category_image_url,
				(SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = pi.id) AS view_count,
				(
					SELECT COALESCE(json_agg(json_build_object(
						'offer_product_id', op2.id,
						'product_name', pi2.sub_category_name,
						'offer_id', p2.id,
						'offer_name', p2.offer_name,
						'discount_rate', p2.discount_rate,
						'description', p2.description,
						'start_date', p2.start_date,
						'end_date', p2.end_date,
						'promotion_category', json_build_object(
							'id', pc2.id,
							'name', pc2.name,
							'shop_id', pc2.shop_id,
							'is_active', pc2.is_active,
							'icon_path', pc2.icon_path,
							'created_at', pc2.created_at,
							'updated_at', pc2.updated_at
						),
						'promotion_type', json_build_object(
							'id', pt2.id,
							'name', pt2.name,
							'is_active', pt2.is_active,
							'shop_id', pt2.shop_id,
							'promotion_category_id', pt2.promotion_category_id,
							'icon_path', pt2.icon_path,
							'created_at', pt2.created_at,
							'updated_at', pt2.updated_at
						)
					) ORDER BY p2.created_at DESC), '[]')
					FROM offer_products op2
					LEFT JOIN product_items pi2 ON pi2.id = (op2.product_item_id::text::bigint)
					INNER JOIN promotions p2 ON p2.id = op2.offer_id
					LEFT JOIN promotion_categories pc2 ON p2.promotion_category_id = pc2.id
					LEFT JOIN promotions_types pt2 ON p2.promotion_type_id = pt2.id
					WHERE (op2.product_item_id::text::bigint) = pi.id
					AND p2.is_active = true
				) AS offer_products
			FROM product_items pi
			LEFT JOIN categories c ON pi.category_id = c.id
			LEFT JOIN departments d ON pi.department_id = d.id
			LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
			INNER JOIN offer_products op ON (op.product_item_id::text::bigint) = pi.id
			WHERE op.offer_id = $1`

	// Add filters if provided
	var filters []string
	var params []interface{}
	params = append(params, offerID)
	paramIndex := 2

	if categoryID > 0 {
		filters = append(filters, fmt.Sprintf("pi.category_id = $%d", paramIndex))
		params = append(params, categoryID)
		paramIndex++
	}
	if departmentID > 0 {
		filters = append(filters, fmt.Sprintf("pi.department_id = $%d", paramIndex))
		params = append(params, departmentID)
		paramIndex++
	}
	if subCategoryID > 0 {
		filters = append(filters, fmt.Sprintf("pi.sub_category_id = $%d", paramIndex))
		params = append(params, subCategoryID)
		paramIndex++
	}

	// Handle location-based filters
	if latStr != "" && lngStr != "" && radiusKm > 0 {
		lat, errLat := strconv.ParseFloat(latStr, 64)
		lng, errLng := strconv.ParseFloat(lngStr, 64)
		if errLat == nil && errLng == nil {
			// Distance calculation using haversine formula
			filters = append(filters, fmt.Sprintf(`(6371 * acos(cos(radians($%d)) * cos(radians(CAST(pi.latitude AS float))) * cos(radians(CAST(pi.longitude AS float)) - radians($%d)) + sin(radians($%d)) * sin(radians(CAST(pi.latitude AS float))))) <= $%d`, paramIndex, paramIndex+1, paramIndex+2, paramIndex+3))
			params = append(params, lat, lng, lat, radiusKm)
			paramIndex += 4
		}
	}

	// Handle pincode filter
	if pincode != "" {
		filters = append(filters, fmt.Sprintf("pi.pincode = $%d", paramIndex))
		params = append(params, pincode)
		paramIndex++
	}

	if len(filters) > 0 {
		offerQuery += " AND " + strings.Join(filters, " AND ")
	}

	// Add pagination
	offerQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramIndex, paramIndex+1)
	params = append(params, limit, offset)

	type offerProductDB struct {
		Name                string    `gorm:"column:sub_category_name"`
		ID                  string    `gorm:"column:id"`
		CategoryID          string    `gorm:"column:category_id"`
		DepartmentID        string    `gorm:"column:department_id"`
		SubCategoryID       string    `gorm:"column:sub_category_id"`
		CategoryName        string    `gorm:"column:category_name"`
		DepartmentName      string    `gorm:"column:department_name"`
		SubCategoryNameRef  string    `gorm:"column:sub_category_name_ref"`
		SubCategoryImageURL string    `gorm:"column:sub_category_image_url"`
		ProductItemImages   string    `gorm:"column:product_item_images"`
		DynamicFields       []byte    `gorm:"column:dynamic_fields"`
		OfferProducts       []byte    `gorm:"column:offer_products"`
		CreatedAt           time.Time `gorm:"column:created_at"`
		UpdatedAt           time.Time `gorm:"column:updated_at"`
		ViewCount           uint      `gorm:"column:view_count"`

		// Offer details
		OfferID      string    `gorm:"column:offer_id"`
		OfferName    string    `gorm:"column:offer_name"`
		DiscountRate uint      `gorm:"column:discount_rate"`
		Description  string    `gorm:"column:description"`
		StartDate    time.Time `gorm:"column:start_date"`
		EndDate      time.Time `gorm:"column:end_date"`

		// Promotion category details
		PromotionCategoryID        string    `gorm:"column:promotion_category_id"`
		PromotionCategoryName      string    `gorm:"column:promotion_category_name"`
		PromotionCategoryShopID    string    `gorm:"column:promotion_category_shop_id"`
		PromotionCategoryIsActive  bool      `gorm:"column:promotion_category_is_active"`
		PromotionCategoryIconPath  string    `gorm:"column:promotion_category_icon_path"`
		PromotionCategoryCreatedAt time.Time `gorm:"column:promotion_category_created_at"`
		PromotionCategoryUpdatedAt time.Time `gorm:"column:promotion_category_updated_at"`

		// Promotion type details
		PromotionTypeID                  string    `gorm:"column:promotion_type_id"`
		PromotionTypeName                string    `gorm:"column:promotion_type_name"`
		PromotionTypeIsActive            bool      `gorm:"column:promotion_type_is_active"`
		PromotionTypeShopID              uint      `gorm:"column:promotion_type_shop_id"`
		PromotionTypePromotionCategoryID string    `gorm:"column:promotion_type_promotion_category_id"`
		PromotionTypeType                string    `gorm:"column:promotion_type_type"`
		PromotionTypeIconPath            string    `gorm:"column:promotion_type_icon_path"`
		PromotionTypeCreatedAt           time.Time `gorm:"column:promotion_type_created_at"`
		PromotionTypeUpdatedAt           time.Time `gorm:"column:promotion_type_updated_at"`
	}

	var dbItems []offerProductDB
	err = db.Raw(offerQuery, params...).Scan(&dbItems).Error
	if err != nil {
		return
	}

	for _, dbItem := range dbItems {
		var images []string
		if dbItem.ProductItemImages != "" {
			imageStr := dbItem.ProductItemImages
			if len(imageStr) > 2 && imageStr[0] == '{' && imageStr[len(imageStr)-1] == '}' {
				imageStr = imageStr[1 : len(imageStr)-1]
				if imageStr != "" {
					images = []string{}
					for _, img := range parsePostgresArray(imageStr) {
						images = append(images, img)
					}
				}
			}
		}
		if images == nil {
			images = []string{}
		}

		item := response.ProductItems{
			ID:                  dbItem.ID,
			Name:                dbItem.Name,
			CategoryName:        dbItem.CategoryName,
			MainCategoryName:    dbItem.DepartmentName,
			SubCategoryImageURL: dbItem.SubCategoryImageURL,
			CategoryID:          dbItem.CategoryID,
			DepartmentID:        dbItem.DepartmentID,
			SubCategoryID:       dbItem.SubCategoryID,
			ProductItemImages:   images,
			CreatedAt:           dbItem.CreatedAt,
			UpdatedAt:           dbItem.UpdatedAt,
			ViewCount:           dbItem.ViewCount,
		}

		if len(dbItem.DynamicFields) > 0 {
			var dynamicFields map[string]interface{}
			if err := json.Unmarshal(dbItem.DynamicFields, &dynamicFields); err == nil {
				item.DynamicFields = dynamicFields
			}
		} else {
			item.DynamicFields = make(map[string]interface{})
		}

		// Unmarshal offer_products
		if len(dbItem.OfferProducts) > 0 {
			var offerProducts []response.OfferProduct
			if err := json.Unmarshal(dbItem.OfferProducts, &offerProducts); err == nil {
				item.OfferProducts = offerProducts
			} else {
				item.OfferProducts = []response.OfferProduct{}
			}
		} else {
			item.OfferProducts = []response.OfferProduct{}
		}

		productItems = append(productItems, item)
	}

	return
}

// CreateCategory inserts a new category with image_url and sort_order.
func (c *productDatabase) CreateCategory(ctx context.Context, departmentID, name, imageURL string, sortOrder int, isActive bool) error {
	query := `INSERT INTO categories (id, department_id, name, image_url, sort_order, is_active)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	return c.DB.Exec(query, domain.NewID(domain.PrefixCategory), departmentID, name, imageURL, sortOrder, isActive).Error
}

// UpdateCategory updates name, image_url, sort_order, is_active for a category.
// If imageURL is empty the existing image_url is preserved.
func (c *productDatabase) UpdateCategory(ctx context.Context, categoryID, name, imageURL string, sortOrder int, isActive bool) error {
	if imageURL != "" {
		query := `UPDATE categories SET name = $1, image_url = $2, sort_order = $3, is_active = $4 WHERE id = $5`
		return c.DB.Exec(query, name, imageURL, sortOrder, isActive, categoryID).Error
	}
	query := `UPDATE categories SET name = $1, sort_order = $2, is_active = $3 WHERE id = $4`
	return c.DB.Exec(query, name, sortOrder, isActive, categoryID).Error
}

func (c *productDatabase) DeleteCategory(ctx context.Context, categoryID string) error {
	return c.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM sub_categories WHERE category_id = $1`, categoryID).Error; err != nil {
			return err
		}
		return tx.Exec(`DELETE FROM categories WHERE id = $1`, categoryID).Error
	})
}

// CreateSubCategory inserts a new sub-category with image_url and sort_order.
func (c *productDatabase) CreateSubCategory(ctx context.Context, departmentID, categoryID, name, imageURL string, sortOrder int, isActive bool) error {
	query := `INSERT INTO sub_categories (id, department_id, category_id, name, image_url, sort_order, is_active)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`
	return c.DB.Exec(query, domain.NewID(domain.PrefixSubCategory), departmentID, categoryID, name, imageURL, sortOrder, isActive).Error
}

func (c *productDatabase) DeleteSubCategory(ctx context.Context, subCategoryID string) error {
	return c.DB.Exec(`DELETE FROM sub_categories WHERE id = $1`, subCategoryID).Error
}

// UpdateSubCategory updates name, image_url, sort_order, is_active for a sub-category.
// If imageURL is empty the existing image_url is preserved.
func (c *productDatabase) UpdateSubCategory(ctx context.Context, subCategoryID, name, imageURL string, sortOrder int, isActive bool) error {
	if imageURL != "" {
		query := `UPDATE sub_categories SET name = $1, image_url = $2, sort_order = $3, is_active = $4 WHERE id = $5`
		return c.DB.Exec(query, name, imageURL, sortOrder, isActive, subCategoryID).Error
	}
	query := `UPDATE sub_categories SET name = $1, sort_order = $2, is_active = $3 WHERE id = $4`
	return c.DB.Exec(query, name, sortOrder, isActive, subCategoryID).Error
}
