package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
)

// LanguageHandler serves the selectable languages list to the client.
type LanguageHandler struct {
	useCase *usecase.LanguageUseCase
}

func NewLanguageHandler(useCase *usecase.LanguageUseCase) *LanguageHandler {
	return &LanguageHandler{useCase: useCase}
}

// GetLanguages godoc
//
//	@Summary		List selectable languages
//	@Security		BearerAuth
//	@Description	Active languages for the onboarding language picker, ordered.
//	@Tags			Languages
//	@Produce		json
//	@Router			/languages [get]
//	@Success		200	{object}	response.Response{}
func (h *LanguageHandler) GetLanguages(ctx *gin.Context) {
	items, err := h.useCase.ListActive(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch languages", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "ok", items)
}

// ---- Admin management (admin-authed) ----

// AdminListLanguages returns all languages (active + inactive) with full fields.
func (h *LanguageHandler) AdminListLanguages(ctx *gin.Context) {
	langs, err := h.useCase.ListAll(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch languages", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "ok", langs)
}

// AdminCreateLanguage inserts a new selectable language.
func (h *LanguageHandler) AdminCreateLanguage(ctx *gin.Context) {
	var req request.LanguageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}
	lang, err := h.useCase.Create(ctx.Request.Context(), req)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, err.Error(), err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Language created", lang)
}

// AdminUpdateLanguage patches an existing language by id.
func (h *LanguageHandler) AdminUpdateLanguage(ctx *gin.Context) {
	id := ctx.Param("id")
	var req request.LanguageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}
	lang, err := h.useCase.Update(ctx.Request.Context(), id, req)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, err.Error(), err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Language updated", lang)
}

// AdminDeleteLanguage removes a language by id.
func (h *LanguageHandler) AdminDeleteLanguage(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := h.useCase.Delete(ctx.Request.Context(), id); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, err.Error(), err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Language deleted", nil)
}
