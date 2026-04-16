package repository

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/elasticsearch"
	"gorm.io/gorm"
)

type searchRepository struct {
	DB            *gorm.DB
	ElasticClient *elasticsearch.ElasticService
}

func NewSearchRepository(db *gorm.DB, elasticClient *elasticsearch.ElasticService) interfaces.SearchRepository {
	return &searchRepository{
		DB:            db,
		ElasticClient: elasticClient,
	}
}

// entityTypeSet returns a set of requested entity types. If empty, all types are included.
func entityTypeSet(types []string) map[string]bool {
	if len(types) == 0 {
		return nil // nil means all
	}
	m := make(map[string]bool, len(types))
	for _, t := range types {
		m[strings.TrimSpace(strings.ToLower(t))] = true
	}
	return m
}

func includeEntity(set map[string]bool, entity string) bool {
	if set == nil {
		return true
	}
	return set[entity]
}

func (r *searchRepository) GlobalSearch(ctx context.Context, query string, entityTypes []string,
	lat, lng, radiusKm float64, pincode string, limit, offset uint) (response.GlobalSearchResult, error) {

	typeSet := entityTypeSet(entityTypes)
	result := response.GlobalSearchResult{}

	// Try Elasticsearch first if client is available
	if r.ElasticClient != nil && r.ElasticClient.Client != nil {
		esResult, err := r.globalSearchES(ctx, query, typeSet, lat, lng, radiusKm, pincode, limit, offset)
		if err == nil {
			return esResult, nil
		}
		log.Printf("Elasticsearch global search failed, falling back to PostgreSQL: %v", err)
	}

	// PostgreSQL fallback — run queries concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex
	errs := make([]error, 0)

	if includeEntity(typeSet, "products") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := r.searchProductsDB(ctx, query, limit, offset)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("products: %w", err))
				return
			}
			result.Products = items
		}()
	}

	if includeEntity(typeSet, "categories") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := r.searchCategoriesDB(ctx, query, limit, offset)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("categories: %w", err))
				return
			}
			result.Categories = items
		}()
	}

	if includeEntity(typeSet, "shops") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := r.searchShopsDB(ctx, query, lat, lng, radiusKm, pincode, limit, offset)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("shops: %w", err))
				return
			}
			result.Shops = items
		}()
	}

	if includeEntity(typeSet, "brands") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := r.searchBrandsDB(ctx, query, limit, offset)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("brands: %w", err))
				return
			}
			result.Brands = items
		}()
	}

	if includeEntity(typeSet, "departments") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := r.searchDepartmentsDB(ctx, query, limit, offset)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("departments: %w", err))
				return
			}
			result.Departments = items
		}()
	}

	if includeEntity(typeSet, "offers") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := r.searchOffersDB(ctx, query, limit, offset)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("offers: %w", err))
				return
			}
			result.Offers = items
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		log.Printf("Global search partial errors: %v", errs)
	}

	// Initialize nil slices to empty
	if result.Products == nil {
		result.Products = []response.ProductSearchItem{}
	}
	if result.Categories == nil {
		result.Categories = []response.CategorySearchItem{}
	}
	if result.Shops == nil {
		result.Shops = []response.ShopSearchItem{}
	}
	if result.Brands == nil {
		result.Brands = []response.BrandSearchItem{}
	}
	if result.Departments == nil {
		result.Departments = []response.DepartmentSearchItem{}
	}
	if result.Offers == nil {
		result.Offers = []response.OfferSearchItem{}
	}

	return result, nil
}

