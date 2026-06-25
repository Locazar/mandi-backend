package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/crypto"
	"github.com/rohit221990/mandi-backend/pkg/utils"
	"gorm.io/gorm"
)

// NotificationService abstracts sending notifications.
type NotificationService interface {
	SendNotification(userID string, message string) error
}

// noopNotificationService is a placeholder implementation; replace with real logic.
type noopNotificationService struct{}

func (n *noopNotificationService) SendNotification(userID string, message string) error {
	// TODO: integrate actual notification provider (email, SMS, push, etc.)
	return nil
}

// notificationService is a package-level variable used by repository methods.
var notificationService NotificationService = &noopNotificationService{}

type adminDatabase struct {
	DB  *gorm.DB
	enc *crypto.Service
}

func NewAdminRepository(DB *gorm.DB, enc *crypto.Service) interfaces.AdminRepository {
	return &adminDatabase{DB: DB, enc: enc}
}

// shopPIIKeys are the UpdateShop map keys whose values must be encrypted at rest.
var shopPIIKeys = map[string]bool{
	"BankAccountNumber":    true,
	"BankIFSC":             true,
	"PanNumber":            true,
	"ITRDocuments":         true,
	"Document_Value":       true,
	"ShopVerificationDocs": true,
}

// encrypt seals a non-empty PII value for storage. Empty values pass through.
func (c *adminDatabase) encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	return c.enc.Encrypt(plain)
}

// decrypt opens a stored PII value. Empty values pass through. Values that are
// not in the keyed-ciphertext format (e.g. legacy plaintext) are returned
// as-is so reads remain resilient during rollout.
func (c *adminDatabase) decrypt(stored string) string {
	if stored == "" {
		return ""
	}
	if plain, err := c.enc.Decrypt(stored); err == nil {
		return plain
	}
	return stored
}

// decryptAdminPII decrypts the encrypted-at-rest fields of an admin in place.
func (c *adminDatabase) decryptAdminPII(admin *domain.Admin) {
	admin.BankAccountNumber = c.decrypt(admin.BankAccountNumber)
	admin.BankIFSC = c.decrypt(admin.BankIFSC)
	admin.PAN = c.decrypt(admin.PAN)
}

// decryptShopPII decrypts the encrypted-at-rest fields of a shop in place.
func (c *adminDatabase) decryptShopPII(shop *domain.ShopDetails) {
	shop.BankAccountNumber = c.decrypt(shop.BankAccountNumber)
	shop.BankIFSC = c.decrypt(shop.BankIFSC)
	shop.PanNumber = c.decrypt(shop.PanNumber)
	shop.ITRDocuments = c.decrypt(shop.ITRDocuments)
	shop.Document_Value = c.decrypt(shop.Document_Value)
	shop.ShopVerificationDocs = c.decrypt(shop.ShopVerificationDocs)
}

func (c *adminDatabase) FindAdminByEmail(ctx context.Context, email string) (domain.Admin, error) {

	var admin domain.Admin
	err := c.DB.Raw("SELECT * FROM admins WHERE email = $1", email).Scan(&admin).Error
	if err != nil {
		return admin, err
	}
	if admin.ID == "" {
		return admin, gorm.ErrRecordNotFound
	}
	c.decryptAdminPII(&admin)
	return admin, nil
}

func (c *adminDatabase) FindAdminByPhone(ctx context.Context, phone string) (domain.Admin, error) {
	fmt.Printf("Finding admin by phone number: %s\n", phone)
	var admin domain.Admin
	err := c.DB.Raw("SELECT * FROM admins WHERE mobile = $1", phone).Scan(&admin).Error
	if err != nil {
		return admin, err
	}
	if admin.ID == "" {
		return admin, gorm.ErrRecordNotFound
	}
	c.decryptAdminPII(&admin)
	return admin, nil
}

