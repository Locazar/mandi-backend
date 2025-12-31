package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/jinzhu/copier"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"googlemaps.github.io/maps"
)

type userUserCase struct {
	userRepo      interfaces.UserRepository
	cartRepo      interfaces.CartRepository
	productRepo   interfaces.ProductRepository
	imageUploader cloud.CloudService
}

// UploadProfileImage implements interfaces.UserUseCase.

type S3ImageUploader struct {
	// add AWS S3 client config here
	uploader *manager.Uploader
	bucket   string
}

func NewUserUseCase(userRepo interfaces.UserRepository, cartRepo interfaces.CartRepository,
	productRepo interfaces.ProductRepository, imageUploader cloud.CloudService) service.UserUseCase {
	return &userUserCase{
		userRepo:      userRepo,
		cartRepo:      cartRepo,
		productRepo:   productRepo,
		imageUploader: imageUploader,
	}
}

func (c *userUserCase) FindProfile(ctx context.Context, userID uint) (domain.User, error) {

	user, err := c.userRepo.FindUserByUserID(ctx, userID)
	if err != nil {
		return domain.User{}, utils.PrependMessageToError(err, "failed to find user details")
	}

	return user, nil
}

func (c *userUserCase) UpdateProfile(ctx context.Context, user domain.User) error {

	// first check any other user exist with this entered unique fields
	checkUser, err := c.userRepo.FindUserByUserNameEmailOrPhoneNotID(ctx, user)
	if err != nil {
		return err
	}
	if checkUser.ID != 0 { // if there is an user exist with given details then make it as error
		err = utils.CompareUserExistingDetails(user, checkUser)
		fmt.Println(user)
		return err
	}

	// if user password given then hash the password
	if user.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
		if err != nil {
			return fmt.Errorf("failed to generate hash password for user")
		}
		user.Password = string(hash)
	}

	err = c.userRepo.UpdateUser(ctx, user)

	if err != nil {
		return err
	}

	return nil
}

// adddress
func (c *userUserCase) SaveAddress(ctx context.Context, userID uint, address domain.Address, isDefault bool) error {

	exist, err := c.userRepo.IsAddressAlreadyExistForUser(ctx, address, userID)
	if err != nil {
		return fmt.Errorf("failed to check address already exist \nerror:%v", err.Error())
	}
	if exist {
		return fmt.Errorf("given address already exist for user")
	}

	// //this address not exist then create it
	fmt.Printf("saving address for user id: %d\n", userID)
	fmt.Printf("address details: %+v\n", address)
	// check the country id is valid or not
	country, err := c.userRepo.FindCountryByID(ctx, address.CountryID)
	if err != nil {
		return err
	} else if country.ID == 0 {
		return errors.New("invalid country id")
	}

	// save the address on database
	addressID, err := c.userRepo.SaveAddress(ctx, address)
	if err != nil {
		return err
	}

	//creating a user address with this given value
	var userAddress = domain.UserAddress{
		UserID:    userID,
		AddressID: addressID,
		IsDefault: isDefault,
	}

	// then update the address with user
	err = c.userRepo.SaveUserAddress(ctx, userAddress)

	if err != nil {
		return err
	}

	return nil
}

func (c *userUserCase) UpdateAddress(ctx context.Context, addressBody request.EditAddress, userID uint) error {

	if exist, err := c.userRepo.IsAddressIDExist(ctx, addressBody.ID); err != nil {
		return err
	} else if !exist {
		return errors.New("invalid address id")
	}

	var address domain.Address
	copier.Copy(&address, &addressBody)

	if err := c.userRepo.UpdateAddress(ctx, address); err != nil {
		return err
	}

	// check the user address need to set default or not if it need then set it as default
	if addressBody.IsDefault != nil && *addressBody.IsDefault {
		userAddress := domain.UserAddress{
			UserID:    userID,
			AddressID: address.ID,
			IsDefault: *addressBody.IsDefault,
		}

		err := c.userRepo.UpdateUserAddress(ctx, userAddress)
		if err != nil {
			return err
		}
	}
	log.Printf("successfully address saved for user with user_id %v", userID)
	return nil
}