// globalSearchES uses Elasticsearch for full-text search across multiple indices.
func (r *searchRepository) globalSearchES(ctx context.Context, query string, typeSet map[string]bool,
	lat, lng, radiusKm float64, pincode string, limit, offset uint) (response.GlobalSearchResult, error) {

	result := response.GlobalSearchResult{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	errs := make([]error, 0)

	intLimit := int(limit)
	intOffset := int(offset)

	if includeEntity(typeSet, "products") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			products, err := r.ElasticClient.SearchProducts(ctx, query, nil, nil, nil, nil, intLimit, intOffset)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("es products: %w", err))
				return
			}
			items := make([]response.ProductSearchItem, 0, len(products))
			for _, p := range products {
				items = append(items, response.ProductSearchItem{
					ID:          p.ID,
					Name:        p.Name,
					Description: p.Description,
					CategoryID:  p.CategoryID,
				})
			}
			result.Products = items
		}()
	}

	if includeEntity(typeSet, "categories") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			categories, err := r.ElasticClient.SearchCategories(ctx, query, nil, intLimit, intOffset)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("es categories: %w", err))
				return
			}
			items := make([]response.CategorySearchItem, 0, len(categories))
			for _, c := range categories {
				items = append(items, response.CategorySearchItem{
					ID:           c.ID,
					Name:         c.Name,
					DepartmentID: c.DepartmentID,
				})
			}
			result.Categories = items
		}()
	}

	if includeEntity(typeSet, "brands") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			brands, err := r.ElasticClient.SearchBrands(ctx, query, intLimit, intOffset)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("es brands: %w", err))
				return
			}
			items := make([]response.BrandSearchItem, 0, len(brands))
			for _, b := range brands {
				items = append(items, response.BrandSearchItem{
					ID:   b.ID,
					Name: b.Name,
				})
			}
			result.Brands = items
		}()
	}

	if includeEntity(typeSet, "departments") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			departments, err := r.ElasticClient.SearchDepartments(ctx, query, intLimit, intOffset)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("es departments: %w", err))
				return
			}
			items := make([]response.DepartmentSearchItem, 0, len(departments))
			for _, d := range departments {
				items = append(items, response.DepartmentSearchItem{
					ID:   d.ID,
					Name: d.Name,
				})
			}
			result.Departments = items
		}()
	}

	if includeEntity(typeSet, "offers") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			offers, err := r.ElasticClient.SearchOffers(ctx, query, intLimit, intOffset)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("es offers: %w", err))
				return
			}
			items := make([]response.OfferSearchItem, 0, len(offers))
			for _, o := range offers {
				items = append(items, response.OfferSearchItem{
					ID:          o.ID,
					Name:        o.Name,
					Description: o.Description,
				})
			}
			result.Offers = items
		}()
	}

	// Shops are not indexed in ES, always use DB for shops
	if includeEntity(typeSet, "shops") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := r.searchShopsDB(ctx, query, lat, lng, radiusKm, pincode, limit, offset)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("shops: %w", err))
				return
			}
			result.Shops = items
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		return result, fmt.Errorf("elasticsearch global search errors: %v", errs)
	}

	// Initialize nil slices to empty
	if result.Products == nil {
		result.Products = []response.ProductSearchItem{}
	}
	if result.Categories == nil {
		result.Categories = []response.CategorySearchItem{}
	}
	if result.Shops == nil {
		result.Shops = []response.ShopSearchItem{}
	}
	if result.Brands == nil {
		result.Brands = []response.BrandSearchItem{}
	}
	if result.Departments == nil {
		result.Departments = []response.DepartmentSearchItem{}
	}
	if result.Offers == nil {
		result.Offers = []response.OfferSearchItem{}
	}

	return result, nil
}

// ---------- PostgreSQL entity search methods ----------

