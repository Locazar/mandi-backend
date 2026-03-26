package response

import (
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

var ResoposeMap map[string]string

// admin
type AdminLogin struct {
	ID    uint   `json:"id" `
	Email string `json:"email"`
}

type AdminWithShopVerification struct {
	Admin            domain.Admin            `json:"admin"`
	ShopVerification domain.ShopVerification `json:"shop_verification"`
}

// Helper function to convert domain objects to response
func ConvertAdminToResponse(admin domain.Admin, shopVerification domain.ShopVerification) AdminWithShopVerification {
	return AdminWithShopVerification{
		Admin:            admin,
		ShopVerification: shopVerification,
	}
}

// reponse for get all variations with its respective category

type SalesReport struct {
	UserID          uint      `json:"user_id"`
	FirstName       string    `json:"first_name"`
	Email           string    `json:"email"`
	ShopOrderID     uint      `json:"order_id"`
	OrderDate       time.Time `json:"order_date"`
	OrderTotalPrice uint      `json:"order_total_price"`
	Discount        uint      `json:"discount_price"`
	OrderStatus     string    `json:"order_status"`
	PaymentType     string    `json:"payment_type"`
}

type Stock struct {
	ProductItemID    uint              `json:"product_item_id"`
	ProductName      string            `json:"product_name"`
	VariationOptions []VariationOption `gorm:"-"`
}