func (c *adminDatabase) FindAdminWithShopVerificationByPhone(ctx context.Context, phone string) (domain.Admin, domain.ShopVerification, error) {
	var admin domain.Admin
	var shopVerification domain.ShopVerification

	// First get admin data
	query := `SELECT a.id, a.full_name, a.email, a.password, a.address_line1, a.address_line2, 
		a.city, a.state, a.country, a.pincode, a.mobile, a.latitude, a.longitude,
		a.payment_status, a.payment_type, a.payment_date, a.start_date, a.expiry_date,
		a.bank_account_number, a.bank_ifsc, a.pan, a.aadhaar_last4, a.aadhaar_verified, a.agree_to_terms,
		a.created_at, a.updated_at, a.verified_seller, a.status
	FROM admins a
	WHERE a.mobile = $1`

	err := c.DB.Raw(query, phone).Scan(&admin).Error
	if err != nil {
		return admin, shopVerification, err
	}
	if admin.ID == "" {
		return admin, shopVerification, gorm.ErrRecordNotFound
	}
	c.decryptAdminPII(&admin)
	// Then get shop verification data
	shopQuery := `SELECT sv.id, sv.admin_id, sv.shop_id, sv.shop_name, sv.verification_status,
		sv.remarks, sv.agent_id, sv.created_at, sv.updated_at
	FROM shop_verifications sv
	WHERE sv.admin_id = $1`

	// admin.ID is already a string typed-prefix ID
	adminIDStr := admin.ID
	shopErr := c.DB.Raw(shopQuery, adminIDStr).Scan(&shopVerification).Error
	// Shop verification might not exist, so don't treat as error
	if shopErr != nil {
	}

	return admin, shopVerification, nil
}

func (c *adminDatabase) SaveAdmin(ctx context.Context, admin domain.Admin) (domain.Admin, error) {
	admin.AdminID = utils.GenerateAdminID()
	if admin.UserName == "" {
		admin.UserName = utils.GenerateRandomUserName("seller")
	}
	encBankAccount, err := c.encrypt(admin.BankAccountNumber)
	if err != nil {
		return domain.Admin{}, err
	}
	encBankIFSC, err := c.encrypt(admin.BankIFSC)
	if err != nil {
		return domain.Admin{}, err
	}
	encPAN, err := c.encrypt(admin.PAN)
	if err != nil {
		return domain.Admin{}, err
	}

	tx := c.DB.Begin()
	if tx.Error != nil {
		return domain.Admin{}, tx.Error
	}
	// Generate typed-prefix ID before INSERT (raw SQL bypasses BeforeCreate hook).
	if admin.ID == "" {
		admin.ID = domain.NewID(domain.PrefixAdmin)
	}

	// First insert into admins table
	query := `INSERT INTO admins (id, full_name, email, mobile, password, user_name,
		address_line1, address_line2, city, state, country, pincode,
		bank_account_number, bank_ifsc, pan, aadhaar_last4, aadhaar_verified,
		verified_seller, status, latitude, longitude, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
		$12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`

	err = tx.Exec(query, admin.ID, admin.FullName, admin.Email, admin.Mobile, admin.Password, admin.UserName,
		admin.AddressLine1, admin.AddressLine2, admin.City, admin.State, admin.Country, admin.Pincode,
		encBankAccount, encBankIFSC, encPAN, admin.AadhaarLast4, admin.AadhaarVerified,
		admin.VerifiedSeller, admin.Status, admin.Latitude, admin.Longitude, time.Now(), time.Now()).Error
	if err != nil {
		tx.Rollback()
		return domain.Admin{}, err
	}

	// Commit transaction after admin insert only
	if err := tx.Commit().Error; err != nil {
		return domain.Admin{}, err
	}

	return admin, nil
}

