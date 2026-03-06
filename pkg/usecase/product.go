package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
	"gorm.io/gorm"
)

type productUseCase struct {
	productRepo  interfaces.ProductRepository
	cloudService cloud.CloudService
	DB           DBQuerier // Add this field for DB access
}

type OfferUseCase struct {
	offerRepo interfaces.OfferRepository
	DB        DBQuerier
}

// DBQuerier abstracts the DB query interface for easier testing and flexibility.
type DBQuerier interface {
	Query(ctx context.Context, sql string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// Rows interface defines the required methods for database rows
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
}

type pgxRowsAdapter struct {
	rows *sql.Rows
}

func (p *pgxRowsAdapter) Next() bool {
	return p.rows.Next()
}

func (p *pgxRowsAdapter) Scan(dest ...interface{}) error {
	return p.rows.Scan(dest...)
}

func (p *pgxRowsAdapter) Close() error {
	return p.rows.Close()
}

func (p *pgxRowsAdapter) Err() error {
	return p.rows.Err()
}

// GormDBAdapter adapts *gorm.DB to implement DBQuerier
type GormDBAdapter struct {
	db *gorm.DB
}

func (g *GormDBAdapter) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	rows, err := g.db.Raw(sql, args...).Rows()
	if err != nil {
		return nil, err
	}
	return &pgxRowsAdapter{rows: rows}, nil
}

func (g *GormDBAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return &pgxRowAdapter{row: g.db.Raw(sql, args...).Row()}
}

// pgxRowAdapter adapts sql.Row to pgx.Row
type pgxRowAdapter struct {
	row *sql.Row
}

func (p *pgxRowAdapter) Scan(dest ...interface{}) error { return p.row.Scan(dest...) }

type ProductFilters struct {
	Categories []Category `json:"categories"`
	Brands     []Brand    `json:"brands"`
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

type Category struct {
	CategoryID       uuid.UUID  `json:"category_id"`
	Name             string     `json:"name"`
	ParentCategoryID *uuid.UUID `json:"parent_category_id,omitempty"`
}

type Brand struct {
	BrandID     uuid.UUID `json:"brand_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
}

// Product struct definition (add fields as per your DB schema)
type Product struct {
	ProductID   uuid.UUID `json:"product_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	BrandID     uuid.UUID `json:"brand_id"`
	CategoryID  uuid.UUID `json:"category_id"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	LocationID  uuid.UUID `json:"location_id"`
}

type Offer struct {
	OfferID         uuid.UUID `json:"offer_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	CategoryID      uuid.UUID `json:"category_id"`
	Discount        float64   `json:"discount"`
	DiscountPercent float64   `json:"discount_percent"`
	StartDate       string    `json:"start_date"`
	EndDate         string    `json:"end_date"`
	Active          bool      `json:"active"`
}

func NewProductUseCase(productRepo interfaces.ProductRepository, cloudService cloud.CloudService, db *gorm.DB) service.ProductUseCase {
	return &productUseCase{
		productRepo:  productRepo,
		cloudService: cloudService,
		DB:           &GormDBAdapter{db: db},
	}
}

func (c *productUseCase) FindAllCategories(ctx context.Context, pagination request.Pagination) ([]response.Category, error) {

	categories, err := c.productRepo.FindAllMainCategories(ctx, pagination)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed find all main categories")
	}

	return categories, nil
}

// Save category
func (c *productUseCase) SaveCategory(ctx context.Context, body request.Category, departmentId string) error {

	err := c.productRepo.SaveCategory(ctx, body, departmentId)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to save category")
	}

	return nil
}

// Save Sub category
func (c *productUseCase) SaveSubCategory(ctx context.Context, body request.SubCategory, departmentId string, category_id string) error {

	err := c.productRepo.SaveSubCategory(ctx, body, departmentId, category_id)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to save sub category")
	}

	return nil
}

// to add new variation for a category
func (c *productUseCase) SaveVariation(ctx context.Context, categoryID uint, variationNames []string) error {

	err := c.productRepo.Transactions(ctx, func(repo interfaces.ProductRepository) error {

		for _, variationName := range variationNames {

			variationExist, err := repo.IsVariationNameExistForCategory(ctx, variationName, categoryID)
			if err != nil {
				return utils.PrependMessageToError(err, "failed to check variation already exist")
			}

			if variationExist {
				return utils.PrependMessageToError(ErrVariationAlreadyExist, "variation name "+variationName)
			}

			err = c.productRepo.SaveVariation(ctx, categoryID, variationName)
			if err != nil {
				return utils.PrependMessageToError(err, "failed to save variation")
			}
		}
		return nil
	})

	return err
}

// to add new variation value for variation
func (c *productUseCase) SaveVariationOption(ctx context.Context, variationID uint, variationOptionValues []string) error {

	err := c.productRepo.Transactions(ctx, func(repo interfaces.ProductRepository) error {
		for _, variationValue := range variationOptionValues {

			valueExist, err := repo.IsVariationValueExistForVariation(ctx, variationValue, variationID)
			if err != nil {
				return utils.PrependMessageToError(err, "failed to check variation already exist")
			}
			if valueExist {
				return utils.PrependMessageToError(ErrVariationOptionAlreadyExist, "variation option value "+variationValue)
			}

			err = repo.SaveVariationOption(ctx, variationID, variationValue)
			if err != nil {
				return utils.PrependMessageToError(err, "failed to save variation option")
			}
		}
		return nil
	})

	return err
}

