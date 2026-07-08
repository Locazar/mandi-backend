package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

func parseShopID(ctx *gin.Context) (string, error) {
	shopID := ctx.Param("shop_id")
	if shopID == "" {
		return "", fmt.Errorf("missing shop_id parameter")
	}
	return shopID, nil
}

func parseOptionalUserID(ctx *gin.Context) (string, error) {
	idStr := ctx.Param("id")
	if idStr == "" {
		return utils.GetUserIdFromContext(ctx), nil
	}
	return idStr, nil
}

// FollowShop godoc
//
//	@Summary		Follow a shop (User)
//	@Security		BearerAuth
//	@Id				FollowShop
//	@Tags			User Social
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/social/follow/shop/{shop_id} [post]
//	@Success		200	{object}	response.Response{}	"Successfully followed shop"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to follow shop"
func (c *UserHandler) FollowShop(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	userID := utils.GetUserIdFromContext(ctx)
	if err := c.userUseCase.FollowShop(ctx, userID, shopID); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to follow shop", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully followed shop", nil)
}

// UnfollowShop godoc
//
//	@Summary		Unfollow a shop (User)
//	@Security		BearerAuth
//	@Id				UnfollowShop
//	@Tags			User Social
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/social/follow/shop/{shop_id} [delete]
//	@Success		200	{object}	response.Response{}	"Successfully unfollowed shop"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to unfollow shop"
func (c *UserHandler) UnfollowShop(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	userID := utils.GetUserIdFromContext(ctx)
	if err := c.userUseCase.UnfollowShop(ctx, userID, shopID); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to unfollow shop", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully unfollowed shop", nil)
}

// IsFollowingShop godoc
//
//	@Summary		Check if user follows a shop (User)
//	@Security		BearerAuth
//	@Id				IsFollowingShop
//	@Tags			User Social
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/social/follow/shop/{shop_id}/is-following [get]
//	@Success		200	{object}	response.Response{}	"Successfully fetched follow status"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get follow status"
func (c *UserHandler) IsFollowingShop(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	userID := utils.GetUserIdFromContext(ctx)
	isFollowing, err := c.userUseCase.IsFollowingShop(ctx, userID, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get follow status", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully fetched follow status", gin.H{"is_following": isFollowing})
}

// GetFollowers godoc
//
//	@Summary		Get followers of a shop (User)
//	@Security		BearerAuth
//	@Id				GetFollowers
//	@Tags			User Social
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/social/follow/shop/{shop_id} [get]
//	@Success		200	{object}	response.Response{}	"Successfully got followers"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get followers"
func (c *UserHandler) GetFollowers(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}

	followers, err := c.userUseCase.GetFollowers(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get followers", err, nil)
		return
	}

	if followers == nil {
		followers = []response.User{}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully got followers", followers)
}

// GetFollowedShops godoc
//
//	@Summary		Get followed shops by user (User)
//	@Security		BearerAuth
//	@Id				GetFollowedShops
//	@Tags			User Social
//	@Param			id	path	int	true	"User ID"
//	@Router			/social/follow/{id} [get]
//	@Success		200	{object}	response.Response{}	"Successfully got followed shops"
//	@Failure		400	{object}	response.Response{}	"Invalid user ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get followed shops"
func (c *UserHandler) GetFollowedShops(ctx *gin.Context) {
	userID, err := parseOptionalUserID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid user ID", err, nil)
		return
	}
	shops, err := c.userUseCase.GetFollowedShops(ctx, userID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get followed shops", err, nil)
		return
	}
	if shops == nil {
		shops = []response.Shop{}
	}
	response.ResolveShopsImages(c.cloudService, shops)
	response.SuccessResponse(ctx, http.StatusOK, "Successfully got followed shops", shops)
}