func (c *adminDatabase) FindAllUser(ctx context.Context, pagination request.Pagination) (users []response.User, err error) {

	limit := pagination.Limit
	offset := pagination.Offset

	query := `SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err = c.DB.Raw(query, limit, offset).Scan(&users).Error

	return users, err
}

// sales report from order // !add  product wise report
func (c *adminDatabase) CreateFullSalesReport(ctc context.Context, salesReq request.SalesReport) (salesReport []response.SalesReport, err error) {

	limit := salesReq.Pagination.Limit
	offset := salesReq.Pagination.Offset

	query := `SELECT u.first_name, u.email,  so.id AS shop_order_id, so.user_id, so.order_date,
	so.order_total_amount_minor AS order_total_price, so.discount_amount_minor AS discount,
	so.status AS order_status, pm.payment_type FROM shop_orders so
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

	// Check if Document_Type is nil before dereferencing
	if Document_Type != nil && *Document_Type != "manual" {
		verificationStatusValue = verificationStatus
	} else {
		// If Document_Type is nil or "manual", use the provided verificationStatus
		verificationStatusValue = verificationStatus
	}

	insertQuery := `INSERT INTO shop_details (id, admin_id, shop_verification_status, photo_shop_verification, business_doc_verification, identity_doc_verification, address_proof_verification, updated_at, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT (admin_id) DO UPDATE SET
	shop_verification_status = EXCLUDED.shop_verification_status,
	photo_shop_verification = EXCLUDED.photo_shop_verification,
	business_doc_verification = EXCLUDED.business_doc_verification,
	identity_doc_verification = EXCLUDED.identity_doc_verification,
	address_proof_verification = EXCLUDED.address_proof_verification,
	updated_at = EXCLUDED.updated_at`
	err = c.DB.Exec(insertQuery, domain.NewID(domain.PrefixShop), adminId, verificationStatusValue, shopVerification.Photo_Shop_Verification, shopVerification.Business_Doc_Verification, shopVerification.Identity_Doc_Verification, shopVerification.Address_Proof_Verification, time.Now(), time.Now()).Error
	if err != nil {
		return fmt.Errorf("failed to upsert shop details for admin %s: %v", adminId, err)
	}
	return nil

}

func (c *adminDatabase) CreateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error) {
	ad.ID = domain.NewID(domain.PrefixAdvertisement)
	if ad.Audience == "" {
		ad.Audience = domain.AdvertisementAudienceCustomer
	}
	query := `INSERT INTO advertisements
		(id, title, content, image_url, target_url, start_date, end_date, created_at, updated_at,
		 created_by_admin, admin_id, area_targeted, pincode_targeted, latitude, longitude, distance_km,
		 status, priority, audience, department_id, category_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`

	err := c.DB.Exec(query,
		ad.ID, ad.Title, ad.Content, ad.ImageURL, ad.TargetURL,
		ad.StartDate, ad.EndDate, time.Now(), time.Now(),
		ad.CreatedByAdmin, ad.AdminID,
		ad.AreaTargeted, ad.PincodeTargeted, ad.Latitude, ad.Longitude, ad.DistanceKM,
		ad.Status, ad.Priority, ad.Audience, ad.DepartmentID, ad.CategoryID,
	).Error

	return ad, err
}

