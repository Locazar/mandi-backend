package repository

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type dealRepository struct {
	db *gorm.DB
}

func NewDealRepository(db *gorm.DB) interfaces.DealRepository {
	return &dealRepository{db: db}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func discountLabel(dealType string, rate float64) string {
	switch {
	case dealType == "bogo":
		return "BOGO"
	case dealType == "free_gift" || dealType == "gift":
		return "Free Gift"
	case dealType == "free_shipping" || dealType == "shipping":
		return "Free Shipping"
	case dealType == "fixed":
		return fmt.Sprintf("₹%.0f OFF", rate)
	default:
		return fmt.Sprintf("%.0f%% OFF", rate)
	}
}

func daysLeft(end time.Time) int {
	return int(math.Ceil(time.Until(end).Hours() / 24))
}

// ── promotionRow: raw DB row shared by all promotion queries ─────────────────
// Uses gorm:"column:..." tags so GORM's raw-SQL scanner maps correctly even
// when the struct field names don't snake-case to the SQL alias.
type promotionRow struct {
	ID                     string   `gorm:"column:id"`
	OfferName              string   `gorm:"column:offer_name"`
	Description            string   `gorm:"column:description"`
	DiscountRate           float64  `gorm:"column:discount_rate"`
	DealType               string   `gorm:"column:deal_type"`
	StartDate              time.Time `gorm:"column:start_date"`
	EndDate                time.Time `gorm:"column:end_date"`
	ShopID                 string   `gorm:"column:shop_id"`
	ShopName               string   `gorm:"column:shop_name"`
	ShopLogo               string   `gorm:"column:shop_logo"`
	IconPath               string   `gorm:"column:icon_path"`
	ProductCount           int      `gorm:"column:product_count"`
	MinimumPurchaseAmount  *float64 `gorm:"column:minimum_purchase_amount"`
	BogoBuyQuantity        *int     `gorm:"column:bogo_buy_quantity"`
	BogoGetQuantity        *int     `gorm:"column:bogo_get_quantity"`
	BogoCombinationEnabled *bool    `gorm:"column:bogo_combination_enabled"`
	GiftDescription        *string  `gorm:"column:gift_description"`
}

func (r promotionRow) toCustomerDeal() response.CustomerDeal {
	now := time.Now()
	return response.CustomerDeal{
		ID:                     r.ID,
		Source:                 "promotion",
		DealType:               r.DealType,
		Title:                  r.OfferName,
		Description:            r.Description,
		DiscountRate:           r.DiscountRate,
		DiscountLabel:          discountLabel(r.DealType, r.DiscountRate),
		StartDate:              r.StartDate,
		EndDate:                r.EndDate,
		IsExpired:              now.After(r.EndDate),
		DaysLeft:               daysLeft(r.EndDate),
		ShopID:                 r.ShopID,
		ShopName:               r.ShopName,
		ShopLogo:               r.ShopLogo,
		IconPath:               r.IconPath,
		ProductCount:           r.ProductCount,
		MinimumPurchaseAmount:  r.MinimumPurchaseAmount,
		BogoBuyQuantity:        r.BogoBuyQuantity,
		BogoGetQuantity:        r.BogoGetQuantity,
		BogoCombinationEnabled: r.BogoCombinationEnabled,
		GiftDescription:        r.GiftDescription,
	}
}

// ── base CTE ──────────────────────────────────────────────────────────────────
// Joins promotions → promotion_types → shop_details. Uses shop_image_url
// (the real column name) aliased as shop_logo for the response field.
// Callers append their own WHERE / ORDER / LIMIT clauses.
const promotionBaseCTE = `
WITH promo AS (
    SELECT
        p.id,
        p.offer_name,
        p.description,
        p.discount_rate,
        COALESCE(pt.type, 'percentage')    AS deal_type,
        p.start_date,
        p.end_date,
        p.shop_id,
        COALESCE(sd.shop_name, '')         AS shop_name,
        COALESCE(sd.shop_image_url, '')    AS shop_logo,
        COALESCE(pt.icon_path, '')         AS icon_path,
        COUNT(DISTINCT pp.product_item_id)::int AS product_count,
        p.minimum_purchase_amount,
        p.bogo_buy_quantity,
        p.bogo_get_quantity,
        p.bogo_combination_enabled,
        p.gift_description
    FROM promotions p
    LEFT JOIN promotions_types   pt ON pt.id = p.promotion_type_id
    LEFT JOIN shop_details       sd ON sd.id = p.shop_id
    LEFT JOIN promotion_products pp ON pp.promotion_id = p.id
    WHERE p.is_active = TRUE
      AND p.start_date IS NOT NULL
      AND p.end_date   IS NOT NULL
      AND p.end_date   > NOW()
    GROUP BY p.id, pt.type, pt.icon_path, sd.shop_name, sd.shop_image_url
)`

// selectFromPromo is the fixed column list used after the CTE.
const selectFromPromo = `
SELECT id, offer_name, description, discount_rate, deal_type,
       start_date, end_date, shop_id, shop_name, shop_logo,
       icon_path, product_count,
       minimum_purchase_amount, bogo_buy_quantity, bogo_get_quantity,
       bogo_combination_enabled, gift_description
FROM promo`

// ── GetAllDeals ───────────────────────────────────────────────────────────────

func (r *dealRepository) GetAllDeals(ctx context.Context, filter request.DealFilter) ([]response.CustomerDeal, int64, error) {
	conditions := []string{}
	args := []interface{}{}

	if filter.DealType != "" {
		conditions = append(conditions, "deal_type = ?")
		args = append(args, filter.DealType)
	}
	if filter.MinDiscount > 0 {
		conditions = append(conditions, "discount_rate >= ?")
		args = append(args, filter.MinDiscount)
	}
	if filter.MaxDiscount > 0 {
		conditions = append(conditions, "discount_rate <= ?")
		args = append(args, filter.MaxDiscount)
	}
	if filter.ShopID != "" {
		conditions = append(conditions, "shop_id = ?")
		args = append(args, filter.ShopID)
	}

	// Category filter: only deals that have at least one product in this category.
	if filter.CategoryID != "" {
		conditions = append(conditions, `id IN (
			SELECT DISTINCT pp.promotion_id
			FROM promotion_products pp
			JOIN product_items pi ON pi.id = pp.product_item_id
			WHERE pi.category_id = ?)`)
		args = append(args, filter.CategoryID)
	}

	// Department filter: only deals touching products in this department.
	if filter.DepartmentID != "" {
		conditions = append(conditions, `id IN (
			SELECT DISTINCT pp.promotion_id
			FROM promotion_products pp
			JOIN product_items pi ON pi.id = pp.product_item_id
			WHERE pi.department_id = ?)`)
		args = append(args, filter.DepartmentID)
	}

	// Location filter: only deals from shops within the given radius.
	if filter.Lat != 0 && filter.Lng != 0 && filter.RadiusKM > 0 {
		conditions = append(conditions, `shop_id IN (
			SELECT id FROM shop_details
			WHERE latitude IS NOT NULL AND longitude IS NOT NULL
			  AND (6371 * acos(LEAST(1, GREATEST(-1,
			        cos(radians(?)) * cos(radians(latitude::double precision)) *
			        cos(radians(longitude::double precision) - radians(?)) +
			        sin(radians(?)) * sin(radians(latitude::double precision))
			      )))) <= ?)`)
		args = append(args, filter.Lat, filter.Lng, filter.Lat, filter.RadiusKM)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	orderSQL := dealOrderClause(filter.SortBy)
	lim := int(filter.Pagination.Limit)
	if lim == 0 {
		lim = 20
	}
	off := int(filter.Pagination.Offset)

	query := fmt.Sprintf(`%s %s %s %s LIMIT %d OFFSET %d`,
		promotionBaseCTE, selectFromPromo, where, orderSQL, lim, off)
	countQuery := fmt.Sprintf(`%s SELECT COUNT(*) FROM promo %s`, promotionBaseCTE, where)

	var rows []promotionRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("GetAllDeals: %w", err)
	}

	var total int64
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	if err := r.db.WithContext(ctx).Raw(countQuery, countArgs...).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("GetAllDeals count: %w", err)
	}

	deals := make([]response.CustomerDeal, len(rows))
	for i, row := range rows {
		deals[i] = row.toCustomerDeal()
	}
	return deals, total, nil
}

