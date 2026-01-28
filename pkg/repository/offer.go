package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type offerDatabase struct {
	DB *gorm.DB
}

func NewOfferRepository(db *gorm.DB) interfaces.OfferRepository {
	return &offerDatabase{
		DB: db,
	}
}

func (c *offerDatabase) Transactions(ctx context.Context, trxFn func(repo interfaces.OfferRepository) error) error {

	trx := c.DB.Begin()

	repo := NewOfferRepository(trx)

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

// Find offer by id
func (c *offerDatabase) FindOfferByID(ctx context.Context, offerID uint) (offer domain.Offer, err error) {

	query := `SELECT * FROM offers WHERE id = $1`
	err = c.DB.Raw(query, offerID).Scan(&offer).Error

	return
}

// Find offer by name
func (c *offerDatabase) FindOfferByName(ctx context.Context, offerName string) (offer domain.Offer, err error) {

	query := `SELECT * FROM offers WHERE name = $1`
	err = c.DB.Raw(query, offer.Name).Scan(&offer).Error

	return
}

// findAll offers
func (c *offerDatabase) FindAllOffers(ctx context.Context,
	pagination request.Pagination) (response.OffersAndPromotions, error) {

	limit := pagination.Limit
	offset := pagination.Offset

	// Fetch promotions
	var promotions []response.Promotion
	queryPromotions := `SELECT p.id, p.promotion_category_id, p.promotion_type_id, p.offer_details, p.shop_id, p.is_active, p.created_at, p.updated_at, pc.name as promotion_category_name, pt.name as promotion_type_name, pt.icon_path, p.offer_details->>'thumbnail' as thumbnail
	          FROM promotions p
	          LEFT JOIN promotion_categories pc ON p.promotion_category_id = pc.id
	          LEFT JOIN promotions_types pt ON p.promotion_type_id = pt.id
	          WHERE p.is_active = true AND p.offer_details->>'start_date' IS NOT NULL AND p.offer_details->>'end_date' IS NOT NULL AND p.offer_details->>'start_date' <= CURRENT_DATE::text AND p.offer_details->>'end_date' >= CURRENT_DATE::text
	          ORDER BY p.created_at DESC
	          LIMIT $1 OFFSET $2`
	err := c.DB.Raw(queryPromotions, limit, offset).Scan(&promotions).Error
	if err != nil {
		return response.OffersAndPromotions{}, err
	}

	return response.OffersAndPromotions{
		Promotions: promotions,
	}, nil
}

// save a new offer
func (c *offerDatabase) SaveOffer(ctx context.Context, offer request.Offer) error {

	query := `INSERT INTO offers (name, description, discount_rate, start_date, end_date, offer_type, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`
	err := c.DB.Exec(query, offer.Name, offer.Description, offer.DiscountRate, offer.StartDate, offer.EndDate, offer.Type).Error

	return err
}

// save a new offer with image URLs
func (c *offerDatabase) SaveOfferWithImages(ctx context.Context, offer request.Offer, imageURL, thumbnailURL string) error {

	query := `INSERT INTO offers (name, description, discount_rate, start_date, end_date, offer_type, image, thumbnail, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`
	err := c.DB.Exec(query, offer.Name, offer.Description, offer.DiscountRate, offer.StartDate, offer.EndDate, offer.Type, imageURL, thumbnailURL).Error

	return err
}

// update an existing offer
func (c *offerDatabase) UpdateOffer(ctx context.Context, offer domain.Offer) error {

	query := `UPDATE offers SET offer_name = $1, description = $2, 
	discount_rate = $3, start_date = $4, end_date = $5 
	WHERE id = $6`
	err := c.DB.Exec(query, offer.Name, offer.Description,
		offer.DiscountRate, offer.StartDate, offer.EndDate, offer.ID).Error

	return err
}

// Delete all product offers related to given offer id
func (c *offerDatabase) DeleteAllProductOffersByOfferID(ctx context.Context, offerID uint) error {

	query := `DELETE FROM offer_products WHERE product_item_id = $1`
	err := c.DB.Exec(query, offerID).Error

	return err
}

// Delete all category offers related to given offer id
func (c *offerDatabase) DeleteAllCategoryOffersByOfferID(ctx context.Context, offerID uint) error {

	query := `DELETE FROM offer_categories WHERE offer_id = $1`
	err := c.DB.Exec(query, offerID).Error

	return err
}

// delete an offer
func (c *offerDatabase) DeleteOffer(ctx context.Context, offerID uint) error {

	query := `DELETE FROM offers WHERE id = $1`
	err := c.DB.Exec(query, offerID).Error

	return err
}

// find offer_category by category_id (for mainly checking this category have an offer existing or not)
func (c *offerDatabase) FindOfferCategoryCategoryID(ctx context.Context,
	categoryID uint) (offerCategory domain.OfferCategory, err error) {

	query := `SELECT * FROM offer_categories WHERE  category_id = ?`
	err = c.DB.Raw(query, categoryID).Scan(&offerCategory).Error

	return
}

// find all offer_category
func (c *offerDatabase) FindAllOfferCategories(ctx context.Context,
	pagination request.Pagination) (offerCategories []response.OfferCategory, err error) {

	query := `SELECT oc.id AS offer_category_id, oc.category_id,c.name AS category_name, 
	oc.offer_id, o.name AS offer_name, o.discount_rate 
	FROM offer_categories oc INNER JOIN categories c ON c.id = oc.category_id 
	INNER JOIN offers o ON oc.offer_id = o.id`

	err = c.DB.Raw(query).Scan(&offerCategories).Error

	return
}

// save a new offer for category
func (c *offerDatabase) SaveCategoryOffer(ctx context.Context,
	categoryOffer request.OfferCategory) (categoryOfferID uint, err error) {

	query := `INSERT INTO offer_categories (offer_id,category_id) VALUES ($1, $2) RETURNING id`
	err = c.DB.Raw(query, categoryOffer.OfferID, categoryOffer.CategoryID).Scan(&categoryOfferID).Error

	return
}

// remove offer_category
func (c *offerDatabase) DeleteCategoryOffer(ctx context.Context, categoryOfferID uint) error {

	query := `DELETE FROM offer_categories WHERE id = $1 `
	err := c.DB.Exec(query, categoryOfferID).Error

	return err
}

// update offer_category
func (c *offerDatabase) UpdateCategoryOffer(ctx context.Context, categoryOfferID, offerID uint) error {

	query := `UPDATE offer_categories SET offer_id = $1 WHERE id = $2`
	err := c.DB.Exec(query, offerID, categoryOfferID).Error

	return err
}

// find product_offer with product_id
func (c *offerDatabase) FindOfferProductByProductID(ctx context.Context,
	productItemID uint) (offerProduct domain.OfferProduct, err error) {

	query := `SELECT op.* FROM offer_products op 
	INNER JOIN promotions p ON op.offer_id = p.id 
	WHERE op.product_item_id = ? AND op.is_active = true 
	AND (p.start_date)::timestamp <= CURRENT_TIMESTAMP 
	AND (p.end_date)::timestamp >= CURRENT_TIMESTAMP`
	err = c.DB.Raw(query, productItemID).Scan(&offerProduct).Error

	return
}

// find all offer_products
func (c *offerDatabase) FindAllOfferProducts(ctx context.Context,
	pagination request.Pagination) (offerProducts []response.OfferProduct, err error) {

	query := `SELECT op.id AS offer_product_id, op.product_id,p.name,op.offer_id,o.name, o.discount_rate  
	FROM offer_products op INNER JOIN products p ON p.id = op.product_id 
	INNER JOIN offers o ON o.id = op.offer_id`
	err = c.DB.Raw(query).Scan(&offerProducts).Error

	return
}

// save a offer for product
func (c *offerDatabase) SaveOfferProduct(ctx context.Context,
	offerProduct domain.OfferProduct) (productOfferId uint, err error) {

	// Check if the promotion exists
	var promotionExists int64
	checkOfferQuery := `SELECT COUNT(id) FROM promotions WHERE id = $1`
	err = c.DB.Raw(checkOfferQuery, offerProduct.OfferID).Scan(&promotionExists).Error
	if err != nil {
		return 0, fmt.Errorf("failed to verify promotion exists: %w", err)
	}
	if promotionExists == 0 {
		return 0, fmt.Errorf("promotion with ID %d does not exist", offerProduct.OfferID)
	}

	// Check if an offer already exists for this product_item_id
	var existingID uint
	checkQuery := `SELECT id FROM offer_products WHERE product_item_id = $1 LIMIT 1`
	err = c.DB.Raw(checkQuery, offerProduct.ProductItemID).Scan(&existingID).Error
	if err == nil && existingID != 0 {
		return 0, fmt.Errorf("an offer already exists for this product item")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	query := `INSERT INTO offer_products (offer_id, product_item_id) VALUES ($1,$2)  RETURNING id`
	err = c.DB.Raw(query, offerProduct.OfferID, offerProduct.ProductItemID).Scan(&productOfferId).Error

	return
}

// delete offer_products
func (c *offerDatabase) DeleteOfferProduct(ctx context.Context, productOfferID uint) error {

	query := `DELETE FROM offer_products WHERE id = $1`
	err := c.DB.Exec(query, productOfferID).Error

	return err
}

// update offer_products
func (c *offerDatabase) UpdateOfferProduct(ctx context.Context, productOfferID, offerID uint) error {

	query := `UPDATE offer_products SET offer_id = $1 WHERE id = $1`
	err := c.DB.Exec(query, offerID, productOfferID).Error

	return err
}

// Update product discount price by check given category offer id
func (c *offerDatabase) UpdateProductsDiscountByCategoryOfferID(ctx context.Context, categoryOfferID uint) error {

	query := `UPDATE products p SET discount_price = (price * (100 - o.discount_rate))/100 
	FROM offer_categories oc 
	INNER JOIN offers o ON o.id = oc.offer_id 
	WHERE p.category_id = oc.category_id AND oc.id = $1`
	err := c.DB.Exec(query, categoryOfferID).Error

	return err
}

// Remove product discount price by check given category offer id
func (c *offerDatabase) RemoveProductsDiscountByCategoryOfferID(ctx context.Context, categoryOfferID uint) error {

	query := `UPDATE products p SET discount_price = 0 
	FROM offer_categories oc 
	INNER JOIN offers o ON o.id = oc.offer_id 
	WHERE p.category_id = oc.category_id AND oc.id = $1`
	err := c.DB.Exec(query, categoryOfferID).Error

	return err
}

// Update product items discount price by check given category offer id
func (c *offerDatabase) UpdateProductItemsDiscountByCategoryOfferID(ctx context.Context,
	categoryOfferID uint) error {

	query := `UPDATE product_items AS pi SET discount_price = (pi.price * (100 - o.discount_rate))/100 
	FROM offer_categories oc 
	INNER JOIN products p ON p.category_id = oc.category_id
	INNER JOIN offers o ON o.id = oc.offer_id 
	WHERE oc.id = $1`
	err := c.DB.Exec(query, categoryOfferID).Error

	return err
}

// Remove product items discount price by check given category offer id
func (c *offerDatabase) RemoveProductItemsDiscountByCategoryOfferID(ctx context.Context,
	categoryOfferID uint) error {

	query := `UPDATE product_items AS pi SET discount_price = 0 
	FROM offer_categories oc 
	INNER JOIN products p ON p.category_id = oc.category_id
	INNER JOIN offers o ON o.id = oc.offer_id 
	WHERE oc.id = $1`
	err := c.DB.Exec(query, categoryOfferID).Error

	return err
}

// Recalculate all product discount price by check given product offer id
func (c *offerDatabase) ApplyOfferToProductItem(ctx context.Context, productOfferID uint) error {

	fmt.Printf("insert offer om product item: %d\n", productOfferID)
	query := `insert into offer_products (offer_id, product_id)`
	err := c.DB.Exec(query, productOfferID).Error
	return err
}

// Recalculate all product discount price by check given product offer id
func (c *offerDatabase) RemoveProductsDiscountByProductOfferID(ctx context.Context, productOfferID uint) error {

	fmt.Printf("Removing product discount for product offer id: %d\n", productOfferID)
	query := `UPDATE products p SET discount_price = (p.price * (100 - o.discount_rate))/100 
	FROM offer_products op
	INNER JOIN  offers o ON op.offer_id = o.id 
	WHERE p.id = op.product_id AND op.id = $1`
	err := c.DB.Exec(query, productOfferID).Error

	return err
}

// Remove  product items discount price by given product offer id
func (c *offerDatabase) UpdateProductItemsDiscountByProductOfferID(ctx context.Context, productOfferID uint) error {

	fmt.Printf("Updating product items discount for product offer id: %d\n", productOfferID)
	query := `UPDATE product_items pi SET discount_price = 0 
	FROM offer_products op
	INNER JOIN offers o ON o.id = op.offer_id  
	WHERE pi.product_id = op.product_id AND op.id = $1`
	err := c.DB.Exec(query, productOfferID).Error

	return err
}

// Recalculate all product items discount price by given product offer id
func (c *offerDatabase) RemoveProductItemsDiscountByProductOfferID(ctx context.Context, productOfferID uint) error {

	query := `UPDATE product_items pi SET discount_price = 0 
	FROM offer_products op
	INNER JOIN offers o ON o.id = op.offer_id  
	WHERE pi.product_id = op.product_id AND op.id = $1`
	err := c.DB.Exec(query, productOfferID).Error

	return err
}

func (c *offerDatabase) ApplyOfferToShop(ctx context.Context, adminID string, body request.ApplyOfferToShop) error {
	// First, get the shop ID using admin ID
	var shopID uint
	shopQuery := `SELECT id FROM shop_details WHERE admin_id = $1`
	err := c.DB.Raw(shopQuery, adminID).Scan(&shopID).Error
	if err != nil {
		return fmt.Errorf("failed to find shop for admin ID %s: %w", adminID, err)
	}

	// Insert into shop_offers with shop_id
	query := `INSERT INTO shop_offers (offer_id, shop_id, admin_id, start_date, end_date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`
	err = c.DB.Exec(query, body.OfferID, shopID, adminID, body.StartDate.Time, body.EndDate.Time).Error
	return err
}

// FindActiveOffers returns offers that are active based on start and end date
func (c *offerDatabase) FindActiveOffers(ctx context.Context) ([]domain.Offer, error) {
	var offers []domain.Offer
	query := `SELECT * FROM promotions WHERE is_active = true AND start_date <= NOW() AND end_date >= NOW()`
	err := c.DB.Raw(query).Scan(&offers).Error
	return offers, err
}

// FindShopOffersByShopIDAndDateRange finds shop offers for a shop within a date range
func (c *offerDatabase) FindShopOffersByShopIDAndDateRange(ctx context.Context, shopID uint, startDate, endDate time.Time) ([]domain.ShopOffer, error) {
	var shopOffers []domain.ShopOffer
	query := `SELECT * FROM shop_offers WHERE shop_id = $1 AND start_date <= $3 AND end_date >= $2`
	err := c.DB.Raw(query, shopID, startDate, endDate).Scan(&shopOffers).Error
	return shopOffers, err
}