func (c *adminDatabase) GetAllAdvertisements(ctx context.Context, pagination request.Pagination) (ads []domain.Advertisement, err error) {
	limit := pagination.Limit
	offset := pagination.Offset

	query := `SELECT * FROM advertisements ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err = c.DB.Raw(query, limit, offset).Scan(&ads).Error

	return ads, err
}

func (c *adminDatabase) UpdateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error) {
	query := `UPDATE advertisements SET
		title = $1, content = $2, image_url = $3, target_url = $4,
		start_date = $5, end_date = $6, updated_at = $7,
		area_targeted = $8, pincode_targeted = $9,
		latitude = $10, longitude = $11, distance_km = $12,
		status = $13, priority = $14,
		audience = $15, department_id = $16, category_id = $17
		WHERE id = $18`

	err := c.DB.Exec(query,
		ad.Title, ad.Content, ad.ImageURL, ad.TargetURL,
		ad.StartDate, ad.EndDate, time.Now(),
		ad.AreaTargeted, ad.PincodeTargeted, ad.Latitude, ad.Longitude, ad.DistanceKM,
		ad.Status, ad.Priority,
		ad.Audience, ad.DepartmentID, ad.CategoryID,
		ad.ID,
	).Error

	return ad, err
}

func (c *adminDatabase) DeleteAdvertisement(ctx context.Context, advertisementID string) error {
	query := `DELETE FROM advertisements WHERE id = $1`
	err := c.DB.Exec(query, advertisementID).Error

	return err
}

func (c *adminDatabase) GetAdvertisementByID(ctx context.Context, advertisementID string) (domain.Advertisement, error) {
	var ad domain.Advertisement
	query := `SELECT * FROM advertisements WHERE id = $1`
	err := c.DB.Raw(query, advertisementID).Scan(&ad).Error
	return ad, err
}

func (c *adminDatabase) GetActiveAdvertisements(ctx context.Context) ([]domain.Advertisement, error) {
	var ads []domain.Advertisement
	// start_date must be <= NOW (ad has already started or starts today)
	// end_date must be >= today midnight (ad is still valid today — not expired yesterday)
	query := `SELECT * FROM advertisements
		WHERE status = 'active'
		  AND start_date <= NOW()
		  AND end_date   >= DATE_TRUNC('day', NOW())
		ORDER BY
		  CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END,
		  created_at DESC`
	err := c.DB.Raw(query).Scan(&ads).Error
	return ads, err
}

func (c *adminDatabase) GetActiveAdvertisementsFiltered(ctx context.Context, filter domain.AdvertisementFilter) ([]domain.Advertisement, error) {
	var ads []domain.Advertisement

	// start_date <= NOW: ad has started
	// end_date >= today midnight: ad is still valid today (not expired yesterday)
	query := `SELECT * FROM advertisements
		WHERE status = 'active'
		  AND start_date <= NOW()
		  AND end_date   >= DATE_TRUNC('day', NOW())`
	args := []interface{}{}
	argIdx := 1

	// Audience / app-type filter: include ads with no audience set OR matching audience.
	if filter.AppType != "" {
		query += fmt.Sprintf(" AND (audience = '' OR audience = $%d)", argIdx)
		args = append(args, string(filter.AppType))
		argIdx++
	}

	// Pincode filter: include ads with no pincode set OR matching pincode.
	if filter.Pincode != "" {
		query += fmt.Sprintf(" AND (pincode_targeted = '' OR pincode_targeted = $%d)", argIdx)
		args = append(args, filter.Pincode)
		argIdx++
	}

	// Geo-radius filter using Haversine: include ads with no location set (lat=0,lng=0)
	// OR within distance_km radius of the provided coordinates.
	if filter.Latitude != 0 && filter.Longitude != 0 && filter.RadiusKM > 0 {
		query += fmt.Sprintf(`
		  AND (
		    (latitude = 0 AND longitude = 0)
		    OR (
		      distance_km = 0
		      OR (6371 * acos(
		        cos(radians($%d)) * cos(radians(latitude)) *
		        cos(radians(longitude) - radians($%d)) +
		        sin(radians($%d)) * sin(radians(latitude))
		      )) <= LEAST(distance_km, $%d)
		    )
		  )`, argIdx, argIdx+1, argIdx+2, argIdx+3)
		args = append(args, filter.Latitude, filter.Longitude, filter.Latitude, filter.RadiusKM)
		argIdx += 4
	}

	query += ` ORDER BY CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END, created_at DESC`

	err := c.DB.Raw(query, args...).Scan(&ads).Error
	return ads, err
}

// Shop Details
func (c *adminDatabase) CreateShop(ctx context.Context, shop domain.ShopDetails) (domain.ShopDetails, error) {
	shop.ShopID = utils.GenerateShopID()

	encBankAccount, err := c.encrypt(shop.BankAccountNumber)
	if err != nil {
		return shop, err
	}
	encBankIFSC, err := c.encrypt(shop.BankIFSC)
	if err != nil {
		return shop, err
	}
	encPAN, err := c.encrypt(shop.PanNumber)
	if err != nil {
		return shop, err
	}
	encITR, err := c.encrypt(shop.ITRDocuments)
	if err != nil {
		return shop, err
	}
	encDocValue, err := c.encrypt(shop.Document_Value)
	if err != nil {
		return shop, err
	}

	tx := c.DB.Begin()
	if tx.Error != nil {
		return shop, tx.Error
	}

	shopID := domain.NewID(domain.PrefixShop)

	query := `INSERT INTO shop_details (id, admin_id, shop_name,owner_name, address_line1, address_line2, email, phone,
	city, state, country, pincode, latitude, longitude, bank_account_number, shop_type, shop_status, bank_ifsc, pan_number, itr_documents, document_type, document_value,  created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
	ON CONFLICT (admin_id) DO UPDATE SET
		shop_name = EXCLUDED.shop_name,
		owner_name = EXCLUDED.owner_name,
		address_line1 = EXCLUDED.address_line1,
		address_line2 = EXCLUDED.address_line2,
		email = EXCLUDED.email,
		phone = EXCLUDED.phone,
		city = EXCLUDED.city,
		state = EXCLUDED.state,
		country = EXCLUDED.country,
		pincode = EXCLUDED.pincode,
		latitude = EXCLUDED.latitude,
		longitude = EXCLUDED.longitude,
		bank_account_number = EXCLUDED.bank_account_number,
		shop_type = EXCLUDED.shop_type,
		shop_status = EXCLUDED.shop_status,
		bank_ifsc = EXCLUDED.bank_ifsc,
		pan_number = EXCLUDED.pan_number,
		itr_documents = EXCLUDED.itr_documents,
		document_type = EXCLUDED.document_type,
		document_value = EXCLUDED.document_value,
		updated_at = EXCLUDED.updated_at
	RETURNING id`

	err = tx.Raw(query, shopID, shop.AdminID, shop.ShopName, shop.OwnerName, shop.AddressLine1,
		shop.AddressLine2, shop.Email, shop.Phone, shop.City, shop.State, shop.Country, shop.Pincode, shop.Latitude, shop.Longitude,
		encBankAccount, shop.ShopType, shop.ShopStatus, encBankIFSC, encPAN, encITR, shop.Document_Type, encDocValue,
		time.Now(), time.Now()).Scan(&shop.ID).Error

	if err != nil {
		tx.Rollback()
		return shop, err
	}

	// Use UPSERT to handle existing admin_id (unique constraint)
	queryVerification := `INSERT INTO shop_verifications (id, shop_id, admin_id, verification_status, remarks, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (admin_id) DO UPDATE SET
		shop_id = EXCLUDED.shop_id,
		verification_status = EXCLUDED.verification_status,
		remarks = EXCLUDED.remarks,
		updated_at = EXCLUDED.updated_at`

	adminIDStr := fmt.Sprintf("%s", shop.AdminID)
	if err := tx.Exec(queryVerification, domain.NewID(domain.PrefixShopVerif), shop.ID, adminIDStr, shop.ShopVerificationStatus, shop.ShopVerificationRemarks, time.Now(), time.Now()).Error; err != nil {
		tx.Rollback()
		return shop, err
	}

	// Insert default shop time record
	queryShopTime := `INSERT INTO shop_times (id, shop_id, status, open_time, close_time, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	if err := tx.Exec(queryShopTime, domain.NewID(domain.PrefixShopTime), shop.ID, "close", "09:00", "21:00", time.Now(), time.Now()).Error; err != nil {
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
	limit := pagination.Limit
	offset := pagination.Offset

	query := `SELECT sd.*, (EXISTS (SELECT 1 FROM shop_offers so WHERE so.shop_id = sd.id)) as has_offers FROM shop_details sd ORDER BY sd.created_at DESC LIMIT $1 OFFSET $2`
	err = c.DB.Raw(query, limit, offset).Scan(&shops).Error
	for i := range shops {
		c.decryptShopPII(&shops[i])
	}
	return shops, err
}

func (c *adminDatabase) GetShopByID(ctx context.Context, shopID string) (shop domain.ShopDetails, err error) {
	query := `SELECT sd.*, (EXISTS (SELECT 1 FROM shop_offers so WHERE so.shop_id = sd.id)) as has_offers FROM shop_details sd WHERE sd.id = $1`
	err = c.DB.Raw(query, shopID).Scan(&shop).Error
	c.decryptShopPII(&shop)
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

		// Encrypt PII fields at rest before persisting.
		if shopPIIKeys[k] {
			if s, ok := v.(string); ok {
				enc, encErr := c.encrypt(s)
				if encErr != nil {
					return nil, encErr
				}
				v = enc
			}
		}

		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", columnName, paramCount))
		values = append(values, v)
		paramCount++
	}

	// Add updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", paramCount))
	values = append(values, time.Now())
	paramCount++

	query := fmt.Sprintf("UPDATE shop_details SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), paramCount)
	values = append(values, shopId)

	result := c.DB.Exec(query, values...)
	if result.Error != nil {
		return nil, result.Error
	}

	return shop, nil
}

func (c *adminDatabase) GetShopByOwnerID(ctx context.Context, ownerID string) (shop domain.ShopDetails, err error) {
	query := `SELECT * FROM shop_details WHERE admin_id = $1`
	err = c.DB.Raw(query, ownerID).Scan(&shop).Error
	if err != nil {
		return shop, err
	}
	if shop.ID == "" {
		return shop, gorm.ErrRecordNotFound
	}
	c.decryptShopPII(&shop)
	return shop, nil
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

func (c *adminDatabase) SendNotificationToUser(ctx context.Context, userID string, message string) error {
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

func (c *adminDatabase) UploadShopDocument(ctx context.Context, shopID string, documentType string, documentValue string) error {
	encDocValue, err := c.encrypt(documentValue)
	if err != nil {
		return err
	}
	query := `UPDATE shop_details SET document_type = $1, document_value = $2, updated_at = $3 WHERE admin_id = $4`
	return c.DB.Exec(query, documentType, encDocValue, time.Now(), shopID).Error
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
	query := `INSERT INTO shop_details (id, admin_id, shop_name, owner_name, phone, address_line1, address_line2, city, state, pincode, latitude, longitude, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
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

	err = c.DB.Exec(query, domain.NewID(domain.PrefixShop), adminId, address.ShopName, address.OwnerName, address.Phone, address.AddressLine1, address.AddressLine2, address.City, address.State, address.Pincode,
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
	query := `SELECT COALESCE(shop_image_url, '') FROM shop_details WHERE id = $1`
	err := c.DB.Raw(query, shopId).Scan(&shopProfileImage).Error
	if err != nil {
		return "", err
	}
	return shopProfileImage, nil
}

func (c *adminDatabase) DeleteRefreshSessionByUserID(ctx context.Context, adminId string) error {
	query := `DELETE FROM refresh_sessions WHERE user_id = $1`
	err := c.DB.Exec(query, adminId).Error
	return err
}

func (c *adminDatabase) GetShopSocialDetails(ctx context.Context, shopID string) ([]domain.ShopSocial, error) {
	var details []domain.ShopSocial
	if err := c.DB.WithContext(ctx).Where("shop_id = ?", shopID).Find(&details).Error; err != nil {
		return nil, err
	}
	return details, nil
}

func (c *adminDatabase) GetAdminByID(ctx context.Context, adminID string) (domain.Admin, error) {
	var admin domain.Admin
	err := c.DB.Raw("SELECT * FROM admins WHERE id = $1", adminID).Scan(&admin).Error
	if err != nil {
		return admin, err
	}
	c.decryptAdminPII(&admin)
	return admin, nil
}

func (a *adminDatabase) GetDashboardStats(ctx context.Context) (domain.DashboardStats, error) {
	var stats domain.DashboardStats

	if err := a.DB.WithContext(ctx).Model(&domain.Admin{}).Count(&stats.TotalSellers).Error; err != nil {
		return stats, err
	}
	if err := a.DB.WithContext(ctx).Model(&domain.Admin{}).Where("status = ?", domain.AdminStatusActive).Count(&stats.ActiveSellers).Error; err != nil {
		return stats, err
	}
	if err := a.DB.WithContext(ctx).Model(&domain.ShopDetails{}).Count(&stats.TotalShops).Error; err != nil {
		return stats, err
	}
	if err := a.DB.WithContext(ctx).Model(&domain.ShopDetails{}).Where("shop_verification_status = ?", true).Count(&stats.VerifiedShops).Error; err != nil {
		return stats, err
	}
	if err := a.DB.WithContext(ctx).Model(&domain.ShopVerification{}).Where("verification_status = ?", false).Count(&stats.PendingVerifications).Error; err != nil {
		return stats, err
	}
	if err := a.DB.WithContext(ctx).Model(&domain.ShopOrder{}).Count(&stats.TotalOrders).Error; err != nil {
		return stats, err
	}
	if err := a.DB.WithContext(ctx).Model(&domain.User{}).Count(&stats.TotalCustomers).Error; err != nil {
		return stats, err
	}

	var revenue *int64
	if err := a.DB.WithContext(ctx).Model(&domain.ShopOrder{}).
		Where("status = ?", domain.StatusOrderDelivered).
		Select("COALESCE(SUM(order_total_amount_minor), 0)").
		Scan(&revenue).Error; err != nil {
		return stats, err
	}
	if revenue != nil {
		stats.TotalRevenue = float64(*revenue)
	}

	if err := a.DB.WithContext(ctx).Model(&domain.ProductItem{}).Count(&stats.TotalProducts).Error; err != nil {
		return stats, err
	}

	return stats, nil
}
