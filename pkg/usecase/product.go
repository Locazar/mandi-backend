package usecase

import (
	"context"
	"database/sql"
	"fmt"

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

	for i, category := range categories {

		subCategory, err := c.productRepo.FindAllSubCategories(ctx, category.ID)
		if err != nil {
			return nil, utils.PrependMessageToError(err, "failed to find sub categories")
		}
		categories[i].SubCategory = subCategory
	}

	return categories, nil
}

// Save category
func (c *productUseCase) SaveCategory(ctx context.Context, categoryName string) error {

	categoryExist, err := c.productRepo.IsCategoryNameExist(ctx, categoryName)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to check category already exist")
	}
	if categoryExist {
		return ErrCategoryAlreadyExist
	}

	err = c.productRepo.SaveCategory(ctx, categoryName)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to save category")
	}

	return nil
}

// Save Sub category
func (c *productUseCase) SaveSubCategory(ctx context.Context, subCategory request.SubCategory) error {

	subCatExist, err := c.productRepo.IsSubCategoryNameExist(ctx, subCategory.Name, subCategory.CategoryID)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to check sub category already exist")
	}
	if subCatExist {
		return ErrCategoryAlreadyExist
	}

	err = c.productRepo.SaveSubCategory(ctx, subCategory.CategoryID, subCategory.Name)
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
func (c *productUseCase) FindAllProducts(ctx context.Context, pagination request.Pagination) ([]response.Product, error) {
	responseProducts, err := c.productRepo.FindAllProducts(ctx, pagination)
	fmt.Printf("%+v\n", responseProducts)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to get product details from database")
	}

	for i, p := range responseProducts {
		url, err := c.cloudService.GetFileUrl(ctx, p.Image)
		if err != nil {
			responseProducts[i].Image = p.Image
		} else {
			responseProducts[i].Image = url
		}
	}

	return responseProducts, nil
}

// to add new product
func (c *productUseCase) SaveProduct(ctx context.Context, product request.Product) error {

	productNameExist, err := c.productRepo.IsProductNameExist(ctx, product.Name)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to check product name already exist")
	}
	if productNameExist {
		return utils.PrependMessageToError(ErrProductAlreadyExist, "product name "+product.Name)
	}

	uploadID, err := c.cloudService.SaveFile(ctx, product.ImageFileHeader)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to save image on cloud storage")
	}

	err = c.productRepo.SaveProduct(ctx, domain.Product{
		Name:        product.Name,
		Description: product.Description,
		CategoryID:  product.CategoryID,
		BrandID:     product.BrandID,
		Price:       product.Price,
		Image:       uploadID,
	})
	if err != nil {
		return utils.PrependMessageToError(err, "failed to save product")
	}
	return nil
}