func (c *productUseCase) FindAllVariationsAndItsValues(ctx context.Context, categoryID uint) ([]response.Variation, error) {

	variations, err := c.productRepo.FindAllVariationsByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to find all variations of category")
	}

	// get all variation values of each variations
	for i, variation := range variations {

		variationOption, err := c.productRepo.FindAllVariationOptionsByVariationID(ctx, variation.ID)
		if err != nil {
			return nil, utils.PrependMessageToError(err, "failed to get variation option")
		}
		variations[i].VariationOptions = variationOption
	}
	return variations, nil
}

// to get all product
func (c *productUseCase) FindAllProducts(ctx context.Context, pagination request.Pagination, search string) ([]response.Product, error) {
	responseProducts, err := c.productRepo.FindAllProducts(ctx, pagination, search)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to get product details from database")
	}

	// Images are already stored as local paths, no need to fetch from cloud
	// Just return the products as-is with their local image paths
	return responseProducts, nil
}

// to get product by ID
func (c *productUseCase) FindProductByID(ctx context.Context, productID uint) (domain.Product, error) {
	product, err := c.productRepo.FindProductByID(ctx, productID)
	if err != nil {
		return product, utils.PrependMessageToError(err, "failed to get product from database")
	}
	return product, nil
}

// to add new product
func (c *productUseCase) SaveProduct(ctx context.Context, product request.Product, adminID string) (productID uint, err error) {

	// productNameExist, err := c.productRepo.IsProductNameExist(ctx, product.Name)
	// if err != nil {
	// 	return 0, utils.PrependMessageToError(err, "failed to check product name already exist")
	// }
	// if productNameExist {
	// 	return 0, utils.PrependMessageToError(ErrProductAlreadyExist, "product name "+product.Name)
	// }

	// Save image to uploads folder in project directory
	localPath, err := utils.SaveFileLocally(product.ImageFileHeader, "uploads/products")
	if err != nil {
		return 0, utils.PrependMessageToError(err, "failed to save image locally")
	}

	productID, err = c.productRepo.SaveProduct(ctx, domain.Product{
		Name:         product.Name,
		Description:  product.Description,
		CategoryID:   product.CategoryID,
		Image:        localPath,
		DepartmentID: product.DepartmentID,
	}, adminID)
	if err != nil {
		return 0, utils.PrependMessageToError(err, "failed to save product")
	}
	return productID, nil
}

// for add new productItem for a specific product
func (c *productUseCase) SaveProductItem(ctx context.Context, productItem request.ProductItem, adminID string, shopID uint) error {
	_, err := c.productRepo.SaveProductItem(ctx, productItem, adminID, shopID)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to save product item")
	}
	return nil
}

func (c *productUseCase) UpdateProductItem(ctx context.Context, productItemID uint, productItem request.ProductItem) error {
	err := c.productRepo.UpdateProductItem(ctx, productItemID, productItem)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to update product item")
	}
	return nil
}

// step 1 : get product_id and and all variation id as function parameter
// step 2 : initialize an map for storing product item id and its count(map[uint]int)
// step 3 : loop through the variation option ids
// step 4 : then find all product items ids with given product id and the loop variation option id
// step 5 : if the product item array length is zero means the configuration not exist return false
// step 6 : then loop through the product items ids array(got from database)
// step 7 : add each id on the map and increment its count
// step 8 : check if any of the product items id's count is greater than the variation options ids length then return true
// step 9 : if the loop exist means product configuration is not exist
func (c *productUseCase) isProductVariationCombinationExist(productID uint, variationOptionIDs []uint) (exist bool, err error) {

	setOfIds := map[uint]int{}

	for _, variationOptionID := range variationOptionIDs {

		productItemIds, err := c.productRepo.FindAllProductItemIDsByProductIDAndVariationOptionID(context.TODO(),
			productID, variationOptionID)
		if err != nil {
			return false, utils.PrependMessageToError(err, "failed to find product item ids from database using product id and variation option id")
		}

		if len(productItemIds) == 0 {
			return false, nil
		}

		for _, productItemID := range productItemIds {

			setOfIds[productItemID]++
			// if any of the ids count is equal to array length it means product item id of this is the existing product item of this configuration
			if setOfIds[productItemID] >= len(variationOptionIDs) {
				return true, nil
			}
		}
	}
	return false, nil
}

// for get all productItem for a specific product
func (c *productUseCase) FindAllProductItems(ctx context.Context, adminId string, keyword string, categoryID, brandID, locationID *string, offer string, sortby string, pagination *request.Pagination, filterByShopID string) ([]response.ProductItems, error) {

	productItems, err := c.productRepo.FindAllProductItems(ctx, adminId, keyword, categoryID, brandID, locationID, offer, sortby, pagination, filterByShopID)
	if err != nil {
		return productItems, err
	}

	return productItems, nil
}

