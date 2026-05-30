package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type OrderUseCase struct {
	orderRepo interfaces.OrderRepository
	cartRepo  interfaces.CartRepository
	userRepo  interfaces.UserRepository
}

func NewOrderUseCase(orderRepo interfaces.OrderRepository, cartRepo interfaces.CartRepository,
	userRepo interfaces.UserRepository,
	paymentRepo interfaces.PaymentRepository) service.OrderUseCase {
	return &OrderUseCase{
		orderRepo: orderRepo,
		cartRepo:  cartRepo,
		userRepo:  userRepo,
	}
}

// get all order statuses
func (c *OrderUseCase) FindAllOrderStatuses(ctx context.Context) ([]domain.OrderStatusType, error) {

	orderStatuses, err := c.orderRepo.FindAllOrderStatuses(ctx)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to find all order statuses")
	}

	return orderStatuses, nil
}

// Save order
func (c *OrderUseCase) SaveOrder(ctx context.Context, userID, addressID string) (string, error) {

	cart, err := c.cartRepo.FindCartByUserID(ctx, userID)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to get user cart")
	}

	if cart.TotalPrice.IsZero() {
		return "", ErrEmptyCart
	}

	// check the cart of user is valid for place order
	valid, err := c.cartRepo.IsCartValidForOrder(ctx, userID)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to check cart is valid for order")
	}

	if !valid {
		return "", ErrOutOfStockOnCart
	}

	orderTotal, err := cart.TotalPrice.Sub(cart.DiscountAmount)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to compute order total")
	}

	shopOrder := domain.ShopOrder{
		UserID:     userID,
		AddressID:  addressID,
		OrderTotal: orderTotal,
		Discount:   cart.DiscountAmount,
		Status:     domain.StatusPaymentPending,
	}

	err = c.orderRepo.Transaction(func(trxRepo interfaces.OrderRepository) error {

		shopOrder.ID, err = trxRepo.SaveShopOrder(ctx, shopOrder)
		if err != nil {
			return utils.PrependMessageToError(err, "failed to save shop order on database")
		}

		cartItems, err := c.cartRepo.FindAllCartItemsByCartID(ctx, cart.ID)
		if err != nil {
			return utils.PrependMessageToError(err, "failed to find all cart items")
		}

		var OrderPrice uint
		// save all order lines
		for _, cartItem := range cartItems {

			if cartItem.DiscountPrice != 0 {
				OrderPrice = cartItem.DiscountPrice
			} else {
				OrderPrice = cartItem.Price
			}

			orderLine := domain.OrderLine{
				ProductItemID: cartItem.ProductItemId,
				ShopOrderID:   shopOrder.ID,
				Qty:           cartItem.Qty,
				Price:         domain.INR(int64(OrderPrice)),
			}
			err = trxRepo.SaveOrderLine(ctx, orderLine)
			if err != nil {
				return utils.PrependMessageToError(err, "failed to save order line on database")
			}
		}
		return nil
	})
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to complete save order")
	}

	return shopOrder.ID, nil
}

// Find all orders of a user
func (c *OrderUseCase) FindUserShopOrder(ctx context.Context, userID string,
	pagination request.Pagination) ([]response.ShopOrder, error) {

	shopOrders, err := c.orderRepo.FindAllShopOrdersByUserID(ctx, userID, pagination)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to find all shop orders by user id")
	}

	for i, order := range shopOrders {

		address, err := c.userRepo.FindAddressByID(ctx, order.AddressID)
		if err != nil {
			return nil, utils.PrependMessageToError(err, "failed to get order address")
		}
		shopOrders[i].Address = address
	}

	return shopOrders, nil
}

// func to Find all shop order
func (c *OrderUseCase) FindAllShopOrders(ctx context.Context, pagination request.Pagination) ([]response.ShopOrder, error) {

	shopOrders, err := c.orderRepo.FindAllShopOrders(ctx, pagination)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to find all shop orders")
	}

	for i, order := range shopOrders {

		address, err := c.userRepo.FindAddressByID(ctx, order.AddressID)
		if err != nil {
			return nil, utils.PrependMessageToError(err, "failed to get order address")
		}
		shopOrders[i].Address = address
	}

	return shopOrders, nil
}

