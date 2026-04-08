package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"gorm.io/gorm"
)

func (c *userDatabase) getShopAdminID(ctx context.Context, shopID uint) (uint, error) {
	var adminID uint
	err := c.DB.WithContext(ctx).
		Model(&domain.ShopDetails{}).
		Where("id = ?", shopID).
		Select("admin_id").
		Scan(&adminID).Error
	if err != nil {
		return 0, err
	}
	if adminID == 0 {
		return 0, fmt.Errorf("shop not found")
	}
	return adminID, nil
}

func (c *userDatabase) ensureShopSocialTable() error {
	if c.DB.Migrator().HasTable(&domain.ShopSocial{}) {
		if !c.DB.Migrator().HasColumn(&domain.ShopSocial{}, "is_liked") {
			return c.DB.Migrator().AddColumn(&domain.ShopSocial{}, "is_liked")
		}
		return nil
	}
	return c.DB.AutoMigrate(&domain.ShopSocial{})
}

func (c *userDatabase) isUndefinedColumnError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "does not exist") || strings.Contains(errMsg, "sqlstate 42703")
}

func (c *userDatabase) upsertUserShopSocial(ctx context.Context, userID uint, shopID uint, mutate func(*domain.ShopSocial)) error {
	var social domain.ShopSocial
	err := c.DB.WithContext(ctx).Where("shop_id = ? AND user_id = ?", shopID, userID).First(&social).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			if c.isUndefinedColumnError(err) {
				if ensureErr := c.ensureShopSocialTable(); ensureErr != nil {
					return ensureErr
				}
				return c.upsertUserShopSocial(ctx, userID, shopID, mutate)
			}
			return err
		}
		adminID, adminErr := c.getShopAdminID(ctx, shopID)
		if adminErr != nil {
			return adminErr
		}
		social = domain.ShopSocial{ShopID: shopID, UserID: userID, AdminID: adminID}
		mutate(&social)
		createErr := c.DB.WithContext(ctx).Create(&social).Error
		if createErr != nil && c.isUndefinedColumnError(createErr) {
			if ensureErr := c.ensureShopSocialTable(); ensureErr != nil {
				return ensureErr
			}
			return c.DB.WithContext(ctx).Create(&social).Error
		}
		return createErr
	}

	mutate(&social)
	return c.DB.WithContext(ctx).Save(&social).Error
}

func (c *userDatabase) FollowShop(ctx context.Context, userID uint, shopID uint) error {
	return c.upsertUserShopSocial(ctx, userID, shopID, func(s *domain.ShopSocial) {
		s.IsFollower = true
	})
}

func (c *userDatabase) UnfollowShop(ctx context.Context, userID uint, shopID uint) error {
	return c.DB.WithContext(ctx).
		Model(&domain.ShopSocial{}).
		Where("shop_id = ? AND user_id = ?", shopID, userID).
		Update("is_follower", false).Error
}

func (c *userDatabase) IsFollowingShop(ctx context.Context, userID uint, shopID uint) (bool, error) {
	var count int64
	err := c.DB.WithContext(ctx).
		Model(&domain.ShopSocial{}).
		Where("shop_id = ? AND user_id = ? AND is_follower = ?", shopID, userID, true).
		Count(&count).Error
	return count > 0, err
}

func (c *userDatabase) GetFollowers(ctx context.Context, shopID uint) ([]response.User, error) {
	var followers []response.User
	query := `SELECT ss.user_id AS id,
		COALESCE(u.first_name, NULLIF(split_part(a.full_name, ' ', 1), ''), '') AS first_name,
		COALESCE(u.last_name, NULLIF(split_part(a.full_name, ' ', 2), ''), '') AS last_name,
		COALESCE(u.email, a.email, '') AS email,
		COALESCE(u.phone, a.mobile, '') AS phone,
		COALESCE(u.verified, a.verified_seller, FALSE) AS verified,
		COALESCE(u.block_status, FALSE) AS block_status,
		COALESCE(u.created_at, a.created_at, ss.created_at) AS created_at,
		COALESCE(u.updated_at, a.updated_at, ss.updated_at) AS updated_at
	FROM shop_socials ss
	LEFT JOIN users u ON u.id = ss.user_id
	LEFT JOIN admins a ON a.id = ss.user_id
	WHERE ss.shop_id = ? AND ss.is_follower = TRUE
	ORDER BY ss.updated_at DESC`
	err := c.DB.WithContext(ctx).Raw(query, shopID).Scan(&followers).Error
	return followers, err
}

