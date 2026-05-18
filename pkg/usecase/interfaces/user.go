package interfaces

import (
	"context"
	"mime/multipart"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type UserUseCase interface {
	FindProfile(ctx context.Context, userId uint) (domain.User, error)
	UpdateProfile(ctx context.Context, user domain.User) error

	// profile side
	// GetProductItemsByDepartment(ctx context.Context, documentID uint) ([]response.ProductItems, error)
	// //address side
	SaveAddress(ctx context.Context, userID uint, address domain.Address, isDefault bool) error // save address
	UpdateAddress(ctx context.Context, addressBody request.EditAddress, userID uint) error
	FindAddresses(ctx context.Context, userID uint) ([]response.Address, error) // to get all address of a user

	// wishlist
	SaveToWishList(ctx context.Context, wishList domain.WishList) error
	RemoveFromWishList(ctx context.Context, userID, productItemID uint) error
	FindAllWishListItems(ctx context.Context, userID uint) ([]response.WishListItem, error)
	UploadProfileImage(ctx context.Context, userID string, file *multipart.FileHeader, fileSize int64, fileName, contentType string) (string, error)
	GetSellersByRadius(ctx context.Context, reqData request.SellerRadiusRequest) (sellers []response.Shop, err error)
	GetSellersByPincode(ctx context.Context, reqData request.SellerPincodeRequest) (sellers []response.Shop, err error)
	SearchShopList(ctx context.Context, reqData request.SearchShopListRequest) (shops []response.Shop, err error)
	GetProductItemsByDepartment(ctx context.Context, documentID uint) ([]response.ProductItems, error)
	GetProductItemsByCategory(ctx context.Context, categoryID uint) ([]response.ProductItems, error)
	GetProductItemsBySubCategory(ctx context.Context, subCategoryID uint) ([]response.ProductItems, error)
	GetProductItemsByShop(ctx context.Context, adminID uint) ([]response.ProductItems, error)
	GetShopByID(ctx context.Context, shopID uint) (response.Shop, error)
	GetShopSocialDetails(ctx context.Context, shopID uint, userID uint) (domain.ShopSocialSummary, error)
	FollowShop(ctx context.Context, userID uint, shopID uint) error
	UnfollowShop(ctx context.Context, userID uint, shopID uint) error
	IsFollowingShop(ctx context.Context, userID uint, shopID uint) (bool, error)
	GetFollowers(ctx context.Context, shopID uint) ([]response.User, error)
	GetFollowedShops(ctx context.Context, userID uint) ([]response.Shop, error)
	LikeShop(ctx context.Context, userID uint, shopID uint) error
	UnlikeShop(ctx context.Context, userID uint, shopID uint) error
	IsLikedShop(ctx context.Context, userID uint, shopID uint) (bool, error)
	GetLikedShops(ctx context.Context, userID uint) ([]response.Shop, error)
	RateShop(ctx context.Context, userID uint, shopID uint, rating uint) error
	GetUserShopRating(ctx context.Context, userID uint, shopID uint) (uint, error)
	GetAllShopRatings(ctx context.Context, shopID uint) ([]domain.ShopSocial, error)
	GetShopAverageRating(ctx context.Context, shopID uint) (float64, error)
	GetShopRatingDistribution(ctx context.Context, shopID uint) ([]domain.ShopRatingDistribution, error)
	ReviewShop(ctx context.Context, userID uint, shopID uint, review string) error
	GetUserShopReview(ctx context.Context, userID uint, shopID uint) (string, error)
	GetAllShopReviews(ctx context.Context, shopID uint) ([]domain.ShopSocial, error)
}
