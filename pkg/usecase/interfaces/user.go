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

	//address side
	SaveAddress(ctx context.Context, userID uint, address domain.Address, isDefault bool) error // save address
	UpdateAddress(ctx context.Context, addressBody request.EditAddress, userID uint) error
	FindAddresses(ctx context.Context, userID uint) ([]response.Address, error) // to get all address of a user

	// wishlist
	SaveToWishList(ctx context.Context, wishList domain.WishList) error
	RemoveFromWishList(ctx context.Context, userID, productItemID uint) error
	FindAllWishListItems(ctx context.Context, userID uint) ([]response.WishListItem, error)
	UploadProfileImage(ctx context.Context, userID string, file *multipart.FileHeader, fileSize int64, fileName, contentType string) (string, error)
}