func (c *productUseCase) FindLowViewProductItems(ctx context.Context, adminId string, keyword string, categoryID, brandID, locationID *string, sortby string, pagination *request.Pagination, filterByShopID *string) ([]response.ProductItems, error) {

	productItems, err := c.productRepo.FindLowViewProductItems(ctx, adminId, keyword, categoryID, brandID, locationID, sortby, pagination, filterByShopID)
	if err != nil {
		return productItems, err
	}

	return productItems, nil
}

func (c *productUseCase) DeleteProductItem(ctx context.Context, productItemID uint) error {
	err := c.productRepo.DeleteProductItem(ctx, productItemID)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to delete product item")
	}
	return nil
}

func (c *productUseCase) UpdateProduct(ctx context.Context, updateDetails domain.Product) error {

	nameExistForOther, err := c.productRepo.IsProductNameExistForOtherProduct(ctx, updateDetails.Name, updateDetails.ID)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to check product name already exist for other product")
	}

	if nameExistForOther {
		return utils.PrependMessageToError(ErrProductAlreadyExist, "product name "+updateDetails.Name)
	}

	// c.productRepo.FindProductByID(ctx, updateDetails.ID)

	err = c.productRepo.UpdateProduct(ctx, updateDetails)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to update product")
	}
	return nil
}

// Helper to convert *uuid.UUID to *string
func uuidToStringPtr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func SafeIntToUint64(i int) (uint64, error) {
	if i < 0 {
		return 0, fmt.Errorf("cannot convert negative int (%d) to uint64", i)
	}
	return uint64(i), nil
}

func (c *productUseCase) SearchProducts(ctx context.Context, keyword string, categoryID *string, brandID *string, locationID *string, shopID *string, latitude, longitude, radius float64, pincode *uint, limit, offset int) ([]response.ProductItems, error) {
	limitUint64, err := SafeIntToUint64(limit)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "invalid limit for pagination")
	}

	offsetUint64, err := SafeIntToUint64(offset)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "invalid offset for pagination")
	}

	pagination := request.Pagination{
		Limit:  limitUint64,
		Offset: offsetUint64,
	}
	resProducts, err := c.productRepo.SearchProducts(ctx, keyword, categoryID, brandID, locationID, shopID, latitude, longitude, radius, pincode, pagination)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to search products")
	}

	return resProducts, nil
}

func (s *productUseCase) GetProductNameSuggestions(ctx context.Context, prefix string) ([]string, error) {
	// Using ILIKE with prefix% for simple autocomplete
	query := `SELECT DISTINCT name FROM products WHERE name ILIKE $1 LIMIT 10`
	rows, err := s.DB.Query(ctx, query, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, name)
	}
	return suggestions, nil
}

func (s *productUseCase) GetProductFilters(ctx context.Context) (response.ProductFilters, error) {
	var filters response.ProductFilters

	// Fetch distinct categories
	categoryQuery := `SELECT DISTINCT c.category_id, c.name FROM categories c JOIN products_items pi ON c.category_id = pi.category_id ORDER BY c.name`
	catRows, err := s.DB.Query(ctx, categoryQuery)
	if err != nil {
		return filters, err
	}
	defer catRows.Close()

	for catRows.Next() {
		var c response.Category
		if err := catRows.Scan(&c.ID, &c.Name); err != nil { // ID and Name field assumed
			return filters, err
		}
		filters.Categories = append(filters.Categories, response.CategoryFilter{
			CategoryID:   c.ID,
			CategoryName: c.Name,
			Count:        0, // You might want to calculate the count based on your data
		})
	}
	if err := catRows.Err(); err != nil {
		return filters, err
	}

	// Fetch distinct brands
	brandQuery := `SELECT DISTINCT b.brand_id, b.name FROM brands b JOIN products_items pi ON b.brand_id = pi.brand_id ORDER BY b.name`
	brandRows, err := s.DB.Query(ctx, brandQuery)
	if err != nil {
		return filters, err
	}
	defer brandRows.Close()

	for brandRows.Next() {
		var b struct {
			BrandID uuid.UUID
			Name    string
		}
		if err := brandRows.Scan(&b.BrandID, &b.Name); err != nil { // Use b.ID here if response.Brand uses ID
			return filters, err
		}
		filters.Brands = append(filters.Brands, response.BrandFilter{
			BrandID:   b.BrandID, // both are uuid.UUID
			BrandName: b.Name,
			Count:     0, // or calculated count
		})
	}
	if err := brandRows.Err(); err != nil {
		return filters, err
	}

	return filters, nil
}

// to get all product locations
func (s *productUseCase) GetProductLocations(ctx context.Context) ([]response.Location, error) {
	query := `
        SELECT DISTINCT l.location_id, l.country, l.state, l.city, l.area, l.pincode
        FROM locations l
        JOIN products p ON l.location_id = p.location_id
        ORDER BY l.country, l.state, l.city, l.area
    `
	rows, err := s.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []response.Location
	for rows.Next() {
		var loc response.Location
		err := rows.Scan(&loc.LocationID, &loc.Country, &loc.State, &loc.City, &loc.Area, &loc.Pincode)
		if err != nil {
			return nil, err
		}
		locations = append(locations, response.Location{
			LocationID: loc.LocationID,
			Country:    loc.Country,
			State:      loc.State,
			City:       loc.City,
			Area:       loc.Area,
			Pincode:    loc.Pincode,
			Latitude:   loc.Latitude,
			Longitude:  loc.Longitude,
		})
	}
	return locations, nil
}

