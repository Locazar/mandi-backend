package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type ShopReviewRepository interface {
	CreateShopReview(ctx context.Context, review domain.ShopReview) (domain.ShopReview, error)
	UpdateShopReview(ctx context.Context, reviewID uint, reviewData map[string]interface{}) (domain.ShopReview, error)
	DeleteShopReview(ctx context.Context, reviewID uint) error
	GetShopReviewByID(ctx context.Context, reviewID uint) (domain.ShopReview, error)
	GetShopReviews(ctx context.Context, shopID string, pagination request.Pagination) ([]domain.ShopReview, int64, error)
	GetShopReviewByUserAndShopID(ctx context.Context, userID, shopID string) (domain.ShopReview, error)
}

type ProductReviewRepository interface {
	CreateProductReview(ctx context.Context, review domain.ProductReview) (domain.ProductReview, error)
	UpdateProductReview(ctx context.Context, reviewID uint, reviewData map[string]interface{}) (domain.ProductReview, error)
	DeleteProductReview(ctx context.Context, reviewID uint) error
	GetProductReviewByID(ctx context.Context, reviewID uint) (domain.ProductReview, error)
	GetProductReviews(ctx context.Context, productItemID string, pagination request.Pagination) ([]domain.ProductReview, int64, error)
	GetProductReviewByUserAndProductID(ctx context.Context, userID, productItemID string) (domain.ProductReview, error)
}

type ReviewImageRepository interface {
	UploadReviewImage(ctx context.Context, image domain.ReviewImage) (domain.ReviewImage, error)
	GetReviewImagesByReviewID(ctx context.Context, reviewType domain.ReviewType, reviewID uint) ([]domain.ReviewImage, error)
	DeleteReviewImage(ctx context.Context, imageID uint) error
	DeleteReviewImagesByReviewID(ctx context.Context, reviewType domain.ReviewType, reviewID uint) error
}

type ReviewCommentRepository interface {
	CreateReviewComment(ctx context.Context, comment domain.ReviewComment) (domain.ReviewComment, error)
	UpdateReviewComment(ctx context.Context, commentID uint, commentText string) (domain.ReviewComment, error)
	DeleteReviewComment(ctx context.Context, commentID uint) error
	GetReviewCommentByID(ctx context.Context, commentID uint) (domain.ReviewComment, error)
	GetReviewComments(ctx context.Context, reviewType domain.ReviewType, reviewID uint, pagination request.Pagination) ([]domain.ReviewComment, int64, error)
	GetChildComments(ctx context.Context, parentCommentID uint, pagination request.Pagination) ([]domain.ReviewComment, int64, error)
}

type ReviewReactionRepository interface {
	CreateReviewReaction(ctx context.Context, reaction domain.ReviewReaction) (domain.ReviewReaction, error)
	DeleteReviewReaction(ctx context.Context, reactionID uint) error
	GetReviewReaction(ctx context.Context, reviewType domain.ReviewType, reviewID uint, userID string, reactionType domain.ReactionType) (domain.ReviewReaction, error)
	GetReviewReactions(ctx context.Context, reviewType domain.ReviewType, reviewID uint, pagination request.Pagination) ([]domain.ReviewReaction, int64, error)
}

type FeedbackRepository interface {
	CreateFeedback(ctx context.Context, feedback domain.Feedback) (domain.Feedback, error)
	GetFeedbackByID(ctx context.Context, feedbackID uint) (domain.Feedback, error)
	GetFeedbacks(ctx context.Context, feedbackType *domain.FeedbackType, shopID, productItemID *string, pagination request.Pagination) ([]domain.Feedback, int64, error)
	UpdateFeedbackStatus(ctx context.Context, feedbackID uint, status domain.FeedbackStatus) (domain.Feedback, error)
}

type ShopRatingSummaryRepository interface {
	GetShopRatingSummary(ctx context.Context, shopID string) (domain.ShopRatingSummary, error)
	UpdateShopRatingSummary(ctx context.Context, shopID string, newRating float32, oldRating *float32) (domain.ShopRatingSummary, error)
	InitShopRatingSummary(ctx context.Context, shopID string) (domain.ShopRatingSummary, error)
}

type ProductRatingSummaryRepository interface {
	GetProductRatingSummary(ctx context.Context, productItemID string) (domain.ProductRatingSummary, error)
	UpdateProductRatingSummary(ctx context.Context, productItemID string, newRating float32, oldRating *float32) (domain.ProductRatingSummary, error)
	InitProductRatingSummary(ctx context.Context, productItemID string) (domain.ProductRatingSummary, error)
}
