package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

func (c *userUserCase) FollowShop(ctx context.Context, userID string, shopID string) error {
	return c.userRepo.FollowShop(ctx, userID, shopID)
}

func (c *userUserCase) UnfollowShop(ctx context.Context, userID string, shopID string) error {
	return c.userRepo.UnfollowShop(ctx, userID, shopID)
}

func (c *userUserCase) IsFollowingShop(ctx context.Context, userID string, shopID string) (bool, error) {
	return c.userRepo.IsFollowingShop(ctx, userID, shopID)
}

func (c *userUserCase) GetFollowers(ctx context.Context, shopID string) ([]response.User, error) {
	return c.userRepo.GetFollowers(ctx, shopID)
}

func (c *userUserCase) GetFollowedShops(ctx context.Context, userID string) ([]response.Shop, error) {
	return c.userRepo.GetFollowedShops(ctx, userID)
}

func (c *userUserCase) LikeShop(ctx context.Context, userID string, shopID string) error {
	return c.userRepo.LikeShop(ctx, userID, shopID)
}

func (c *userUserCase) UnlikeShop(ctx context.Context, userID string, shopID string) error {
	return c.userRepo.UnlikeShop(ctx, userID, shopID)
}

func (c *userUserCase) IsLikedShop(ctx context.Context, userID string, shopID string) (bool, error) {
	return c.userRepo.IsLikedShop(ctx, userID, shopID)
}

func (c *userUserCase) GetLikedShops(ctx context.Context, userID string) ([]response.Shop, error) {
	return c.userRepo.GetLikedShops(ctx, userID)
}

func (c *userUserCase) RateShop(ctx context.Context, userID string, shopID string, rating uint) error {
	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating should be between 1 and 5")
	}
	return c.userRepo.RateShop(ctx, userID, shopID, rating)
}

func (c *userUserCase) GetUserShopRating(ctx context.Context, userID string, shopID string) (uint, error) {
	return c.userRepo.GetUserShopRating(ctx, userID, shopID)
}

func (c *userUserCase) GetAllShopRatings(ctx context.Context, shopID string) ([]domain.ShopSocial, error) {
	return c.userRepo.GetAllShopRatings(ctx, shopID)
}

func (c *userUserCase) GetShopAverageRating(ctx context.Context, shopID string) (float64, error) {
	return c.userRepo.GetShopAverageRating(ctx, shopID)
}

func (c *userUserCase) GetShopRatingDistribution(ctx context.Context, shopID string) ([]domain.ShopRatingDistribution, error) {
	return c.userRepo.GetShopRatingDistribution(ctx, shopID)
}

func (c *userUserCase) ReviewShop(ctx context.Context, userID string, shopID string, review string) error {
	review = strings.TrimSpace(review)
	if review == "" {
		return fmt.Errorf("review cannot be empty")
	}
	return c.userRepo.ReviewShop(ctx, userID, shopID, review)
}

func (c *userUserCase) SaveShopFeedback(ctx context.Context, userID string, shopID string, rating *uint, review *string, comments *string) error {
	if rating != nil && (*rating < 1 || *rating > 5) {
		return fmt.Errorf("rating should be between 1 and 5")
	}

	if review != nil {
		trimmed := strings.TrimSpace(*review)
		if trimmed == "" {
			return fmt.Errorf("review cannot be empty")
		}
		review = &trimmed
	}

	if comments != nil {
		trimmed := strings.TrimSpace(*comments)
		if trimmed == "" {
			return fmt.Errorf("comments cannot be empty")
		}
		comments = &trimmed
	}

	return c.userRepo.SaveShopFeedback(ctx, userID, shopID, rating, review, comments)
}

func (c *userUserCase) GetUserShopReview(ctx context.Context, userID string, shopID string) (string, error) {
	return c.userRepo.GetUserShopReview(ctx, userID, shopID)
}

func (c *userUserCase) GetUserShopFeedback(ctx context.Context, userID string, shopID string) (domain.ShopSocial, bool, error) {
	return c.userRepo.GetUserShopFeedback(ctx, userID, shopID)
}

func (c *userUserCase) GetAllShopReviews(ctx context.Context, shopID string) ([]domain.ShopSocial, error) {
	return c.userRepo.GetAllShopReviews(ctx, shopID)
}

func (c *userUserCase) GetAllShopComments(ctx context.Context, shopID string) ([]domain.ShopSocial, error) {
	return c.userRepo.GetAllShopComments(ctx, shopID)
}

func (c *userUserCase) GetShopFeedbackList(ctx context.Context, shopID string) ([]domain.ShopSocial, error) {
	return c.userRepo.GetShopFeedbackList(ctx, shopID)
}

func (c *userUserCase) GetShopSocialDetails(ctx context.Context, shopID string, userID string) (domain.ShopSocialSummary, error) {
	details, err := c.userRepo.GetShopSocialDetails(ctx, shopID, userID)
	if err != nil {
		return domain.ShopSocialSummary{}, fmt.Errorf("failed to get shop social details: %w", err)
	}
	return details, nil
}
