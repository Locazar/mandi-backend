package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

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
	fmt.Println(admin)
	query := `INSERT INTO admins (user_name, email, mobile, password, shop_name, gstin, shop_id,
		address_line1, address_line2, city, state, country, pincode,
		bank_account_number, bank_ifsc, pan, aadhar, agree_to_terms,
		verified, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`

	err := c.DB.Exec(query, admin.UserName, admin.Email, admin.Mobile, admin.Password, admin.ShopName, admin.GSTIN, admin.ShopID,
		admin.AddressLine1, admin.AddressLine2, admin.City, admin.State, admin.Country, admin.Pincode,
		admin.BankAccountNumber, admin.BankIFSC, admin.PAN, admin.Aadhar, admin.AgreeToTerms,
		admin.Verified, admin.Status, time.Now(), time.Now()).Error

	return err
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