// GetMyFollowedShops godoc
//
//	@Summary		Get followed shops of logged-in user (User)
//	@Security		BearerAuth
//	@Id				GetMyFollowedShops
//	@Tags			User Social
//	@Router			/social/follow/my-shops [get]
//	@Success		200	{object}	response.Response{}	"Successfully got followed shops"
//	@Failure		401	{object}	response.Response{}	"Unauthorized"
//	@Failure		500	{object}	response.Response{}	"Failed to get followed shops"
func (c *UserHandler) GetMyFollowedShops(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	if userID == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Unauthorized", errors.New("user id not found in token"), nil)
		return
	}

	shops, err := c.userUseCase.GetFollowedShops(ctx, userID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get followed shops", err, nil)
		return
	}
	if shops == nil {
		shops = []response.Shop{}
	}

	response.ResolveShopsImages(c.cloudService, shops)

	response.SuccessResponse(ctx, http.StatusOK, "Successfully got followed shops", shops)
}

// RateShop godoc
//
//	@Summary		Create or update user rating for a shop (User)
//	@Security		BearerAuth
//	@Id				RateShop
//	@Tags			User Rating
//	@Accept			json
//	@Param			shop_id	path	int					true	"Shop ID"
//	@Param			input	body	request.ShopRatingRequest	true	"Rating payload"
//	@Router			/rating/shop/{shop_id} [post]
//	@Success		200	{object}	response.Response{}	"Successfully saved rating"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Failed to save rating"
func (c *UserHandler) RateShop(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	var body request.ShopRatingRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}
	userID := utils.GetUserIdFromContext(ctx)
	if err := c.userUseCase.RateShop(ctx, userID, shopID, body.Rating); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to save rating", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully saved rating", nil)
}

// GetUserShopRating godoc
//
//	@Summary		Get current user rating for a shop (User)
//	@Security		BearerAuth
//	@Id				GetUserShopRating
//	@Tags			User Rating
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/rating/shop/{shop_id}/user-rating [get]
//	@Success		200	{object}	response.Response{}	"Successfully got user rating"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get user rating"
func (c *UserHandler) GetUserShopRating(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	ratings, err := c.userUseCase.GetAllShopRatings(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get shop ratings", err, nil)
		return
	}
	averageRating, err := c.userUseCase.GetShopAverageRating(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get average rating", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully got shop ratings", gin.H{
		"shop_id":        shopID,
		"average_rating": averageRating,
		"ratings":        ratings,
	})
}

// GetAllShopRatings godoc
//
//	@Summary		Get all ratings for a shop (User)
//	@Security		BearerAuth
//	@Id				GetAllShopRatings
//	@Tags			User Rating
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/rating/shop/{shop_id}/ratings [get]
//	@Success		200	{object}	response.Response{}	"Successfully got shop ratings"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get ratings"
func (c *UserHandler) GetAllShopRatings(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	ratings, err := c.userUseCase.GetAllShopRatings(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get ratings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully got shop ratings", ratings)
}

// GetShopAverageRating godoc
//
//	@Summary		Get average rating of a shop (User)
//	@Security		BearerAuth
//	@Id				GetShopAverageRating
//	@Tags			User Rating
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/rating/shop/{shop_id}/average-rating [get]
//	@Success		200	{object}	response.Response{}	"Successfully got average rating"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get average rating"
func (c *UserHandler) GetShopAverageRating(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	avg, err := c.userUseCase.GetShopAverageRating(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get average rating", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully got average rating", gin.H{"average_rating": avg})
}

// GetShopRatingDistribution godoc
//
//	@Summary		Get rating distribution of a shop (User)
//	@Security		BearerAuth
//	@Id				GetShopRatingDistribution
//	@Tags			User Rating
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/rating/shop/{shop_id}/rating-distribution [get]
//	@Success		200	{object}	response.Response{}	"Successfully got rating distribution"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get rating distribution"
func (c *UserHandler) GetShopRatingDistribution(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	distribution, err := c.userUseCase.GetShopRatingDistribution(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get rating distribution", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully got rating distribution", distribution)
}

// ReviewShop godoc
//
//	@Summary		Create or update user review for a shop (User)
//	@Security		BearerAuth
//	@Id				ReviewShop
//	@Tags			User Review
//	@Accept			json
//	@Param			shop_id	path	int					true	"Shop ID"
//	@Param			input	body	request.ShopReviewRequest	true	"Review payload"
//	@Router			/review/shop/{shop_id} [post]
//	@Success		200	{object}	response.Response{}	"Successfully saved review"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Failed to save review"
func (c *UserHandler) ReviewShop(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	var body request.ShopReviewRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}
	userID := utils.GetUserIdFromContext(ctx)
	if err := c.userUseCase.ReviewShop(ctx, userID, shopID, body.Review); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to save review", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully saved review", nil)
}