// services/product_service.go

func (s *productUseCase) GetAllCategories(ctx context.Context) ([]Category, error) {
	query := `SELECT category_id, name, parent_category_id FROM categories ORDER BY name`
	rows, err := s.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		err := rows.Scan(&c.CategoryID, &c.Name, &c.ParentCategoryID)
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// services/product_service.go

func (s *productUseCase) GetProductsByCategory(ctx context.Context, categoryID int, limit, offset int) ([]response.Product, error) {
	query := `
        SELECT product_id, name, description, brand_id, category_id, price, stock, location_id
        FROM products
        WHERE category_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := s.DB.Query(ctx, query, categoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []response.Product
	for rows.Next() {
		var p response.Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.BrandID,
			&p.CategoryID,
			&p.Price,
			&p.DiscountPrice,
			&p.CategoryName,
			&p.MainCategoryName,
			&p.BrandName,
			&p.Image,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, response.Product{
			ID:            p.ID,
			Name:          p.Name,
			Description:   p.Description,
			BrandID:       p.BrandID,
			CategoryID:    p.CategoryID,
			Price:         p.Price,
			DiscountPrice: p.DiscountPrice,
			Stock:         p.Stock,
			LocationID:    p.LocationID,
		})
	}
	return products, nil
}

// services/product_service.go

func (s *productUseCase) GetAllBrands(ctx context.Context) ([]response.Brand, error) {
	query := `SELECT brand_id, name, description FROM brands ORDER BY name`
	rows, err := s.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []response.Brand
	for rows.Next() {
		var b Brand
		if err := rows.Scan(&b.BrandID, &b.Name, &b.Description); err != nil {
			return nil, err
		}
		// Convert Brand to response.Brand (assuming same fields)
		brands = append(brands, response.Brand{
			BrandID:     b.BrandID,
			Name:        b.Name,
			Description: b.Description,
		})
	}

	return brands, nil
}

// services/product_service.go

func (s *productUseCase) GetProductsByBrand(ctx context.Context, brandID int, limit, offset int) ([]response.Product, error) {
	query := `
        SELECT product_id, name, description, brand_id, category_id, price, stock, location_id
        FROM products
        WHERE brand_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := s.DB.Query(ctx, query, brandID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []response.Product
	for rows.Next() {
		var p response.Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.BrandID,
			&p.CategoryID,
			&p.Price,
			&p.DiscountPrice,
			&p.CategoryName,
			&p.MainCategoryName,
			&p.BrandName,
			&p.Image,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, response.Product{
			ID:            p.ID,
			Name:          p.Name,
			Description:   p.Description,
			BrandID:       p.BrandID,
			CategoryID:    p.CategoryID,
			Price:         p.Price,
			DiscountPrice: p.DiscountPrice,
			Stock:         p.Stock,
			LocationID:    p.LocationID,
		})
	}
	return products, nil
}

// services/product_service.go

func (s *productUseCase) GetCategoryFilters(ctx context.Context) ([]response.Category, error) {
	query := `SELECT category_id, name FROM categories ORDER BY name`
	rows, err := s.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []response.Category
	for rows.Next() {
		var c struct {
			CategoryID uint
			Name       string
		}
		if err := rows.Scan(&c.CategoryID, &c.Name); err != nil {
			return nil, err
		}

		// Map scanned result to response.Category
		categories = append(categories, response.Category{
			ID:   c.CategoryID,
			Name: c.Name,
		})
	}
	return categories, nil
}

// services/product_service.go

func (s *productUseCase) GetBrandFilters(ctx context.Context) ([]response.Brand, error) {
	query := `SELECT brand_id, name FROM brands ORDER BY name`
	rows, err := s.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []response.Brand
	for rows.Next() {
		var b Brand
		if err := rows.Scan(&b.BrandID, &b.Name, &b.Description); err != nil {
			return nil, err
		}
		// Convert Brand to response.Brand (assuming same fields)
		brands = append(brands, response.Brand{
			BrandID:     b.BrandID,
			Name:        b.Name,
			Description: b.Description,
		})
	}
	return brands, nil
}

// services/product_service.go

func (s *productUseCase) GetLocationFilter(ctx context.Context) ([]response.Location, error) {
	query := `
        SELECT DISTINCT location_id, country, state, city, area, pincode 
        FROM locations
        ORDER BY country, state, city, area
    `
	rows, err := s.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []response.Location
	for rows.Next() {
		var loc Location
		if err := rows.Scan(&loc.LocationID, &loc.Country, &loc.State, &loc.City, &loc.Area, &loc.Pincode); err != nil {
			return nil, err
		}

		// Map internal Location to response.Location
		locations = append(locations, response.Location{
			LocationID: loc.LocationID,
			Country:    loc.Country,
			State:      loc.State,
			City:       loc.City,
			Area:       loc.Area,
			Pincode:    loc.Pincode,
			Latitude:   loc.Latitude,
			Longitude:  loc.Longitude,
		})
	}
	return locations, nil
}

// services/product_service.go

func (s *productUseCase) GetProductsByLocation(ctx context.Context, locationID int, limit, offset int) ([]response.Product, error) {
	query := `
        SELECT product_id, name, description, brand_id, category_id, price, stock, location_id
        FROM products
        WHERE location_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := s.DB.Query(ctx, query, locationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []response.Product
	for rows.Next() {
		var p response.Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.BrandID,
			&p.CategoryID,
			&p.Price,
			&p.DiscountPrice,
			&p.CategoryName,
			&p.MainCategoryName,
			&p.BrandName,
			&p.Image,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, response.Product{
			ID:            p.ID,
			Name:          p.Name,
			Description:   p.Description,
			BrandID:       p.BrandID,
			CategoryID:    p.CategoryID,
			Price:         p.Price,
			DiscountPrice: p.DiscountPrice,
			Stock:         p.Stock,
			LocationID:    p.LocationID,
		})
	}
	return products, nil
}

// services/product_service.go

func (uc *productUseCase) GetAllAreas(ctx context.Context) ([]response.Area, error) {
	// fetch area strings as before
	areaNames := []string{"Area1", "Area2"}

	// map to []response.Area
	areas := make([]response.Area, len(areaNames))
	for i, name := range areaNames {
		areas[i] = response.Area{
			Name: name,
			// set other fields if needed
		}
	}
	return areas, nil
}

func (s *productUseCase) GetAllCities(ctx context.Context) ([]string, error) {
	return s.getDistinctLocationField(ctx, "city")
}

func (s *productUseCase) GetAllStates(ctx context.Context) ([]string, error) {
	return s.getDistinctLocationField(ctx, "state")
}

func (s *productUseCase) GetAllCountries(ctx context.Context) ([]string, error) {
	return s.getDistinctLocationField(ctx, "country")
}

func (s *productUseCase) GetAllPincodes(ctx context.Context) ([]string, error) {
	return s.getDistinctLocationField(ctx, "pincode")
}

// Helper:
func (s *productUseCase) getDistinctLocationField(ctx context.Context, field string) ([]string, error) {
	query := fmt.Sprintf(`SELECT DISTINCT %s FROM locations WHERE %s IS NOT NULL ORDER BY %s`, field, field, field)

	rows, err := s.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err != nil {
			return nil, err
		}
		results = append(results, val)
	}
	return results, nil
}

// services/product_service.go

func (s *productUseCase) GetCitiesByState(ctx context.Context, state string) ([]string, error) {
	query := `SELECT DISTINCT city FROM locations WHERE state = $1 ORDER BY city`
	return s.fetchLocations(ctx, query, state)
}

func (s *productUseCase) GetAreasByCity(ctx context.Context, city string) ([]string, error) {
	query := `SELECT DISTINCT area FROM locations WHERE city = $1 ORDER BY area`
	return s.fetchLocations(ctx, query, city)
}

func (s *productUseCase) GetPincodesByArea(ctx context.Context, area string) ([]string, error) {
	query := `SELECT DISTINCT pincode FROM locations WHERE area = $1 ORDER BY pincode`
	return s.fetchLocations(ctx, query, area)
}

func (s *productUseCase) fetchLocations(ctx context.Context, query string, arg string) ([]string, error) {
	rows, err := s.DB.Query(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err != nil {
			return nil, err
		}
		results = append(results, val)
	}
	return results, nil
}

func (s *productUseCase) GetLocationByPincode(ctx context.Context, pincodeID string) (response.Location, error) {
	query := `
      SELECT location_id, country, state, city, area, pincode, latitude, longitude 
      FROM locations WHERE pincode = $1 LIMIT 1
    `
	row := s.DB.QueryRow(ctx, query, pincodeID)

	var loc Location
	err := row.Scan(&loc.LocationID, &loc.Country, &loc.State, &loc.City, &loc.Area, &loc.Pincode, &loc.Latitude, &loc.Longitude)
	if err == pgx.ErrNoRows {
		// Return zero value and no error if not found
		return response.Location{}, nil
	}
	if err != nil {
		return response.Location{}, err
	}

	// Map internal Location to response.Location
	location := response.Location{
		LocationID: loc.LocationID,
		Country:    loc.Country,
		State:      loc.State,
		City:       loc.City,
		Area:       loc.Area,
		Pincode:    loc.Pincode,
		Latitude:   loc.Latitude,
		Longitude:  loc.Longitude,
	}

	return location, nil
}

// services/product_service.go

const DefaultRadiusMeters = 5000 // or get from config/env

func (s *productUseCase) GetNearbyProductsByPincode(ctx context.Context, pincode string, limit, offset int) ([]response.ProductItems, error) {

	// Check if pincode exists in shop_details
	checkQuery := `SELECT COUNT(*) FROM shop_details WHERE pincode = $1`
	var count int
	row := s.DB.QueryRow(ctx, checkQuery, pincode)
	err := row.Scan(&count)
	if err != nil || count == 0 {
		return []response.ProductItems{}, nil
	}

	// Get products from shops with the exact pincode
	query := `
		SELECT DISTINCT
			pi.id, 
			pi.shop_id,
			pi.sub_category_name, 
			pi.category_id, 
			pi.department_id, 
			pi.sub_category_id,
			pi.product_item_images, 
			pi.dynamic_fields, 
			pi.created_at, 
			pi.updated_at,
			COALESCE(c.name, '') AS category_name,
			COALESCE(d.name, '') AS main_category_name,
			COALESCE(sc.name, '') AS sub_category_name_ref,
			COALESCE(sc.image_url, '') AS sub_category_image_url
		FROM product_items pi
		INNER JOIN shop_details sd ON pi.shop_id = sd.id
		LEFT JOIN categories c ON pi.category_id = c.id
		LEFT JOIN departments d ON pi.department_id = d.id
		LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
		WHERE sd.pincode = $1
		ORDER BY pi.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.DB.Query(ctx, query, pincode, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var products []response.ProductItems
	for rows.Next() {
		var id uint
		var shopID uint
		var name string
		var categoryID uint
		var deptID uint
		var subCatID uint
		var images string
		var dynamicFields []byte
		var createdAt time.Time
		var updatedAt time.Time
		var categoryName string
		var deptName string
		var subCatName string
		var subCatImageURL string

		if err := rows.Scan(
			&id,
			&shopID,
			&name,
			&categoryID,
			&deptID,
			&subCatID,
			&images,
			&dynamicFields,
			&createdAt,
			&updatedAt,
			&categoryName,
			&deptName,
			&subCatName,
			&subCatImageURL,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}

		var dynamicFieldsMap map[string]interface{}
		if len(dynamicFields) > 0 {
			if err := json.Unmarshal(dynamicFields, &dynamicFieldsMap); err != nil {
				return nil, utils.PrependMessageToError(err, "failed to unmarshal dynamicFields")
			}
		}

		// Convert images string to []string (try JSON unmarshal, fallback to comma split)
		var imagesSlice []string
		if err := json.Unmarshal([]byte(images), &imagesSlice); err != nil {
			// fallback: treat as comma-separated string
			if images != "" {
				imagesSlice = append(imagesSlice, images)
			}
		}

		products = append(products, response.ProductItems{
			ID:                  id,
			Name:                name,
			CategoryID:          categoryID,
			CategoryName:        categoryName,
			DepartmentID:        deptID,
			SubCategoryID:       subCatID,
			SubCategoryImageURL: subCatImageURL,
			DynamicFields:       dynamicFieldsMap,
			ProductItemImages:   imagesSlice,
			MainCategoryName:    deptName,
			CreatedAt:           createdAt,
			UpdatedAt:           updatedAt,
		})
	}

	if len(products) == 0 {
		return []response.ProductItems{}, nil
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return products, nil
}

// services/product_service.go

func (c *productUseCase) GetProductsByRadius(ctx context.Context, latitude float64, longitude float64, radiusKm float64, limit, offset int) ([]response.ProductItems, error) {
	// default radius if not provided
	if radiusKm <= 0 {
		radiusKm = 10.0 // default 10 km
	}


	query := `
		SELECT * FROM (
			SELECT pi.id, pi.shop_id, pi.sub_category_name, pi.category_id, pi.department_id, pi.sub_category_id,
				pi.product_item_images, pi.dynamic_fields, pi.created_at, pi.updated_at,
				COALESCE(c.name, '') AS category_name,
				COALESCE(d.name, '') AS main_category_name,
				COALESCE(sc.image_url, '') AS sub_category_image_url,
				(6371 * acos(
					cos(radians($1)) * cos(radians(sd.latitude)) *
					cos(radians(sd.longitude) - radians($2)) +
					sin(radians($1)) * sin(radians(sd.latitude))
				)) AS distance_km
			FROM product_items pi
			INNER JOIN shop_details sd ON pi.shop_id = sd.id
			LEFT JOIN categories c ON pi.category_id = c.id
			LEFT JOIN departments d ON pi.department_id = d.id
			LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
			WHERE sd.latitude IS NOT NULL AND sd.longitude IS NOT NULL
		) AS subquery
		WHERE distance_km <= $3
		ORDER BY distance_km
		LIMIT $4 OFFSET $5;
	`

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
	}

	rows, err := c.DB.Query(ctx, query, latitude, longitude, radiusKm, limit, offset)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to get products by radius")
	}
	defer rows.Close()

	var products []response.ProductItems
	count := 0
	for rows.Next() {
		var id uint
		var shopID uint
		var name string
		var categoryID uint
		var deptID uint
		var subCatID uint
		var images string
		var dynamicFields []byte
		var createdAt time.Time
		var updatedAt time.Time
		var categoryName string
		var mainCategoryName string
		var subCatImageURL string
		var distanceKm float64

		if err := rows.Scan(
			&id,
			&shopID,
			&name,
			&categoryID,
			&deptID,
			&subCatID,
			&images,
			&dynamicFields,
			&createdAt,
			&updatedAt,
			&categoryName,
			&mainCategoryName,
			&subCatImageURL,
			&distanceKm,
		); err != nil {
			return nil, utils.PrependMessageToError(err, "failed to scan product row")
		}


		var dynamicFieldsMap map[string]interface{}
		if len(dynamicFields) > 0 {
			if err := json.Unmarshal(dynamicFields, &dynamicFieldsMap); err != nil {
				return nil, utils.PrependMessageToError(err, "failed to unmarshal dynamicFields")
			}
		}

		// Convert images string to []string (try JSON unmarshal, fallback to comma split)
		var imagesSlice []string
		if err := json.Unmarshal([]byte(images), &imagesSlice); err != nil {
			// fallback: treat as comma-separated string
			if images != "" {
				imagesSlice = append(imagesSlice, images)
			}
		}

		products = append(products, response.ProductItems{
			ID:                  id,
			Name:                name,
			CategoryID:          categoryID,
			DepartmentID:        deptID,
			SubCategoryID:       subCatID,
			ProductItemImages:   imagesSlice,
			DynamicFields:       dynamicFieldsMap,
			CategoryName:        categoryName,
			MainCategoryName:    mainCategoryName,
			SubCategoryImageURL: subCatImageURL,
			CreatedAt:           createdAt,
			UpdatedAt:           updatedAt,
			// Optionally: DistanceKm: distanceKm,
		})
		count++
	}
	if count == 0 {
	}
	return products, nil
}

func (c *productUseCase) SaveDepartment(ctx context.Context, departmentName string) error {
	err := c.productRepo.SaveDepartment(ctx, departmentName)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to save department")
	}
	return nil
}

