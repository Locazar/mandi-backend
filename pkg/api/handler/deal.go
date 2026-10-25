package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// DealHandler serves the customer-facing deals & promotions discovery API.
type DealHandler struct {
	dealUseCase usecaseInterfaces.DealUseCase
}

func NewDealHandler(dealUseCase usecaseInterfaces.DealUseCase) *DealHandler {
	return &DealHandler{dealUseCase: dealUseCase}
}

// GetAllDeals godoc
//
//	@Summary     Browse all active deals
//	@Tags        Customer Deals
//	@Produce     json
//	@Param       deal_type    query string  false "Filter by type: percentage|fixed|bogo|bundle|flash|free_gift|free_shipping|loyalty"
//	@Param       min_discount query number  false "Minimum discount rate (e.g. 10)"
//	@Param       max_discount query number  false "Maximum discount rate (0 = no cap)"
//	@Param       category_id  query string  false "Filter by product category"
//	@Param       department_id query string false "Filter by department"
//	@Param       shop_id      query string  false "Filter by shop"
//	@Param       lat          query number  false "Customer latitude (enables nearby-only deals)"
//	@Param       lng          query number  false "Customer longitude"
//	@Param       radius_km    query number  false "Radius in km (default 10)"
//	@Param       sort_by      query string  false "discount_desc|discount_asc|ending_soon|newest"
//	@Param       page         query int     false "Page number (default 1)"
//	@Param       limit        query int     false "Page size (default 20, max 100)"
//	@Success     200 {object} response.PagedDeals
//	@Router      /deals/ [get]
func (h *DealHandler) GetAllDeals(ctx *gin.Context) {
	filter := request.ParseDealFilter(ctx)
	deals, total, err := h.dealUseCase.GetAllDeals(ctx, filter)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve deals", err, nil)
		return
	}
	lim := int(filter.Pagination.Limit)
	if lim == 0 {
		lim = 20
	}
	off := int(filter.Pagination.Offset)
	response.SuccessResponse(ctx, http.StatusOK, "Deals retrieved successfully", response.PagedDeals{
		Data:    deals,
		Total:   total,
		Page:    off/lim + 1,
		Limit:   lim,
		HasMore: int64(off+lim) < total,
	})
}

// GetDealFilters godoc
//
//	@Summary     Get filter metadata for the deals listing
//	@Description Returns deal-type chips, discount range buckets, and categories that currently have deals.
//	@Tags        Customer Deals
//	@Produce     json
//	@Success     200 {object} response.DealFilters
//	@Router      /deals/filters [get]
func (h *DealHandler) GetDealFilters(ctx *gin.Context) {
	filters, err := h.dealUseCase.GetDealFilters(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve deal filters", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Deal filters retrieved", filters)
}

// GetDealByID godoc
//
//	@Summary     Get a single deal by ID
//	@Tags        Customer Deals
//	@Produce     json
//	@Param       deal_id path string true "Deal ID"
//	@Success     200 {object} response.CustomerDeal
//	@Router      /deals/{deal_id} [get]
func (h *DealHandler) GetDealByID(ctx *gin.Context) {
	dealID := ctx.Param("deal_id")
	if dealID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "deal_id is required", nil, nil)
		return
	}
	deal, err := h.dealUseCase.GetDealByID(ctx, dealID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "Deal not found", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Deal retrieved", deal)
}

// GetProductsByDeal godoc
//
//	@Summary     Get all products under a specific deal
//	@Tags        Customer Deals
//	@Produce     json
//	@Param       deal_id path string true  "Deal ID"
//	@Param       page    query int   false "Page number (default 1)"
//	@Param       limit   query int   false "Page size (default 20)"
//	@Success     200 {object} response.PagedDealProducts
//	@Router      /deals/{deal_id}/products [get]
func (h *DealHandler) GetProductsByDeal(ctx *gin.Context) {
	dealID := ctx.Param("deal_id")
	if dealID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "deal_id is required", nil, nil)
		return
	}
	pagination := request.GetPagination(ctx)
	products, total, err := h.dealUseCase.GetProductsByDeal(ctx, dealID, pagination)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve products", err, nil)
		return
	}
	lim := int(pagination.Limit)
	if lim == 0 {
		lim = 20
	}
	off := int(pagination.Offset)
	response.SuccessResponse(ctx, http.StatusOK, "Products retrieved", response.PagedDealProducts{
		Data:    products,
		Total:   total,
		Page:    off/lim + 1,
		Limit:   lim,
		HasMore: int64(off+lim) < total,
	})
}

// GetProductsWithDeals godoc
//
//	@Summary     Browse products that have active deals
//	@Description Supports filtering by deal type, discount range, category, department, shop, and location.
//	@Tags        Customer Deals
//	@Produce     json
//	@Param       deal_type     query string false "percentage|fixed|bogo|bundle|flash|free_gift|free_shipping|loyalty"
//	@Param       min_discount  query number false "Min discount (e.g. 10)"
//	@Param       max_discount  query number false "Max discount (0 = no cap)"
//	@Param       category_id   query string false "Filter by category"
//	@Param       department_id query string false "Filter by department"
//	@Param       shop_id       query string false "Filter by shop"
//	@Param       lat           query number false "Customer latitude"
//	@Param       lng           query number false "Customer longitude"
//	@Param       radius_km     query number false "Radius in km (default 10)"
//	@Param       page          query int    false "Page number (default 1)"
//	@Param       limit         query int    false "Page size (default 20)"
//	@Success     200 {object} response.PagedDealProducts
//	@Router      /deals/products [get]
func (h *DealHandler) GetProductsWithDeals(ctx *gin.Context) {
	filter := request.ParseDealFilter(ctx)
	products, total, err := h.dealUseCase.GetProductsWithDeals(ctx, filter)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve products with deals", err, nil)
		return
	}
	lim := int(filter.Pagination.Limit)
	if lim == 0 {
		lim = 20
	}
	off := int(filter.Pagination.Offset)
	response.SuccessResponse(ctx, http.StatusOK, "Products with deals retrieved", response.PagedDealProducts{
		Data:    products,
		Total:   total,
		Page:    off/lim + 1,
		Limit:   lim,
		HasMore: int64(off+lim) < total,
	})
}