func dealOrderClause(sortBy string) string {
	switch sortBy {
	case "discount_asc":
		return "ORDER BY discount_rate ASC"
	case "ending_soon":
		return "ORDER BY end_date ASC"
	case "newest":
		return "ORDER BY id DESC"
	default: // "discount_desc"
		return "ORDER BY discount_rate DESC"
	}
}

// ── GetDealFilters ────────────────────────────────────────────────────────────

func (r *dealRepository) GetDealFilters(ctx context.Context) (response.DealFilters, error) {
	var filters response.DealFilters

	// Deal-type counts + icon
	typeQuery := fmt.Sprintf(`%s
SELECT deal_type, COUNT(*) AS count, MAX(icon_path) AS icon_path
FROM promo GROUP BY deal_type ORDER BY count DESC`, promotionBaseCTE)

	type typeRow struct {
		DealType string `gorm:"column:deal_type"`
		Count    int    `gorm:"column:count"`
		IconPath string `gorm:"column:icon_path"`
	}
	var typeRows []typeRow
	if err := r.db.WithContext(ctx).Raw(typeQuery).Scan(&typeRows).Error; err != nil {
		return filters, fmt.Errorf("GetDealFilters types: %w", err)
	}

	dealTypeLabels := map[string]string{
		"percentage":    "% Off",
		"fixed":         "Flat Off",
		"bogo":          "Buy 1 Get 1",
		"bundle":        "Bundle Deal",
		"flash":         "Flash Sale",
		"free_gift":     "Free Gift",
		"free_shipping": "Free Shipping",
		"loyalty":       "Loyalty Deal",
		"seasonal":      "Seasonal",
	}
	var totalDeals int
	for _, tr := range typeRows {
		label := dealTypeLabels[tr.DealType]
		if label == "" {
			label = tr.DealType
		}
		filters.DealTypes = append(filters.DealTypes, response.DealTypeFilter{
			Type:     tr.DealType,
			Label:    label,
			IconPath: tr.IconPath,
			Count:    tr.Count,
		})
		totalDeals += tr.Count
	}
	filters.TotalDeals = totalDeals

	// Discount range counts (percentage + fixed only)
	rangeQuery := fmt.Sprintf(`%s
SELECT
    COUNT(*) FILTER (WHERE discount_rate >  0  AND discount_rate < 10)  AS under10,
    COUNT(*) FILTER (WHERE discount_rate >= 10 AND discount_rate < 30)  AS r10_30,
    COUNT(*) FILTER (WHERE discount_rate >= 30 AND discount_rate < 50)  AS r30_50,
    COUNT(*) FILTER (WHERE discount_rate >= 50)                         AS above50
FROM promo WHERE deal_type IN ('percentage','fixed')`, promotionBaseCTE)

	type rangeRow struct {
		Under10 int `gorm:"column:under10"`
		R10_30  int `gorm:"column:r10_30"`
		R30_50  int `gorm:"column:r30_50"`
		Above50 int `gorm:"column:above50"`
	}
	var rr rangeRow
	if err := r.db.WithContext(ctx).Raw(rangeQuery).Scan(&rr).Error; err != nil {
		return filters, fmt.Errorf("GetDealFilters ranges: %w", err)
	}
	filters.DiscountRanges = []response.DiscountRangeFilter{
		{Label: "Under 10%", Min: 0, Max: 10, Count: rr.Under10},
		{Label: "10–30%", Min: 10, Max: 30, Count: rr.R10_30},
		{Label: "30–50%", Min: 30, Max: 50, Count: rr.R30_50},
		{Label: "50% +", Min: 50, Max: 0, Count: rr.Above50},
	}

	// Category breakdown — categories that have at least one active deal
	catQuery := `
SELECT c.id AS category_id, c.name AS category_name, COUNT(DISTINCT p.id) AS count
FROM promotions p
JOIN promotion_products pp ON pp.promotion_id = p.id
JOIN product_items       pi ON pi.id = pp.product_item_id
JOIN categories          c  ON c.id  = pi.category_id
WHERE p.is_active = TRUE AND p.end_date > NOW()
GROUP BY c.id, c.name ORDER BY count DESC LIMIT 20`

	type catRow struct {
		CategoryID   string `gorm:"column:category_id"`
		CategoryName string `gorm:"column:category_name"`
		Count        int    `gorm:"column:count"`
	}
	var catRows []catRow
	if err := r.db.WithContext(ctx).Raw(catQuery).Scan(&catRows).Error; err != nil {
		return filters, fmt.Errorf("GetDealFilters categories: %w", err)
	}
	for _, cr := range catRows {
		filters.Categories = append(filters.Categories, response.DealCategoryFilter{
			CategoryID:   cr.CategoryID,
			CategoryName: cr.CategoryName,
			Count:        cr.Count,
		})
	}

	return filters, nil
}

