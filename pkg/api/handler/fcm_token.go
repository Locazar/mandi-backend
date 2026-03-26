package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type FcmTokenHandler struct {
	usecase interfaces.FcmTokenUseCase
}

func NewFcmTokenHandler(usecase interfaces.FcmTokenUseCase) *FcmTokenHandler {
	return &FcmTokenHandler{usecase}
}

func (h *FcmTokenHandler) SaveFcmToken(c *gin.Context) {
	var fcmToken domain.FcmToken
	if err := c.ShouldBindJSON(&fcmToken); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fcmToken, err := h.usecase.SaveFcmToken(fcmToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fcmToken)
}
