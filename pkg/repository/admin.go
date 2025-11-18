package repository

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

// NotificationService abstracts sending notifications.
type NotificationService interface {
	SendNotification(userID uint, message string) error
}

// noopNotificationService is a placeholder implementation; replace with real logic.
type noopNotificationService struct{}

func (n *noopNotificationService) SendNotification(userID uint, message string) error {
	// TODO: integrate actual notification provider (email, SMS, push, etc.)
	return nil
}

// notificationService is a package-level variable used by repository methods.
var notificationService NotificationService = &noopNotificationService{}

type adminDatabase struct {
	DB *gorm.DB
}

func NewAdminRepository(DB *gorm.DB) interfaces.AdminRepository {
	return &adminDatabase{DB: DB}
}

func (c *adminDatabase) FindAdminByEmail(ctx context.Context, email string) (domain.Admin, error) {

	var admin domain.Admin
	err := c.DB.Raw("SELECT * FROM admins WHERE email = $1", email).Scan(&admin).Error

	return admin, err
}

func (c *adminDatabase) FindAdminByUserName(ctx context.Context, userName string) (domain.Admin, error) {

	var admin domain.Admin
	err := c.DB.Raw("SELECT * FROM admins WHERE user_name = $1", userName).Scan(&admin).Error

	return admin, err
}

