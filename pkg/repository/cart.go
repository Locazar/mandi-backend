package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type cartDatabase struct {
	DB *gorm.DB
}

func NewCartRepository(db *gorm.DB) interfaces.CartRepository {
	return &cartDatabase{
		DB: db,
	}
}

// find a cartItem
func (c *cartDatabase) FindCartByUserID(ctx context.Context, userID string) (cart domain.Cart, err error) {

	query := `SELECT * FROM carts WHERE user_id = ?`
	err = c.DB.Raw(query, userID).Scan(&cart).Error

	return
}

// save cart for user
func (c *cartDatabase) SaveCart(ctx context.Context, userID string) (cartID string, err error) {
	cartID = domain.NewID(domain.PrefixCart)
	query := `INSERT INTO carts (id, user_id,total_price_amount_minor,total_price_currency) VALUES($1, $2, $3, $4)`
	err = c.DB.Exec(query, cartID, userID, 0, domain.CurrencyINR).Error

	return cartID, err
}

func (c *cartDatabase) UpdateCart(ctx context.Context, cartId string, discountAmount string, couponID string) error {

	query := `UPDATE carts SET discount_amount_amount_minor = $1, discount_amount_currency = $2, applied_coupon_id = $3 WHERE id = $4`
	err := c.DB.Exec(query, discountAmount, domain.CurrencyINR, couponID, cartId).Error

	return err
}

// find cart_items
func (c *cartDatabase) FindCartItemByID(ctx context.Context, cartItemID string) (cartItem domain.CartItem, err error) {
	query := `SELECT * FROM cart_items WHERE id = ?`
	err = c.DB.Raw(query, cartItemID).Scan(&cartItem).Error

	return
}

func (c *cartDatabase) FindCartItemByCartAndProductItemID(ctx context.Context, cartID, productItemID string) (cartItem domain.CartItem, err error) {
	query := `SELECT * FROM cart_items WHERE cart_id = $1 AND product_item_id = $2`
	err = c.DB.Raw(query, cartID, productItemID).Scan(&cartItem).Error

	return cartItem, err
}

func (c *cartDatabase) SaveCartItem(ctx context.Context, cartId, productItemId string) error {
	query := `INSERT INTO cart_items (id, cart_id, product_item_id, qty) VALUES ($1, $2, $3, $4)`
	err := c.DB.Exec(query, domain.NewID(domain.PrefixCartItem), cartId, productItemId, 1).Error

	return err
}

func (c *cartDatabase) DeleteCartItem(ctx context.Context, cartItemID string) error {

	query := `DELETE FROM cart_items WHERE id = $1`
	err := c.DB.Exec(query, cartItemID).Error

	return err
}

func (c *cartDatabase) DeleteAllCartItemsByCartID(ctx context.Context, cartID string) error {

	query := ` DELETE FROM cart_items WHERE cart_id = $1`
	err := c.DB.Exec(query, cartID).Error
	return err
}

func (c *cartDatabase) UpdateCartItemQty(ctx context.Context, cartItemId string, qty uint) error {

	query := `UPDATE cart_items SET qty = $1 WHERE id = $2`
	err := c.DB.Exec(query, qty, cartItemId).Error

	return err
}

func (c *cartDatabase) FindAllCartItemsByCartID(ctx context.Context, cartID string) (cartItems []response.CartItem, err error) {

	// get the cartItem of all user with subtotal
	query := `SELECT ci.product_item_id, p.name AS product_name, ci.qty,pi.price ,
	 pi.discount_price,
	 CASE WHEN pi.discount_price > 0 THEN pi.discount_price * ci.qty ELSE pi.price * ci.qty END AS sub_total   
	 FROM cart_items ci INNER JOIN product_items pi ON ci.product_item_id = pi.id 
	 INNER JOIN products p ON pi.product_id = p.id AND ci.cart_id=?`

	err = c.DB.Raw(query, cartID).Scan(&cartItems).Error

	return
}

func (c *cartDatabase) IsCartValidForOrder(ctx context.Context, userID string) (valid bool, err error) {

	var outOfStockExist bool
	query := `SELECT 
		EXISTS( SELECT DISTINCT pi.id FROM product_items pi 
		INNER JOIN cart_items ci ON pi.id = ci.product_item_id 
		INNER JOIN carts c ON ci.cart_id = c.id 
		WHERE c.user_id = $1) AS valid FROM carts`

	err = c.DB.Raw(query, userID).Scan(&outOfStockExist).Error

	// if error or a product is found a product is out
	if err != nil || outOfStockExist {
		return false, err
	}

	return true, nil
}
