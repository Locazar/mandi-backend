package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type CartRepository interface {
	FindCartByUserID(ctx context.Context, userID string) (cart domain.Cart, err error)
	SaveCart(ctx context.Context, userID string) (cartID string, err error)
	UpdateCart(ctx context.Context, cartId string, discountAmount string, couponID string) error

	FindCartItemByCartAndProductItemID(ctx context.Context, cartID, productItemID string) (cartItem domain.CartItem, err error)
	FindAllCartItemsByCartID(ctx context.Context, cartID string) (cartItems []response.CartItem, err error)
	SaveCartItem(ctx context.Context, cartId, productItemId string) error
	DeleteCartItem(ctx context.Context, cartItemID string) error
	DeleteAllCartItemsByCartID(ctx context.Context, cartID string) error
	UpdateCartItemQty(ctx context.Context, cartItemId string, qty uint) error

	IsCartValidForOrder(ctx context.Context, userID string) (valid bool, err error)
}
