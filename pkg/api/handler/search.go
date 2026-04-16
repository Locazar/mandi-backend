package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type SearchHandler struct {
	searchUseCase usecaseInterface.SearchUseCase
}

func NewSearchHandler(searchUseCase usecaseInterface.SearchUseCase) interfaces.SearchHandler {
	return &SearchHandler{searchUseCase: searchUseCase}
}

// GlobalSearch godoc
// @Summary		Global search across products, categories, shops, brands, departments, and offers
// @Description	Searches across all entity types simultaneously. Returns grouped results.
// @Tags		Global Search
// @Accept		json
// @Produce		json
// @Param		q		query	string	true	"Search query (min 1 char)"
// @Param		types	query	string	false	"Comma-separated entity types filter (products,categories,shops,brands,departments,offers)"
// @Param		lat		query	number	false	"Latitude for geo-filtering shops"
// @Param		lng		query	number	false	"Longitude for geo-filtering shops"
// @Param		radius	query	number	false	"Radius in km for geo-filtering shops"
// @Param		pincode	query	string	false	"Pincode for location-based shop filtering"
// @Param		limit	query	int		false	"Results per entity type (default 5, max 50)"
// @Param		offset	query	int		false	"Offset for pagination (default 0)"
// @Success		200	{object}	response.Response{data=response.GlobalSearchResult}
// @Failure		400	{object}	response.Response
// @Router		/global-search [get]
func (h *SearchHandler) GlobalSearch(ctx *gin.Context) {
	var req request.GlobalSearchRequest

	req.Query = ctx.Query("q")
	if req.Query == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "search query 'q' is required", nil, nil)
		return
	}

	req.Types = ctx.Query("types")
	req.Pincode = ctx.Query("pincode")

	if lat := ctx.Query("lat"); lat != "" {
		if v, err := strconv.ParseFloat(lat, 64); err == nil {
			req.Latitude = v
		}
	}
	if lng := ctx.Query("lng"); lng != "" {
		if v, err := strconv.ParseFloat(lng, 64); err == nil {
			req.Longitude = v
		}
	}
	if radius := ctx.Query("radius"); radius != "" {
		if v, err := strconv.ParseFloat(radius, 64); err == nil {
			req.RadiusKm = v
		}
	}
	if limit := ctx.Query("limit"); limit != "" {
		if v, err := strconv.ParseUint(limit, 10, 64); err == nil {
			req.Limit = uint(v)
		}
	}
	if offset := ctx.Query("offset"); offset != "" {
		if v, err := strconv.ParseUint(offset, 10, 64); err == nil {
			req.Offset = uint(v)
		}
	}

	result, err := h.searchUseCase.GlobalSearch(ctx.Request.Context(), req)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "failed to perform global search", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "global search results retrieved successfully", result)
}

// Autocomplete godoc
// @Summary		Autocomplete suggestions across all entity types
// @Description	Returns fast autocomplete suggestions from products, categories, shops, brands, departments, and offers.
// @Tags		Global Search
// @Accept		json
// @Produce		json
// @Param		q		query	string	true	"Search prefix (min 1 char)"
// @Param		limit	query	int		false	"Max suggestions (default 10, max 20)"
// @Success		200	{object}	response.Response{data=[]response.AutocompleteSuggestion}
// @Failure		400	{object}	response.Response
// @Router		/global-search/autocomplete [get]
func (h *SearchHandler) Autocomplete(ctx *gin.Context) {
	var req request.AutocompleteRequest

	req.Query = ctx.Query("q")
	if req.Query == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "search prefix 'q' is required", nil, nil)
		return
	}

	if limit := ctx.Query("limit"); limit != "" {
		if v, err := strconv.ParseUint(limit, 10, 64); err == nil {
			req.Limit = uint(v)
		}
	}

	suggestions, err := h.searchUseCase.Autocomplete(ctx.Request.Context(), req)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "failed to get autocomplete suggestions", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "autocomplete suggestions retrieved successfully", suggestions)
}