// get all address
func (c *userUserCase) FindAddresses(ctx context.Context, userID uint) ([]response.Address, error) {

	return c.userRepo.FindAllAddressByUserID(ctx, userID)
}

// to add new productItem to wishlist
func (c *userUserCase) SaveToWishList(ctx context.Context, wishList domain.WishList) error {

	// check the productItem already exist on wishlist for user
	checkWishList, err := c.userRepo.FindWishListItem(ctx, wishList.ProductItemID, wishList.UserID)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to check product item already exist on wish list")
	}
	if checkWishList.ID != 0 {
		return ErrExistWishListProductItem
	}

	err = c.userRepo.SaveWishListItem(ctx, wishList)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to save product item on wish list")
	}

	return nil
}

// remove from wishlist
func (c *userUserCase) RemoveFromWishList(ctx context.Context, userID, productItemID uint) error {

	err := c.userRepo.RemoveWishListItem(ctx, userID, productItemID)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to remove product item form wish list")
	}

	return nil
}

func (c *userUserCase) FindAllWishListItems(ctx context.Context, userID uint) ([]response.WishListItem, error) {

	wishListItems, err := c.userRepo.FindAllWishListItemsByUserID(ctx, userID)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to find wish list product items")
	}

	return wishListItems, nil
}

func (c *userUserCase) FindLocation(ctx context.Context, lat string, long string) {

	apiKey := "YOUR_API_KEY" // Replace with your actual API key
	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	// Example: Reverse Geocoding coordinates to an address
	r := &maps.GeocodingRequest{
		LatLng: &maps.LatLng{
			Lat: 34.052235,
			Lng: -118.243683,
		},
	}
	resp, err := client.Geocode(context.Background(), r)
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	if len(resp) > 0 {
		address := resp[0]
		fmt.Printf("Formatted Address: %s\n", address.FormattedAddress)
		for _, component := range address.AddressComponents {
			for _, t := range component.Types {
				switch t {
				case "locality":
					fmt.Printf("City: %s\n", component.LongName)
				case "administrative_area_level_1":
					fmt.Printf("State: %s\n", component.LongName)
				case "country":
					fmt.Printf("Country: %s\n", component.LongName)
				}
			}
		}
	} else {
		fmt.Println("No results found.")
	}
}

func (u *userUserCase) UploadProfileImage(ctx context.Context, userID string, fileHeader *multipart.FileHeader, imageSize int64, filename string, headerContent string) (string, error) {
	// Use the S3ImageUploader to upload the file
	image_path, err := u.imageUploader.SaveFile(ctx, fileHeader)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to save image on cloud storage")
	}
	return image_path, nil
}

func (c *userUserCase) GetSellersByRadius(ctx context.Context, reqData request.SellerRadiusRequest) (sellers []response.Shop, err error) {
	sellers, err = c.userRepo.FindSellersByRadius(ctx, reqData)

	if err != nil {
		return sellers, fmt.Errorf("failed to get sellers by radius \nerror:%v", err.Error())
	}

	log.Printf("successfully got sellers within %v km radius", reqData.RadiusKm)

	return sellers, nil
}

func (c *userUserCase) GetProductItemsByDepartment(ctx context.Context, documentID uint) ([]response.ProductItems, error) {
	return c.productRepo.GetProductItemsByDepartment(ctx, documentID)
}

func (c *userUserCase) GetProductItemsByCategory(ctx context.Context, categoryID uint) ([]response.ProductItems, error) {
	return c.productRepo.GetProductItemsByCategory(ctx, categoryID)
}

func (c *userUserCase) GetProductItemsBySubCategory(ctx context.Context, subCategoryID uint) ([]response.ProductItems, error) {
	return c.productRepo.GetProductItemsBySubCategory(ctx, subCategoryID)
}

func (c *userUserCase) GetProductItemsByShop(ctx context.Context, adminID uint) ([]response.ProductItems, error) {
	return c.productRepo.GetProductItemsByShop(ctx, adminID)
}
