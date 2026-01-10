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
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/elasticsearch"
	"gorm.io/gorm"
)

type productDatabase struct {
	DB            *gorm.DB
	ElasticClient *elasticsearch.ElasticService
}

// DeleteProductItem deletes a product item by its ID.
func (c *productDatabase) DeleteProductItem(ctx context.Context, productItemID uint) error {
	query := `DELETE FROM product_items WHERE id = $1`
	return c.DB.Exec(query, productItemID).Error
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
func (c *productDatabase) IsCategoryNameExist(ctx context.Context, name string, departmentId uint) (exist bool, err error) {

	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1 AND department_id = $2)`
	err = c.DB.Raw(query, name, departmentId).Scan(&exist).Error

	return
}

// Save Category
func (c *productDatabase) SaveCategory(ctx context.Context, category request.Category, departmentId string) (err error) {

	query := `INSERT INTO categories (name, department_id) VALUES ($1, $2)`
	err = c.DB.Exec(query, category.Name, departmentId).Error

	return err
}

// To check the sub category name already exist for the category
func (c *productDatabase) IsSubCategoryNameExist(ctx context.Context, name string, departmentId uint) (exist bool, err error) {

	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1 AND department_id = $2)`
	err = c.DB.Raw(query, name, departmentId).Scan(&exist).Error

	return
}

// Save Category as sub category
func (c *productDatabase) SaveSubCategory(ctx context.Context, body request.SubCategory, brandID string, categoryID string) (err error) {

	print("department id in repo", brandID, "category id in repo", categoryID)
	query := `INSERT INTO sub_categories (department_id, category_id, name) VALUES ($1, $2, $3)`
	err = c.DB.Exec(query, brandID, categoryID, body.Name).Error

	return err
}

// Find all main category(its not have a category_id)
func (c *productDatabase) FindAllMainCategories(ctx context.Context,
	pagination request.Pagination) (categories []response.Category, err error) {

	limit := pagination.Count
	offset := (pagination.PageNumber - 1) * limit

	query := `SELECT id, name FROM categories 
	LIMIT $1 OFFSET $2`
	err = c.DB.Raw(query, limit, offset).Scan(&categories).Error

	return
}

// Find all sub categories of a category
func (c *productDatabase) FindAllSubCategories(ctx context.Context,
	categoryID uint) (subCategories []response.SubCategory, err error) {

	query := `SELECT id, name FROM sub_categories WHERE category_id = $1`
	err = c.DB.Raw(query, categoryID).Scan(&subCategories).Error

	return
}

// Find all variations which related to given category id
func (c *productDatabase) FindAllVariationsByCategoryID(ctx context.Context,
	categoryID uint) (variations []response.Variation, err error) {

	query := `SELECT id, name FROM variations WHERE category_id = $1`
	err = c.DB.Raw(query, categoryID).Scan(&variations).Error

	return
}

// Find all variation options which related to given variation id
func (c productDatabase) FindAllVariationOptionsByVariationID(ctx context.Context,
	variationID uint) (variationOptions []response.VariationOption, err error) {

	query := `SELECT id, value FROM variation_options WHERE variation_id = $1`
	err = c.DB.Raw(query, variationID).Scan(&variationOptions).Error

	return
}

// To check a variation exist for the given category
func (c *productDatabase) IsVariationNameExistForCategory(ctx context.Context,
	name string, categoryID uint) (exist bool, err error) {

	query := `SELECT EXISTS(SELECT 1 FROM variations WHERE name = $1 AND category_id = $2)`
	err = c.DB.Raw(query, name, categoryID).Scan(&exist).Error

	return
}

// To check a variation value exist for the given variation
func (c *productDatabase) IsVariationValueExistForVariation(ctx context.Context,
	value string, variationID uint) (exist bool, err error) {

	query := `SELECT EXISTS(SELECT 1 FROM variation_options WHERE value = $1 AND variation_id = $2)`
	err = c.DB.Raw(query, value, variationID).Scan(&exist).Error

	return
}

// Save Variation for category
func (c *productDatabase) SaveVariation(ctx context.Context, categoryID uint, variationName string) error {

	query := `INSERT INTO variations (category_id, name) VALUES($1, $2)`
	err := c.DB.Exec(query, categoryID, variationName).Error

	return err
}