// GetUserShopReview godoc
//
//	@Summary		Get all shop reviews with user_id (User)
//	@Security		BearerAuth
//	@Id				GetUserShopReview
//	@Tags			User Review
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/review/shop/{shop_id} [get]
//	@Success		200	{object}	response.Response{}	"Successfully got shop reviews"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get reviews"
func (c *UserHandler) GetUserShopReview(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	reviews, err := c.userUseCase.GetAllShopReviews(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get reviews", err, nil)
		return
	}
	if reviews == nil {
		reviews = []domain.ShopSocial{}
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully got shop reviews", gin.H{
		"shop_id":     shopID,
		"total_count": len(reviews),
		"reviews":     reviews,
	})
}

// GetAllShopReviews godoc
//
//	@Summary		Get all reviews for a shop (User)
//	@Security		BearerAuth
//	@Id				GetAllShopReviews
//	@Tags			User Review
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/review/shop/{shop_id}/reviews [get]
//	@Success		200	{object}	response.Response{}	"Successfully got shop reviews"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get reviews"
func (c *UserHandler) GetAllShopReviews(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	reviews, err := c.userUseCase.GetAllShopReviews(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get reviews", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully got shop reviews", reviews)
}

// LikeShop godoc
//
//	@Summary		Like a shop (User)
//	@Security		BearerAuth
//	@Id				LikeShop
//	@Tags			User Like
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/like/shop/{shop_id} [post]
//	@Success		200	{object}	response.Response{}	"Successfully liked shop"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to like shop"
func (c *UserHandler) LikeShop(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	userID := utils.GetUserIdFromContext(ctx)
	if err := c.userUseCase.LikeShop(ctx, userID, shopID); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to like shop", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully liked shop", nil)
}

// UnlikeShop godoc
//
//	@Summary		Unlike a shop (User)
//	@Security		BearerAuth
//	@Id				UnlikeShop
//	@Tags			User Like
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/like/shop/{shop_id} [delete]
//	@Success		200	{object}	response.Response{}	"Successfully unliked shop"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to unlike shop"
func (c *UserHandler) UnlikeShop(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	userID := utils.GetUserIdFromContext(ctx)
	if err := c.userUseCase.UnlikeShop(ctx, userID, shopID); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to unlike shop", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully unliked shop", nil)
}

// IsLikedShop godoc
//
//	@Summary		Check if user liked a shop (User)
//	@Security		BearerAuth
//	@Id				IsLikedShop
//	@Tags			User Like
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/like/shop/{shop_id}/is-liked [get]
//	@Success		200	{object}	response.Response{}	"Successfully fetched like status"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get like status"
func (c *UserHandler) IsLikedShop(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}
	userID := utils.GetUserIdFromContext(ctx)
	isLiked, err := c.userUseCase.IsLikedShop(ctx, userID, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get like status", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully fetched like status", gin.H{"is_liked": isLiked})
}

// GetLikedShops godoc
//
//	@Summary		Get liked shops by user (User)
//	@Security		BearerAuth
//	@Id				GetLikedShops
//	@Tags			User Like
//	@Param			id	path	int	true	"User ID"
//	@Router			/like/{id} [get]
//	@Success		200	{object}	response.Response{}	"Successfully got liked shops"
//	@Failure		400	{object}	response.Response{}	"Invalid user ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get liked shops"
func (c *UserHandler) GetLikedShops(ctx *gin.Context) {
	userID, err := parseOptionalUserID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid user ID", err, nil)
		return
	}
	shops, err := c.userUseCase.GetLikedShops(ctx, userID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get liked shops", err, nil)
		return
	}
	if shops == nil {
		shops = []response.Shop{}
	}
	response.ResolveShopsImages(c.cloudService, shops)
	response.SuccessResponse(ctx, http.StatusOK, "Successfully got liked shops", shops)
}