func (r *searchRepository) searchProductsDB(ctx context.Context, query string, limit, offset uint) ([]response.ProductSearchItem, error) {
	likeQuery := "%" + query + "%"
	rows, err := r.DB.WithContext(ctx).Raw(`
		SELECT p.id, p.name, p.description, p.category_id,
			COALESCE(c.name, '') AS category_name,
			COALESCE(p.image, '') AS image_url,
			p.shop_id,
			COALESCE(sd.shop_name, '') AS shop_name
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN shop_details sd ON sd.id = p.shop_id
		WHERE p.name ILIKE $1
			OR p.description ILIKE $1
			OR c.name ILIKE $1
		ORDER BY
			CASE WHEN p.name ILIKE $2 THEN 0 ELSE 1 END,
			p.created_at DESC
		LIMIT $3 OFFSET $4
	`, likeQuery, query+"%", limit, offset).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]response.ProductSearchItem, 0)
	for rows.Next() {
		var item response.ProductSearchItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.CategoryID,
			&item.CategoryName, &item.ImageURL, &item.ShopID, &item.ShopName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *searchRepository) searchCategoriesDB(ctx context.Context, query string, limit, offset uint) ([]response.CategorySearchItem, error) {
	likeQuery := "%" + query + "%"
	rows, err := r.DB.WithContext(ctx).Raw(`
		SELECT c.id, c.name, c.department_id,
			COALESCE(d.name, '') AS department_name,
			COALESCE(c.image_url, '') AS image_url
		FROM categories c
		LEFT JOIN departments d ON d.id = c.department_id
		WHERE c.name ILIKE $1
			AND c.is_active = true
		ORDER BY
			CASE WHEN c.name ILIKE $2 THEN 0 ELSE 1 END,
			c.sort_order ASC
		LIMIT $3 OFFSET $4
	`, likeQuery, query+"%", limit, offset).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]response.CategorySearchItem, 0)
	for rows.Next() {
		var item response.CategorySearchItem
		if err := rows.Scan(&item.ID, &item.Name, &item.DepartmentID, &item.DepartmentName, &item.ImageURL); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *searchRepository) searchShopsDB(ctx context.Context, query string, lat, lng, radiusKm float64,
	pincode string, limit, offset uint) ([]response.ShopSearchItem, error) {

	likeQuery := "%" + query + "%"
	var rows *gorm.DB

	hasGeo := lat != 0 && lng != 0 && radiusKm > 0

	if hasGeo {
		// Haversine formula for geo-filtering with text search
		rows = r.DB.WithContext(ctx).Raw(`
			SELECT sd.id, sd.shop_name, sd.city, sd.state, sd.pincode,
				COALESCE(sd.shop_image_url, '') AS shop_image_url,
				sd.latitude, sd.longitude, COALESCE(sd.shop_type, '') AS shop_type
			FROM shop_details sd
			WHERE (sd.shop_name ILIKE $1 OR sd.city ILIKE $1 OR sd.owner_name ILIKE $1)
				AND sd.shop_verification_status = true
				AND (6371 * acos(
					cos(radians($5)) * cos(radians(sd.latitude)) *
					cos(radians(sd.longitude) - radians($6)) +
					sin(radians($5)) * sin(radians(sd.latitude))
				)) <= $7
			ORDER BY
				CASE WHEN sd.shop_name ILIKE $2 THEN 0 ELSE 1 END,
				sd.created_at DESC
			LIMIT $3 OFFSET $4
		`, likeQuery, query+"%", limit, offset, lat, lng, radiusKm)
	} else if pincode != "" {
		rows = r.DB.WithContext(ctx).Raw(`
			SELECT sd.id, sd.shop_name, sd.city, sd.state, sd.pincode,
				COALESCE(sd.shop_image_url, '') AS shop_image_url,
				sd.latitude, sd.longitude, COALESCE(sd.shop_type, '') AS shop_type
			FROM shop_details sd
			WHERE (sd.shop_name ILIKE $1 OR sd.city ILIKE $1 OR sd.owner_name ILIKE $1)
				AND sd.shop_verification_status = true
				AND sd.pincode = $5
			ORDER BY
				CASE WHEN sd.shop_name ILIKE $2 THEN 0 ELSE 1 END,
				sd.created_at DESC
			LIMIT $3 OFFSET $4
		`, likeQuery, query+"%", limit, offset, pincode)
	} else {
		rows = r.DB.WithContext(ctx).Raw(`
			SELECT sd.id, sd.shop_name, sd.city, sd.state, sd.pincode,
				COALESCE(sd.shop_image_url, '') AS shop_image_url,
				sd.latitude, sd.longitude, COALESCE(sd.shop_type, '') AS shop_type
			FROM shop_details sd
			WHERE (sd.shop_name ILIKE $1 OR sd.city ILIKE $1 OR sd.owner_name ILIKE $1)
				AND sd.shop_verification_status = true
			ORDER BY
				CASE WHEN sd.shop_name ILIKE $2 THEN 0 ELSE 1 END,
				sd.created_at DESC
			LIMIT $3 OFFSET $4
		`, likeQuery, query+"%", limit, offset)
	}

	sqlRows, err := rows.Rows()
	if err != nil {
		return nil, err
	}
	defer sqlRows.Close()

	items := make([]response.ShopSearchItem, 0)
	for sqlRows.Next() {
		var item response.ShopSearchItem
		if err := sqlRows.Scan(&item.ID, &item.ShopName, &item.City, &item.State, &item.Pincode,
			&item.ShopImageURL, &item.Latitude, &item.Longitude, &item.ShopType); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *searchRepository) searchBrandsDB(ctx context.Context, query string, limit, offset uint) ([]response.BrandSearchItem, error) {
	likeQuery := "%" + query + "%"
	rows, err := r.DB.WithContext(ctx).Raw(`
		SELECT id, name FROM brands
		WHERE name ILIKE $1 AND is_active = true
		ORDER BY
			CASE WHEN name ILIKE $2 THEN 0 ELSE 1 END,
			sort_order ASC
		LIMIT $3 OFFSET $4
	`, likeQuery, query+"%", limit, offset).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]response.BrandSearchItem, 0)
	for rows.Next() {
		var item response.BrandSearchItem
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *searchRepository) searchDepartmentsDB(ctx context.Context, query string, limit, offset uint) ([]response.DepartmentSearchItem, error) {
	likeQuery := "%" + query + "%"
	rows, err := r.DB.WithContext(ctx).Raw(`
		SELECT id, name, COALESCE(image_url, '') AS image_url FROM departments
		WHERE name ILIKE $1 AND is_active = true
		ORDER BY
			CASE WHEN name ILIKE $2 THEN 0 ELSE 1 END,
			sort_order ASC
		LIMIT $3 OFFSET $4
	`, likeQuery, query+"%", limit, offset).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]response.DepartmentSearchItem, 0)
	for rows.Next() {
		var item response.DepartmentSearchItem
		if err := rows.Scan(&item.ID, &item.Name, &item.ImageURL); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *searchRepository) searchOffersDB(ctx context.Context, query string, limit, offset uint) ([]response.OfferSearchItem, error) {
	likeQuery := "%" + query + "%"
	rows, err := r.DB.WithContext(ctx).Raw(`
		SELECT id, name, description, discount_rate, offer_type, start_date, end_date,
			COALESCE(image, '') AS image_url
		FROM offers
		WHERE (name ILIKE $1 OR description ILIKE $1)
			AND is_active = true
			AND end_date > NOW()
		ORDER BY
			CASE WHEN name ILIKE $2 THEN 0 ELSE 1 END,
			start_date DESC
		LIMIT $3 OFFSET $4
	`, likeQuery, query+"%", limit, offset).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]response.OfferSearchItem, 0)
	for rows.Next() {
		var item response.OfferSearchItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.DiscountRate,
			&item.OfferType, &item.StartDate, &item.EndDate, &item.Image); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// ---------- Autocomplete ----------