func (c *productUseCase) GetAllDepartments(ctx context.Context) ([]response.Department, error) {
	departments, err := c.productRepo.GetAllDepartments(ctx)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to get departments")
	}
	// loop through departments and map to response.Department
	resDepartments := make([]response.Department, len(departments))
	for i, dept := range departments {
		resDepartments[i] = response.Department{
			ID:       dept.ID,
			Name:     dept.Name,
			ImageUrl: dept.ImageUrl,
		}
	}
	return resDepartments, nil
}

func (c *productUseCase) GetDepartmentByID(ctx context.Context, departmentID uint) (response.Department, error) {
	department, err := c.productRepo.GetDepartmentByID(ctx, departmentID)
	if err != nil {
		return response.Department{}, utils.PrependMessageToError(err, "failed to get department by id")
	}
	return response.Department{
		ID:   department.ID,
		Name: department.Name,
	}, nil
}

func (c *productUseCase) GetAllSubCategories(ctx context.Context) ([]response.SubCategory, error) {
	subCategories, err := c.productRepo.GetAllSubCategories(ctx)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to get sub-categories")
	}
	return subCategories, nil
}

func (c *productUseCase) GetAllCategoriesByDepartmentID(ctx context.Context, departmentID uint) ([]response.Category, error) {
	categories, err := c.productRepo.GetAllCategoriesByDepartmentID(ctx, departmentID)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to get categories by department id")
	}
	return categories, nil
}

