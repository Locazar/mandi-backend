package social

import (
	"context"
	"fmt"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type reviewImageDatabase struct {
	DB *gorm.DB
}

func NewReviewImageRepository(DB *gorm.DB) repoInterface.ReviewImageRepository {
	return &reviewImageDatabase{DB: DB}
}

// UploadReviewImage uploads a new review image.
func (r *reviewImageDatabase) UploadReviewImage(ctx context.Context, image domain.ReviewImage) (domain.ReviewImage, error) {
	if err := r.DB.WithContext(ctx).Create(&image).Error; err != nil {
		return domain.ReviewImage{}, fmt.Errorf("failed to upload review image: %w", err)
	}
	return image, nil
}

// GetReviewImagesByReviewID retrieves review images for a given review ID and type.
func (r *reviewImageDatabase) GetReviewImagesByReviewID(ctx context.Context, reviewType domain.ReviewType, reviewID uint) ([]domain.ReviewImage, error) {
	var images []domain.ReviewImage
	if err := r.DB.WithContext(ctx).Where("review_type = ? AND review_id = ?", reviewType, reviewID).Find(&images).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve review images: %w", err)
	}
	return images, nil
}

// DeleteReviewImage deletes a review image by its ID.
func (r *reviewImageDatabase) DeleteReviewImage(ctx context.Context, imageID uint) error {
	result := r.DB.WithContext(ctx).Delete(&domain.ReviewImage{}, imageID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete review image: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("review image with ID %d not found", imageID)
	}
	return nil
}

// DeleteReviewImagesByReviewID deletes all review images for a given review ID and type.
func (r *reviewImageDatabase) DeleteReviewImagesByReviewID(ctx context.Context, reviewType domain.ReviewType, reviewID uint) error {
	result := r.DB.WithContext(ctx).Where("review_type = ? AND review_id = ?", reviewType, reviewID).Delete(&domain.ReviewImage{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete review images for review %d: %w", reviewID, result.Error)
	}
	return nil
}
