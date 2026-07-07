package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
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

func (c *adminDatabase) GetAllAdvertisements(ctx context.Context, pagination request.Pagination, filter domain.AdvertisementFilter) (ads []domain.Advertisement, err error) {
	var conditions []string
	var args []interface{}

	if filter.DepartmentID != "" {
		conditions = append(conditions, fmt.Sprintf("department_id = $%d", len(args)+1))
		args = append(args, filter.DepartmentID)
	}
	if filter.CategoryID != "" {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", len(args)+1))
		args = append(args, filter.CategoryID)
	}
	if filter.PincodeTargeted != "" {
		conditions = append(conditions, fmt.Sprintf("pincode_targeted = $%d", len(args)+1))
		args = append(args, filter.PincodeTargeted)
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, string(filter.Status))
	}
	if filter.Audience != "" {
		conditions = append(conditions, fmt.Sprintf("audience = $%d", len(args)+1))
		args = append(args, string(filter.Audience))
	}
	if filter.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", len(args)+1))
		args = append(args, string(filter.Priority))
	}
	if filter.AdminID != "" {
		conditions = append(conditions, fmt.Sprintf("admin_id = $%d", len(args)+1))
		args = append(args, filter.AdminID)
	}
	if !filter.StartDateFrom.IsZero() {
		conditions = append(conditions, fmt.Sprintf("start_date >= $%d", len(args)+1))
		args = append(args, filter.StartDateFrom)
	}
	if !filter.EndDateTo.IsZero() {
		conditions = append(conditions, fmt.Sprintf("end_date <= $%d", len(args)+1))
		args = append(args, filter.EndDateTo)
	}
	if filter.FilterLatitude != 0 && filter.FilterLongitude != 0 && filter.DistanceKM > 0 {
		n := len(args) + 1
		conditions = append(conditions, fmt.Sprintf(
			`(6371 * acos(cos(radians($%d)) * cos(radians(latitude)) * cos(radians(longitude) - radians($%d)) + sin(radians($%d)) * sin(radians(latitude)))) <= $%d`,
			n, n+1, n+2, n+3,
		))
		args = append(args, filter.FilterLatitude, filter.FilterLongitude, filter.FilterLatitude, filter.DistanceKM)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	limitN := len(args) + 1
	offsetN := len(args) + 2
	query := fmt.Sprintf(
		`SELECT * FROM advertisements %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, limitN, offsetN,
	)
	args = append(args, int(pagination.Limit), int(pagination.Offset))

	log.Printf("[GetAllAdvertisements] query: %s | args: %v", query, args)
	err = c.DB.WithContext(ctx).Raw(query, args...).Scan(&ads).Error
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

	// Audience filter:
	// - 'seller'   ads → only seller app
	// - 'customer' ads → customer app AND seller app (it's the default/general audience)
	// - '' / NULL  ads → everyone
	// So: seller app sees seller + customer + empty; customer app sees customer + empty.
	if filter.AppType == domain.AdvertisementAudienceSeller {
		// Seller app: show only seller-targeted ads.
		query += " AND audience = 'seller'"
	} else if filter.AppType == domain.AdvertisementAudienceCustomer {
		// Customer app: show only customer-targeted ads.
		query += " AND audience = 'customer'"
	}
	// If AppType is empty, no audience filter applied — return all.

	// Location filter logic:
	//   An ad passes if ANY of these is true:
	//   1. Ad has no geo and no pincode targeting (global ad — shown everywhere)
	//   2. Geo provided and ad is within range (Haversine)
	//   3. Pincode provided and ad's pincode_targeted matches
	//
	// This means an ad targeted by pincode will always match when the user's
	// pincode matches, even when lat/lng are also present, and vice-versa.
	hasGeo := filter.Latitude != 0 && filter.Longitude != 0 && filter.RadiusKM > 0
	hasPincode := filter.Pincode != ""

	if hasGeo && hasPincode {
		query += fmt.Sprintf(`
		  AND (
		    -- Global ad: no geo and no pincode targeting
		    ((latitude = 0 AND longitude = 0) AND (pincode_targeted IS NULL OR pincode_targeted = ''))
		    OR
		    -- Geo match: ad has a location and user is within range
		    (latitude != 0 AND longitude != 0 AND (
		      distance_km = 0
		      OR (6371 * acos(
		        cos(radians($%d)) * cos(radians(latitude)) *
		        cos(radians(longitude) - radians($%d)) +
		        sin(radians($%d)) * sin(radians(latitude))
		      )) <= LEAST(distance_km, $%d)
		    ))
		    OR
		    -- Pincode match: ad targets this pincode
		    (pincode_targeted IS NOT NULL AND pincode_targeted != '' AND pincode_targeted = $%d)
		  )`, argIdx, argIdx+1, argIdx+2, argIdx+3, argIdx+4)
		args = append(args, filter.Latitude, filter.Longitude, filter.Latitude, filter.RadiusKM, filter.Pincode)
		argIdx += 5
	} else if hasGeo {
		query += fmt.Sprintf(`
		  AND (
		    ((latitude = 0 AND longitude = 0) AND (pincode_targeted IS NULL OR pincode_targeted = ''))
		    OR (latitude != 0 AND longitude != 0 AND (
		      distance_km = 0
		      OR (6371 * acos(
		        cos(radians($%d)) * cos(radians(latitude)) *
		        cos(radians(longitude) - radians($%d)) +
		        sin(radians($%d)) * sin(radians(latitude))
		      )) <= LEAST(distance_km, $%d)
		    ))
		  )`, argIdx, argIdx+1, argIdx+2, argIdx+3)
		args = append(args, filter.Latitude, filter.Longitude, filter.Latitude, filter.RadiusKM)
		argIdx += 4
	} else if hasPincode {
		// Pincode-only: ads with no pincode AND no geo targeting are global and always included.
		query += fmt.Sprintf(`
		  AND (
		    ((pincode_targeted IS NULL OR pincode_targeted = '') AND (latitude = 0 AND longitude = 0))
		    OR pincode_targeted = $%d
		  )`, argIdx)
		args = append(args, filter.Pincode)
		argIdx++
	}
	// If neither geo nor pincode is provided, no location filter is applied.

	query += ` ORDER BY CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END, created_at DESC`

	err := c.DB.Raw(query, args...).Scan(&ads).Error
	return ads, err
}

// Advertisement Requests (seller-raised)

func (c *adminDatabase) CreateAdvertisementRequest(ctx context.Context, req domain.AdvertisementRequest) (domain.AdvertisementRequest, error) {
	req.ID = domain.NewID(domain.PrefixAdvertRequest)
	if req.Status == "" {
		req.Status = domain.AdvertRequestStatusPending
	}
	query := `INSERT INTO advertisement_requests
		(id, admin_id, shop_id, title, content, start_date, end_date, plan_key, price_minor, status, admin_comment, created_at, updated_at)
		VALUES ($1,$2,$3,NULLIF($4,''),$5,$6,$7,$8,$9,$10,$11,NOW(),NOW())`
	err := c.DB.WithContext(ctx).Exec(query,
		req.ID, req.AdminID, req.ShopID, req.Title, req.Content,
		req.StartDate, req.EndDate, req.PlanKey, req.PriceMinor,
		req.Status, req.AdminComment,
	).Error
	return req, err
}

func (c *adminDatabase) GetAllAdvertisementRequests(ctx context.Context, pagination request.Pagination, adminID string) ([]domain.AdvertisementRequest, error) {
	var requests []domain.AdvertisementRequest
	query := `SELECT ar.*, COALESCE(sd.shop_name, '') AS shop_name
		FROM advertisement_requests ar
		LEFT JOIN shop_details sd ON sd.id = ar.shop_id
		WHERE ($1 = '' OR ar.admin_id = $1)
		ORDER BY ar.created_at DESC
		LIMIT $2 OFFSET $3`
	err := c.DB.WithContext(ctx).Raw(query, adminID, pagination.Limit, pagination.Offset).Scan(&requests).Error
	return requests, err
}

func (c *adminDatabase) GetAdvertisementRequestByID(ctx context.Context, requestID string) (domain.AdvertisementRequest, error) {
	var req domain.AdvertisementRequest
	query := `SELECT ar.*, COALESCE(sd.shop_name, '') AS shop_name
		FROM advertisement_requests ar
		LEFT JOIN shop_details sd ON sd.id = ar.shop_id
		WHERE ar.id = $1`
	err := c.DB.WithContext(ctx).Raw(query, requestID).Scan(&req).Error
	if err == nil && req.ID == "" {
		return req, gorm.ErrRecordNotFound
	}
	return req, err
}

func (c *adminDatabase) UpdateAdvertisementRequest(ctx context.Context, req domain.AdvertisementRequest) (domain.AdvertisementRequest, error) {
	query := `UPDATE advertisement_requests
		SET title = $2, content = $3, start_date = $4, end_date = $5,
		    plan_key = $6, price_minor = $7, status = $8, admin_comment = $9, updated_at = NOW()
		WHERE id = $1`
	result := c.DB.WithContext(ctx).Exec(query,
		req.ID, req.Title, req.Content, req.StartDate, req.EndDate,
		req.PlanKey, req.PriceMinor, req.Status, req.AdminComment,
	)
	if result.Error != nil {
		return req, result.Error
	}
	if result.RowsAffected == 0 {
		return req, gorm.ErrRecordNotFound
	}
	return c.GetAdvertisementRequestByID(ctx, req.ID)
}

func (c *adminDatabase) DeleteAdvertisementRequest(ctx context.Context, requestID string) error {
	result := c.DB.WithContext(ctx).Exec(`DELETE FROM advertisement_requests WHERE id = $1`, requestID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (c *adminDatabase) SetAdvertisementRequestPaymentOrder(ctx context.Context, requestID, orderID string) error {
	result := c.DB.WithContext(ctx).Exec(`UPDATE advertisement_requests
		SET payment_order_id = $2, payment_status = 'pending', updated_at = NOW()
		WHERE id = $1 AND payment_status <> 'paid'`, requestID, orderID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (c *adminDatabase) GetAdvertisementRequestByPaymentOrderID(ctx context.Context, orderID string) (domain.AdvertisementRequest, error) {
	var req domain.AdvertisementRequest
	err := c.DB.WithContext(ctx).Raw(
		`SELECT * FROM advertisement_requests WHERE payment_order_id = $1`, orderID,
	).Scan(&req).Error
	if err == nil && req.ID == "" {
		return req, gorm.ErrRecordNotFound
	}
	return req, err
}

func (c *adminDatabase) MarkAdvertisementRequestPaid(ctx context.Context, requestID, paymentID string) error {
	result := c.DB.WithContext(ctx).Exec(`UPDATE advertisement_requests
		SET payment_status = 'paid', payment_id = $2, paid_at = NOW(), updated_at = NOW()
		WHERE id = $1`, requestID, paymentID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (c *adminDatabase) MarkAdvertisementRequestPaymentFailed(ctx context.Context, requestID string) error {
	// Never downgrade a paid request.
	result := c.DB.WithContext(ctx).Exec(`UPDATE advertisement_requests
		SET payment_status = 'failed', updated_at = NOW()
		WHERE id = $1 AND payment_status <> 'paid'`, requestID)
	return result.Error
}

// Advertisement pricing configuration (admin-managed)

func (c *adminDatabase) ListAdvertisementPlans(ctx context.Context, activeOnly bool) ([]domain.AdvertisementPlanConfig, error) {
	var plans []domain.AdvertisementPlanConfig
	query := `SELECT * FROM advertisement_price_plans`
	if activeOnly {
		query += ` WHERE is_active = TRUE`
	}
	query += ` ORDER BY sort_order ASC, created_at ASC`
	err := c.DB.WithContext(ctx).Raw(query).Scan(&plans).Error
	return plans, err
}

func (c *adminDatabase) GetAdvertisementPlanByKey(ctx context.Context, planKey string) (domain.AdvertisementPlanConfig, error) {
	var plan domain.AdvertisementPlanConfig
	err := c.DB.WithContext(ctx).Raw(
		`SELECT * FROM advertisement_price_plans WHERE plan_key = $1`, planKey,
	).Scan(&plan).Error
	if err == nil && plan.ID == "" {
		return plan, gorm.ErrRecordNotFound
	}
	return plan, err
}

func (c *adminDatabase) CreateAdvertisementPlan(ctx context.Context, plan domain.AdvertisementPlanConfig) (domain.AdvertisementPlanConfig, error) {
	plan.ID = domain.NewID(domain.PrefixAdvertPlan)
	err := c.DB.WithContext(ctx).Exec(`INSERT INTO advertisement_price_plans
		(id, plan_key, name, description, rate_per_day_minor, sort_order, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NOW(),NOW())`,
		plan.ID, plan.PlanKey, plan.Name, plan.Description,
		plan.RatePerDayMinor, plan.SortOrder, plan.IsActive,
	).Error
	return plan, err
}

func (c *adminDatabase) UpdateAdvertisementPlan(ctx context.Context, plan domain.AdvertisementPlanConfig) (domain.AdvertisementPlanConfig, error) {
	result := c.DB.WithContext(ctx).Exec(`UPDATE advertisement_price_plans
		SET plan_key = $2, name = $3, description = $4, rate_per_day_minor = $5,
		    sort_order = $6, is_active = $7, updated_at = NOW()
		WHERE id = $1`,
		plan.ID, plan.PlanKey, plan.Name, plan.Description,
		plan.RatePerDayMinor, plan.SortOrder, plan.IsActive,
	)
	if result.Error != nil {
		return plan, result.Error
	}
	if result.RowsAffected == 0 {
		return plan, gorm.ErrRecordNotFound
	}
	return plan, nil
}

func (c *adminDatabase) DeleteAdvertisementPlan(ctx context.Context, planID string) error {
	result := c.DB.WithContext(ctx).Exec(
		`DELETE FROM advertisement_price_plans WHERE id = $1`, planID,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (c *adminDatabase) GetAdvertisementPricingConfig(ctx context.Context) (domain.AdvertisementPricingConfig, error) {
	var cfg domain.AdvertisementPricingConfig
	err := c.DB.WithContext(ctx).Raw(
		`SELECT * FROM advertisement_pricing_config ORDER BY id LIMIT 1`,
	).Scan(&cfg).Error
	if err == nil && cfg.ID == "" {
		return cfg, gorm.ErrRecordNotFound
	}
	return cfg, err
}

// Feature flags

func (c *adminDatabase) ListFeatureFlags(ctx context.Context) ([]domain.FeatureFlag, error) {
	var flags []domain.FeatureFlag
	err := c.DB.WithContext(ctx).Raw(
		`SELECT * FROM feature_flags ORDER BY flag_key ASC`,
	).Scan(&flags).Error
	return flags, err
}

func (c *adminDatabase) CreateFeatureFlag(ctx context.Context, flag domain.FeatureFlag) (domain.FeatureFlag, error) {
	flag.ID = domain.NewID(domain.PrefixFeatureFlag)
	err := c.DB.WithContext(ctx).Exec(`INSERT INTO feature_flags
		(id, flag_key, enabled, description, created_at, updated_at)
		VALUES ($1,$2,$3,$4,NOW(),NOW())`,
		flag.ID, flag.FlagKey, flag.Enabled, flag.Description,
	).Error
	return flag, err
}

func (c *adminDatabase) UpdateFeatureFlag(ctx context.Context, flag domain.FeatureFlag) (domain.FeatureFlag, error) {
	result := c.DB.WithContext(ctx).Exec(`UPDATE feature_flags
		SET flag_key = $2, enabled = $3, description = $4, updated_at = NOW()
		WHERE id = $1`,
		flag.ID, flag.FlagKey, flag.Enabled, flag.Description,
	)
	if result.Error != nil {
		return flag, result.Error
	}
	if result.RowsAffected == 0 {
		return flag, gorm.ErrRecordNotFound
	}
	return flag, nil
}

func (c *adminDatabase) DeleteFeatureFlag(ctx context.Context, flagID string) error {
	result := c.DB.WithContext(ctx).Exec(`DELETE FROM feature_flags WHERE id = $1`, flagID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// App configs

func (c *adminDatabase) ListAppConfigs(ctx context.Context) ([]domain.AppConfig, error) {
	var cfgs []domain.AppConfig
	err := c.DB.WithContext(ctx).Raw(
		`SELECT * FROM app_configs ORDER BY config_key ASC`,
	).Scan(&cfgs).Error
	return cfgs, err
}

func (c *adminDatabase) GetAppConfigByKey(ctx context.Context, configKey string) (domain.AppConfig, error) {
	var cfg domain.AppConfig
	err := c.DB.WithContext(ctx).Raw(
		`SELECT * FROM app_configs WHERE config_key = $1`, configKey,
	).Scan(&cfg).Error
	if err == nil && cfg.ID == "" {
		return cfg, gorm.ErrRecordNotFound
	}
	return cfg, err
}

func (c *adminDatabase) CreateAppConfig(ctx context.Context, cfg domain.AppConfig) (domain.AppConfig, error) {
	cfg.ID = domain.NewID(domain.PrefixAppConfig)
	err := c.DB.WithContext(ctx).Exec(`INSERT INTO app_configs
		(id, config_key, value, description, enabled, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,NOW(),NOW())`,
		cfg.ID, cfg.ConfigKey, cfg.Value, cfg.Description, cfg.Enabled,
	).Error
	return cfg, err
}

func (c *adminDatabase) UpdateAppConfig(ctx context.Context, cfg domain.AppConfig) (domain.AppConfig, error) {
	result := c.DB.WithContext(ctx).Exec(`UPDATE app_configs
		SET config_key = $2, value = $3, description = $4, enabled = $5, updated_at = NOW()
		WHERE id = $1`,
		cfg.ID, cfg.ConfigKey, cfg.Value, cfg.Description, cfg.Enabled,
	)
	if result.Error != nil {
		return cfg, result.Error
	}
	if result.RowsAffected == 0 {
		return cfg, gorm.ErrRecordNotFound
	}
	return cfg, nil
}

func (c *adminDatabase) DeleteAppConfig(ctx context.Context, configID string) error {
	result := c.DB.WithContext(ctx).Exec(`DELETE FROM app_configs WHERE id = $1`, configID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Subscription plans (admin panel)

func (c *adminDatabase) ListSubscriptionPlans(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	var plans []domain.SubscriptionPlan
	err := c.DB.WithContext(ctx).Raw(
		`SELECT * FROM subscription_plans ORDER BY duration_days ASC, name ASC`,
	).Scan(&plans).Error
	return plans, err
}

func (c *adminDatabase) CreateSubscriptionPlan(ctx context.Context, plan domain.SubscriptionPlan) (domain.SubscriptionPlan, error) {
	plan.ID = domain.NewID(domain.PrefixSubscPlan)
	if plan.PriceMonthly.Currency == "" {
		plan.PriceMonthly.Currency = "INR"
	}
	err := c.DB.WithContext(ctx).Exec(`INSERT INTO subscription_plans
		(id, name, price_monthly_amount_minor, price_monthly_currency, duration_days, is_active)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		plan.ID, plan.Name, plan.PriceMonthly.AmountMinor, plan.PriceMonthly.Currency,
		plan.DurationDays, plan.IsActive,
	).Error
	return plan, err
}

