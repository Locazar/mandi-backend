package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type CartUseCase interface {
	SaveProductItemToCart(ctx context.Context, userID, productItemId string) error         // save product_item to cart
	RemoveProductItemFromCartItem(ctx context.Context, userID, productItemId string) error // remove product_item from cart
	UpdateCartItem(ctx context.Context, updateDetails request.UpdateCartItem) error        // edit cartItems( quantity change )
	GetUserCart(ctx context.Context, userID string) (cart domain.Cart, err error)
	GetUserCartItems(ctx context.Context, cartId string) (cartItems []response.CartItem, err error)
}
