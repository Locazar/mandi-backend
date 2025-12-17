package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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

func (c *adminDatabase) FindAdminByPhone(ctx context.Context, phone string) (domain.Admin, error) {

	var admin domain.Admin
	err := c.DB.Raw("SELECT * FROM admins WHERE mobile = $1", phone).Scan(&admin).Error

	return admin, err
}

func (c *adminDatabase) FindAdminWithShopVerificationByPhone(ctx context.Context, phone string) (domain.Admin, domain.ShopVerification, error) {
	var admin domain.Admin
	var shopVerification domain.ShopVerification

	// First get admin data
	query := `SELECT a.id, a.full_name, a.email, a.password, a.address_line1, a.address_line2, 
		a.city, a.state, a.country, a.pincode, a.mobile, a.latitude, a.longitude,
		a.payment_status, a.payment_type, a.payment_date, a.start_date, a.expiry_date,
		a.bank_account_number, a.bank_ifsc, a.pan, a.aadhar, a.agree_to_terms,
		a.created_at, a.updated_at, a.verified_seller, a.status
	FROM admins a 
	WHERE a.mobile = $1`

	err := c.DB.Raw(query, phone).Scan(&admin).Error
	if err != nil {
		return admin, shopVerification, err
	}

	// Then get shop verification data
	shopQuery := `SELECT sv.id, sv.admin_id, sv.shop_id, sv.shop_name, sv.verification_status, 
		sv.remarks, sv.agent_id, sv.created_at, sv.updated_at
	FROM shop_verifications sv 
	WHERE sv.admin_id = $1`

	// Use string conversion of admin ID
	adminIDStr := fmt.Sprintf("%d", admin.ID)
	shopErr := c.DB.Raw(shopQuery, adminIDStr).Scan(&shopVerification).Error
	// Shop verification might not exist, so don't treat as error
	if shopErr != nil {
		fmt.Printf("Shop verification not found for admin %d: %v\n", admin.ID, shopErr)
	}

	fmt.Printf("FindAdminWithShopVerificationByPhone - Admin: %+v, Shop: %+v\n", admin, shopVerification)
	return admin, shopVerification, nil
}

