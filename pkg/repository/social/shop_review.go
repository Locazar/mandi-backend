package social

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

type shopReviewDatabase struct {
	DB *gorm.DB
}

func NewShopReviewRepository(DB *gorm.DB) repoInterface.ShopReviewRepository {
	return &shopReviewDatabase{DB: DB}
}

// CreateShopReview creates a new shop review.
func (s *shopReviewDatabase) CreateShopReview(ctx context.Context, review domain.ShopReview) (domain.ShopReview, error) {
	// Check if the user has already reviewed this shop.
	var existingReview domain.ShopReview
	err := s.DB.WithContext(ctx).Where("shop_id = ? AND user_id = ? AND deleted_at IS NULL", review.ShopID, review.UserID).First(&existingReview).Error

	if err == nil { // A review already exists
		return domain.ShopReview{}, errors.New("user has already submitted a review for this shop")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ShopReview{}, fmt.Errorf("failed to check for existing review: %w", err)
	}

	if err := s.DB.WithContext(ctx).Create(&review).Error; err != nil {
		return domain.ShopReview{}, fmt.Errorf("failed to create shop review: %w", err)
	}
	return review, nil
}

// UpdateShopReview updates an existing shop review.
func (s *shopReviewDatabase) UpdateShopReview(ctx context.Context, reviewID uint, reviewData map[string]interface{}) (domain.ShopReview, error) {
	var review domain.ShopReview
	if err := s.DB.WithContext(ctx).First(&review, reviewID).Error; err != nil {
		return domain.ShopReview{}, fmt.Errorf("shop review with ID %d not found: %w", reviewID, err)
	}

	if err := s.DB.WithContext(ctx).Model(&review).Where("id = ? AND deleted_at IS NULL", reviewID).Updates(reviewData).Error; err != nil {
		return domain.ShopReview{}, fmt.Errorf("failed to update shop review: %w", err)
	}

	return review, nil
}

// DeleteShopReview soft deletes a shop review.
func (s *shopReviewDatabase) DeleteShopReview(ctx context.Context, reviewID uint) error {
	var review domain.ShopReview
	if err := s.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", reviewID).First(&review).Error; err != nil {
		return fmt.Errorf("shop review with ID %d not found or already deleted: %w", reviewID, err)
	}

	if err := s.DB.WithContext(ctx).Delete(&review).Error; err != nil {
		return fmt.Errorf("failed to soft delete shop review: %w", err)
	}
	return nil
}

// GetShopReviewByID retrieves a shop review by its ID.
func (s *shopReviewDatabase) GetShopReviewByID(ctx context.Context, reviewID uint) (domain.ShopReview, error) {
	var review domain.ShopReview
	if err := s.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", reviewID).First(&review).Error; err != nil {
		return domain.ShopReview{}, fmt.Errorf("shop review with ID %d not found: %w", reviewID, err)
	}
	return review, nil
}

// GetShopReviews retrieves all shop reviews for a given shop with pagination.
func (s *shopReviewDatabase) GetShopReviews(ctx context.Context, shopID string, pagination request.Pagination) ([]domain.ShopReview, int64, error) {
	var reviews []domain.ShopReview
	var totalCount int64

	query := s.DB.WithContext(ctx).Where("shop_id = ? AND deleted_at IS NULL", shopID)

	if err := query.Model(&domain.ShopReview{}).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count shop reviews: %w", err)
	}

	if pagination.Offset > 0 {
		query = query.Offset(int(pagination.Offset))
	}

	if pagination.Limit > 0 {
		query = query.Limit(int(pagination.Limit))
	}

	if err := query.Find(&reviews).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve shop reviews: %w", err)
	}

	return reviews, totalCount, nil
}

// GetShopReviewByUserAndShopID retrieves a shop review by user and shop ID.
func (s *shopReviewDatabase) GetShopReviewByUserAndShopID(ctx context.Context, userID, shopID string) (domain.ShopReview, error) {
	var review domain.ShopReview
	err := s.DB.WithContext(ctx).Where("user_id = ? AND shop_id = ? AND deleted_at IS NULL", userID, shopID).First(&review).Error
	return review, err
}
