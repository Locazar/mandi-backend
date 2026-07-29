package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
)

// ShopUpdateHandler serves the per-shop "What's New" advertisement feature:
// customer read (GET /api/shop-updates) + admin CRUD (/api/admin/shop-updates).
type ShopUpdateHandler struct {
	useCase      *usecase.ShopUpdateUseCase
	cloudService cloud.CloudService
}

func NewShopUpdateHandler(useCase *usecase.ShopUpdateUseCase, cloudService cloud.CloudService) *ShopUpdateHandler {
	return &ShopUpdateHandler{useCase: useCase, cloudService: cloudService}
}

// ── Customer read ────────────────────────────────────────────────────────────

// GetShopUpdatesForUser godoc
//
//	@Summary		Shop Updates for a customer (What's New nearby)
//	@Security		BearerAuth
//	@Description	Returns enabled shop-update cards, sorted by proximity when lat/long+radius are given, else by pincode, else generic/latest. Empty data is valid.
//	@Tags			Shop Updates
//	@Produce		json
//	@Param			lat		query	number	false	"Latitude"
//	@Param			long	query	number	false	"Longitude"
//	@Param			radius	query	number	false	"Radius in km"
//	@Param			pincode	query	string	false	"Pincode (used when lat/long absent)"
//	@Router			/shop-updates [get]
//	@Success		200	{object}	response.Response{}
func (h *ShopUpdateHandler) GetShopUpdatesForUser(ctx *gin.Context) {
	lat, _ := strconv.ParseFloat(ctx.Query("lat"), 64)
	lng, _ := strconv.ParseFloat(ctx.Query("long"), 64)
	radius, _ := strconv.ParseFloat(ctx.Query("radius"), 64)
	pincode := strings.TrimSpace(ctx.Query("pincode"))

	cards, err := h.useCase.GetForUser(ctx.Request.Context(), lat, lng, radius, pincode)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch shop updates", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "ok", cards)
}

// ── Admin CRUD ───────────────────────────────────────────────────────────────

func (h *ShopUpdateHandler) CreateShopUpdate(ctx *gin.Context) {
	su := domain.ShopUpdate{}
	if err := h.bindEntryForm(ctx, &su, true); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, err.Error(), err, nil)
		return
	}
	if su.ShopID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "shop_id is required", fmt.Errorf("missing shop_id"), nil)
		return
	}
	created, err := h.useCase.Create(ctx.Request.Context(), su)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create shop update", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Shop update created", created)
}

func (h *ShopUpdateHandler) ListShopUpdates(ctx *gin.Context) {
	updates, err := h.useCase.List(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list shop updates", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "ok", updates)
}

func (h *ShopUpdateHandler) GetShopUpdate(ctx *gin.Context) {
	su, err := h.useCase.GetByID(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "Shop update not found", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "ok", su)
}

func (h *ShopUpdateHandler) UpdateShopUpdate(ctx *gin.Context) {
	id := ctx.Param("id")
	existing, err := h.useCase.GetByID(ctx.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "Shop update not found", err, nil)
		return
	}
	if err := h.bindEntryForm(ctx, &existing, false); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, err.Error(), err, nil)
		return
	}
	updated, err := h.useCase.Update(ctx.Request.Context(), existing)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update shop update", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Shop update updated", updated)
}

// bindEntryForm reads shop-update entry fields from multipart/form-data. On
// create, unset fields take their defaults (is_active=true, action_label
// filled by the repo); on update, only present fields overwrite existing
// values. An attached "image" file is uploaded and its bare key stored as the
// per-entry shop-image override.
func (h *ShopUpdateHandler) bindEntryForm(ctx *gin.Context, su *domain.ShopUpdate, isCreate bool) error {
	if v := strings.TrimSpace(ctx.PostForm("shop_id")); v != "" {
		su.ShopID = v
	}
	if v := strings.TrimSpace(ctx.PostForm("action_label")); v != "" {
		su.ActionLabel = v
	}
	if v := ctx.PostForm("sort_order"); v != "" {
		su.SortOrder, _ = strconv.Atoi(v)
	}
	if isCreate {
		su.IsActive = true
	}
	if v := ctx.PostForm("is_active"); v != "" {
		su.IsActive = v == "true" || v == "1"
	}

	if fh, err := ctx.FormFile("image"); err == nil && fh != nil {
		if fh.Size > 10<<20 {
			return fmt.Errorf("image exceeds 10MB limit")
		}
		contentType := fh.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		key, uploadErr := h.cloudService.SaveFile(ctx, fh, cloud.SaveOptions{
			Namespace:   "shop-updates/entries",
			Visibility:  cloud.VisibilityPublic,
			ContentType: contentType,
			Filename:    fmt.Sprintf("%d-%s", time.Now().Unix(), fh.Filename),
		})
		if uploadErr != nil {
			return fmt.Errorf("failed to upload image: %w", uploadErr)
		}
		su.ImageURL = key
	} else if v := strings.TrimSpace(ctx.PostForm("image_url")); v != "" {
		su.ImageURL = v
	}
	return nil
}

