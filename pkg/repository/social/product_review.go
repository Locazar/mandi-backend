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

type productReviewDatabase struct {
	DB *gorm.DB
}

func NewProductReviewRepository(DB *gorm.DB) repoInterface.ProductReviewRepository {
	return &productReviewDatabase{DB: DB}
}

// CreateProductReview creates a new product review.
func (p *productReviewDatabase) CreateProductReview(ctx context.Context, review domain.ProductReview) (domain.ProductReview, error) {
	// Check if the user has already reviewed this product.
	var existingReview domain.ProductReview
	err := p.DB.WithContext(ctx).Where("product_item_id = ? AND user_id = ? AND deleted_at IS NULL", review.ProductItemID, review.UserID).First(&existingReview).Error

	if err == nil { // A review already exists
		return domain.ProductReview{}, errors.New("user has already submitted a review for this product")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ProductReview{}, fmt.Errorf("failed to check for existing review: %w", err)
	}

	if err := p.DB.WithContext(ctx).Create(&review).Error; err != nil {
		return domain.ProductReview{}, fmt.Errorf("failed to create product review: %w", err)
	}
	return review, nil
}

// UpdateProductReview updates an existing product review.
func (p *productReviewDatabase) UpdateProductReview(ctx context.Context, reviewID uint, reviewData map[string]interface{}) (domain.ProductReview, error) {
	var review domain.ProductReview
	if err := p.DB.WithContext(ctx).First(&review, reviewID).Error; err != nil {
		return domain.ProductReview{}, fmt.Errorf("product review with ID %d not found: %w", reviewID, err)
	}

	if err := p.DB.WithContext(ctx).Model(&review).Where("id = ? AND deleted_at IS NULL", reviewID).Updates(reviewData).Error; err != nil {
		return domain.ProductReview{}, fmt.Errorf("failed to update product review: %w", err)
	}

	return review, nil
}

// DeleteProductReview soft deletes a product review.
func (p *productReviewDatabase) DeleteProductReview(ctx context.Context, reviewID uint) error {
	var review domain.ProductReview
	if err := p.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", reviewID).First(&review).Error; err != nil {
		return fmt.Errorf("product review with ID %d not found or already deleted: %w", reviewID, err)
	}

	if err := p.DB.WithContext(ctx).Delete(&review).Error; err != nil {
		return fmt.Errorf("failed to soft delete product review: %w", err)
	}
	return nil
}

// GetProductReviewByID retrieves a product review by its ID.
func (p *productReviewDatabase) GetProductReviewByID(ctx context.Context, reviewID uint) (domain.ProductReview, error) {
	var review domain.ProductReview
	if err := p.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", reviewID).First(&review).Error; err != nil {
		return domain.ProductReview{}, fmt.Errorf("product review with ID %d not found: %w", reviewID, err)
	}
	return review, nil
}

// GetProductReviews retrieves all product reviews for a given product with pagination.
func (p *productReviewDatabase) GetProductReviews(ctx context.Context, productItemID string, pagination request.Pagination) ([]domain.ProductReview, int64, error) {
	var reviews []domain.ProductReview
	var totalCount int64

	query := p.DB.WithContext(ctx).Where("product_item_id = ? AND deleted_at IS NULL", productItemID)

	if err := query.Model(&domain.ProductReview{}).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count product reviews: %w", err)
	}

	if pagination.Offset > 0 {
		query = query.Offset(int(pagination.Offset))
	}

	if pagination.Limit > 0 {
		query = query.Limit(int(pagination.Limit))
	}

	if err := query.Find(&reviews).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve product reviews: %w", err)
	}

	return reviews, totalCount, nil
}

// GetProductReviewByUserAndProductID retrieves a product review by user and product ID.
func (p *productReviewDatabase) GetProductReviewByUserAndProductID(ctx context.Context, userID, productItemID string) (domain.ProductReview, error) {
	var review domain.ProductReview
	err := p.DB.WithContext(ctx).Where("user_id = ? AND product_item_id = ? AND deleted_at IS NULL", userID, productItemID).First(&review).Error
	return review, err
}