func (r *searchRepository) AutocompleteSuggestions(ctx context.Context, prefix string, limit uint) ([]response.AutocompleteSuggestion, error) {
	prefixLike := prefix + "%"
	containsLike := "%" + prefix + "%"

	// Distribute limits across entities for diversity
	productLimit := limit/2 + 1 // products get more suggestions
	otherLimit := limit / 4
	if otherLimit < 1 {
		otherLimit = 1
	}

	rows, err := r.DB.WithContext(ctx).Raw(`
		(
			SELECT name AS text, 'product' AS entity_type, id, COALESCE(image, '') AS image_url
			FROM products
			WHERE name ILIKE $1 OR name ILIKE $2
			ORDER BY CASE WHEN name ILIKE $1 THEN 0 ELSE 1 END, name
			LIMIT $3
		)
		UNION ALL
		(
			SELECT name AS text, 'category' AS entity_type, id, COALESCE(image_url, '') AS image_url
			FROM categories
			WHERE (name ILIKE $1 OR name ILIKE $2) AND is_active = true
			ORDER BY CASE WHEN name ILIKE $1 THEN 0 ELSE 1 END, sort_order
			LIMIT $4
		)
		UNION ALL
		(
			SELECT shop_name AS text, 'shop' AS entity_type, id, COALESCE(shop_image_url, '') AS image_url
			FROM shop_details
			WHERE (shop_name ILIKE $1 OR shop_name ILIKE $2) AND shop_verification_status = true
			ORDER BY CASE WHEN shop_name ILIKE $1 THEN 0 ELSE 1 END, shop_name
			LIMIT $4
		)
		UNION ALL
		(
			SELECT name AS text, 'brand' AS entity_type, id, '' AS image_url
			FROM brands
			WHERE (name ILIKE $1 OR name ILIKE $2) AND is_active = true
			ORDER BY CASE WHEN name ILIKE $1 THEN 0 ELSE 1 END, sort_order
			LIMIT $4
		)
		UNION ALL
		(
			SELECT name AS text, 'department' AS entity_type, id, COALESCE(image_url, '') AS image_url
			FROM departments
			WHERE (name ILIKE $1 OR name ILIKE $2) AND is_active = true
			ORDER BY CASE WHEN name ILIKE $1 THEN 0 ELSE 1 END, sort_order
			LIMIT $4
		)
		UNION ALL
		(
			SELECT name AS text, 'offer' AS entity_type, id, COALESCE(image, '') AS image_url
			FROM offers
			WHERE (name ILIKE $1 OR name ILIKE $2) AND is_active = true AND end_date > NOW()
			ORDER BY CASE WHEN name ILIKE $1 THEN 0 ELSE 1 END, name
			LIMIT $4
		)
		LIMIT $5
	`, prefixLike, containsLike, productLimit, otherLimit, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	suggestions := make([]response.AutocompleteSuggestion, 0)
	for rows.Next() {
		var s response.AutocompleteSuggestion
		if err := rows.Scan(&s.Text, &s.EntityType, &s.EntityID, &s.ImageURL); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, s)
	}
	return suggestions, nil
}
