package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

//go:generate mockgen -destination=../../mock/mockrepo/user_mock.go -package=mockrepo . UserRepository
type UserRepository interface {
	FindUserByUserID(ctx context.Context, userID string) (user domain.User, err error)
	FindUserByEmail(ctx context.Context, email string) (user domain.User, err error)
	FindUserByUserName(ctx context.Context, userName string) (user domain.User, err error)
	FindUserByPhoneNumber(ctx context.Context, phoneNumber string) (user domain.User, err error)
	FindUserByUserNameEmailOrPhoneNotID(ctx context.Context, user domain.User) (domain.User, error)
	UpdateAdminVerified(ctx context.Context, adminID string) error
	SaveUser(ctx context.Context, user domain.User) (userID string, err error)
	UpdateVerified(ctx context.Context, userID string) error
	UpdateUser(ctx context.Context, user domain.User) (err error)
	UpdateBlockStatus(ctx context.Context, userID string, blockStatus bool) error

	//address
	FindCountryByID(ctx context.Context, countryID string) (domain.Country, error)
	FindAddressByID(ctx context.Context, addressID string) (response.Address, error)
	IsAddressIDExist(ctx context.Context, addressID string) (exist bool, err error)
	IsAddressAlreadyExistForUser(ctx context.Context, address domain.Address, userID string) (bool, error)
	FindAllAddressByUserID(ctx context.Context, userID string) ([]response.Address, error)
	SaveAddress(ctx context.Context, address domain.Address) (addressID string, err error)
	UpdateAddress(ctx context.Context, address domain.Address) error
	// address join table
	SaveUserAddress(ctx context.Context, userAdress domain.UserAddress) error
	UpdateUserAddress(ctx context.Context, userAddress domain.UserAddress) error

	//wishlist
	FindWishListItem(ctx context.Context, productID, userID string) (domain.WishList, error)
	FindAllWishListItemsByUserID(ctx context.Context, userID string) ([]response.WishListItem, error)
	SaveWishListItem(ctx context.Context, wishList domain.WishList) error
	RemoveWishListItem(ctx context.Context, userID, productItemID string) error
	FindSellersByRadius(ctx context.Context, reqData request.SellerRadiusRequest) (sellers []response.Shop, err error)
	FindSellersByPincode(ctx context.Context, reqData request.SellerPincodeRequest) (sellers []response.Shop, err error)
	SearchShopList(ctx context.Context, reqData request.SearchShopListRequest) (shops []response.Shop, err error)
	PublicListShops(ctx context.Context, city, category string, limit, offset int) (shops []response.PublicShop, err error)
	PublicListProducts(ctx context.Context, city, category, shopID string, limit, offset int) (products []response.PublicProduct, err error)
	DeleteRefreshSessionByUserID(ctx context.Context, adminID string, userType string) error
	FindShopByID(ctx context.Context, shopID string) (response.Shop, error)
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
	// GetShopSocialDetails(ctx context.Context, shopID string) ([]domain.ShopSocial, error)
	UpdateTrialUsed(ctx context.Context, userID string) error
	IsTrialUsed(ctx context.Context, userID string) (bool, error)
}
