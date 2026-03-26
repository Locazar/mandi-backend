package repository

import (
	"context"
	"fmt"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type promotionRepository struct {
	db *gorm.DB
}

func NewPromotionRepository(db *gorm.DB) *promotionRepository {
	return &promotionRepository{
		db: db,
	}
}

func (r *promotionRepository) Transactions(ctx context.Context, trxFn func(repo interfaces.PromotionRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		repo := NewPromotionRepository(tx)
		return trxFn(repo)
	})
}

func (r *promotionRepository) FindAllPromotionCategories(ctx context.Context, pagination request.Pagination) ([]response.PromotionCategory, error) {
	var categories []response.PromotionCategory

	offset := pagination.Offset

	query := r.db.Model(&domain.PromotionCategory{}).
		Select("id, name, shop_id, is_active, icon_path, created_at, updated_at").
		Order("created_at DESC").
		Limit(int(pagination.Limit)).
		Offset(int(offset))

	if err := query.Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to find promotion categories: %w", err)
	}

	return categories, nil
}

func (r *promotionRepository) FindPromotionCategoryByID(ctx context.Context, categoryID uint) (response.PromotionCategory, error) {
	var category response.PromotionCategory

	if err := r.db.Model(&domain.PromotionCategory{}).
		Select("id, name, shop_id, is_active, icon_path, created_at, updated_at").
		Where("id = ?", categoryID).
		First(&category).Error; err != nil {
		return category, fmt.Errorf("failed to find promotion category by ID: %w", err)
	}

	return category, nil
}

func (r *promotionRepository) FindAllPromotionTypes(ctx context.Context, pagination request.Pagination) ([]response.PromotionsType, error) {
	var types []response.PromotionsType

	offset := pagination.Offset

	query := `SELECT pt.id, pt.name, pt.is_active, pt.shop_id, pt.promotion_category_id, pt.promotion_offer_details, pt.icon_path, pt.created_at, pt.updated_at, pt.type, pc.name as category_name
	          FROM promotions_types pt
	          LEFT JOIN promotion_categories pc ON pt.promotion_category_id = pc.id
	          ORDER BY pt.created_at DESC
	          LIMIT ? OFFSET ?`

	if err := r.db.Raw(query, pagination.Limit, offset).Find(&types).Error; err != nil {
		return nil, fmt.Errorf("failed to find promotion types: %w", err)
	}

	return types, nil
}

func (r *promotionRepository) FindPromotionTypesByCategoryID(ctx context.Context, categoryID uint, pagination request.Pagination) ([]response.PromotionsType, error) {
	var types []response.PromotionsType

	offset := pagination.Offset

	query := `SELECT pt.id, pt.name, pt.is_active, pt.shop_id, pt.promotion_category_id, pt.promotion_offer_details, pt.icon_path, pt.created_at, pt.updated_at, pt.type, pc.name as category_name
	          FROM promotions_types pt
	          LEFT JOIN promotion_categories pc ON pt.promotion_category_id = pc.id
	          WHERE pt.promotion_category_id = ?
	          ORDER BY pt.created_at ASC
	          LIMIT ? OFFSET ?`

	if err := r.db.Raw(query, categoryID, pagination.Limit, offset).Find(&types).Error; err != nil {
		return nil, fmt.Errorf("failed to find promotion types by category ID: %w", err)
	}

	return types, nil
}

func (r *promotionRepository) FindPromotionTypeByID(ctx context.Context, typeID uint) (response.PromotionsType, error) {
	var promotionType response.PromotionsType

	query := `SELECT pt.id, pt.name, pt.is_active, pt.shop_id, pt.promotion_category_id, pt.promotion_offer_details, pt.icon_path, pt.created_at, pt.updated_at, pt.type, pc.name as category_name
	          FROM promotions_types pt
	          LEFT JOIN promotion_categories pc ON pt.promotion_category_id = pc.id
	          WHERE pt.id = ?`

	if err := r.db.Raw(query, typeID).First(&promotionType).Error; err != nil {
		return promotionType, fmt.Errorf("failed to find promotion type by ID: %w", err)
	}

	return promotionType, nil
}