func (c *productUseCase) GetAllSubCategoriesByCategoryID(ctx context.Context, categoryID uint) ([]response.SubCategory, error) {
	subCategories, err := c.productRepo.GetAllSubCategoriesByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to get sub-categories by category id")
	}
	return subCategories, nil
}

// SaveSubTypeAttribute saves a new sub type attribute
func (c *productUseCase) SaveSubTypeAttribute(ctx context.Context, subCategoryID uint, attribute request.SubTypeAttribute) error {
	// Convert request to domain model
	domainAttribute := domain.SubTypeAttributes{
		FieldName:  attribute.FieldName,
		FieldType:  attribute.FieldType,
		IsRequired: attribute.IsRequired,
		SortOrder:  attribute.SortOrder,
	}

	return c.productRepo.SaveSubTypeAttribute(ctx, subCategoryID, domainAttribute)
}

// GetAllSubTypeAttributes retrieves all sub type attributes for a subcategory
func (c *productUseCase) GetAllSubTypeAttributes(ctx context.Context, subCategoryID uint) ([]response.SubTypeAttribute, error) {
	attributes, err := c.productRepo.GetAllSubTypeAttributes(ctx, subCategoryID)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to get sub type attributes")
	}
	return attributes, nil
}

// GetSubTypeAttributeByID retrieves a single sub type attribute by ID
func (c *productUseCase) GetSubTypeAttributeByID(ctx context.Context, attributeID uint) (response.SubTypeAttribute, error) {
	attribute, err := c.productRepo.GetSubTypeAttributeByID(ctx, attributeID)
	if err != nil {
		return response.SubTypeAttribute{}, utils.PrependMessageToError(err, "failed to get sub type attribute by id")
	}
	return attribute, nil
}

