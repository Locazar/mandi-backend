package repository

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// find wallet by userID
func (c *OrderDatabase) FindWalletByUserID(ctx context.Context, userID string) (wallet domain.Wallet, err error) {
	query := `SELECT * FROM wallets WHERE user_id = $1`
	err = c.DB.Raw(query, userID).Scan(&wallet).Error

	return
}

// create a new wallet for user
func (c *OrderDatabase) SaveWallet(ctx context.Context, userID string) (walletID string, err error) {
	walletID = domain.NewID(domain.PrefixWallet)
	query := `INSERT INTO wallets (id, user_id, total_amount_amount_minor, total_amount_currency) VALUES ($1, $2, $3, $4) RETURNING id`
	err = c.DB.Raw(query, walletID, userID, 0, domain.CurrencyINR).Scan(&walletID).Error

	return
}

func (c *OrderDatabase) UpdateWallet(ctx context.Context, walletID string, upateTotalAmount uint) error {
	query := `UPDATE wallets SET total_amount_amount_minor = $1, total_amount_currency = $2 WHERE id = $3`
	err := c.DB.Exec(query, upateTotalAmount, domain.CurrencyINR, walletID).Error

	return err
}

func (c *OrderDatabase) SaveWalletTransaction(ctx context.Context, walletTrx domain.Transaction) error {
	trxDate := time.Now()
	query := `INSERT INTO transactions (wallet_id, transaction_date, amount_amount_minor, amount_currency, transaction_type)
	VALUES ($1, $2, $3, $4, $5)`
	err := c.DB.Exec(query, walletTrx.WalletID, trxDate,
		walletTrx.Amount.AmountMinor, walletTrx.Amount.Currency, walletTrx.TransactionType).Error

	return err
}

// find wallet transaction history

func (c *OrderDatabase) FindWalletTransactions(ctx context.Context, walletID string,
	pagination request.Pagination,
) (transaction []domain.Transaction, err error) {
	limit := pagination.Limit
	offset := pagination.Offset

	query := `SELECT * FROM transactions WHERE wallet_id = $1
	ORDER BY transaction_date DESC LIMIT $2 OFFSET $3`

	err = c.DB.Raw(query, walletID, limit, offset).Scan(&transaction).Error

	return
}
