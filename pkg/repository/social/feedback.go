package social

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

type feedbackDatabase struct {
	DB *gorm.DB
}

func NewFeedbackRepository(DB *gorm.DB) repoInterface.FeedbackRepository {
	return &feedbackDatabase{DB: DB}
}

// CreateFeedback creates a new feedback entry.
func (f *feedbackDatabase) CreateFeedback(ctx context.Context, feedback domain.Feedback) (domain.Feedback, error) {
	if err := f.DB.WithContext(ctx).Create(&feedback).Error; err != nil {
		return domain.Feedback{}, fmt.Errorf("failed to create feedback: %w", err)
	}
	return feedback, nil
}

// GetFeedbackByID retrieves a feedback entry by its ID.
func (f *feedbackDatabase) GetFeedbackByID(ctx context.Context, feedbackID uint) (domain.Feedback, error) {
	var feedback domain.Feedback
	if err := f.DB.WithContext(ctx).First(&feedback, feedbackID).Error; err != nil {
		return domain.Feedback{}, fmt.Errorf("feedback with ID %d not found: %w", feedbackID, err)
	}
	return feedback, nil
}

// GetFeedbacks retrieves all feedback entries with optional filtering and pagination.
func (f *feedbackDatabase) GetFeedbacks(ctx context.Context, feedbackType *domain.FeedbackType, shopID, productItemID *string, pagination request.Pagination) ([]domain.Feedback, int64, error) {
	var feedbacks []domain.Feedback
	var totalCount int64

	query := f.DB.WithContext(ctx).Model(&domain.Feedback{})

	if feedbackType != nil {
		query = query.Where("feedback_type = ?", *feedbackType)
	}

	if shopID != nil {
		query = query.Where("shop_id = ?", *shopID)
	}

	if productItemID != nil {
		query = query.Where("product_item_id = ?", *productItemID)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count feedbacks: %w", err)
	}

	if pagination.Offset > 0 {
		query = query.Offset(int(pagination.Offset))
	}

	if pagination.Limit > 0 {
		query = query.Limit(int(pagination.Limit))
	}

	if err := query.Find(&feedbacks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve feedbacks: %w", err)
	}

	return feedbacks, totalCount, nil
}

// UpdateFeedbackStatus updates the status of a feedback entry.
func (f *feedbackDatabase) UpdateFeedbackStatus(ctx context.Context, feedbackID uint, status domain.FeedbackStatus) (domain.Feedback, error) {
	var feedback domain.Feedback
	if err := f.DB.WithContext(ctx).First(&feedback, feedbackID).Error; err != nil {
		return domain.Feedback{}, fmt.Errorf("feedback with ID %d not found: %w", feedbackID, err)
	}

	feedback.Status = status
	if err := f.DB.WithContext(ctx).Save(&feedback).Error; err != nil {
		return domain.Feedback{}, fmt.Errorf("failed to update feedback status: %w", err)
	}

	return feedback, nil
}