// ── GetDealByID ───────────────────────────────────────────────────────────────

func (r *dealRepository) GetDealByID(ctx context.Context, dealID string) (response.CustomerDeal, error) {
	query := fmt.Sprintf(`%s %s WHERE id = ?`, promotionBaseCTE, selectFromPromo)
	var row promotionRow
	if err := r.db.WithContext(ctx).Raw(query, dealID).First(&row).Error; err != nil {
		return response.CustomerDeal{}, fmt.Errorf("deal %s not found: %w", dealID, err)
	}
	return row.toCustomerDeal(), nil
}

// ── GetProductsByDeal ─────────────────────────────────────────────────────────

func (r *dealRepository) GetProductsByDeal(ctx context.Context, dealID string, pagination request.Pagination) ([]response.CustomerDealProduct, int64, error) {
	lim := int(pagination.Limit)
	if lim == 0 {
		lim = 20
	}
	off := int(pagination.Offset)

	query := fmt.Sprintf(`
SELECT
    pi.id                                     AS product_item_id,
    COALESCE(pi.name, pi.sub_category_name)   AS product_name,
    pi.sub_category_name,
    pi.shop_id,
    COALESCE(sd.shop_name, '')                AS shop_name,
    COALESCE(sd.shop_image_url, '')           AS shop_logo,
    COALESCE(pi.product_item_images[1], '')   AS image_url,
    pi.stock                                  AS in_stock,
    p.id                                      AS deal_id,
    COALESCE(pt.type, 'percentage')           AS deal_type,
    p.discount_rate,
    p.end_date                                AS deal_ends_at
FROM promotion_products pp
JOIN product_items    pi ON pi.id  = pp.product_item_id
JOIN promotions       p  ON p.id   = pp.promotion_id
LEFT JOIN promotions_types pt ON pt.id = p.promotion_type_id
LEFT JOIN shop_details     sd ON sd.id = pi.shop_id
WHERE pp.promotion_id = ?
  AND p.is_active = TRUE
  AND p.end_date > NOW()
ORDER BY pi.created_at DESC
LIMIT %d OFFSET %d`, lim, off)

	var rows []productScanRow
	if err := r.db.WithContext(ctx).Raw(query, dealID).Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("GetProductsByDeal: %w", err)
	}

	var total int64
	r.db.WithContext(ctx).Raw(`
SELECT COUNT(*) FROM promotion_products pp
JOIN promotions p ON p.id = pp.promotion_id
WHERE pp.promotion_id = ? AND p.is_active = TRUE AND p.end_date > NOW()`, dealID).Scan(&total)

	return productScanRowsToResponse(rows), total, nil
}