// SaveShopRatingAndReview godoc
//
//	@Summary		Create or update shop rating and review in a single API call (User)
//	@Security		BearerAuth
//	@Id				SaveShopRatingAndReview
//	@Tags			User Feedback
//	@Accept			json
//	@Param			shop_id	path	int										true	"Shop ID"
//	@Param			input	body	request.ShopRatingAndReviewRequest	true	"Rating and Review payload (all fields optional)"
//	@Router			/social/feedback/shop/{shop_id} [post]
//	@Success		200	{object}	response.Response{}	"Successfully saved rating and review"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Failed to save rating and review"
func (c *UserHandler) SaveShopRatingAndReview(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}

	var body request.ShopRatingAndReviewRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}

	// Validate that at least one field is provided
	if body.Rating == nil && body.Review == nil && body.Comments == nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "At least one of rating, review, or comments must be provided", errors.New("no data provided"), nil)
		return
	}

	userID := utils.GetUserIdFromContext(ctx)

	// Get customer ID from request or use user ID
	customerID := userID
	if body.CustomerID != nil && *body.CustomerID != "" {
		customerID = *body.CustomerID
	}

	if err := c.userUseCase.SaveShopFeedback(ctx, customerID, shopID, body.Rating, body.Review, body.Comments); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to save feedback", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully saved feedback", gin.H{
		"shop_id":     shopID,
		"customer_id": customerID,
		"rating":      body.Rating,
		"review":      body.Review,
		"comments":    body.Comments,
	})
}

// GetShopFeedback godoc
//
//	@Summary		Get combined feedback (rating and review) for a shop (User)
//	@Security		BearerAuth
//	@Id				GetShopFeedback
//	@Tags			User Feedback
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/social/feedback/shop/{shop_id} [get]
//	@Success		200	{object}	response.Response{}	"Successfully got shop feedback"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get feedback"
func (c *UserHandler) GetShopFeedback(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}

	// Get ratings
	ratings, err := c.userUseCase.GetAllShopRatings(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get ratings", err, nil)
		return
	}

	// Get reviews
	reviews, err := c.userUseCase.GetAllShopReviews(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get reviews", err, nil)
		return
	}

	// Get comments
	comments, err := c.userUseCase.GetAllShopComments(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get comments", err, nil)
		return
	}

	if ratings == nil {
		ratings = []domain.ShopSocial{}
	}
	if reviews == nil {
		reviews = []domain.ShopSocial{}
	}
	if comments == nil {
		comments = []domain.ShopSocial{}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully got shop feedback", gin.H{
		"shop_id":  shopID,
		"ratings":  ratings,
		"reviews":  reviews,
		"comments": comments,
	})
}

// GetShopFeedbackSummary godoc
//
//	@Summary		Get feedback summary for a shop (average rating and count) (User)
//	@Security		BearerAuth
//	@Id				GetShopFeedbackSummary
//	@Tags			User Feedback
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/social/feedback/shop/{shop_id}/summary [get]
//	@Success		200	{object}	response.Response{}	"Successfully got feedback summary"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get feedback summary"
func (c *UserHandler) GetShopFeedbackSummary(ctx *gin.Context) {
	shopID, err := parseShopID(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}

	// Get average rating
	avgRating, err := c.userUseCase.GetShopAverageRating(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get average rating", err, nil)
		return
	}

	// Get rating distribution
	distribution, err := c.userUseCase.GetShopRatingDistribution(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get rating distribution", err, nil)
		return
	}

	// Get reviews count
	reviews, err := c.userUseCase.GetAllShopReviews(ctx, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get reviews", err, nil)
		return
	}

	reviewCount := 0
	if reviews != nil {
		reviewCount = len(reviews)
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully got feedback summary", gin.H{
		"shop_id":             shopID,
		"average_rating":      avgRating,
		"rating_distribution": distribution,
		"total_reviews":       reviewCount,
	})
}
