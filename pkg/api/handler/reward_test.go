package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	"github.com/stretchr/testify/assert"
)

func TestRewardSentinelStatuses(t *testing.T) {
	want := map[error]int{
		usecase.ErrRewardProgramDisabled:  http.StatusServiceUnavailable,
		usecase.ErrRewardCustomerOnly:     http.StatusForbidden,
		usecase.ErrRewardShopNotFound:     http.StatusNotFound,
		usecase.ErrShopNotOptedIn:         http.StatusUnprocessableEntity,
		usecase.ErrShopLocationMissing:    http.StatusUnprocessableEntity,
		usecase.ErrTooFarFromShop:         http.StatusUnprocessableEntity,
		usecase.ErrSelfPurchase:           http.StatusForbidden,
		usecase.ErrRepeatPurchaseWindow:   http.StatusConflict,
		usecase.ErrShopDailyCapReached:    http.StatusTooManyRequests,
		usecase.ErrBelowMinBill:           http.StatusUnprocessableEntity,
		usecase.ErrInvalidBillAmount:      http.StatusBadRequest,
		usecase.ErrInvalidLocation:        http.StatusBadRequest,
		usecase.ErrPurchaseNotFound:       http.StatusNotFound,
		usecase.ErrPurchaseNotPending:     http.StatusConflict,
		usecase.ErrInsufficientPoints:     http.StatusUnprocessableEntity,
		usecase.ErrInvalidDiscountPercent: http.StatusBadRequest,
		usecase.ErrInvalidRewardConfig:    http.StatusBadRequest,
		usecase.ErrInvalidAdjustment:      http.StatusBadRequest,
		usecase.ErrRewardAccountNotFound:  http.StatusNotFound,
		usecase.ErrRewardAdminOnly:        http.StatusForbidden,
	}
	for err, code := range want {
		assert.Equal(t, code, sentinelStatus[err], err.Error())
	}
}

func rewardCtx(method, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userId", "usr_1")
	return c, w
}

// Binding failures must be rejected before the usecase is touched (nil here).
func TestRewardHandler_SubmitPurchaseBadBodies(t *testing.T) {
	h := NewRewardHandler(nil)
	for name, body := range map[string]string{
		"not json":           `{`,
		"missing latitude":   `{"shop_id":"shp_1","bill_amount_paise":50000,"longitude":77.6,"client_request_id":"a"}`,
		"negative bill":      `{"shop_id":"shp_1","bill_amount_paise":-1,"latitude":12.9,"longitude":77.6,"client_request_id":"a"}`,
		"missing request id": `{"shop_id":"shp_1","bill_amount_paise":50000,"latitude":12.9,"longitude":77.6}`,
	} {
		t.Run(name, func(t *testing.T) {
			c, w := rewardCtx(http.MethodPost, body)
			h.SubmitPurchase(c)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestRewardHandler_EligibilityRequiresCoordinates(t *testing.T) {
	h := NewRewardHandler(nil)
	c, w := rewardCtx(http.MethodGet, "")
	c.Params = gin.Params{{Key: "shop_id", Value: "shp_1"}}
	c.Request.URL.RawQuery = "lat=abc&lng=77.6"
	h.GetEligibility(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRewardHandler_UpdateSettingsRequiresEnabledFlag(t *testing.T) {
	h := NewRewardHandler(nil)
	c, w := rewardCtx(http.MethodPut, `{"discount_percent":10}`)
	h.UpdateSellerSettings(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRewardHandler_AdjustRejectsOutOfRangeDelta(t *testing.T) {
	h := NewRewardHandler(nil)
	for name, body := range map[string]string{
		"too large":           `{"delta_points":100001,"reason":"typo with extra zero"}`,
		"too small":           `{"delta_points":-100001,"reason":"typo with extra zero"}`,
		"request id too long": `{"delta_points":10,"reason":"goodwill credit","client_request_id":"` + strings.Repeat("x", 41) + `"}`,
	} {
		t.Run(name, func(t *testing.T) {
			c, w := rewardCtx(http.MethodPost, body)
			c.Params = gin.Params{{Key: "account_id", Value: "rwa_1"}}
			h.AdjustAccount(c)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