// ── GetProductsWithDeals ──────────────────────────────────────────────────────

func (r *dealRepository) GetProductsWithDeals(ctx context.Context, filter request.DealFilter, pagination request.Pagination) ([]response.CustomerDealProduct, int64, error) {
	lim := int(pagination.Limit)
	if lim == 0 {
		lim = 20
	}
	off := int(pagination.Offset)

	conditions := []string{
		"p.is_active = TRUE",
		"p.end_date > NOW()",
	}
	args := []interface{}{}

	if filter.DealType != "" {
		conditions = append(conditions, "COALESCE(pt.type, 'percentage') = ?")
		args = append(args, filter.DealType)
	}
	if filter.MinDiscount > 0 {
		conditions = append(conditions, "p.discount_rate >= ?")
		args = append(args, filter.MinDiscount)
	}
	if filter.MaxDiscount > 0 {
		conditions = append(conditions, "p.discount_rate <= ?")
		args = append(args, filter.MaxDiscount)
	}
	if filter.CategoryID != "" {
		conditions = append(conditions, "pi.category_id = ?")
		args = append(args, filter.CategoryID)
	}
	if filter.DepartmentID != "" {
		conditions = append(conditions, "pi.department_id = ?")
		args = append(args, filter.DepartmentID)
	}
	if filter.ShopID != "" {
		conditions = append(conditions, "pi.shop_id = ?")
		args = append(args, filter.ShopID)
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	query := fmt.Sprintf(`
SELECT
    pi.id                                     AS product_item_id,
    COALESCE(pi.name, pi.sub_category_name)   AS product_name,
    pi.sub_category_name,
    pi.shop_id,
    COALESCE(sd.shop_name, '')                AS shop_name,
    COALESCE(sd.shop_image_url, '')           AS shop_logo,
    COALESCE(pi.product_item_images[1], '')   AS image_url,
    pi.stock                                  AS in_stock,
    p.id                                      AS deal_id,
    COALESCE(pt.type, 'percentage')           AS deal_type,
    p.discount_rate,
    p.end_date                                AS deal_ends_at
FROM promotion_products pp
JOIN product_items    pi ON pi.id  = pp.product_item_id
JOIN promotions       p  ON p.id   = pp.promotion_id
LEFT JOIN promotions_types pt ON pt.id = p.promotion_type_id
LEFT JOIN shop_details     sd ON sd.id = pi.shop_id
%s
ORDER BY p.discount_rate DESC, p.end_date ASC
LIMIT %d OFFSET %d`, where, lim, off)

	countQuery := fmt.Sprintf(`
SELECT COUNT(*) FROM promotion_products pp
JOIN product_items  pi ON pi.id  = pp.product_item_id
JOIN promotions     p  ON p.id   = pp.promotion_id
LEFT JOIN promotions_types pt ON pt.id = p.promotion_type_id
%s`, where)

	var rows []productScanRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("GetProductsWithDeals: %w", err)
	}

	var total int64
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	r.db.WithContext(ctx).Raw(countQuery, countArgs...).Scan(&total)

	return productScanRowsToResponse(rows), total, nil
}

