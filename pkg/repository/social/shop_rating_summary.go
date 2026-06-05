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

type shopRatingSummaryDatabase struct {
	DB *gorm.DB
}

func NewShopRatingSummaryRepository(DB *gorm.DB) repoInterface.ShopRatingSummaryRepository {
	return &shopRatingSummaryDatabase{DB: DB}
}

// GetShopRatingSummary retrieves the rating summary for a shop.
func (s *shopRatingSummaryDatabase) GetShopRatingSummary(ctx context.Context, shopID string) (domain.ShopRatingSummary, error) {
	var summary domain.ShopRatingSummary
	if err := s.DB.WithContext(ctx).Where("shop_id = ?", shopID).First(&summary).Error; err != nil {
		return domain.ShopRatingSummary{}, fmt.Errorf("shop rating summary for shop ID %s not found: %w", shopID, err)
	}
	return summary, nil
}

// UpdateShopRatingSummary updates the rating summary for a shop.
// newRating: the rating from the new or updated review.
// oldRating: if updating a review, the previous rating; nil if creating a new review or deleting one.
func (s *shopRatingSummaryDatabase) UpdateShopRatingSummary(ctx context.Context, shopID string, newRating float32, oldRating *float32) (domain.ShopRatingSummary, error) {
	var summary domain.ShopRatingSummary

	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Use For UPDATE to lock the row during transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("shop_id = ?", shopID).First(&summary).Error; err != nil {
			return fmt.Errorf("failed to get shop rating summary for update: %w", err)
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
			return fmt.Errorf("failed to update shop rating summary in transaction: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.ShopRatingSummary{}, err
	}

	return summary, nil
}

// InitShopRatingSummary initializes a new shop rating summary entry if it doesn't exist.
func (s *shopRatingSummaryDatabase) InitShopRatingSummary(ctx context.Context, shopID string) (domain.ShopRatingSummary, error) {
	summary := domain.ShopRatingSummary{ShopID: shopID, AverageRating: 0.0, TotalReviews: 0}
	err := s.DB.WithContext(ctx).Where("shop_id = ?", shopID).FirstOrCreate(&summary).Error
	if err != nil {
		return domain.ShopRatingSummary{}, fmt.Errorf("failed to initialize shop rating summary: %w", err)
	}
	return summary, nil
}