func (c *userDatabase) GetFollowedShops(ctx context.Context, userID uint) ([]response.Shop, error) {
	var shops []response.Shop
	query := `SELECT sd.id, sd.shop_name, sd.email, sd.phone, sd.address_line1, sd.address_line2,
		sd.city, sd.state, sd.country, sd.pincode, sd.shop_type, sd.shop_verification_status,
		sd.shop_image_url, sd.latitude, sd.longitude, sd.created_at, sd.updated_at
	FROM shop_details sd
	INNER JOIN shop_socials ss ON ss.shop_id = sd.id
	WHERE ss.user_id = ? AND ss.is_follower = TRUE
	ORDER BY ss.updated_at DESC`
	err := c.DB.WithContext(ctx).Raw(query, userID).Scan(&shops).Error
	return shops, err
}

func (c *userDatabase) LikeShop(ctx context.Context, userID uint, shopID uint) error {
	return c.upsertUserShopSocial(ctx, userID, shopID, func(s *domain.ShopSocial) {
		s.IsLiked = true
	})
}

func (c *userDatabase) UnlikeShop(ctx context.Context, userID uint, shopID uint) error {
	return c.DB.WithContext(ctx).
		Model(&domain.ShopSocial{}).
		Where("shop_id = ? AND user_id = ?", shopID, userID).
		Update("is_liked", false).Error
}

func (c *userDatabase) IsLikedShop(ctx context.Context, userID uint, shopID uint) (bool, error) {
	var count int64
	err := c.DB.WithContext(ctx).
		Model(&domain.ShopSocial{}).
		Where("shop_id = ? AND user_id = ? AND is_liked = ?", shopID, userID, true).
		Count(&count).Error
	return count > 0, err
}

func (c *userDatabase) GetLikedShops(ctx context.Context, userID uint) ([]response.Shop, error) {
	var shops []response.Shop
	query := `SELECT sd.id, sd.shop_name, sd.email, sd.phone, sd.address_line1, sd.address_line2,
		sd.city, sd.state, sd.country, sd.pincode, sd.shop_type, sd.shop_verification_status,
		sd.shop_image_url, sd.latitude, sd.longitude, sd.created_at, sd.updated_at
	FROM shop_details sd
	INNER JOIN shop_socials ss ON ss.shop_id = sd.id
	WHERE ss.user_id = ? AND ss.is_liked = TRUE
	ORDER BY ss.updated_at DESC`
	err := c.DB.WithContext(ctx).Raw(query, userID).Scan(&shops).Error
	return shops, err
}

func (c *userDatabase) RateShop(ctx context.Context, userID uint, shopID uint, rating uint) error {
	return c.upsertUserShopSocial(ctx, userID, shopID, func(s *domain.ShopSocial) {
		s.Rating = rating
	})
}

func (c *userDatabase) GetUserShopRating(ctx context.Context, userID uint, shopID uint) (uint, error) {
	var social domain.ShopSocial
	err := c.DB.WithContext(ctx).Where("shop_id = ? AND user_id = ?", shopID, userID).First(&social).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return social.Rating, nil
}

func (c *userDatabase) GetAllShopRatings(ctx context.Context, shopID uint) ([]domain.ShopSocial, error) {
	var ratings []domain.ShopSocial
	err := c.DB.WithContext(ctx).
		Where("shop_id = ? AND rating > 0", shopID).
		Order("updated_at DESC").
		Find(&ratings).Error
	return ratings, err
}