// add variation option
func (c *productDatabase) SaveVariationOption(ctx context.Context, variationID uint, variationValue string) error {

	query := `INSERT INTO variation_options (variation_id, value) VALUES($1, $2)`
	err := c.DB.Exec(query, variationID, variationValue).Error

	return err
}

// find product by id
func (c *productDatabase) FindProductByID(ctx context.Context, productID uint) (product domain.Product, err error) {

	query := `SELECT * FROM products WHERE id = $1`
	err = c.DB.Raw(query, productID).Scan(&product).Error

	return
}

func (c *productDatabase) IsProductNameExistForOtherProduct(ctx context.Context,
	name string, productID uint) (exist bool, err error) {

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
func (c *productDatabase) SaveProduct(ctx context.Context, product domain.Product, adminID string) (productID uint, err error) {
	// Get the shop Id and Shop name using adminID

	fmt.Printf("Saving product: %+v for adminID: %s\n", product, adminID)
	var shopDetails struct {
		ShopID   string `gorm:"column:id"`
		ShopName string `gorm:"column:shop_name"`
	}

	query := `SELECT id, shop_name FROM shop_details WHERE admin_id = $1`
	err = c.DB.Raw(query, adminID).Scan(&shopDetails).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to fetch shop details for admin %s: %v", adminID, err)
	}

	fmt.Printf("Shop ID: %v, Shop Name: %v\n", shopDetails.ShopID, shopDetails.ShopName)

	// Check if product with shop_id and category_id already exists
	checkQuery := `SELECT id FROM products WHERE shop_id = $1 AND category_id = $2 LIMIT 1`
	err = c.DB.Raw(checkQuery, shopDetails.ShopID, product.CategoryID).Scan(&productID).Error

	fmt.Printf("Checked existing product ID: %d for shop_id: %v and category_id: %d\n", productID, shopDetails.ShopID, product.CategoryID)
	if err == nil && productID != 0 {
		// Product already exists, return existing product ID
		fmt.Printf("Product already exists with ID: %d for shop_id: %v and category_id: %d\n",
			productID, shopDetails.ShopID, product.CategoryID)
		return productID, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check if product exists: %v", err)
	}

	// Validate department_id exists
	if product.DepartmentID != 0 {
		var deptExists bool
		deptQuery := `SELECT EXISTS(SELECT 1 FROM departments WHERE id = $1)`
		err = c.DB.Raw(deptQuery, product.DepartmentID).Scan(&deptExists).Error
		if err != nil {
			return 0, fmt.Errorf("failed to check if department exists: %v", err)
		}
		if !deptExists {
			return 0, fmt.Errorf("department with id %d does not exist", product.DepartmentID)
		}
	}

	// Insert new product
	query = `INSERT INTO products (name, description, category_id, image, department_id, shop_id, created_at) 
	VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	createdAt := time.Now()
	err = c.DB.Raw(query, product.Name, product.Description, product.CategoryID, product.Image, product.DepartmentID, shopDetails.ShopID, createdAt).Scan(&productID).Error
	if err != nil {
		return 0, fmt.Errorf("failed to insert product: %v", err)
	}

	fmt.Printf("New product created with ID: %d\n", productID)
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

	limit := pagination.Count
	offset := (pagination.PageNumber - 1) * limit

	query := `SELECT p.*, c.name AS category_name, c.image_url AS category_image_url
	FROM products p
	LEFT JOIN categories c ON p.category_id = c.id
	ORDER BY p.created_at DESC LIMIT $1 OFFSET $2`

	// Temporary struct for scanning (without ProductItems slice)
	type productDB struct {
		ID               uint       `gorm:"column:id"`
		CategoryID       uint       `gorm:"column:category_id"`
		Price            uint       `gorm:"column:price"`
		DiscountPrice    uint       `gorm:"column:discount_price"`
		Name             string     `gorm:"column:name"`
		Description      string     `gorm:"column:description"`
		CategoryName     string     `gorm:"column:category_name"`
		CategoryImageURL string     `gorm:"column:category_image_url"`
		MainCategoryName string     `gorm:"column:main_category_name"`
		BrandID          uint       `gorm:"column:brand_id"`
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
func (c *productDatabase) findProductItemsByProductID(ctx context.Context, productID uint) (productItems []response.ProductItems, err error) {
	query := `SELECT pi.sub_category_name, pi.id, pi.category_id, 
		   sc.name AS category_name, mc.name AS main_category_name
	       FROM product_items pi 
	       LEFT JOIN categories sc ON pi.category_id = sc.id 
	       LEFT JOIN categories mc ON pi.category_id = mc.id 
	       WHERE pi.id = $1`

	type productItemDB struct {
		Name             string `gorm:"column:sub_category_name"`
		ID               uint   `gorm:"column:id"`
		CategoryID       uint   `gorm:"column:category_id"`
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
func (c *productDatabase) FindProductItemByID(ctx context.Context, productItemID uint) (productItem domain.ProductItem, err error) {

	query := `SELECT * FROM product_items WHERE id = $1`
	err = c.DB.Raw(query, productItemID).Scan(&productItem).Error

	return productItem, err
}

// to get how many variations are available for a product
func (c *productDatabase) FindVariationCountForProduct(ctx context.Context, productID uint) (variationCount uint, err error) {

	fmt.Printf("Finding variation count for product ID: %d\n", productID) // Debugging line
	query := `SELECT COUNT(v.id) FROM variations v
	INNER JOIN categories c ON c.id = v.category_id 
	INNER JOIN products p ON p.category_id = v.category_id 
	WHERE p.id = $1`

	err = c.DB.Raw(query, productID).Scan(&variationCount).Error

	return
}

// To find all product item ids which related to the given product id and variation option id
func (c *productDatabase) FindAllProductItemIDsByProductIDAndVariationOptionID(ctx context.Context, productID,
	variationOptionID uint) (productItemIDs []uint, err error) {

	query := `SELECT id FROM product_items pi 
		INNER JOIN product_configurations pc ON pi.id = pc.product_item_id 
		WHERE pi.product_id = $1 AND variation_option_id = $2`
	err = c.DB.Raw(query, productID, variationOptionID).Scan(&productItemIDs).Error

	return
}

func (c *productDatabase) SaveProductConfiguration(ctx context.Context, productItemID, variationOptionID uint) error {

	query := `INSERT INTO product_configurations (product_item_id, variation_option_id) VALUES ($1, $2)`
	err := c.DB.Exec(query, productItemID, variationOptionID).Error

	return err
}

func (c *productDatabase) SaveProductItem(ctx context.Context, productItem request.ProductItem, adminID string) (productItemID uint, err error) {

	query := `INSERT INTO product_items (admin_id, sub_category_name, dynamic_fields, product_item_images, category_id, department_id, sub_category_id, created_at, updated_at) 
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	createdAt := time.Now()

	// Marshal DynamicFields to JSON for JSONB column
	dynamicFieldsJSON, err := json.Marshal(productItem.DynamicFields)
	if err != nil {
		return 0, err
	}

	err = c.DB.Raw(query, adminID, productItem.SubCategoryName, dynamicFieldsJSON, productItem.ProductItemImages, productItem.CategoryID, productItem.DepartmentID, productItem.SubCategoryID, createdAt, createdAt).Scan(&productItemID).Error

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
		}
		go c.ElasticClient.IndexProductItem(ctx, domainItem) // index asynchronously
	}

	return productItemID, err
}