// ── shared product scan helpers ───────────────────────────────────────────────

// productScanRow holds a flat, GORM-scannable representation of a deal product.
// Uses pi.product_item_images[1] (PostgreSQL 1-based array index) so the first
// image comes back as a plain string rather than a text[] that GORM can't parse.
type productScanRow struct {
	ProductItemID   string    `gorm:"column:product_item_id"`
	ProductName     string    `gorm:"column:product_name"`
	SubCategoryName string    `gorm:"column:sub_category_name"`
	ShopID          string    `gorm:"column:shop_id"`
	ShopName        string    `gorm:"column:shop_name"`
	ShopLogo        string    `gorm:"column:shop_logo"`
	ImageURL        string    `gorm:"column:image_url"`
	InStock         bool      `gorm:"column:in_stock"`
	DealID          string    `gorm:"column:deal_id"`
	DealType        string    `gorm:"column:deal_type"`
	DiscountRate    float64   `gorm:"column:discount_rate"`
	DealEndsAt      time.Time `gorm:"column:deal_ends_at"`
}

func productScanRowsToResponse(rows []productScanRow) []response.CustomerDealProduct {
	out := make([]response.CustomerDealProduct, len(rows))
	for i, row := range rows {
		out[i] = response.CustomerDealProduct{
			ProductItemID:   row.ProductItemID,
			ProductName:     row.ProductName,
			SubCategoryName: row.SubCategoryName,
			ShopID:          row.ShopID,
			ShopName:        row.ShopName,
			ShopLogo:        row.ShopLogo,
			ImageURL:        row.ImageURL,
			InStock:         row.InStock,
			DealID:          row.DealID,
			DealType:        row.DealType,
			DiscountRate:    row.DiscountRate,
			DiscountLabel:   discountLabel(row.DealType, row.DiscountRate),
			DealEndsAt:      row.DealEndsAt,
		}
	}
	return out
}