func (c *OrderUseCase) FindOrderItems(ctx context.Context, shopOrderID string,
	pagination request.Pagination) (orderItems []response.OrderItem, err error) {

	orderItems, err = c.orderRepo.FindAllOrdersItemsByShopOrderID(ctx, shopOrderID, pagination)
	if err != nil {
		return nil, utils.PrependMessageToError(err, "failed to find order items using shop order id")
	}

	return orderItems, nil
}

func (c *OrderUseCase) CancelOrder(ctx context.Context, shopOrderID string) error {

	shopOrder, err := c.orderRepo.FindShopOrderByShopOrderID(ctx, shopOrderID)
	if err != nil {
		return err
	}

	if shopOrder.Status != domain.StatusOrderPlaced {
		return fmt.Errorf("order is ' %s ' \ncan't cancel the order", shopOrder.Status)
	}

	err = c.orderRepo.UpdateShopOrderStatus(ctx, shopOrder.ID, domain.StatusOrderCancelled)
	if err != nil {
		return fmt.Errorf("failed to cancel the order %v", err.Error())
	}

	return nil
}

// update order
func (c *OrderUseCase) UpdateOrderStatus(ctx context.Context, shopOrderID string, newStatus domain.OrderStatusType) error {

	if !newStatus.IsValid() {
		return fmt.Errorf("invalid order status: %s", newStatus)
	}

	shopOrder, err := c.orderRepo.FindShopOrderByShopOrderID(ctx, shopOrderID)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to find shop order")
	}

	switch shopOrder.Status {
	case domain.StatusOrderPlaced: // if order status is placed then change status should be order delivered
		if newStatus != domain.StatusOrderDelivered {
			return fmt.Errorf("order status is 'order placed' \nchange status should be 'order delivered'")
		}
	default:
		return fmt.Errorf("order status %s can't change to %s ", shopOrder.Status, newStatus)
	}

	err = c.orderRepo.UpdateShopOrderStatus(ctx, shopOrder.ID, newStatus)
	if err != nil {
		return fmt.Errorf("failed to change order status %v", err.Error())
	}
	return nil
}

// to get pending order returns
func (c *OrderUseCase) FindAllPendingOrderReturns(ctx context.Context, pagination request.Pagination) ([]response.OrderReturn, error) {

	pendingOrderReturns, err := c.orderRepo.FindAllPendingOrderReturns(ctx, pagination)
	if err != nil {
		return pendingOrderReturns, fmt.Errorf("failed to Find pendin order returns \nerror:%v", err.Error())
	}
	return pendingOrderReturns, nil
}

// to get all order return
func (c *OrderUseCase) FindAllOrderReturns(ctx context.Context, pagination request.Pagination) ([]response.OrderReturn, error) {

	orderReturns, err := c.orderRepo.FindAllOrderReturns(ctx, pagination)
	if err != nil {
		return orderReturns, fmt.Errorf("failed to Find all order returns \nerror:%v", err.Error())
	}
	return orderReturns, nil
}