// for get all products items for a product filtered by admin_id and additional filters
func (c *productDatabase) FindAllProductItems(ctx context.Context,
	adminID string, keyword string, categoryID *string, brandID *string, locationID *string, offer string, sortby string, pagination *request.Pagination) (productItems []response.ProductItems, err error) {

	log.Printf("FindAllProductItems called with adminID: %s, offer: %s", adminID, offer)

	var ids []uint
	if keyword != "" && c.ElasticClient != nil {
		limit := 100
		offset := 0
		if pagination != nil {
			limit = int(pagination.Count)
			offset = int((pagination.PageNumber - 1) * pagination.Count)
		}
		var err error
		ids, err = c.ElasticClient.SearchProductItems(ctx, keyword, categoryID, limit, offset)
		if err != nil {
			log.Printf("ES search failed, falling back to PG: %v", err)
		} else if len(ids) == 0 {
			return []response.ProductItems{}, nil
		}
	}

	// Define offerSubquery conditionally
	var offerSubquery string
	if offer == "false" {
		offerSubquery = "'[]'::json"
	} else {
		offerSubquery = `COALESCE(json_agg(json_build_object(
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
		WHERE op2.product_item_id = pi.id`
	}
	log.Printf("offerSubquery: %s", offerSubquery)

	query := `SELECT pi.sub_category_name, pi.id, pi.category_id, pi.department_id, pi.sub_category_id,
				pi.product_item_images, pi.dynamic_fields, pi.created_at, pi.updated_at,
				c.name AS category_name, d.name AS department_name, sc.name AS sub_category_name_ref,
			(SELECT MAX(o3.discount_rate) FROM offer_products op3 INNER JOIN offers o3 ON o3.id = op3.offer_id WHERE op3.product_item_id = pi.id) AS discount_rate,
				sc.image_url AS sub_category_image_url,
				(SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = pi.id) AS view_count,
				(SELECT ` + offerSubquery + `) AS offer_products
			FROM product_items pi 
			LEFT JOIN categories c ON pi.category_id = c.id 
			LEFT JOIN departments d ON pi.department_id = d.id
			LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
			WHERE (pi.admin_id ->> 'id' = @adminID OR pi.admin_id #>> '{}' = @adminID)`
	// Add offer filter
	if offer == "true" {
		query += " AND EXISTS (SELECT 1 FROM offer_products op WHERE op.product_item_id = pi.id)"
	} else if offer == "false" {
		query += " AND NOT EXISTS (SELECT 1 FROM offer_products op WHERE op.product_item_id = pi.id)"
	}
	fmt.Printf("Final query: %s", query)
	// Add filters dynamically
	params := map[string]interface{}{
		"adminID": adminID,
	}
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
	if pagination != nil {
		query += " ORDER BY " + orderBy + " DESC LIMIT @limit OFFSET @offset"
		params["limit"] = pagination.Count
		params["offset"] = (pagination.PageNumber - 1) * pagination.Count
	} else {
		query += " ORDER BY " + orderBy + " DESC"
	}

	// Internal struct for scanning DB result
	type productItemDB struct {
		Name                string        `gorm:"column:sub_category_name"`
		ID                  uint          `gorm:"column:id"`
		CategoryID          uint          `gorm:"column:category_id"`
		DepartmentID        uint          `gorm:"column:department_id"`
		SubCategoryID       uint          `gorm:"column:sub_category_id"`
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
	}
	var dbItems []productItemDB
	log.Printf("Query: %s, Params: %v", query, params)
	err = c.DB.Raw(query, params).Scan(&dbItems).Error
	if err != nil {
		return
	}

	log.Printf("Number of dbItems scanned: %d", len(dbItems))
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
		}

		// Unmarshal offer_products if present
		if len(dbItem.OfferProducts) > 0 {
			var offerProducts []response.OfferProduct
			if err := json.Unmarshal(dbItem.OfferProducts, &offerProducts); err == nil {
				item.OfferProducts = offerProducts // Make sure OfferProducts in response.ProductItems is []response.OfferProduct
				if len(offerProducts) > 0 {
					var offerNames []string
					for _, op := range offerProducts {
						offerNames = append(offerNames, op.OfferName)
					}
					// If you want to keep offerNames, assign to a []string field, not item.Offer
					// For example, if item has OfferNames []string, use:
					// item.OfferNames = offerNames
					// Otherwise, remove this assignment to avoid the type error.
				}
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
	fmt.Printf("Retrieved %d product items\n", len(productItems)) // Debugging line
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

// Find all variation and value of a product item
func (c *productDatabase) FindAllVariationValuesOfProductItem(ctx context.Context,
	productItemID uint) (productVariationsValues []response.ProductVariationValue, err error) {

	query := `SELECT v.id AS variation_id, v.name, vo.id AS variation_option_id, vo.value 
	FROM  product_configurations pc 
	INNER JOIN variation_options vo ON vo.id = pc.variation_option_id 
	INNER JOIN variations v ON v.id = vo.variation_id 
	WHERE pc.product_item_id = $1`
	err = c.DB.Raw(query, productItemID).Scan(&productVariationsValues).Error

	return
}

// To save image for product item
func (c *productDatabase) SaveProductItemImage(ctx context.Context, productItemID uint, image domain.ProductItemImage) error {

	query := `INSERT INTO product_images (product_item_id, image) VALUES ($1, $2)`
	err := c.DB.Exec(query, productItemID, image).Error

	return err
}

// To find all images of a product item
func (c *productDatabase) FindAllProductItemImages(ctx context.Context, productItemID uint) (images []string, err error) {

	query := `SELECT product_item_images FROM product_items WHERE id = $1`

	err = c.DB.Raw(query, productItemID).Scan(&images).Error

	return
}

// SearchProducts implements interfaces.ProductRepository.
func (c *productDatabase) SearchProducts(ctx context.Context, keyword string, categoryID, brandID, locationID *string, pagination request.Pagination) (products []response.ProductItems, err error) {
	pageNumber := pagination.PageNumber
	if pageNumber < 1 {
		pageNumber = 1
	}
	offset := int((pageNumber - 1) * pagination.Count)
	limit := int(pagination.Count)

	var ids []uint
	if keyword != "" && c.ElasticClient != nil {
		// Use Elasticsearch for search
		ids, err = c.ElasticClient.SearchProductItems(ctx, keyword, categoryID, limit, offset)
		if err != nil {
			log.Printf("ES search failed, falling back to PG: %v", err)
		} else if len(ids) == 0 {
			return []response.ProductItems{}, nil
		}
	}

	// Build the base query
	baseQuery := `SELECT pi.sub_category_name, pi.id, pi.category_id, pi.department_id, pi.sub_category_id,
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
		LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id`

	params := []interface{}{}
	paramIndex := 1

	// If we have IDs from Elasticsearch, filter by them
	if len(ids) > 0 {
		placeholders := make([]string, len(ids))
		for i, id := range ids {
			placeholders[i] = fmt.Sprintf("$%d", paramIndex)
			params = append(params, id)
			paramIndex++
		}
		baseQuery += " WHERE pi.id IN (" + strings.Join(placeholders, ",") + ")"
		baseQuery += " ORDER BY pi.created_at DESC"
	} else {
		// Fallback to original search logic
		whereClause := " WHERE (pi.sub_category_name ILIKE $1 OR pi.dynamic_fields::text ILIKE $1 OR c.name::text ILIKE $1 OR sc.name::text ILIKE $1)"
		params = append(params, "%"+keyword+"%")
		paramIndex = 2

		if categoryID != nil {
			if cid, err := strconv.ParseUint(*categoryID, 10, 64); err == nil {
				whereClause += fmt.Sprintf(" AND pi.category_id = $%d", paramIndex)
				params = append(params, cid)
				paramIndex++
			}
		}

		baseQuery += whereClause + " ORDER BY pi.created_at DESC LIMIT $" + fmt.Sprint(paramIndex) + " OFFSET $" + fmt.Sprint(paramIndex+1)
		params = append(params, limit, offset)
	}

	// Log the final SQL and parameters for debugging
	fmt.Printf("SearchProducts SQL: %s\n", baseQuery)
	fmt.Printf("SearchProducts Params: %#v\n", params)

	// Scan into internal DB struct to correctly parse JSONB and array columns
	type productItemDB struct {
		Name                string    `gorm:"column:sub_category_name"`
		ID                  uint      `gorm:"column:id"`
		CategoryID          uint      `gorm:"column:category_id"`
		DepartmentID        uint      `gorm:"column:department_id"`
		SubCategoryID       uint      `gorm:"column:sub_category_id"`
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
	err = c.DB.Raw(baseQuery, params...).Scan(&dbItems).Error
	if err != nil {
		fmt.Printf("Executed Query Error: %v\n", err) // Debugging line
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
			ProductItemImages:   images,
			CreatedAt:           dbItem.CreatedAt,
			UpdatedAt:           dbItem.UpdatedAt,
			ViewCount:           dbItem.ViewCount,
		}

		// Unmarshal offer_products if present. Handle both raw JSON bytes and JSON-as-string.
		if len(dbItem.OfferProducts) > 0 {
			var offerProducts []response.OfferProduct
			if err := json.Unmarshal(dbItem.OfferProducts, &offerProducts); err != nil {
				// try interpreting as string containing JSON
				rawStr := string(dbItem.OfferProducts)
				if rawStr != "" {
					if err2 := json.Unmarshal([]byte(rawStr), &offerProducts); err2 != nil {
						fmt.Printf("SearchProducts: failed to unmarshal offer_products: %v %v\n", err, err2)
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

func (c *productDatabase) SaveDepartment(ctx context.Context, departmentName string) error {

	query := `INSERT INTO departments (name) VALUES ($1)`
	err := c.DB.Exec(query, departmentName).Error

	return err
}

func (c *productDatabase) GetAllDepartments(ctx context.Context) (departments []response.Department, err error) {

	query := `SELECT id, name, image_url FROM departments`
	err = c.DB.Raw(query).Scan(&departments).Error
	return
}

func (c *productDatabase) GetDepartmentByID(ctx context.Context, brandID uint) (department response.Department, err error) {

	query := `SELECT id, name, image_url FROM departments WHERE id = $1`
	err = c.DB.Raw(query, brandID).Scan(&department).Error

	return
}

func (c *productDatabase) GetAllSubCategories(ctx context.Context) (subCategories []response.SubCategory, err error) {

	query := `SELECT id, name FROM sub_categories`
	err = c.DB.Raw(query).Scan(&subCategories).Error

	return
}

func (c *productDatabase) GetAllCategoriesByDepartmentID(ctx context.Context, brandID uint) (categories []response.Category, err error) {

	query := `SELECT id, name, image_url FROM categories WHERE department_id = $1`
	err = c.DB.Raw(query, brandID).Scan(&categories).Error

	return
}

func (c *productDatabase) GetAllSubCategoriesByCategoryID(ctx context.Context, categoryID uint) (subCategories []response.SubCategory, err error) {

	query := `SELECT id, name, image_url FROM sub_categories WHERE category_id = $1`
	err = c.DB.Raw(query, categoryID).Scan(&subCategories).Error

	return
}

// SaveSubTypeAttribute saves a new sub type attribute for a subcategory
func (c *productDatabase) SaveSubTypeAttribute(ctx context.Context, locationID uint, attribute domain.SubTypeAttributes) error {
	attribute.SubCategoryID = locationID
	return c.DB.Create(&attribute).Error
}

// GetAllSubTypeAttributes retrieves all sub type attributes for a subcategory
func (c *productDatabase) GetAllSubTypeAttributes(ctx context.Context, locationID uint) (attributes []response.SubTypeAttribute, err error) {
	query := `SELECT id, sub_category_id, field_name, field_type, is_required, sort_order 
	          FROM sub_type_attributes 
	          WHERE sub_category_id = $1 
	          ORDER BY sort_order ASC`
	err = c.DB.Raw(query, locationID).Scan(&attributes).Error
	return
}

// GetSubTypeAttributeByID retrieves a single sub type attribute by ID
func (c *productDatabase) GetSubTypeAttributeByID(ctx context.Context, attributeID uint) (attribute response.SubTypeAttribute, err error) {
	query := `SELECT id, sub_category_id, field_name, field_type, is_required, sort_order 
	          FROM sub_type_attributes 
	          WHERE id = $1`
	err = c.DB.Raw(query, attributeID).Scan(&attribute).Error
	return
}

// SaveSubTypeAttributeOption saves a new option for a sub type attribute
func (c *productDatabase) SaveSubTypeAttributeOption(ctx context.Context, attributeID uint, option domain.SubTypeAttributeOptions) error {
	option.SubTypeAttributeID = attributeID
	return c.DB.Create(&option).Error
}

// GetAllSubTypeAttributeOptions retrieves all options for a sub type attribute
func (c *productDatabase) GetAllSubTypeAttributeOptions(ctx context.Context, attributeID uint) (options []response.SubTypeAttributeOption, err error) {
	query := `SELECT id, sub_type_attribute_id, option_value, sort_order 
	          FROM sub_type_attribute_options 
	          WHERE sub_type_attribute_id = $1 
	          ORDER BY sort_order ASC`
	err = c.DB.Raw(query, attributeID).Scan(&options).Error
	return
}

// GetSubTypeAttributeOptionByID retrieves a single option by ID
func (c *productDatabase) GetSubTypeAttributeOptionByID(ctx context.Context, optionID uint) (option response.SubTypeAttributeOption, err error) {
	query := `SELECT id, sub_type_attribute_id, option_value, sort_order 
	          FROM sub_type_attribute_options 
	          WHERE id = $1`
	err = c.DB.Raw(query, optionID).Scan(&option).Error
	return
}

// SaveCategoryImage saves a new category image
func (c *productDatabase) SaveCategoryImage(ctx context.Context, categoryID uint, image domain.CategoryImage) error {
	image.CategoryID = categoryID
	query := `INSERT INTO category_images (category_id, image_url, alt_text, sort_order, is_active) 
	          VALUES ($1, $2, $3, $4, $5)`
	return c.DB.Exec(query, image.CategoryID, image.ImageURL, image.AltText, image.SortOrder, image.IsActive).Error
}

// GetAllCategoryImages retrieves all images for a category
func (c *productDatabase) GetAllCategoryImages(ctx context.Context, categoryID uint) (images []response.CategoryImage, err error) {
	query := `SELECT id, category_id, image_url, alt_text, sort_order, is_active 
	          FROM category_images 
	          WHERE category_id = $1 AND is_active = true 
	          ORDER BY sort_order ASC, id ASC`
	err = c.DB.Raw(query, categoryID).Scan(&images).Error
	return
}

// GetCategoryImageByID retrieves a single category image by ID
func (c *productDatabase) GetCategoryImageByID(ctx context.Context, imageID uint) (image response.CategoryImage, err error) {
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
func (c *productDatabase) DeleteCategoryImage(ctx context.Context, imageID uint) error {
	query := `UPDATE category_images SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	return c.DB.Exec(query, imageID).Error
}

func (c *productDatabase) GetProductItemByID(ctx context.Context, productItemID uint) (productItem response.ProductItems, err error) {
	// First, get product item details (excluding images)
	query := `SELECT pi.id, pi.sub_category_name, pi.category_id, 
	           sc.name AS category_name, mc.name AS main_category_name, 
	           pi.dynamic_fields, pi.created_at, pi.updated_at,
	           (SELECT COALESCE(SUM(view_count), 0) FROM product_item_views WHERE product_item_id = pi.id) AS view_count
	       FROM product_items pi 
	       LEFT JOIN categories sc ON pi.category_id = sc.id 
	       LEFT JOIN categories mc ON pi.category_id = mc.id 
	       WHERE pi.id = $1;`

	var dbItem struct {
		ID               uint
		SubCategoryName  string
		ProductID        uint
		CategoryID       uint
		CategoryName     string
		MainCategoryName string
		DynamicFields    []byte
		CreatedAt        time.Time
		UpdatedAt        time.Time
		ViewCount        uint
	}

	err = c.DB.Raw(query, productItemID).Scan(&dbItem).Error
	if err != nil {
		return
	}

	productItem.ID = dbItem.ID
	productItem.Name = dbItem.SubCategoryName
	productItem.CategoryName = dbItem.CategoryName
	productItem.MainCategoryName = dbItem.MainCategoryName
	productItem.CreatedAt = dbItem.CreatedAt
	productItem.UpdatedAt = dbItem.UpdatedAt
	productItem.ViewCount = dbItem.ViewCount

	// Fetch images from product_images table
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

	// Fetch offers for this product item
	offerQuery := `SELECT op.id as offer_product_id, pi.sub_category_name as product_name,
	                  o.id as offer_id, o.name as offer_name, o.discount_rate, o.description,
	                  o.start_date, o.end_date, o.image, o.thumbnail
	               FROM offer_products op
	               INNER JOIN offers o ON o.id = op.offer_id
	               LEFT JOIN product_items pi ON pi.id = op.product_item_id
	               WHERE op.product_item_id = $1`

	var offerRows []struct {
		OfferProductID uint      `gorm:"column:offer_product_id"`
		ProductName    string    `gorm:"column:product_name"`
		OfferID        uint      `gorm:"column:offer_id"`
		OfferName      string    `gorm:"column:offer_name"`
		DiscountRate   uint      `gorm:"column:discount_rate"`
		Description    string    `gorm:"column:description"`
		StartDate      time.Time `gorm:"column:start_date"`
		EndDate        time.Time `gorm:"column:end_date"`
		Image          string    `gorm:"column:image"`
		Thumbnail      string    `gorm:"column:thumbnail"`
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
				Image:          row.Image,
				Thumbnail:      row.Thumbnail,
			}
		}
		productItem.OfferProducts = offerProducts
	}

	return
}

func (c *productDatabase) IncrementProductItemViewCount(ctx context.Context, productItemID uint, adminID string) error {
	// Get shop ID using admin ID
	var shopID string
	shopQuery := `SELECT id FROM shop_details WHERE admin_id = $1`
	err := c.DB.Raw(shopQuery, adminID).Scan(&shopID).Error
	if err != nil {
		return err
	}

	// First, try to update existing record
	updateQuery := `UPDATE product_item_views SET view_count = view_count + 1, viewed_at = CURRENT_TIMESTAMP WHERE product_item_id = $1 AND shop_id = $2`
	result := c.DB.Exec(updateQuery, productItemID, shopID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// No existing record, insert new one
		insertQuery := `INSERT INTO product_item_views (product_item_id, shop_id, admin_id, view_count, created_at, viewed_at) VALUES ($1, $2, $3, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
		return c.DB.Exec(insertQuery, productItemID, shopID, adminID).Error
	}
	return nil
}
func (c *productDatabase) GetProductItemViewCount(ctx context.Context, productItemID uint, adminID string) (viewCount uint, err error) {
	// Get shop ID using admin ID
	var shopID string
	shopQuery := `SELECT id FROM shop_details WHERE admin_id = $1`
	err = c.DB.Raw(shopQuery, adminID).Scan(&shopID).Error
	if err != nil {
		return 0, err
	}

	query := `SELECT view_count FROM product_item_views WHERE product_item_id = $1 AND shop_id = $2`
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
		ID                uint      `gorm:"column:id"`
		CategoryID        uint      `gorm:"column:category_id"`
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
func (c *productDatabase) GetProductItemsByDepartment(ctx context.Context, brandID uint) (productItems []response.ProductItems, err error) {
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
		ID                  uint          `gorm:"column:id"`
		CategoryID          uint          `gorm:"column:category_id"`
		DepartmentID        uint          `gorm:"column:department_id"`
		SubCategoryID       uint          `gorm:"column:sub_category_id"`
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
func (c *productDatabase) GetProductItemsByCategory(ctx context.Context, categoryID uint) (productItems []response.ProductItems, err error) {
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
		ID                  uint      `gorm:"column:id"`
		CategoryID          uint      `gorm:"column:category_id"`
		DepartmentID        uint      `gorm:"column:department_id"`
		SubCategoryID       uint      `gorm:"column:sub_category_id"`
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
func (c *productDatabase) GetProductItemsBySubCategory(ctx context.Context, locationID uint) (productItems []response.ProductItems, err error) {
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
		ID                  uint      `gorm:"column:id"`
		CategoryID          uint      `gorm:"column:category_id"`
		DepartmentID        uint      `gorm:"column:department_id"`
		SubCategoryID       uint      `gorm:"column:sub_category_id"`
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
func (c *productDatabase) GetProductItemsByShop(ctx context.Context, adminID uint) (productItems []response.ProductItems, err error) {
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
		ID                  uint      `gorm:"column:id"`
		CategoryID          uint      `gorm:"column:category_id"`
		DepartmentID        uint      `gorm:"column:department_id"`
		SubCategoryID       uint      `gorm:"column:sub_category_id"`
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
	log.Printf("GetProductItemsByShop adminID param: %d", adminID)

	var dbItems []productItemDB
	err = c.DB.Raw(query, adminID).Scan(&dbItems).Error
	if err != nil {
		return
	}

	log.Printf("GetProductItemsByShop rows returned: %d", len(dbItems))

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

func (c *productDatabase) FindProductItemFilters(ctx context.Context, adminID string) ([]domain.ProductItemFilterType, error) {
	var shopID uint
	shopQuery := `SELECT id FROM shop_details WHERE admin_id = $1`
	err := c.DB.Raw(shopQuery, adminID).Scan(&shopID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shop ID for admin %s: %v", adminID, err)
	}

	var filters []domain.ProductItemFilterType
	query := `SELECT id, filter_name, shop_id FROM product_item_filter_types ORDER BY id ASC`
	err = c.DB.Raw(query).Scan(&filters).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch product item filters: %v", err)
	}

	// Get products for the admin
	var products []struct {
		CategoryID uint `json:"category_id"`
	}
	productQuery := `SELECT category_id FROM product_items WHERE admin_id = $1`
	err = c.DB.Raw(productQuery, adminID).Scan(&products).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %v", err)
	}

	// Get unique category IDs
	categoryIDMap := make(map[uint]bool)
	for _, p := range products {
		categoryIDMap[p.CategoryID] = true
	}

	// Get category names and ids
	var categories []struct {
		ID   uint
		Name string
	}
	if len(categoryIDMap) > 0 {
		categoryIDs := make([]uint, 0, len(categoryIDMap))
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

	fmt.Printf("Fetched %d product item filters for admin %s\n", len(filters), adminID)

	return filters, nil
}
