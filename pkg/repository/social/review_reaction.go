package social

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

type reviewReactionDatabase struct {
	DB *gorm.DB
}

func NewReviewReactionRepository(DB *gorm.DB) repoInterface.ReviewReactionRepository {
	return &reviewReactionDatabase{DB: DB}
}

// CreateReviewReaction creates a new review reaction.
func (r *reviewReactionDatabase) CreateReviewReaction(ctx context.Context, reaction domain.ReviewReaction) (domain.ReviewReaction, error) {
	// Check for existing reaction of the same type by the same user on the same review.
	var existingReaction domain.ReviewReaction
	err := r.DB.WithContext(ctx).Where(
		"review_type = ? AND review_id = ? AND user_id = ? AND reaction_type = ?",
		reaction.ReviewType, reaction.ReviewID, reaction.UserID, reaction.ReactionType,
	).First(&existingReaction).Error

	if err == nil { // Reaction already exists
		return domain.ReviewReaction{}, fmt.Errorf("user has already submitted this reaction type for this review")
	}

	if err := r.DB.WithContext(ctx).Create(&reaction).Error; err != nil {
		return domain.ReviewReaction{}, fmt.Errorf("failed to create review reaction: %w", err)
	}
	return reaction, nil
}

// DeleteReviewReaction deletes a review reaction.
func (r *reviewReactionDatabase) DeleteReviewReaction(ctx context.Context, reactionID uint) error {
	result := r.DB.WithContext(ctx).Delete(&domain.ReviewReaction{}, reactionID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete review reaction: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("review reaction with ID %d not found", reactionID)
	}
	return nil
}

// GetReviewReaction retrieves a specific review reaction.
func (r *reviewReactionDatabase) GetReviewReaction(ctx context.Context, reviewType domain.ReviewType, reviewID uint, userID string, reactionType domain.ReactionType) (domain.ReviewReaction, error) {
	var reaction domain.ReviewReaction
	err := r.DB.WithContext(ctx).Where(
		"review_type = ? AND review_id = ? AND user_id = ? AND reaction_type = ?",
		reviewType, reviewID, userID, reactionType,
	).First(&reaction).Error
	return reaction, err
}

// GetReviewReactions retrieves all reactions for a given review with pagination.
func (r *reviewReactionDatabase) GetReviewReactions(ctx context.Context, reviewType domain.ReviewType, reviewID uint, pagination request.Pagination) ([]domain.ReviewReaction, int64, error) {
	var reactions []domain.ReviewReaction
	var totalCount int64

	query := r.DB.WithContext(ctx).Where("review_type = ? AND review_id = ?", reviewType, reviewID)

	if err := query.Model(&domain.ReviewReaction{}).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count review reactions: %w", err)
	}

	if pagination.Offset > 0 {
		query = query.Offset(int(pagination.Offset))
	}

	if pagination.Limit > 0 {
		query = query.Limit(int(pagination.Limit))
	}

	if err := query.Find(&reactions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve review reactions: %w", err)
	}

	return reactions, totalCount, nil
}