func (c *userDatabase) GetShopAverageRating(ctx context.Context, shopID uint) (float64, error) {
	var avg float64
	err := c.DB.WithContext(ctx).
		Model(&domain.ShopSocial{}).
		Where("shop_id = ? AND rating > 0", shopID).
		Select("COALESCE(AVG(rating::numeric), 0)").
		Scan(&avg).Error
	return avg, err
}

func (c *userDatabase) GetShopRatingDistribution(ctx context.Context, shopID uint) ([]domain.ShopRatingDistribution, error) {
	var distribution []domain.ShopRatingDistribution
	err := c.DB.WithContext(ctx).
		Model(&domain.ShopSocial{}).
		Select("rating, COUNT(*) as count").
		Where("shop_id = ? AND rating > 0", shopID).
		Group("rating").
		Order("rating DESC").
		Scan(&distribution).Error
	return distribution, err
}

func (c *userDatabase) ReviewShop(ctx context.Context, userID uint, shopID uint, review string) error {
	review = strings.TrimSpace(review)
	return c.upsertUserShopSocial(ctx, userID, shopID, func(s *domain.ShopSocial) {
		s.Review = review
	})
}

func (c *userDatabase) GetUserShopReview(ctx context.Context, userID uint, shopID uint) (string, error) {
	var social domain.ShopSocial
	err := c.DB.WithContext(ctx).Where("shop_id = ? AND user_id = ?", shopID, userID).First(&social).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return social.Review, nil
}

func (c *userDatabase) GetAllShopReviews(ctx context.Context, shopID uint) ([]domain.ShopSocial, error) {
	var reviews []domain.ShopSocial
	err := c.DB.WithContext(ctx).
		Raw("SELECT * FROM shop_socials WHERE shop_id = ? AND review IS NOT NULL AND review <> '' ORDER BY updated_at DESC", shopID).
		Scan(&reviews).Error
	return reviews, err
}

func (c *userDatabase) GetShopSocialDetails(ctx context.Context, shopID uint, userID uint) (domain.ShopSocialSummary, error) {
	summary := domain.ShopSocialSummary{ShopID: shopID}

	if err := c.DB.WithContext(ctx).Model(&domain.ShopSocial{}).Where("shop_id = ? AND is_follower = ?", shopID, true).Count(&summary.FollowerCount).Error; err != nil {
		return summary, err
	}
	if err := c.DB.WithContext(ctx).Model(&domain.ShopSocial{}).Where("user_id = ? AND is_follower = ?", userID, true).Count(&summary.FollowingCount).Error; err != nil {
		return summary, err
	}
	if err := c.DB.WithContext(ctx).Model(&domain.ShopSocial{}).Where("shop_id = ? AND is_liked = ?", shopID, true).Count(&summary.LikeCount).Error; err != nil {
		return summary, err
	}
	if err := c.DB.WithContext(ctx).Model(&domain.ShopSocial{}).Where("shop_id = ? AND rating > 0", shopID).Count(&summary.RatingCount).Error; err != nil {
		return summary, err
	}
	if err := c.DB.WithContext(ctx).Model(&domain.ShopSocial{}).Where("shop_id = ? AND review IS NOT NULL AND TRIM(review) <> ''", shopID).Count(&summary.ReviewCount).Error; err != nil {
		return summary, err
	}

	avg, err := c.GetShopAverageRating(ctx, shopID)
	if err != nil {
		return summary, err
	}
	summary.AverageRating = avg

	isFollowing, err := c.IsFollowingShop(ctx, userID, shopID)
	if err != nil {
		return summary, err
	}
	summary.IsFollowing = isFollowing

	isLiked, err := c.IsLikedShop(ctx, userID, shopID)
	if err != nil {
		return summary, err
	}
	summary.IsLiked = isLiked

	userRating, err := c.GetUserShopRating(ctx, userID, shopID)
	if err != nil {
		return summary, err
	}
	summary.UserRating = userRating

	userReview, err := c.GetUserShopReview(ctx, userID, shopID)
	if err != nil {
		return summary, err
	}
	summary.UserReview = userReview

	return summary, nil
}
