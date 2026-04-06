package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type fcmTokenPayload struct {
	Token     string `json:"token" form:"token"`
	FCMToken  string `json:"fcm_token" form:"fcm_token"`
	FcmToken  string `json:"fcmToken" form:"fcmToken"`
	DeviceTok string `json:"device_token" form:"device_token"`
	RegToken  string `json:"registration_token" form:"registration_token"`
	Device    string `json:"device" form:"device"`
	DeviceTyp string `json:"device_type" form:"device_type"`
	Platform  string `json:"platform" form:"platform"`
	ShopID    uint   `json:"shop_id" form:"shop_id"`
	AdminID   uint   `json:"admin_id" form:"admin_id"`
	OwnerID   string `form:"owner_id"`
	OwnerType string `json:"owner_type" form:"owner_type"`
}

func coerceString(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatUint(uint64(typed), 10)
	case float32:
		return strconv.FormatUint(uint64(typed), 10)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case uint:
		return strconv.FormatUint(uint64(typed), 10)
	case uint64:
		return strconv.FormatUint(typed, 10)
	default:
		return ""
	}
}

func coerceUint(value interface{}) uint {
	switch typed := value.(type) {
	case string:
		parsed, err := strconv.ParseUint(strings.TrimSpace(typed), 10, 64)
		if err == nil {
			return uint(parsed)
		}
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			return uint(parsed)
		}
	case float64:
		return uint(typed)
	case float32:
		return uint(typed)
	case int:
		return uint(typed)
	case int64:
		return uint(typed)
	case uint:
		return typed
	case uint64:
		return uint(typed)
	}
	return 0
}

type FcmTokenHandler struct {
	usecase interfaces.FcmTokenUseCase
}

func NewFcmTokenHandler(usecase interfaces.FcmTokenUseCase) *FcmTokenHandler {
	return &FcmTokenHandler{usecase}
}

// SaveFcmToken godoc
//
//	@Summary		Save an FCM device token
//	@Security		BearerAuth
//	@ID				SaveFcmToken
//	@Tags			Admin FCM
//	@Accept			json
//	@Produce		json
//	@Param			input	body	domain.FcmToken	true	"FCM token payload"
//	@Router			/admin/fcm/token [post]
//	@Success		200	{object}	domain.FcmToken	"Token saved"
//	@Failure		400	{object}	map[string]string	"Invalid input"
//	@Failure		500	{object}	map[string]string	"Internal server error"
func (h *FcmTokenHandler) SaveFcmToken(c *gin.Context) {
	var payload fcmTokenPayload
	var raw map[string]interface{}

	bodyBytes, _ := c.GetRawData()
	if len(bodyBytes) > 0 {
		decoder := json.NewDecoder(strings.NewReader(string(bodyBytes)))
		decoder.UseNumber()
		if err := decoder.Decode(&raw); err == nil {
			payload.Token = coerceString(raw["token"])
			payload.FCMToken = coerceString(raw["fcm_token"])
			payload.FcmToken = coerceString(raw["fcmToken"])
			payload.DeviceTok = coerceString(raw["device_token"])
			payload.RegToken = coerceString(raw["registration_token"])
			payload.Device = coerceString(raw["device"])
			payload.DeviceTyp = coerceString(raw["device_type"])
			payload.Platform = coerceString(raw["platform"])
			payload.ShopID = coerceUint(raw["shop_id"])
			payload.AdminID = coerceUint(raw["admin_id"])
			payload.OwnerID = coerceString(raw["owner_id"])
			payload.OwnerType = coerceString(raw["owner_type"])
		}
	}

	if payload.Token == "" {
		payload.Token = c.PostForm("token")
	}
	if payload.FCMToken == "" {
		payload.FCMToken = c.PostForm("fcm_token")
	}
	if payload.FcmToken == "" {
		payload.FcmToken = c.PostForm("fcmToken")
	}
	if payload.DeviceTok == "" {
		payload.DeviceTok = c.PostForm("device_token")
	}
	if payload.RegToken == "" {
		payload.RegToken = c.PostForm("registration_token")
	}
	if payload.Device == "" {
		payload.Device = c.PostForm("device")
	}
	if payload.DeviceTyp == "" {
		payload.DeviceTyp = c.PostForm("device_type")
	}
	if payload.Platform == "" {
		payload.Platform = c.PostForm("platform")
	}
	if payload.OwnerID == "" {
		payload.OwnerID = c.PostForm("owner_id")
	}
	if payload.OwnerType == "" {
		payload.OwnerType = c.PostForm("owner_type")
	}
	if payload.ShopID == 0 {
		payload.ShopID = coerceUint(c.PostForm("shop_id"))
	}
	if payload.AdminID == 0 {
		payload.AdminID = coerceUint(c.PostForm("admin_id"))
	}

	token := strings.TrimSpace(payload.Token)
	if token == "" {
		token = strings.TrimSpace(payload.FCMToken)
	}
	if token == "" {
		token = strings.TrimSpace(payload.FcmToken)
	}
	if token == "" {
		token = strings.TrimSpace(payload.DeviceTok)
	}
	if token == "" {
		token = strings.TrimSpace(payload.RegToken)
	}
	if token == "" {
		token = strings.TrimSpace(c.Query("token"))
	}
	if token == "" {
		token = strings.TrimSpace(c.GetHeader("X-FCM-Token"))
	}
	if token == "" {
		token = strings.TrimSpace(c.GetHeader("X-Device-Token"))
	}

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing FCM token. Expected one of: token, fcm_token, fcmToken, device_token, registration_token (or query token / header X-FCM-Token)",
			"example": gin.H{
				"token":      "<fcm-token>",
				"platform":   "android",
				"admin_id":   1,
				"owner_type": "seller",
			},
		})
		return
	}

	device := strings.TrimSpace(payload.Device)
	if device == "" {
		device = strings.TrimSpace(payload.DeviceTyp)
	}

	fcmToken := domain.FcmToken{
		Token:     token,
		Device:    device,
		Platform:  strings.TrimSpace(payload.Platform),
		ShopID:    payload.ShopID,
		AdminID:   payload.AdminID,
		OwnerID:   strings.TrimSpace(payload.OwnerID),
		OwnerType: strings.TrimSpace(payload.OwnerType),
	}

	if fcmToken.OwnerID == "" {
		if fcmToken.ShopID != 0 {
			fcmToken.OwnerID = strconv.FormatUint(uint64(fcmToken.ShopID), 10)
		} else if fcmToken.AdminID != 0 {
			fcmToken.OwnerID = strconv.FormatUint(uint64(fcmToken.AdminID), 10)
		} else {
			ctxUserID := utils.GetUserIdFromContext(c)
			if ctxUserID != 0 {
				fcmToken.AdminID = ctxUserID
				fcmToken.OwnerID = strconv.FormatUint(uint64(ctxUserID), 10)
			}
		}
	}

	if fcmToken.OwnerType == "" {
		fcmToken.OwnerType = "seller"
	}

	if fcmToken.Token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing FCM token: provide token or fcm_token or fcmToken"})
		return
	}

	if fcmToken.OwnerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing owner context: provide owner_id/shop_id/admin_id or valid auth token"})
		return
	}

	fcmToken, err := h.usecase.SaveFcmToken(fcmToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fcmToken)
}
