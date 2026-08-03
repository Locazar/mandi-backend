package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type shopUpdateDatabase struct {
	DB *gorm.DB
}

func NewShopUpdateRepository(db *gorm.DB) repo.ShopUpdateRepository {
	return &shopUpdateDatabase{DB: db}
}

// ── Admin CRUD ──────────────────────────────────────────────────────────────

func (c *shopUpdateDatabase) Create(ctx context.Context, su domain.ShopUpdate) (domain.ShopUpdate, error) {
	su.ID = domain.NewID(domain.PrefixShopUpdate)
	if su.ActionLabel == "" {
		su.ActionLabel = "Visit"
	}
	err := c.DB.WithContext(ctx).Omit("Products").Create(&su).Error
	return su, err
}

func (c *shopUpdateDatabase) List(ctx context.Context) ([]domain.ShopUpdate, error) {
	var updates []domain.ShopUpdate
	if err := c.DB.WithContext(ctx).Order("sort_order ASC, created_at DESC").Find(&updates).Error; err != nil {
		return nil, err
	}
	for i := range updates {
		products, err := c.ListProducts(ctx, updates[i].ID)
		if err != nil {
			return nil, err
		}
		updates[i].Products = products
	}
	return updates, nil
}

func (c *shopUpdateDatabase) GetByID(ctx context.Context, id string) (domain.ShopUpdate, error) {
	var su domain.ShopUpdate
	if err := c.DB.WithContext(ctx).First(&su, "id = ?", id).Error; err != nil {
		return su, err
	}
	products, err := c.ListProducts(ctx, id)
	if err != nil {
		return su, err
	}
	su.Products = products
	return su, nil
}

func (c *shopUpdateDatabase) Update(ctx context.Context, su domain.ShopUpdate) (domain.ShopUpdate, error) {
	su.UpdatedAt = time.Now()
	err := c.DB.WithContext(ctx).Model(&domain.ShopUpdate{}).
		Where("id = ?", su.ID).
		Updates(map[string]interface{}{
			"shop_id":      su.ShopID,
			"image_url":    su.ImageURL,
			"action_label": su.ActionLabel,
			"is_active":    su.IsActive,
			"sort_order":   su.SortOrder,
			"updated_at":   su.UpdatedAt,
		}).Error
	if err != nil {
		return su, err
	}
	return c.GetByID(ctx, su.ID)
}

func (c *shopUpdateDatabase) Delete(ctx context.Context, id string) error {
	return c.DB.WithContext(ctx).Delete(&domain.ShopUpdate{}, "id = ?", id).Error
}

// ── Product items ───────────────────────────────────────────────────────────

func (c *shopUpdateDatabase) AddProduct(ctx context.Context, p domain.ShopUpdateProduct) (domain.ShopUpdateProduct, error) {
	p.ID = domain.NewID(domain.PrefixShopUpdateProduct)
	err := c.DB.WithContext(ctx).Create(&p).Error
	return p, err
}

func (c *shopUpdateDatabase) UpdateProduct(ctx context.Context, p domain.ShopUpdateProduct) (domain.ShopUpdateProduct, error) {
	p.UpdatedAt = time.Now()
	err := c.DB.WithContext(ctx).Model(&domain.ShopUpdateProduct{}).
		Where("id = ?", p.ID).
		Updates(map[string]interface{}{
			"product_item_id": p.ProductItemID,
			"title":           p.Title,
			"image_url":       p.ImageURL,
			"attribute":       p.Attribute,
			"sort_order":      p.SortOrder,
			"is_active":       p.IsActive,
			"updated_at":      p.UpdatedAt,
		}).Error
	return p, err
}

func (c *shopUpdateDatabase) DeleteProduct(ctx context.Context, id string) error {
	return c.DB.WithContext(ctx).Delete(&domain.ShopUpdateProduct{}, "id = ?", id).Error
}

func (c *shopUpdateDatabase) ListProducts(ctx context.Context, shopUpdateID string) ([]domain.ShopUpdateProduct, error) {
	products := []domain.ShopUpdateProduct{}
	err := c.DB.WithContext(ctx).
		Where("shop_update_id = ?", shopUpdateID).
		Order("sort_order ASC, created_at ASC").
		Find(&products).Error
	return products, err
}

// ── Customer read ───────────────────────────────────────────────────────────

func (c *shopUpdateDatabase) GetForUser(ctx context.Context, lat, lng, radius float64, pincode string) ([]response.ShopUpdateCard, error) {
	// Reuse the product search's geo helper for the distance column + radius
	// filter against shop_details' coordinates. We build our own ORDER BY (the
	// helper's references pi.created_at, which isn't in scope here).
	distanceExpr, geoFilter, _, geoArgs, nextParam := buildGeoDistanceQuery(
		lat, lng, radius, 1, "sd.latitude", "sd.longitude", "")

	params := []interface{}{}
	params = append(params, geoArgs...)

	where := " WHERE su.is_active = true" + geoFilter
	orderBy := " ORDER BY su.sort_order ASC, su.created_at DESC"

	useGeo := lat != 0 && lng != 0
	if useGeo {
		orderBy = " ORDER BY distance_km ASC NULLS LAST, su.sort_order ASC"
	} else if pincode != "" {
		where += fmt.Sprintf(" AND sd.pincode = $%d", nextParam)
		params = append(params, pincode)
		nextParam++
	}

	query := fmt.Sprintf(`
		SELECT
			sd.id            AS shop_id,
			sd.shop_name     AS shop_name,
			COALESCE(NULLIF(su.image_url, ''), sd.shop_image_url) AS shop_image_url,
			su.action_label  AS action_label,
			%s,
			(
				SELECT COALESCE(json_agg(json_build_object(
					'id', sup.id,
					'title', COALESCE(NULLIF(sup.title, ''), pi.sub_category_name, ''),
					'image_url', COALESCE(NULLIF(sup.image_url, ''), pi.product_item_images[1]),
					'attribute', NULLIF(sup.attribute, '')
				) ORDER BY sup.sort_order ASC), '[]')
				FROM shop_update_products sup
				LEFT JOIN product_items pi ON pi.id = sup.product_item_id
				WHERE sup.shop_update_id = su.id AND sup.is_active = true
			) AS products
		FROM shop_updates su
		JOIN shop_details sd ON sd.id = su.shop_id
		%s
		%s`, distanceExpr, where, orderBy)

	type rowT struct {
		ShopID       string   `gorm:"column:shop_id"`
		ShopName     string   `gorm:"column:shop_name"`
		ShopImageURL string   `gorm:"column:shop_image_url"`
		ActionLabel  string   `gorm:"column:action_label"`
		DistanceKM   *float64 `gorm:"column:distance_km"`
		Products     []byte   `gorm:"column:products"`
	}

	var rows []rowT
	if err := c.DB.WithContext(ctx).Raw(query, params...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	cards := []response.ShopUpdateCard{}
	for _, r := range rows {
		products := []response.ShopUpdateProductCard{}
		if len(r.Products) > 0 {
			_ = json.Unmarshal(r.Products, &products)
		}
		label := r.ActionLabel
		if label == "" {
			label = "Visit"
		}
		cards = append(cards, response.ShopUpdateCard{
			ShopID:      r.ShopID,
			ShopName:    r.ShopName,
			ImageURL:    r.ShopImageURL,
			DistanceKM:  r.DistanceKM,
			ActionLabel: label,
			Products:    products,
		})
	}
	return cards, nil
}