// SaveSubTypeAttributeOption saves a new option for a sub type attribute
func (c *productUseCase) SaveSubTypeAttributeOption(ctx context.Context, attributeID uint, option request.SubTypeAttributeOption) error {
	// Convert request to domain model
	domainOption := domain.SubTypeAttributeOptions{
		OptionValue: option.OptionValue,
		SortOrder:   option.SortOrder,
	}

	return c.productRepo.SaveSubTypeAttributeOption(ctx, attributeID, domainOption)
}

// GetAllSubTypeAttributeOptions retrieves all options for a sub type attribute
func (c *productUseCase) GetAllSubTypeAttributeOptions(ctx context.Context, attributeID uint) ([]response.SubTypeAttributeOption, error) {
	options, err := c.productRepo.GetAllSubTypeAttributeOptions(ctx, attributeID)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to get sub type attribute options")
	}
	return options, nil
}

// GetSubTypeAttributeOptionByID retrieves a single option by ID
func (c *productUseCase) GetSubTypeAttributeOptionByID(ctx context.Context, optionID uint) (response.SubTypeAttributeOption, error) {
	option, err := c.productRepo.GetSubTypeAttributeOptionByID(ctx, optionID)
	if err != nil {
		return response.SubTypeAttributeOption{}, utils.PrependMessageToError(err, "failed to get sub type attribute option by id")
	}
	return option, nil
}

