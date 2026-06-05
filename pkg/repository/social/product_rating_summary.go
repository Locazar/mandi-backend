package social

import (
	"context"
	"fmt"
	"math"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

type productRatingSummaryDatabase struct {
	DB *gorm.DB
}

func NewProductRatingSummaryRepository(DB *gorm.DB) repoInterface.ProductRatingSummaryRepository {
	return &productRatingSummaryDatabase{DB: DB}
}

// GetProductRatingSummary retrieves the rating summary for a product.
func (p *productRatingSummaryDatabase) GetProductRatingSummary(ctx context.Context, productItemID string) (domain.ProductRatingSummary, error) {
	var summary domain.ProductRatingSummary
	if err := p.DB.WithContext(ctx).Where("product_item_id = ?", productItemID).First(&summary).Error; err != nil {
		return domain.ProductRatingSummary{}, fmt.Errorf("product rating summary for product item ID %s not found: %w", productItemID, err)
	}
	return summary, nil
}

// UpdateProductRatingSummary updates the rating summary for a product.
// newRating: the rating from the new or updated review.
// oldRating: if updating a review, the previous rating; nil if creating a new review or deleting one.
func (p *productRatingSummaryDatabase) UpdateProductRatingSummary(ctx context.Context, productItemID string, newRating float32, oldRating *float32) (domain.ProductRatingSummary, error) {
	var summary domain.ProductRatingSummary

	err := p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Use For UPDATE to lock the row during transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("product_item_id = ?", productItemID).First(&summary).Error; err != nil {
			return fmt.Errorf("failed to get product rating summary for update: %w", err)
		}

		if oldRating == nil { // New review (or review restored from soft-delete)
			summary.TotalReviews++
		} else { // Existing review updated
			// Decrement count for old rating
			switch uint(*oldRating) {
			case 1:
				summary.Rating1Count--
			case 2:
				summary.Rating2Count--
			case 3:
				summary.Rating3Count--
			case 4:
				summary.Rating4Count--
			case 5:
				summary.Rating5Count--
			}
		}

		// Increment count for new rating
		switch uint(newRating) {
		case 1:
			summary.Rating1Count++
		case 2:
			summary.Rating2Count++
		case 3:
			summary.Rating3Count++
		case 4:
			summary.Rating4Count++
		case 5:
			summary.Rating5Count++
		}

		// Recalculate average rating
		var totalRatingSum float32
		totalRatingSum = float32(summary.Rating1Count*1 + summary.Rating2Count*2 + summary.Rating3Count*3 + summary.Rating4Count*4 + summary.Rating5Count*5)

		if summary.TotalReviews > 0 {
			summary.AverageRating = float32(math.Round(float64(totalRatingSum)/float64(summary.TotalReviews)*10) / 10)
		} else {
			summary.AverageRating = 0.0
		}

		if err := tx.Save(&summary).Error; err != nil {
			return fmt.Errorf("failed to update product rating summary in transaction: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.ProductRatingSummary{}, err
	}

	return summary, nil
}

// InitProductRatingSummary initializes a new product rating summary entry if it doesn't exist.
func (p *productRatingSummaryDatabase) InitProductRatingSummary(ctx context.Context, productItemID string) (domain.ProductRatingSummary, error) {
	summary := domain.ProductRatingSummary{ProductItemID: productItemID, AverageRating: 0.0, TotalReviews: 0}
	err := p.DB.WithContext(ctx).Where("product_item_id = ?", productItemID).FirstOrCreate(&summary).Error
	if err != nil {
		return domain.ProductRatingSummary{}, fmt.Errorf("failed to initialize product rating summary: %w", err)
	}
	return summary, nil
}
