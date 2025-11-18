package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"golang.org/x/crypto/bcrypt"
)

type adminUseCase struct {
	adminRepo interfaces.AdminRepository
	userRepo  interfaces.UserRepository
}

func NewAdminUseCase(repo interfaces.AdminRepository, userRepo interfaces.UserRepository) service.AdminUseCase {

	return &adminUseCase{
		adminRepo: repo,
		userRepo:  userRepo,
	}
}

func (c *adminUseCase) SignUp(ctx context.Context, loginDetails domain.Admin) error {

	existAdmin, err := c.adminRepo.FindAdminByEmail(ctx, loginDetails.Email)
	if err != nil {
		return err
	} else if existAdmin.ID != 0 {
		return errors.New("can't save admin \nan admin already exist with this email")
	}

	existAdmin, err = c.adminRepo.FindAdminByUserName(ctx, loginDetails.UserName)
	if err != nil {
		return err
	} else if existAdmin.ID != 0 {
		return errors.New("can't save admin \nan admin already exist with this user_name")
	}

	// generate a hashed password for admin
	hashPass, err := bcrypt.GenerateFromPassword([]byte(loginDetails.Password), 10)

	if err != nil {
		return errors.New("failed to generate hashed password for admin")
	}
	// set the hashed password on the admin
	loginDetails.Password = string(hashPass)

	return c.adminRepo.SaveAdmin(ctx, loginDetails)
}

func (c *adminUseCase) FindAllUser(ctx context.Context, pagination request.Pagination) (users []response.User, err error) {

	users, err = c.adminRepo.FindAllUser(ctx, pagination)

	return users, err
}

// Block User
func (c *adminUseCase) BlockOrUnBlockUser(ctx context.Context, blockDetails request.BlockUser) error {

	userToBlock, err := c.userRepo.FindUserByUserID(ctx, blockDetails.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user \nerror:%w", err)
	}

	if userToBlock.BlockStatus == blockDetails.Block {
		return ErrSameBlockStatus
	}

	err = c.userRepo.UpdateBlockStatus(ctx, blockDetails.UserID, blockDetails.Block)
	if err != nil {
		return fmt.Errorf("failed to update user block status \nerror:%v", err.Error())
	}
	return nil
}

func (c *adminUseCase) GetFullSalesReport(ctx context.Context, requestData request.SalesReport) (salesReport []response.SalesReport, err error) {
	salesReport, err = c.adminRepo.CreateFullSalesReport(ctx, requestData)

	if err != nil {
		return salesReport, err
	}

	log.Printf("successfully got sales report from %v to %v of limit %v",
		requestData.StartDate, requestData.EndDate, requestData.Pagination.Count)

	return salesReport, nil
}

func (c *adminUseCase) VerifyShop(ctx context.Context, verify domain.ShopVerification) error {
	err := c.adminRepo.VerifyShop(ctx, verify)

	if err != nil {
		return fmt.Errorf("failed to update shop verification status \nerror:%v", err.Error())
	}
	return nil
}

func (c *adminUseCase) CreateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error) {
	createdAd, err := c.adminRepo.CreateAdvertisement(ctx, ad)
	if err != nil {
		return domain.Advertisement{}, fmt.Errorf("failed to create advertisement \nerror:%v", err.Error())
	}
	return createdAd, nil
}

func (c *adminUseCase) GetAllAdvertisements(ctx context.Context, pagination request.Pagination) (ads []domain.Advertisement, err error) {
	ads, err = c.adminRepo.GetAllAdvertisements(ctx, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get all advertisements \nerror:%v", err.Error())
	}
	return ads, nil
}

func (c *adminUseCase) UpdateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error) {
	updatedAd, err := c.adminRepo.UpdateAdvertisement(ctx, ad)
	if err != nil {
		return domain.Advertisement{}, fmt.Errorf("failed to update advertisement \nerror:%v", err.Error())
	}
	return updatedAd, nil
}

func (c *adminUseCase) DeleteAdvertisement(ctx context.Context, advertisementID string) error {
	err := c.adminRepo.DeleteAdvertisement(ctx, advertisementID)
	if err != nil {
		return fmt.Errorf("failed to delete advertisement \nerror:%v", err.Error())
	}
	return nil
}

func (c *adminUseCase) CreateShop(ctx context.Context, shop domain.ShopDetails) (domain.ShopDetails, error) {
	createdShop, err := c.adminRepo.CreateShop(ctx, shop)
	if err != nil {
		return domain.ShopDetails{}, fmt.Errorf("failed to create shop \nerror:%v", err.Error())
	}
	return createdShop, nil
}

func (c *adminUseCase) GetAllShops(ctx context.Context, pagination request.Pagination) (shops []domain.ShopDetails, err error) {
	shops, err = c.adminRepo.GetAllShops(ctx, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get all shops \nerror:%v", err.Error())
	}
	return shops, nil
}

func (c *adminUseCase) GetShopByID(ctx context.Context, shopID uint) (shop domain.ShopDetails, err error) {
	shop, err = c.adminRepo.GetShopByID(ctx, shopID)
	if err != nil {
		return domain.ShopDetails{}, fmt.Errorf("failed to get shop by id \nerror:%v", err.Error())
	}
	return shop, nil
}
func (c *adminUseCase) UpdateShop(ctx context.Context, shop domain.ShopDetails) (domain.ShopDetails, error) {
	updatedShop, err := c.adminRepo.UpdateShop(ctx, shop)
	if err != nil {
		return domain.ShopDetails{}, fmt.Errorf("failed to update shop \nerror:%v", err.Error())
	}
	return updatedShop, nil
}

func (c *adminUseCase) GetShopByOwnerID(ctx context.Context, ownerID uint) (shop domain.ShopDetails, err error) {
	shop, err = c.adminRepo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		return domain.ShopDetails{}, fmt.Errorf("failed to get shop by owner id \nerror:%v", err.Error())
	}
	return shop, nil
}

func (c *adminUseCase) SendNotificationToUsersInRadius(ctx context.Context, requestData request.NotificationRadiusRequest) error {
	err := c.adminRepo.SendNotificationToUsersInRadius(ctx, requestData)
	if err != nil {
		return fmt.Errorf("failed to send notification to users in radius \nerror:%v", err.Error())
	}
	return nil
}

func (c *adminUseCase) SendNotificationToUser(ctx context.Context, userID uint, message string) error {
	err := c.adminRepo.SendNotificationToUser(ctx, userID, message)
	if err != nil {
		return fmt.Errorf("failed to send notification to user \nerror:%v", err.Error())
	}
	return nil
}
