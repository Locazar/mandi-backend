package response

<<<<<<< HEAD
import "github.com/rohit221990/mandi-backend/pkg/domain"
=======
import "github.com/nikhilnarayanan623/ecommerce-gin-clean-arch/pkg/domain"
>>>>>>> b9ab446 (Initial commit)

type OrderPayment struct {
	PaymentType  domain.PaymentType `json:"payment_type"`
	PaymentOrder any                `json:"payment_order"`
}
