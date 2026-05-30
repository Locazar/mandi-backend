package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type OrderRepository interface {
	Transaction(callBack func(transactionRepo OrderRepository) error) error

	SaveOrderLine(ctx context.Context, orderLine domain.OrderLine) error

	UpdateShopOrderStatus(ctx context.Context, shopOrderID string, status domain.OrderStatusType) error
	UpdateShopOrderStatusAndSavePaymentMethod(ctx context.Context, shopOrderID string, status domain.OrderStatusType, paymentID string) error

	// shop order order
	SaveShopOrder(ctx context.Context, shopOrder domain.ShopOrder) (shopOrderID string, err error)
	FindShopOrderByShopOrderID(ctx context.Context, shopOrderID string) (domain.ShopOrder, error)
	FindAllShopOrders(ctx context.Context, pagination request.Pagination) (shopOrders []response.ShopOrder, err error)
	FindAllShopOrdersByUserID(ctx context.Context, userID string, pagination request.Pagination) ([]response.ShopOrder, error)

	// find shop order items
	FindAllOrdersItemsByShopOrderID(ctx context.Context,
		shopOrderID string, pagination request.Pagination) (orderItems []response.OrderItem, err error)

	// order status
	FindOrderStatusByShopOrderID(ctx context.Context, shopOrderID string) (domain.OrderStatusType, error)
	FindAllOrderStatuses(ctx context.Context) ([]domain.OrderStatusType, error)

	//order return
	FindOrderReturnByReturnID(ctx context.Context, orderReturnID string) (domain.OrderReturn, error)
	FindOrderReturnByShopOrderID(ctx context.Context, shopOrderID string) (orderReturn domain.OrderReturn, err error)
	FindAllOrderReturns(ctx context.Context, pagination request.Pagination) ([]response.OrderReturn, error)
	FindAllPendingOrderReturns(ctx context.Context, pagination request.Pagination) ([]response.OrderReturn, error)
	SaveOrderReturn(ctx context.Context, orderReturn domain.OrderReturn) error
	UpdateOrderReturn(ctx context.Context, orderReturn domain.OrderReturn) error

	// wallet
	FindWalletByUserID(ctx context.Context, userID string) (wallet domain.Wallet, err error)
	SaveWallet(ctx context.Context, userID string) (walletID string, err error)
	UpdateWallet(ctx context.Context, walletID string, updateTotalAmount uint) error
	SaveWalletTransaction(ctx context.Context, walletTrx domain.Transaction) error

	FindWalletTransactions(ctx context.Context, walletID string,
		pagination request.Pagination) (transaction []domain.Transaction, err error)

	SaveShoppingFeedback(ctx context.Context, feedback request.ShoppingFeedback) error
}