func (h *ShopUpdateHandler) DeleteShopUpdate(ctx *gin.Context) {
	if err := h.useCase.Delete(ctx.Request.Context(), ctx.Param("id")); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete shop update", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Shop update deleted", nil)
}

// ── Product items (multipart: fields + optional image file) ──────────────────

func (h *ShopUpdateHandler) AddProduct(ctx *gin.Context) {
	shopUpdateID := ctx.Param("id")
	p := domain.ShopUpdateProduct{ShopUpdateID: shopUpdateID}
	if err := h.bindProductForm(ctx, &p, true); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, err.Error(), err, nil)
		return
	}
	created, err := h.useCase.AddProduct(ctx.Request.Context(), p)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to add product", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Product added", created)
}

func (h *ShopUpdateHandler) UpdateProduct(ctx *gin.Context) {
	p := domain.ShopUpdateProduct{
		ID:           ctx.Param("productId"),
		ShopUpdateID: ctx.Param("id"),
	}
	if err := h.bindProductForm(ctx, &p, false); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, err.Error(), err, nil)
		return
	}
	updated, err := h.useCase.UpdateProduct(ctx.Request.Context(), p)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update product", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Product updated", updated)
}

func (h *ShopUpdateHandler) DeleteProduct(ctx *gin.Context) {
	if err := h.useCase.DeleteProduct(ctx.Request.Context(), ctx.Param("productId")); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete product", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Product deleted", nil)
}

// bindProductForm reads product fields from the multipart/form-data body and,
// if an "image" file is attached, uploads it to object storage (bare key
// stored). isCreate controls the is_active default (true on create).
func (h *ShopUpdateHandler) bindProductForm(ctx *gin.Context, p *domain.ShopUpdateProduct, isCreate bool) error {
	p.ProductItemID = strings.TrimSpace(ctx.PostForm("product_item_id"))
	p.Title = strings.TrimSpace(ctx.PostForm("title"))
	p.Attribute = strings.TrimSpace(ctx.PostForm("attribute"))
	if v := ctx.PostForm("sort_order"); v != "" {
		p.SortOrder, _ = strconv.Atoi(v)
	}

	p.IsActive = isCreate // default true on create, false on update unless set
	if v := ctx.PostForm("is_active"); v != "" {
		p.IsActive = v == "true" || v == "1"
	}

	// Optional image: file upload takes precedence over an image_url form value.
	if fh, err := ctx.FormFile("image"); err == nil && fh != nil {
		if fh.Size > 10<<20 {
			return fmt.Errorf("image exceeds 10MB limit")
		}
		contentType := fh.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		key, uploadErr := h.cloudService.SaveFile(ctx, fh, cloud.SaveOptions{
			Namespace:   fmt.Sprintf("shop-updates/products/%s", p.ShopUpdateID),
			Visibility:  cloud.VisibilityPublic,
			ContentType: contentType,
			Filename:    fmt.Sprintf("%d-%s", time.Now().Unix(), fh.Filename),
		})
		if uploadErr != nil {
			return fmt.Errorf("failed to upload image: %w", uploadErr)
		}
		p.ImageURL = key
	} else if v := strings.TrimSpace(ctx.PostForm("image_url")); v != "" {
		p.ImageURL = v
	}

	if isCreate && p.ProductItemID == "" && p.Title == "" {
		return fmt.Errorf("either product_item_id or title is required")
	}
	return nil
}