func (c *adminDatabase) UpdateSubscriptionPlan(ctx context.Context, plan domain.SubscriptionPlan) (domain.SubscriptionPlan, error) {
	if plan.PriceMonthly.Currency == "" {
		plan.PriceMonthly.Currency = "INR"
	}
	result := c.DB.WithContext(ctx).Exec(`UPDATE subscription_plans
		SET name = $2, price_monthly_amount_minor = $3, price_monthly_currency = $4,
		    duration_days = $5, is_active = $6
		WHERE id = $1`,
		plan.ID, plan.Name, plan.PriceMonthly.AmountMinor, plan.PriceMonthly.Currency,
		plan.DurationDays, plan.IsActive,
	)
	if result.Error != nil {
		return plan, result.Error
	}
	if result.RowsAffected == 0 {
		return plan, gorm.ErrRecordNotFound
	}
	return plan, nil
}

func (c *adminDatabase) DeleteSubscriptionPlan(ctx context.Context, planID string) error {
	// subscription_orders / user_subscriptions reference plans with
	// ON DELETE RESTRICT, so plans with history cannot be deleted — the FK
	// error is surfaced to the caller (deactivate instead).
	result := c.DB.WithContext(ctx).Exec(`DELETE FROM subscription_plans WHERE id = $1`, planID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (c *adminDatabase) UpdateAdvertisementPricingConfig(ctx context.Context, cfg domain.AdvertisementPricingConfig) (domain.AdvertisementPricingConfig, error) {
	// Upsert the singleton row.
	err := c.DB.WithContext(ctx).Exec(`INSERT INTO advertisement_pricing_config
		(id, gst_rate_percent, platform_fee_percent, updated_at)
		VALUES ('advcfg_default', $1, $2, NOW())
		ON CONFLICT (id) DO UPDATE
		SET gst_rate_percent = EXCLUDED.gst_rate_percent,
		    platform_fee_percent = EXCLUDED.platform_fee_percent,
		    updated_at = NOW()`,
		cfg.GSTRatePercent, cfg.PlatformFeePercent,
	).Error
	if err != nil {
		return cfg, err
	}
	return c.GetAdvertisementPricingConfig(ctx)
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

// SearchShops filters shops by phone, pincode, city, and/or radius (km) around a point.
// Empty filters are ignored; all provided filters are ANDed together.
func (c *adminDatabase) SearchShops(ctx context.Context, filter request.ShopSearch) (shops []domain.ShopDetails, err error) {
	query := `SELECT sd.*, (EXISTS (SELECT 1 FROM shop_offers so WHERE so.shop_id = sd.id)) as has_offers FROM shop_details sd WHERE 1=1`
	args := []interface{}{}
	i := 1

	if filter.Phone != "" {
		query += fmt.Sprintf(" AND sd.phone ILIKE $%d", i)
		args = append(args, "%"+filter.Phone+"%")
		i++
	}
	if filter.Pincode != "" {
		query += fmt.Sprintf(" AND sd.pincode = $%d", i)
		args = append(args, filter.Pincode)
		i++
	}
	if filter.City != "" {
		query += fmt.Sprintf(" AND sd.city ILIKE $%d", i)
		args = append(args, "%"+filter.City+"%")
		i++
	}
	if filter.Search != "" {
		query += fmt.Sprintf(" AND (sd.shop_name ILIKE $%d OR sd.owner_name ILIKE $%d)", i, i+1)
		args = append(args, "%"+filter.Search+"%", "%"+filter.Search+"%")
		i += 2
	}
	if filter.RadiusKm > 0 && filter.Latitude != 0 && filter.Longitude != 0 {
		// Haversine distance in km
		query += fmt.Sprintf(` AND sd.latitude != 0 AND sd.longitude != 0 AND (
			6371 * acos(least(1.0,
				cos(radians($%d)) * cos(radians(sd.latitude)) *
				cos(radians(sd.longitude) - radians($%d)) +
				sin(radians($%d)) * sin(radians(sd.latitude))
			))
		) <= $%d`, i, i+1, i+2, i+3)
		args = append(args, filter.Latitude, filter.Longitude, filter.Latitude, filter.RadiusKm)
		i += 4
	}

	query += fmt.Sprintf(" ORDER BY sd.created_at DESC LIMIT $%d", i)
	args = append(args, filter.Limit)

	err = c.DB.Raw(query, args...).Scan(&shops).Error
	for j := range shops {
		c.decryptShopPII(&shops[j])
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