// SaveCategoryImage saves a new category image
func (c *productUseCase) SaveCategoryImage(ctx context.Context, categoryID uint, image request.CategoryImage) error {
	categoryImage := domain.CategoryImage{
		CategoryID: categoryID,
		ImageURL:   image.ImageURL,
		AltText:    image.AltText,
		SortOrder:  image.SortOrder,
		IsActive:   image.IsActive,
	}
	if categoryImage.IsActive == false {
		categoryImage.IsActive = true
	}
	err := c.productRepo.SaveCategoryImage(ctx, categoryID, categoryImage)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to save category image")
	}
	return nil
}

// GetAllCategoryImages retrieves all images for a category
func (c *productUseCase) GetAllCategoryImages(ctx context.Context, categoryID uint) ([]response.CategoryImage, error) {
	images, err := c.productRepo.GetAllCategoryImages(ctx, categoryID)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to get all category images")
	}
	return images, nil
}

// GetCategoryImageByID retrieves a single category image
func (c *productUseCase) GetCategoryImageByID(ctx context.Context, imageID uint) (response.CategoryImage, error) {
	image, err := c.productRepo.GetCategoryImageByID(ctx, imageID)
	if err != nil {
		return response.CategoryImage{}, utils.PrependMessageToError(err, "failed to get category image by id")
	}
	return image, nil
}

// UpdateCategoryImage updates an existing category image
func (c *productUseCase) UpdateCategoryImage(ctx context.Context, imageID uint, image request.CategoryImage) error {
	categoryImage := domain.CategoryImage{
		ID:        imageID,
		ImageURL:  image.ImageURL,
		AltText:   image.AltText,
		SortOrder: image.SortOrder,
		IsActive:  image.IsActive,
	}
	err := c.productRepo.UpdateCategoryImage(ctx, categoryImage)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to update category image")
	}
	return nil
}

// DeleteCategoryImage soft deletes a category image
func (c *productUseCase) DeleteCategoryImage(ctx context.Context, imageID uint) error {
	err := c.productRepo.DeleteCategoryImage(ctx, imageID)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to delete category image")
	}
	return nil
}

func (c *productUseCase) GetProductItemByID(ctx context.Context, productItemID uint) (response.ProductItems, error) {
	productItem, err := c.productRepo.GetProductItemByID(ctx, productItemID)
	if err != nil {
		return response.ProductItems{}, utils.PrependMessageToError(err, "failed to get product item by id")
	}
	return productItem, nil
}

func (c *productUseCase) IncrementProductItemViewCount(ctx context.Context, productItemID uint, adminID string) error {
	err := c.productRepo.IncrementProductItemViewCount(ctx, productItemID, adminID)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to increment product item view count")
	}
	return nil
}

func (c *productUseCase) GetProductItemViewCount(ctx context.Context, productItemID uint, adminID string) (uint, error) {
	count, err := c.productRepo.GetProductItemViewCount(ctx, productItemID, adminID)
	if err != nil {
		return 0, utils.PrependMessageToError(err, "failed to get product item view count")
	}
	return count, nil
}

func (s *productUseCase) FindProductItemFilters(ctx context.Context, adminID string, shopID uint) ([]domain.ProductItemFilterType, error) {
	filters, err := s.productRepo.FindProductItemFilters(ctx, adminID, shopID)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to find product item filters")
	}
	return filters, nil
}

func (s *productUseCase) GetProductItemsByOfferID(ctx context.Context, offerID uint, categoryID int, departmentID int, subCategoryID int, latStr string, lngStr string, pincode string, radiusKm float64, limit int, offset int) ([]response.ProductItems, error) {
	products, err := s.productRepo.GetProductItemsByOfferID(ctx, offerID, categoryID, departmentID, subCategoryID, latStr, lngStr, pincode, radiusKm, limit, offset)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to get product items by offer id")
	}
	return products, nil
}

// services/product_service.go
