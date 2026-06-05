package social

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

type reviewCommentDatabase struct {
	DB *gorm.DB
}

func NewReviewCommentRepository(DB *gorm.DB) repoInterface.ReviewCommentRepository {
	return &reviewCommentDatabase{DB: DB}
}

// CreateReviewComment creates a new review comment.
func (r *reviewCommentDatabase) CreateReviewComment(ctx context.Context, comment domain.ReviewComment) (domain.ReviewComment, error) {
	if err := r.DB.WithContext(ctx).Create(&comment).Error; err != nil {
		return domain.ReviewComment{}, fmt.Errorf("failed to create review comment: %w", err)
	}
	return comment, nil
}

// UpdateReviewComment updates an existing review comment.
func (r *reviewCommentDatabase) UpdateReviewComment(ctx context.Context, commentID uint, commentText string) (domain.ReviewComment, error) {
	var comment domain.ReviewComment
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", commentID).First(&comment).Error; err != nil {
		return domain.ReviewComment{}, fmt.Errorf("review comment with ID %d not found or already deleted: %w", commentID, err)
	}

	comment.Comment = commentText
	if err := r.DB.WithContext(ctx).Save(&comment).Error; err != nil {
		return domain.ReviewComment{}, fmt.Errorf("failed to update review comment: %w", err)
	}

	return comment, nil
}

// DeleteReviewComment soft deletes a review comment.
func (r *reviewCommentDatabase) DeleteReviewComment(ctx context.Context, commentID uint) error {
	var comment domain.ReviewComment
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", commentID).First(&comment).Error; err != nil {
		return fmt.Errorf("review comment with ID %d not found or already deleted: %w", commentID, err)
	}

	if err := r.DB.WithContext(ctx).Delete(&comment).Error; err != nil {
		return fmt.Errorf("failed to soft delete review comment: %w", err)
	}
	return nil
}

// GetReviewCommentByID retrieves a review comment by its ID.
func (r *reviewCommentDatabase) GetReviewCommentByID(ctx context.Context, commentID uint) (domain.ReviewComment, error) {
	var comment domain.ReviewComment
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", commentID).First(&comment).Error; err != nil {
		return domain.ReviewComment{}, fmt.Errorf("review comment with ID %d not found: %w", commentID, err)
	}
	return comment, nil
}

// GetReviewComments retrieves all comments for a given review with pagination.
func (r *reviewCommentDatabase) GetReviewComments(ctx context.Context, reviewType domain.ReviewType, reviewID uint, pagination request.Pagination) ([]domain.ReviewComment, int64, error) {
	var comments []domain.ReviewComment
	var totalCount int64

	query := r.DB.WithContext(ctx).Where("review_type = ? AND review_id = ? AND parent_comment_id IS NULL AND deleted_at IS NULL", reviewType, reviewID)

	if err := query.Model(&domain.ReviewComment{}).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count review comments: %w", err)
	}

	if pagination.Offset > 0 {
		query = query.Offset(int(pagination.Offset))
	}

	if pagination.Limit > 0 {
		query = query.Limit(int(pagination.Limit))
	}

	if err := query.Order("created_at ASC").Find(&comments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve review comments: %w", err)
	}

	return comments, totalCount, nil
}

// GetChildComments retrieves child comments for a given parent comment with pagination.
func (r *reviewCommentDatabase) GetChildComments(ctx context.Context, parentCommentID uint, pagination request.Pagination) ([]domain.ReviewComment, int64, error) {
	var comments []domain.ReviewComment
	var totalCount int64

	query := r.DB.WithContext(ctx).Where("parent_comment_id = ? AND deleted_at IS NULL", parentCommentID)

	if err := query.Model(&domain.ReviewComment{}).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count child comments: %w", err)
	}

	if pagination.Offset > 0 {
		query = query.Offset(int(pagination.Offset))
	}

	if pagination.Limit > 0 {
		query = query.Limit(int(pagination.Limit))
	}

	if err := query.Order("created_at ASC").Find(&comments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve child comments: %w", err)
	}

	return comments, totalCount, nil
}