func (c *OrderUseCase) SubmitReturnRequest(ctx context.Context, returnDetails request.Return) error {

	shopOrder, err := c.orderRepo.FindShopOrderByShopOrderID(ctx, returnDetails.ShopOrderID)
	if err != nil {
		return err
	}

	if shopOrder.Status != domain.StatusOrderDelivered {
		return fmt.Errorf("order is ' %s '\ncan't a make return request for this order", shopOrder.Status)
	}

	orderReturn := domain.OrderReturn{
		ShopOrderID:  returnDetails.ShopOrderID,
		ReturnReason: returnDetails.ReturnReason,
		RequestDate:  time.Now(),
		RefundAmount: shopOrder.OrderTotal,
	}

	err = c.orderRepo.Transaction(func(trxRepo interfaces.OrderRepository) error {

		err := trxRepo.SaveOrderReturn(ctx, orderReturn)
		if err != nil {
			return fmt.Errorf("failed to submit order return \nerror:%v", err.Error())
		}

		err = trxRepo.UpdateShopOrderStatus(ctx, shopOrder.ID, domain.StatusReturnRequested)
		if err != nil {
			return fmt.Errorf("failed to update order status \n error:%v", err.Error())
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to save order return \nerror:%v", err.Error())
	}
	log.Println("successfully order return request submitted")
	return nil
}

func (c *OrderUseCase) UpdateReturnDetails(ctx context.Context, updateDetails request.UpdateOrderReturn) error {

	if !updateDetails.OrderStatus.IsValid() {
		return fmt.Errorf("invalid order status: %s", updateDetails.OrderStatus)
	}

	orderReturn, err := c.orderRepo.FindOrderReturnByReturnID(ctx, updateDetails.OrderReturnID)
	if err != nil {
		return fmt.Errorf("failed to Find order \nerror:%v", err.Error())
	}

	shopOrder, err := c.orderRepo.FindShopOrderByShopOrderID(ctx, orderReturn.ShopOrderID)
	if err != nil {
		return fmt.Errorf("failed to Find order details \nerror:%v", err.Error())
	}

	newStatus := updateDetails.OrderStatus

	switch shopOrder.Status {

	case domain.StatusReturnRequested:
		if newStatus == domain.StatusReturnApproved {
			if time.Since(updateDetails.ReturnDate) > 0 {
				return fmt.Errorf("given return date is invalid \nto update 'return approved' return date should be greater than cuurent time")
			}
			orderReturn.ApprovalDate = time.Now()
			orderReturn.IsApproved = true
			orderReturn.ReturnDate = updateDetails.ReturnDate
		} else if newStatus == domain.StatusReturnCancelled {
			// nothing extra update on order return may be in future when adding new statuses
		} else {
			return errors.New("order staus is return requested \nchange status must be return approved or return cancelled")
		}

	case domain.StatusReturnApproved:
		if newStatus != domain.StatusOrderReturned {
			return errors.New(" change status must be order returned")
		}
		if time.Since(updateDetails.ReturnDate) <= 0 {
			return fmt.Errorf("given return date is invalid \nto update 'order returned' return should be less than current time")
		} else {
			orderReturn.ReturnDate = updateDetails.ReturnDate
		}

	default:
		return fmt.Errorf("order status %s can't change to %s ", shopOrder.Status, newStatus)
	}

	orderReturn.AdminComment = updateDetails.AdminComment
	err = c.orderRepo.Transaction(func(trxRepo interfaces.OrderRepository) error {

		err := trxRepo.UpdateOrderReturn(ctx, orderReturn)
		if err != nil {
			return fmt.Errorf("failed to update orders return \nerror:%v", err.Error())
		}

		err = trxRepo.UpdateShopOrderStatus(ctx, shopOrder.ID, newStatus)
		if err != nil {
			return fmt.Errorf("failed to update order status \nerror:%v", err.Error())
		}

		// if order changing to order return then return the order amount to use wallet
		if newStatus == domain.StatusOrderReturned {
			// get user wallet
			wallet, err := trxRepo.FindWalletByUserID(ctx, shopOrder.UserID)
			if err != nil {
				return fmt.Errorf("failed to get user wallet for refund amount \nerror:%v", err.Error())
			}
			// if user have no wallet then create a new wallet for user
			if wallet.ID == "" {
				wallet.ID, err = c.orderRepo.SaveWallet(ctx, shopOrder.UserID)
				if err != nil {
					return fmt.Errorf("failed to create a wallet for user")
				}
			}

			// calculate wallet amount and update
			newWalletTotal, err := wallet.TotalAmount.Add(shopOrder.OrderTotal)
			if err != nil {
				return fmt.Errorf("failed to credit refund to user wallet \nerror:%v", err.Error())
			}
			err = trxRepo.UpdateWallet(ctx, wallet.ID, uint(newWalletTotal.AmountMinor))
			if err != nil {
				return fmt.Errorf("failed to update return amount to user wallet \nerror:%v", err.Error())
			}

			// wallet transaction
			transaction := domain.Transaction{
				WalletID:        wallet.ID,
				TransactionDate: time.Now(),
				TransactionType: domain.Credit,
				Amount:          shopOrder.OrderTotal,
			}
			err = trxRepo.SaveWalletTransaction(ctx, transaction)

			if err != nil {
				return fmt.Errorf("failed to save wallet transaction \nerror:%v", err.Error())
			}
		}
		return nil

	})

	if err != nil {
		return fmt.Errorf("failed to update order return \nerror:%v", err.Error())
	}

	log.Printf("successfully updated order return request for shop_order_id %v", shopOrder.ID)
	return nil
}

func (c *OrderUseCase) SubmitShoppingFeedback(ctx context.Context, feedbackDetails request.ShoppingFeedback) error {

	err := c.orderRepo.SaveShoppingFeedback(ctx, feedbackDetails)
	if err != nil {
		return utils.PrependMessageToError(err, "failed to save shopping feedback")
	}

	return nil
}