func (c *adminDatabase) SaveAdmin(ctx context.Context, admin domain.Admin) error {
	tx := c.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	// First insert into admins table
	query := `INSERT INTO admins (full_name, email, mobile, password,
		address_line1, address_line2, city, state, country, pincode,
		bank_account_number, bank_ifsc, pan, aadhar, agree_to_terms,
		verified_seller, status, latitude, longitude, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21) RETURNING id`

	var adminID = uuid.NewString()
	err := tx.Raw(query, admin.FullName, admin.Email, admin.Mobile, admin.Password,
		admin.AddressLine1, admin.AddressLine2, admin.City, admin.State, admin.Country, admin.Pincode,
		admin.BankAccountNumber, admin.BankIFSC, admin.PAN, admin.Aadhar, admin.AgreeToTerms,
		admin.VerifiedSeller, admin.Status, admin.Latitude, admin.Longitude, time.Now(), time.Now()).Scan(&adminID).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	shopVerification := domain.ShopVerification{
		AdminID: adminID,
		Remarks: "Shop registration under review",
	}

	queryShops := `INSERT INTO shop_verifications (admin_id, verification_status, remarks, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	if err := tx.Exec(queryShops, adminID, shopVerification.VerificationStatus, shopVerification.Remarks, time.Now(), time.Now()).Error; err != nil {
		tx.Rollback()
		return err
	}

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
	query := `SELECT p.name AS product_name, vo.value AS variation_value  
	FROM product_items pi 
	INNER JOIN products p ON p.id = pi.product_id 
	INNER JOIN product_configurations pc ON pc.product_item_id = pi.id 
	INNER JOIN variation_options vo ON vo.id = pc.variation_option_id
	WHERE pi.sku = $1`

	err = c.DB.Raw(query, sku).Scan(&stock).Error

	return stock, err
}

func (c *adminDatabase) VerifyShop(ctx context.Context, shopVerification request.ShopVerification, adminId string, verificationStatus bool) error {
	// Get shop Id and shop name using admin Id and Insert the table firsttime and next time just update the status
	var verificationStatusValue bool
	query := `SELECT id, shop_name, document_type FROM shop_details WHERE admin_id = $1`
	var shopID *string
	var shopName *string
	var Document_Type *string
	err := c.DB.Raw(query, adminId).Scan(&struct {
		ShopID        *string `gorm:"column:id"`
		ShopName      *string `gorm:"column:shop_name"`
		Document_Type *string `gorm:"column:document_type"`
	}{shopID, shopName, Document_Type}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to fetch shop details for admin %s: %v", adminId, err)
	}

	fmt.Printf("Document_Type for admin %s: %v\n", adminId, Document_Type)

	// Check if Document_Type is nil before dereferencing
	if Document_Type != nil && *Document_Type != "manual" {
		verificationStatusValue = verificationStatus
	} else {
		// If Document_Type is nil or "manual", use the provided verificationStatus
		verificationStatusValue = verificationStatus
	}

	fmt.Printf("Fetched shop details for admin %s: shopID=%v, shopName=%v, verificationStatusValue=%v\n", adminId, shopID, shopName, verificationStatusValue)

	insertQuery := `INSERT INTO shop_details (admin_id, shop_verification_status, photo_shop_verification, business_doc_verification, identity_doc_verification, address_proof_verification, updated_at, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (admin_id) DO UPDATE SET
	shop_verification_status = EXCLUDED.shop_verification_status,
	photo_shop_verification = EXCLUDED.photo_shop_verification,
	business_doc_verification = EXCLUDED.business_doc_verification,
	identity_doc_verification = EXCLUDED.identity_doc_verification,
	address_proof_verification = EXCLUDED.address_proof_verification,
	updated_at = EXCLUDED.updated_at`
	err = c.DB.Exec(insertQuery, adminId, verificationStatusValue, shopVerification.Photo_Shop_Verification, shopVerification.Business_Doc_Verification, shopVerification.Identity_Doc_Verification, shopVerification.Address_Proof_Verification, time.Now(), time.Now()).Error
	if err != nil {
		return fmt.Errorf("failed to upsert shop details for admin %s: %v", adminId, err)
	}
	return nil

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

	query := `INSERT INTO shop_details (owner_id, shop_name, address_line1, address_line2, email, mobile,
	city, state, country, pincode, bank_account_number, shop_type, shop_status, bank_ifsc, pan, itr_documents, document_type, document_value, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22) RETURNING id`

	err := tx.Exec(query, shop.AdminID, shop.ShopName, shop.AddressLine1,
		shop.AddressLine2, shop.Email, shop.Phone, shop.City, shop.State, shop.Country, shop.Pincode,
		shop.BankAccountNumber, shop.ShopType, shop.ShopStatus, shop.BankIFSC, shop.PanNumber, shop.ITRDocuments, shop.Document_Type, shop.Document_Value,
		time.Now(), time.Now()).Scan(&shop.ID).Error

	queryShops := `INSERT INTO shop_verifications (shop_id, admin_id, verification_status, remarks, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	if err := tx.Exec(queryShops, shop.AdminID, shop.ShopVerificationStatus, shop.ShopVerificationRemarks, time.Now(), time.Now()).Error; err != nil {
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
	err = c.DB.Model(&domain.ShopDetails{}).Where("id = ?", shopID).First(&shop).Error
	return shop, err
}

func (c *adminDatabase) UpdateShop(ctx context.Context, shop map[string]interface{}, shopId string) (map[string]interface{}, error) {
	// Build dynamic SET clause
	setClauses := []string{}
	values := []interface{}{}
	paramCount := 1

	// Map API keys to DB column names and build SET clause
	for k, v := range shop {
		var columnName string

		print("---------------------", k, v)

		switch k {
		case "AdminID":
			columnName = "admin_id"
		case "ShopName":
			columnName = "shop_name"
		case "OwnerName":
			columnName = "owner_name"
		case "AddressLine1":
			columnName = "address_line1"
		case "AddressLine2":
			columnName = "address_line2"
		case "City":
			columnName = "city"
		case "State":
			columnName = "state"
		case "Country":
			columnName = "country"
		case "Pincode":
			columnName = "pincode"
		case "Email":
			columnName = "email"
		case "Phone":
			columnName = "mobile"
		case "BankAccountNumber":
			columnName = "bank_account_number"
		case "ShopType":
			columnName = "shop_type"
		case "ShopStatus":
			columnName = "shop_status"
		case "BankIFSC":
			columnName = "bank_ifsc"
		case "PanNumber":
			columnName = "pan"
		case "ITRDocuments":
			columnName = "itr_documents"
		case "Document_Type":
			columnName = "document_type"
		case "Document_Value":
			columnName = "document_value"
		default:
			columnName = k // fallback: use as-is
		}

		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", columnName, paramCount))
		values = append(values, v)
		paramCount++
	}

	// Add updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", paramCount))
	values = append(values, time.Now())
	paramCount++

	query := fmt.Sprintf("UPDATE shop_details SET %s WHERE id = %s",
		strings.Join(setClauses, ", "), shopId)

	fmt.Printf("-------------------------Executing query: %s\nWith values: %+v\n", query, values)

	result := c.DB.Exec(query, values...)
	if result.Error != nil {
		fmt.Printf("Error updating shop: %v\n", result.Error)
		return nil, result.Error
	}

	fmt.Printf("Successfully updated shop with ID %s, rows affected: %d\n", shopId, result.RowsAffected)

	return shop, nil
}

func (c *adminDatabase) GetShopByOwnerID(ctx context.Context, ownerID uint) (shop domain.ShopDetails, err error) {
	query := `SELECT * FROM shop_details WHERE admin_id = $1`
	err = c.DB.Raw(query, ownerID).Scan(&shop).Error

	fmt.Printf("GetShopByOwnerID - shop: %+v, err: %v\n", shop, err)
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

func (c *adminDatabase) UploadAdminProfileImage(ctx context.Context, adminID string, imagePath string, shopId string) (string, error) {
	var idToUpdate string
	if shopId != "" {
		idToUpdate = shopId
	} else {
		idToUpdate = adminID
	}

	query := `UPDATE shop_details SET shop_image_url = $1, updated_at = $2 WHERE id = $3`
	err := c.DB.Exec(query, imagePath, time.Now(), idToUpdate).Error
	return imagePath, err
}

func (c *adminDatabase) UploadShopDocument(ctx context.Context, shopID uint, documentType string, documentValue string) error {
	query := `UPDATE shop_details SET document_type = $1, document_value = $2, updated_at = $3 WHERE admin_id = $4`
	err := c.DB.Exec(query, documentType, documentValue, time.Now(), shopID).Error
	return err
}

func (c *adminDatabase) UploadAddress(ctx context.Context, adminId string, address request.AddressRequest) error {
	// Parse latitude and longitude from string to float64
	latitude, err := strconv.ParseFloat(address.Latitude, 64)
	if err != nil {
		return fmt.Errorf("invalid latitude format: %v", err)
	}

	longitude, err := strconv.ParseFloat(address.Longitude, 64)
	if err != nil {
		return fmt.Errorf("invalid longitude format: %v", err)
	}

	//insert or update address in shop_details table
	query := `INSERT INTO shop_details (admin_id, shop_name, owner_name, phone, address_line1, address_line2, city, state, pincode, latitude, longitude, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	ON CONFLICT (admin_id) DO UPDATE SET
		shop_name = EXCLUDED.shop_name,
		owner_name = EXCLUDED.owner_name,
		phone = EXCLUDED.phone,
		address_line1 = EXCLUDED.address_line1,
		address_line2 = EXCLUDED.address_line2,
		city = EXCLUDED.city,
		state = EXCLUDED.state,
		pincode = EXCLUDED.pincode,
		latitude = EXCLUDED.latitude,
		longitude = EXCLUDED.longitude,
		updated_at = EXCLUDED.updated_at`

	err = c.DB.Exec(query, adminId, address.ShopName, address.OwnerName, address.Phone, address.AddressLine1, address.AddressLine2, address.City, address.State, address.Pincode,
		latitude, longitude, time.Now(), time.Now()).Error

	return err
}

func (c *adminDatabase) UploadAdminDocumentOtpSend(ctx context.Context, adminID string, documentType string, documentValue string) error {
	// For simplicity, assuming OTP verification is done elsewhere
	var value string
	if documentType == "Pan" {
		value = documentType
	} else {
		value = documentType
	}

	// documentValue is a column name of admins table
	query := `UPDATE admins SET ` + value + ` = $1, updated_at = $2 WHERE id = $3`
	err := c.DB.Exec(query, value, time.Now(), adminID).Error
	return err
}

func (c *adminDatabase) GetVerificationStatus(ctx context.Context, adminId string) (domain.Admin, domain.ShopVerification, error) {
	var admin domain.Admin
	var shopVerification domain.ShopVerification

	// Get admin verification status
	adminQuery := `SELECT verified_seller FROM admins WHERE id = $1`
	err := c.DB.Raw(adminQuery, adminId).Scan(&admin).Error
	if err != nil {
		return admin, shopVerification, err
	}

	// Get shop details (shop_verification_status and document_type)
	var shopDetails struct {
		ShopVerificationStatus bool   `gorm:"column:shop_verification_status"`
		DocumentType           string `gorm:"column:document_type"`
	}
	shopDetailsQuery := `SELECT shop_verification_status, document_type FROM shop_details WHERE admin_id = $1`
	shopDetailsErr := c.DB.Raw(shopDetailsQuery, adminId).Scan(&shopDetails).Error
	if shopDetailsErr != nil && !errors.Is(shopDetailsErr, gorm.ErrRecordNotFound) {
		return admin, shopVerification, shopDetailsErr
	}

	// Get shop verification status
	shopQuery := `SELECT verification_status FROM shop_verifications WHERE admin_id = $1`
	err = c.DB.Raw(shopQuery, adminId).Scan(&shopVerification).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return admin, shopVerification, err
	}

	// Set the shop details values in the shopVerification struct if found
	if shopDetailsErr == nil {
		shopVerification.VerificationStatus = shopDetails.ShopVerificationStatus
	}

	return admin, shopVerification, nil
}

func (c *adminDatabase) GetShopProfileImageById(ctx context.Context, shopId string) (string, error) {
	var shopProfileImage string
	query := `SELECT shop_image_url FROM shop_details WHERE id = $1`
	err := c.DB.Raw(query, shopId).Scan(&shopProfileImage).Error
	if err != nil {
		return "", err
	}
	return shopProfileImage, nil
}
