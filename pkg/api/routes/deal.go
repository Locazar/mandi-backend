package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler"
)

// DealRoutes registers the customer-facing deals & promotions discovery API.
// All endpoints are public (no auth required) so the customer app can browse
// deals before logging in.
func DealRoutes(api *gin.RouterGroup, dealHandler *handler.DealHandler) {
	deals := api.Group("/deals")
	{
		// GET /api/deals/
		// Browse all active deals with optional filters:
		//   ?deal_type=percentage|fixed|bogo|bundle|flash|free_gift|free_shipping|loyalty
		//   ?min_discount=10&max_discount=50
		//   ?category_id=cat_xxx
		//   ?shop_id=shp_xxx
		//   ?sort_by=discount_desc|discount_asc|ending_soon|newest
		//   ?page=1&limit=20
		deals.GET("/", dealHandler.GetAllDeals)

		// GET /api/deals/filters
		// Returns deal-type chips, discount range buckets, and categories with active deals.
		// Customer app uses this to populate the filter bar without a separate count query.
		deals.GET("/filters", dealHandler.GetDealFilters)

		// GET /api/deals/products
		// Products that have at least one active deal, filterable by deal type / discount band.
		// Powers "Products on Sale" screen.
		deals.GET("/products", dealHandler.GetProductsWithDeals)

		// GET /api/deals/:deal_id
		// Single deal detail — title, discount, dates, shop, product count, type-specific fields.
		deals.GET("/:deal_id", dealHandler.GetDealByID)

		// GET /api/deals/:deal_id/products
		// Paginated list of all products carrying this deal.
		deals.GET("/:deal_id/products", dealHandler.GetProductsByDeal)
	}
}