func (r *promotionRepository) CreatePromotion(ctx context.Context, promotion domain.Promotion) (response.Promotion, error) {
	//var existingPromotion domain.Promotion
	// if err := r.db.Where("promotion_type_id = ?", promotion.PromotionTypeID).First(&existingPromotion).Error; err == nil {
	// 	return response.Promotion{}, fmt.Errorf("promotion with this promotion_type_id already exists")
	// }

	if err := r.db.Create(&promotion).Error; err != nil {
		fmt.Printf("Error creating promotion: %v\n", err) // Log the error for debugging
		return response.Promotion{}, fmt.Errorf("failed to create promotion: %w", err)
	}

	var createdPromotion response.Promotion
	query := `SELECT p.id, p.promotion_category_id, p.promotion_type_id, p.shop_id, p.is_active, p.created_at, p.updated_at,
	          pc.name as promotion_category__name, pc.shop_id as promotion_category__shop_id, pc.is_active as promotion_category__is_active, pc.icon_path as promotion_category__icon_path, pc.created_at as promotion_category__created_at, pc.updated_at as promotion_category__updated_at,
	          pt.id as promotion_type__id, pt.name as promotion_type__name, pt.is_active as promotion_type__is_active, pt.shop_id as promotion_type__shop_id, pt.promotion_category_id as promotion_type__promotion_category_id, pt.promotion_offer_details as promotion_type__promotion_offer_details, pt.type as promotion_type__type, pt.icon_path as promotion_type__icon_path, pt.created_at as promotion_type__created_at, pt.updated_at as promotion_type__updated_at
	          FROM promotions p
	          LEFT JOIN promotion_categories pc ON p.promotion_category_id = pc.id
	          LEFT JOIN promotions_types pt ON p.promotion_type_id = pt.id
	          WHERE p.id = ?`

	if err := r.db.Raw(query, promotion.ID).First(&createdPromotion).Error; err != nil {
		return response.Promotion{}, fmt.Errorf("failed to retrieve created promotion: %w", err)
	}

	return createdPromotion, nil
}

func (r *promotionRepository) GetAllPromotions(ctx context.Context, pagination request.Pagination) ([]response.Promotion, error) {
	var promotions []response.Promotion

	offset := pagination.Offset
	limit := pagination.Limit

	query := r.db.
		Preload("PromotionCategory").
		Preload("PromotionType").
		Where("is_active = true").
		Where("start_date IS NOT NULL").
		Where("end_date IS NOT NULL").
		Where("(end_date)::timestamp >= CURRENT_TIMESTAMP").
		Order("created_at DESC").
		Limit(int(limit)).
		Offset(int(offset))

	if err := query.Find(&promotions).Error; err != nil {
		return nil, fmt.Errorf("failed to find promotions: %w", err)
	}

	// Populate IconPath from PromotionType and dereference pointer fields
	for i := range promotions {
		if promotions[i].PromotionType.IconPath != "" {
			promotions[i].IconPath = promotions[i].PromotionType.IconPath
		}
		// Dereference pointer fields to ensure they display correctly
		// (omitempty will handle nil values automatically)
	}

	return promotions, nil
}

func (r *promotionRepository) GetPromotionByID(ctx context.Context, promotionID uint) (response.Promotion, error) {
	var promotion response.Promotion

	query := `SELECT p.id, p.promotion_category_id, p.promotion_type_id, p.shop_id, p.is_active, p.created_at, p.updated_at,
	          pc.name as promotion_category__name, pc.shop_id as promotion_category__shop_id, pc.is_active as promotion_category__is_active, pc.icon_path as promotion_category__icon_path, pc.created_at as promotion_category__created_at, pc.updated_at as promotion_category__updated_at,
	          pt.id as promotion_type__id, pt.name as promotion_type__name, pt.is_active as promotion_type__is_active, pt.shop_id as promotion_type__shop_id, pt.promotion_category_id as promotion_type__promotion_category_id, pt.promotion_offer_details as promotion_type__promotion_offer_details, pt.type as promotion_type__type, pt.icon_path as promotion_type__icon_path, pt.created_at as promotion_type__created_at, pt.updated_at as promotion_type__updated_at
	          FROM promotions p
	          LEFT JOIN promotion_categories pc ON p.promotion_category_id = pc.id
	          LEFT JOIN promotions_types pt ON p.promotion_type_id = pt.id
	          WHERE p.id = ?`

	if err := r.db.Raw(query, promotionID).First(&promotion).Error; err != nil {
		return promotion, fmt.Errorf("failed to find promotion by ID: %w", err)
	}

	return promotion, nil
}

func (r *promotionRepository) DeletePromotion(ctx context.Context, promotionID uint) error {
	if err := r.db.Delete(&domain.Promotion{}, promotionID).Error; err != nil {
		return fmt.Errorf("failed to delete promotion: %w", err)
	}
	return nil
}
