package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type paymentDatabase struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) interfaces.PaymentRepository {
	return &paymentDatabase{
		db: db,
	}
}

func (c *paymentDatabase) FindPaymentMethodByID(ctx context.Context, paymentMethodID string) (paymentMethods domain.PaymentMethod, err error) {

	query := `SELECT * FROM payment_methods WHERE id = $1`

	err = c.db.Raw(query, paymentMethodID).Scan(&paymentMethods).Error

	return paymentMethods, err
}

// find payment_method by name
func (c *paymentDatabase) FindPaymentMethodByType(ctx context.Context,
	paymentType domain.PaymentType) (paymentMethod domain.PaymentMethod, err error) {

	query := `SELECT * FROM payment_methods WHERE name = $1`
	err = c.db.Raw(query, paymentType).Scan(&paymentMethod).Error

	return paymentMethod, err
}

func (c *paymentDatabase) FindAllPaymentMethods(ctx context.Context) (paymentMethods []domain.PaymentMethod, err error) {

	query := `SELECT * FROM payment_methods`
	err = c.db.Raw(query).Scan(&paymentMethods).Error

	return paymentMethods, err
}

func (c *paymentDatabase) SavePaymentMethod(ctx context.Context, paymentMethod domain.PaymentMethod) (paymentMethodID string, err error) {
	paymentMethodID = domain.NewID(domain.PrefixPaymentMethod)
	query := `INSERT INTO payment_methods (id, name, block_status, maximum_amount_amount_minor, maximum_amount_currency)
	VALUES ($1, $2, $3, $4, $5)`
	err = c.db.Exec(query, paymentMethodID, paymentMethod.Name, paymentMethod.BlockStatus,
		paymentMethod.MaximumAmount.AmountMinor, paymentMethod.MaximumAmount.Currency).Error

	return paymentMethodID, err
}
func (c *paymentDatabase) UpdatePaymentMethod(ctx context.Context, paymentMethod request.PaymentMethodUpdate) error {

	query := `UPDATE payment_methods SET  block_status = $1,
	maximum_amount_amount_minor = $2, maximum_amount_currency = $3 WHERE id = $4`

	err := c.db.Exec(query, paymentMethod.BlockStatus,
		paymentMethod.MaximumAmount, domain.CurrencyINR, paymentMethod.ID).Error

	return err
}
