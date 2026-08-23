package interfaces

import (
	"context"
	"mime/multipart"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type UserUseCase interface {
	FindProfile(ctx context.Context, userId string) (domain.User, error)
	UpdateProfile(ctx context.Context, user domain.User) error

	// profile side
	// GetProductItemsByDepartment(ctx context.Context, documentID string) ([]response.ProductItems, error)
	// //address side
	SaveAddress(ctx context.Context, userID string, address domain.Address, isDefault bool) error // save address
	UpdateAddress(ctx context.Context, addressBody request.EditAddress, userID string) error
	FindAddresses(ctx context.Context, userID string) ([]response.Address, error) // to get all address of a user

	// wishlist
	SaveToWishList(ctx context.Context, wishList domain.WishList) error
	RemoveFromWishList(ctx context.Context, userID, productItemID string) error
	FindAllWishListItems(ctx context.Context, userID string) ([]response.WishListItem, error)
	UploadProfileImage(ctx context.Context, userID string, file *multipart.FileHeader, fileSize int64, fileName, contentType string) (string, error)
	GetSellersByRadius(ctx context.Context, reqData request.SellerRadiusRequest) (sellers []response.Shop, err error)
	GetSellersByPincode(ctx context.Context, reqData request.SellerPincodeRequest) (sellers []response.Shop, err error)
	SearchShopList(ctx context.Context, reqData request.SearchShopListRequest) (shops []response.Shop, err error)
	GetProductItemsByDepartment(ctx context.Context, documentID string) ([]response.ProductItems, error)
	GetProductItemsByCategory(ctx context.Context, categoryID string) ([]response.ProductItems, error)
	GetProductItemsBySubCategory(ctx context.Context, subCategoryID string) ([]response.ProductItems, error)
	GetProductItemsByShop(ctx context.Context, adminID string) ([]response.ProductItems, error)
	GetShopByID(ctx context.Context, shopID string) (response.Shop, error)
	GetShopSocialDetails(ctx context.Context, shopID string, userID string) (domain.ShopSocialSummary, error)
	FollowShop(ctx context.Context, userID string, shopID string) error
	UnfollowShop(ctx context.Context, userID string, shopID string) error
	IsFollowingShop(ctx context.Context, userID string, shopID string) (bool, error)
	GetFollowers(ctx context.Context, shopID string) ([]response.User, error)
	GetFollowedShops(ctx context.Context, userID string) ([]response.Shop, error)
	LikeShop(ctx context.Context, userID string, shopID string) error
	UnlikeShop(ctx context.Context, userID string, shopID string) error
	IsLikedShop(ctx context.Context, userID string, shopID string) (bool, error)
	GetLikedShops(ctx context.Context, userID string) ([]response.Shop, error)
	RateShop(ctx context.Context, userID string, shopID string, rating uint) error
	GetUserShopRating(ctx context.Context, userID string, shopID string) (uint, error)
	GetAllShopRatings(ctx context.Context, shopID string) ([]domain.ShopSocial, error)
	GetShopAverageRating(ctx context.Context, shopID string) (float64, error)
	GetShopRatingDistribution(ctx context.Context, shopID string) ([]domain.ShopRatingDistribution, error)
	ReviewShop(ctx context.Context, userID string, shopID string, review string) error
	SaveShopFeedback(ctx context.Context, userID string, shopID string, rating *uint, review *string, comments *string) error
	GetUserShopReview(ctx context.Context, userID string, shopID string) (string, error)
	GetUserShopFeedback(ctx context.Context, userID string, shopID string) (domain.ShopSocial, bool, error)
	GetAllShopReviews(ctx context.Context, shopID string) ([]domain.ShopSocial, error)
	GetAllShopComments(ctx context.Context, shopID string) ([]domain.ShopSocial, error)
	GetShopFeedbackList(ctx context.Context, shopID string) ([]domain.ShopSocial, error)
}