func (c *adminDatabase) SaveAdmin(ctx context.Context, admin domain.Admin) error {
	tx := c.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	// First insert into admins table
	query := `INSERT INTO admins (user_name, email, mobile, password,
		address_line1, address_line2, city, state, country, pincode,
		bank_account_number, bank_ifsc, pan, aadhar, agree_to_terms,
		verified, status, latitude, longitude, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)`

	if err := tx.Exec(query, admin.UserName, admin.Email, admin.Mobile, admin.Password,
		admin.AddressLine1, admin.AddressLine2, admin.City, admin.State, admin.Country, admin.Pincode,
		admin.BankAccountNumber, admin.BankIFSC, admin.PAN, admin.Aadhar, admin.AgreeToTerms,
		admin.Verified, admin.Status, admin.Latitude, admin.Longitude, time.Now(), time.Now()).Error; err != nil {
		tx.Rollback()
		return err
	}

	// shopVerification := domain.ShopVerification{
	// 	AdminID:            admin.ID,
	// 	ShopID:             admin.ShopID,
	// 	VerificationStatus: "under_review",
	// 	Remarks:            "Shop registration under review",
	// }

	// queryShops := `INSERT INTO shop_verifications (shop_id, verification_status, remarks, created_at, updated_at)
	// 	VALUES ($1, $2, $3, $4, $5)`

	// if err := tx.Exec(queryShops, admin.ShopID, shopVerification.VerificationStatus, shopVerification.Remarks, time.Now(), time.Now()).Error; err != nil {
	// 	tx.Rollback()
	// 	return err
	// }

	// Commit transaction if both inserts succeed
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (c *adminDatabase) FindAllUser(ctx context.Context, pagination request.Pagination) (users []response.User, err error) {

	limit := pagination.Count
	offset := (pagination.PageNumber - 1) * limit

	query := `SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err = c.DB.Raw(query, limit, offset).Scan(&users).Error

	return users, err
}

// sales report from order // !add  product wise report
func (c *adminDatabase) CreateFullSalesReport(ctc context.Context, salesReq request.SalesReport) (salesReport []response.SalesReport, err error) {

	limit := salesReq.Pagination.Count
	offset := (salesReq.Pagination.PageNumber - 1) * limit

	query := `SELECT u.first_name, u.email,  so.id AS shop_order_id, so.user_id, so.order_date, 
	so.order_total_price, so.discount, os.status AS order_status, pm.payment_type FROM shop_orders so
	INNER JOIN order_statuses os ON so.order_status_id = os.id 
	INNER JOIN  payment_methods pm ON so.payment_method_id = pm.id 
	INNER JOIN users u ON so.user_id = u.id 
	WHERE order_date >= $1 AND order_date <= $2
	ORDER BY so.order_date LIMIT  $3 OFFSET $4`

	err = c.DB.Raw(query, salesReq.StartDate, salesReq.EndDate, limit, offset).Scan(&salesReport).Error

	return
}

// stock side
func (c *adminDatabase) FindStockBySKU(ctx context.Context, sku string) (stock response.Stock, err error) {
	query := `SELECT pi.sku, pi.qty_in_stock, pi.price, p.name AS product_name, vo.value AS variation_value  
	FROM product_items pi 
	INNER JOIN products p ON p.id = pi.product_id 
	INNER JOIN product_configurations pc ON pc.product_item_id = pi.id 
	INNER JOIN variation_options vo ON vo.id = pc.variation_option_id
	WHERE pi.sku = $1`

	err = c.DB.Raw(query, sku).Scan(&stock).Error

	return stock, err
}

func (c *adminDatabase) VerifyShop(ctx context.Context, shopVerification domain.ShopVerification) error {
	query := `UPDATE admins SET shop_verification_status = $1, updated_at = $2 WHERE id = $3`
	err := c.DB.Exec(query, shopVerification.VerificationStatus, time.Now(), shopVerification.ShopID).Error

	return err
}

func (c *adminDatabase) CreateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error) {
	query := `INSERT INTO advertisements (title, content, image_url, start_date, end_date, created_at, updated_at, created_by_admin, admin_id, area_targeted, pincode_targeted, latitude, longitude, distance_km, status, distance_km, priority)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING id`

	err := c.DB.Raw(query, ad.Title, ad.Content, ad.ImageURL, ad.StartDate, ad.EndDate, time.Now(), time.Now(),
		ad.CreatedByAdmin, ad.AdminID, ad.AreaTargeted, ad.PincodeTargeted, ad.Latitude, ad.Longitude,
		ad.DistanceKM, ad.Status, ad.Priority).Scan(&ad.ID).Error

	return ad, err
}

func (c *adminDatabase) GetAllAdvertisements(ctx context.Context, pagination request.Pagination) (ads []domain.Advertisement, err error) {
	limit := pagination.Count
	offset := (pagination.PageNumber - 1) * limit

	query := `SELECT * FROM advertisements ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err = c.DB.Raw(query, limit, offset).Scan(&ads).Error

	return ads, err
}

func (c *adminDatabase) UpdateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error) {
	query := `UPDATE advertisements SET title = $1, content = $2, image_url = $3, target_url = $4,
	start_date = $5, end_date = $6, updated_at = $7, area_targeted = $8, pincode_targeted = $9,
	latitude = $10, longitude = $11, distance_km = $12, status = $13, priority = $14 WHERE id = $15`

	err := c.DB.Exec(query, ad.Title, ad.Content, ad.ImageURL, ad.TargetURL, ad.StartDate, ad.EndDate,
		time.Now(), ad.AreaTargeted, ad.PincodeTargeted, ad.Latitude, ad.Longitude, ad.DistanceKM,
		ad.Status, ad.Priority, ad.ID).Error

	return ad, err
}

func (c *adminDatabase) DeleteAdvertisement(ctx context.Context, advertisementID string) error {
	query := `DELETE FROM advertisements WHERE id = $1`
	err := c.DB.Exec(query, advertisementID).Error

	return err
}