// for add new productItem for a specific product
func (c *productUseCase) SaveProductItem(ctx context.Context, productID uint, productItem request.ProductItem) error {

	variationCount, err := c.productRepo.FindVariationCountForProduct(ctx, productID)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to get variation count of product from database")
	}

	fmt.Printf("productItem VariationOptionIDs: %+v\n", productItem.VariationOptionIDs)
	fmt.Printf("Expected variation count: %d\n", variationCount)

	if len(productItem.VariationOptionIDs) != int(variationCount) {
		return ErrNotEnoughVariations
	}

	// check the given all combination already exist (Color:Red with Size:M)
	productItemExist, err := c.isProductVariationCombinationExist(productID, productItem.VariationOptionIDs)
	if err != nil {
		return err
	}
	if productItemExist {
		return ErrProductItemAlreadyExist
	}

	err = c.productRepo.Transactions(ctx, func(trxRepo interfaces.ProductRepository) error {

		sku := utils.GenerateSKU()
		newProductItem := domain.ProductItem{
			ProductID:  productID,
			QtyInStock: productItem.QtyInStock,
			Price:      productItem.Price,
			SKU:        sku,
		}

		productItemID, err := trxRepo.SaveProductItem(ctx, newProductItem)
		if err != nil {
			return utils.PrependMessageToError(err, "failed to save product item")
		}

		errChan := make(chan error, 2)
		newCtx, cancel := context.WithCancel(ctx) // for any of one of goroutine get error then cancel the working of other also
		defer cancel()

		go func() {
			// save all product configurations based on given variation option id
			for _, variationOptionID := range productItem.VariationOptionIDs {

				select {
				case <-newCtx.Done():
					return
				default:
					err = trxRepo.SaveProductConfiguration(ctx, productItemID, variationOptionID)
					if err != nil {
						errChan <- utils.PrependMessageToError(err, "failed to save product_item configuration")
						return
					}
				}
			}
			errChan <- nil
		}()

		go func() {
			// save all images for the given product item
			for _, imageFile := range productItem.ImageFileHeaders {

				select {
				case <-newCtx.Done():
					return
				default:
					// upload image on cloud
					uploadID, err := c.cloudService.SaveFile(ctx, imageFile)
					if err != nil {
						errChan <- utils.PrependMessageToError(err, "failed to upload image to cloud")
						return
					}
					// save upload id on database
					err = trxRepo.SaveProductItemImage(ctx, productItemID, uploadID)
					if err != nil {
						errChan <- utils.PrependMessageToError(err, "failed to save image for product item on database")
						return
					}
				}
			}
			errChan <- nil
		}()

		// wait for the both go routine to complete
		for i := 1; i <= 2; i++ {

			select {
			case <-ctx.Done():
				return nil
			case err := <-errChan:
				if err != nil { // if any of the goroutine send error then return the error
					return err
				}
				// no error then continue for the next check of select
			}
		}

		return nil
	})

	if err != nil {
		return err
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
func (c *productUseCase) FindAllProductItems(ctx context.Context, productID uint) ([]response.ProductItems, error) {

	productItems, err := c.productRepo.FindAllProductItems(ctx, productID)
	if err != nil {
		return productItems, err
	}

	errChan := make(chan error, 2)
	newCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {

		// get all variation values of each product items
		for i := range productItems {

			select { // checking each time context is cancelled or not
			case <-ctx.Done():
				return
			default:
				variationValues, err := c.productRepo.FindAllVariationValuesOfProductItem(ctx, productItems[i].ID)
				if err != nil {
					errChan <- utils.PrependMessageToError(err, "failed to find variation values product item")
					return
				}
				productItems[i].VariationValues = variationValues
			}
		}
		errChan <- nil
	}()

	go func() {
		// get all images of each product items
		for i := range productItems {

			select { // checking each time context is cancelled or not
			case <-newCtx.Done():
				return
			default:
				images, err := c.productRepo.FindAllProductItemImages(ctx, productItems[i].ID)

				imageUrls := make([]string, len(images))

				for j := range images {

					url, err := c.cloudService.GetFileUrl(ctx, images[j])
					if err != nil {
						errChan <- utils.PrependMessageToError(err, "failed to get image url from could service")
					}
					imageUrls[j] = url
				}

				if err != nil {
					errChan <- utils.PrependMessageToError(err, "failed to find images of product item")
					return
				}
				productItems[i].Images = imageUrls
			}
		}
		errChan <- nil
	}()

	// wait for the two routine to complete
	for i := 1; i <= 2; i++ {

		select {
		case <-ctx.Done():
			return nil, nil
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
			// no error then continue for the next check
		}
	}

	return productItems, nil
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

func (c *productUseCase) SearchProducts(ctx context.Context, keyword string, categoryID *string, brandID *string, locationID *string, limit, offset int) ([]response.Product, error) {
	// Assuming request.Pagination looks like:
	// In your function:
	// Assuming request.Pagination looks like:
	pageNumber, err := SafeIntToUint64(offset)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "invalid offset for pagination")
	}

	limitUint64, err := SafeIntToUint64(limit)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "invalid limit for pagination")
	}

	pagination := request.Pagination{
		PageNumber: pageNumber,
		Count:      limitUint64,
	}
	resProducts, err := c.productRepo.SearchProducts(ctx, keyword, categoryID, brandID, locationID, pagination)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to search products")
	}

	// convert from []response.Product to []domain.Product
	domainProducts := make([]response.Product, 0, len(resProducts))
	for _, p := range resProducts {
		domainProducts = append(domainProducts, response.Product{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			// map other fields...
		})
	}

	return domainProducts, nil
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
	categoryQuery := `SELECT DISTINCT c.category_id, c.name FROM categories c JOIN products p ON c.category_id = p.category_id ORDER BY c.name`
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
	brandQuery := `SELECT DISTINCT b.brand_id, b.name FROM brands b JOIN products p ON b.brand_id = p.brand_id ORDER BY b.name`
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
			ID:          c.CategoryID,
			Name:        c.Name,
			SubCategory: nil, // Populate this if needed later
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

