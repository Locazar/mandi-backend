package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