// Shop Details
func (c *adminDatabase) CreateShop(ctx context.Context, shop domain.ShopDetails) (domain.ShopDetails, error) {
	tx := c.DB.Begin()
	if tx.Error != nil {
		return shop, tx.Error
	}

	query := `INSERT INTO shop_details (owner_id, shop_name, gstin, shop_id, address_line1, address_line2, email, mobile,
	city, state, country, pincode, bank_account_number, shop_type, shop_status, bank_ifsc, pan, gstin, msme_registration_number, electricity_bill, itr_documents, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22) RETURNING id`

	err := tx.Exec(query, shop.OwnerID, shop.ShopName, shop.GSTIN, shop.ShopID, shop.AddressLine1,
		shop.AddressLine2, shop.Email, shop.Mobile, shop.City, shop.State, shop.Country, shop.Pincode,
		shop.BankAccountNumber, shop.ShopType, shop.ShopStatus, shop.BankIFSC, shop.PanNumber, shop.GSTIN, shop.MSMERegistrationNumber, shop.ElectricityBill, shop.ITRDocuments,
		time.Now(), time.Now()).Scan(&shop.ID).Error

	queryShops := `INSERT INTO shop_verifications (shop_id, admin_id, verification_status, remarks, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	if err := tx.Exec(queryShops, shop.ShopID, shop.OwnerID, shop.ShopVerificationStatus, shop.ShopVerificationRemarks, time.Now(), time.Now()).Error; err != nil {
		tx.Rollback()
		return shop, err
	}

	// Commit transaction if both inserts succeed
	if err := tx.Commit().Error; err != nil {
		return shop, err
	}

	return shop, err
}

func (c *adminDatabase) GetAllShops(ctx context.Context, pagination request.Pagination) (shops []domain.ShopDetails, err error) {
	limit := pagination.Count
	offset := (pagination.PageNumber - 1) * limit

	query := `SELECT * FROM shop_details ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err = c.DB.Raw(query, limit, offset).Scan(&shops).Error

	return shops, err
}

func (c *adminDatabase) GetShopByID(ctx context.Context, shopID uint) (shop domain.ShopDetails, err error) {
	query := `SELECT * FROM shop_details WHERE id = $1`
	err = c.DB.Raw(query, shopID).Scan(&shop).Error

	return shop, err
}

func (c *adminDatabase) UpdateShop(ctx context.Context, shop domain.ShopDetails) (domain.ShopDetails, error) {
	query := `UPDATE shop_details SET owner_id = $1, shop_name = $2, gstin = $3, shop_id = $4,
	address_line1 = $5, address_line2 = $6, city = $7, state = $8, country = $9, pincode = $10,
	bank_account_number = $11, bank_ifsc = $12, pan = $13, latitude = $14, longitude = $15, gstin = $16, MSME_registration_number = $17, electricity_bill = $18, itr_documents = $19, shop_verification_status = $20, shop_verification_remarks = $21, updated_at = $22 WHERE id = $23`

	err := c.DB.Exec(query, shop.OwnerID, shop.ShopName, shop.GSTIN, shop.ShopID,
		shop.AddressLine1, shop.AddressLine2, shop.City, shop.State, shop.Country, shop.Pincode,
		shop.BankAccountNumber, shop.BankIFSC, shop.PanNumber, shop.Latitude, shop.Longitude, shop.GSTIN, shop.MSMERegistrationNumber, shop.ElectricityBill, shop.ITRDocuments, shop.ShopVerificationStatus, shop.ShopVerificationRemarks, time.Now(), shop.ID).Error
	return shop, err
}

func (c *adminDatabase) GetShopByOwnerID(ctx context.Context, ownerID uint) (shop domain.ShopDetails, err error) {
	query := `SELECT * FROM shop_details WHERE owner_id = $1`
	err = c.DB.Raw(query, ownerID).Scan(&shop).Error

	return shop, err
}

func (c *adminDatabase) SendNotificationToUsersInRadius(ctx context.Context, requestData request.NotificationRadiusRequest) error {
	query := `SELECT id FROM users
	 WHERE earth_distance(ll_to_earth($1, $2), ll_to_earth(latitude, longitude)) <= $3`

	var userIDs []uint
	err := c.DB.Raw(query, requestData.Latitude, requestData.Longitude, requestData.RadiusM*1000).Scan(&userIDs).Error
	if err != nil {
		return err
	}

	// Here, you would integrate with your notification service to send notifications to the userIDs
	// For example:
	// for _, userID := range userIDs {
	//     err := notificationService.SendNotification(userID, requestData.Message)
	//     if err != nil {
	//         // Handle notification sending error
	//     }
	// }

	return nil
}

func (c *adminDatabase) SendNotificationToUser(ctx context.Context, userID uint, message string) error {
	// Here, you would integrate with your notification service to send a notification to the userID
	// For example:
	err := notificationService.SendNotification(userID, message)
	if err != nil {
		return err
	}
	return nil
}