func (s *productUseCase) GetNearbyProductsByPincode(ctx context.Context, pincode string, limit, offset int) ([]response.Product, error) {
	const DefaultRadiusMeters = 5000

	query := `
        SELECT
            p.product_id, p.name, p.description, p.brand_id, p.category_id,
            p.price, p.discount_price,
            c.name AS category_name,
            mc.name AS main_category_name,
            b.name AS brand_name,
            p.image,
            p.created_at, p.updated_at
        FROM products p
        JOIN locations l ON p.location_id = l.location_id
        LEFT JOIN categories c ON p.category_id = c.category_id
        LEFT JOIN categories mc ON c.parent_id = mc.category_id -- assuming main category is parent
        LEFT JOIN brands b ON p.brand_id = b.brand_id
        WHERE l.geog IS NOT NULL
          AND (
            SELECT geog FROM locations WHERE pincode = $1 LIMIT 1
          ) IS NOT NULL
          AND ST_DWithin(l.geog, (SELECT geog FROM locations WHERE pincode = $1 LIMIT 1), $2)
        ORDER BY ST_Distance(l.geog, (SELECT geog FROM locations WHERE pincode = $1 LIMIT 1))
        LIMIT $3 OFFSET $4
    `

	rows, err := s.DB.Query(ctx, query, pincode, DefaultRadiusMeters, limit, offset)
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

func (c *productUseCase) GetProductsByRadius(ctx context.Context, latitude int, longitude int, radiusMeters int, limit, offset int) ([]response.Product, error) {
	// default radius if not provided
	if radiusMeters <= 0 {
		radiusMeters = DefaultRadiusMeters
	}

	query := `
		SELECT * FROM (
    SELECT p.id, p.name, p.description, p.brand_id, p.category_id, p.price, p.discount_price,
           c.name, b.name, p.image, p.created_at, p.updated_at,
           (6371 * acos(
                cos(radians($1)) * cos(radians(p.latitude)) *
                cos(radians(p.longitude) - radians($2)) +
                sin(radians($1)) * sin(radians(p.latitude))
           )) AS distance_km
		FROM products p
		JOIN categories c ON p.category_id = c.category_id
		JOIN brands b ON p.brand_id = b.id
		WHERE p.latitude IS NOT NULL AND p.longitude IS NOT NULL
	) AS subquery
	WHERE distance_km <= $3
	ORDER BY distance_km;
	`

	fmt.Printf("Executing GetProductsByRadius with lat: %d, long: %d, radiusMeters: %d\n", latitude, longitude, radiusMeters)
	rows, err := c.DB.Query(ctx, query, latitude, longitude, float64(radiusMeters)/1000)
	fmt.Println("Query executed, checking for errors...")
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to get products by radius")
	}
	fmt.Printf("Rows returned: %+v\n", rows)
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
			&p.BrandName,
			&p.Image,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, utils.PrependMessageToError(err, "failed to scan product row")
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
