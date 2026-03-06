package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type AdminRepository interface {
	FindAdminByEmail(ctx context.Context, email string) (domain.Admin, error)
	FindAdminByPhone(ctx context.Context, userName string) (domain.Admin, error)
	FindAdminWithShopVerificationByPhone(ctx context.Context, phone string) (domain.Admin, domain.ShopVerification, error)
	SaveAdmin(ctx context.Context, admin domain.Admin) error

	FindAllUser(ctx context.Context, pagination request.Pagination) (users []response.User, err error)

	CreateFullSalesReport(ctc context.Context, reqData request.SalesReport) (salesReport []response.SalesReport, err error)

	//stock side
	FindStockBySKU(ctx context.Context, sku string) (stock response.Stock, err error)
	VerifyShop(ctx context.Context, shopVerification request.ShopVerification, adminId string, verificationStatus bool) error
	// Advertisement Management
	CreateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error)
	GetAllAdvertisements(ctx context.Context, pagination request.Pagination) (ads []domain.Advertisement, err error)
	UpdateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error)
	DeleteAdvertisement(ctx context.Context, advertisementID string) error

	//Shop Details
	CreateShop(ctx context.Context, shop domain.ShopDetails) (domain.ShopDetails, error)
	GetAllShops(ctx context.Context, pagination request.Pagination) (shops []domain.ShopDetails, err error)
	GetShopByID(ctx context.Context, shopID uint) (shop domain.ShopDetails, err error)
	UpdateShop(ctx context.Context, shop map[string]interface{}, shopId string) (map[string]interface{}, error)
	GetShopByOwnerID(ctx context.Context, ownerID uint) (shop domain.ShopDetails, err error)

	SendNotificationToUsersInRadius(ctx context.Context, requestData request.NotificationRadiusRequest) error
	SendNotificationToUser(ctx context.Context, userID uint, message string) error
	UploadAdminProfileImage(ctx context.Context, adminID string, imagePath string, shopId string) (string, error)
	UploadShopDocument(ctx context.Context, shopID uint, documentType string, documentValue string) error
	UploadAddress(ctx context.Context, adminId string, address request.AddressRequest) error
	UploadAdminDocumentOtpSend(ctx context.Context, adminID string, documentType string, documentValue string) error

	GetVerificationStatus(ctx context.Context, adminId string) (domain.Admin, domain.ShopVerification, error)
	GetShopProfileImageById(ctx context.Context, shopId string) (string, error)
	DeleteRefreshSessionByUserID(ctx context.Context, adminId string) error
	GetShopSocialDetails(ctx context.Context, shopID uint) ([]domain.ShopSocial, error)
}
