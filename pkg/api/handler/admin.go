package handler

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type adminHandler struct {
	adminUseCase    usecaseInterface.AdminUseCase
	shopTimeUseCase usecaseInterface.ShopTimeUseCase
}

// UserLogout implements the UserLogout method required by the AdminHandler interface.
func (a *adminHandler) UserLogout(ctx *gin.Context) {
	// Implementation goes here, or just a stub if not needed yet
	response.SuccessResponse(ctx, http.StatusOK, "Successfully logged out", nil)
}

func NewAdminHandler(adminUsecase usecaseInterface.AdminUseCase, shopTimeUseCase usecaseInterface.ShopTimeUseCase) interfaces.AdminHandler {
	return &adminHandler{
		adminUseCase:    adminUsecase,
		shopTimeUseCase: shopTimeUseCase,
	}
}

// AdminSignUp godoc
// @summary api for admin to login
// @id AdminSignUp
// @tags Admin SignUp
// @Param input body domain.Admin{} true "inputs"
// @Router /admin/signUp [post]
// @Success 200 {object} response.Response{} "successfully logged in"
// @Failure 400 {object} response.Response{} "invalid input"
// @Failure 500 {object} response.Response{} "failed to generate jwt token"
func (a *adminHandler) AdminSignUp(ctx *gin.Context) {

	var body domain.Admin

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	// Validate mobile number
	if body.Mobile == "" || body.Mobile == "null" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Mobile number is required", nil, nil)
		return
	}
	// Validate mobile number (should be 10 digits)
	if matched, _ := regexp.MatchString(`^\d{10}$`, body.Mobile); !matched {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid mobile number. Must be 10 digits.", nil, nil)
		return
	}

	otpID, err := a.adminUseCase.SignUp(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to create account for admin", err, nil)
		return
	}

	responseData := map[string]interface{}{
		"otp_id":  otpID,
		"message": "OTP sent to mobile number for verification",
	}

	response.SuccessResponse(ctx, 200, "Successfully account created for admin", responseData)
}

// GetAdminWithShopVerificationByPhone godoc
// @summary api to get admin with shop verification data by phone
// @id GetAdminWithShopVerificationByPhone
// @tags Admin
// @Param phone query string true "Admin phone number"
// @Router /admin/profile [get]
// @Success 200 {object} response.Response{} "successfully retrieved admin with shop verification"
// @Failure 400 {object} response.Response{} "invalid phone number"
// @Failure 500 {object} response.Response{} "failed to retrieve admin data"
func (a *adminHandler) GetAdminWithShopVerificationByPhone(ctx *gin.Context) {
	phone := ctx.Query("phone")
	if phone == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Phone number is required", nil, nil)
		return
	}

	adminData, shopVerificationData, err := a.adminUseCase.GetAdminWithShopVerificationByPhone(ctx, phone)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve admin data", err, nil)
		return
	}

	responseData := response.ConvertAdminToResponse(adminData, shopVerificationData)
	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved admin with shop verification", responseData)
}

func (a *adminHandler) AdminSignUpVerify(ctx *gin.Context) {
	var body request.OTPVerify
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	// get the user using loginOtp useCase
	userID, err := a.adminUseCase.AdminSignUpOtpVerify(ctx, body)
	println("userID:", userID)
	if err != nil {
		var statusCode int
		switch {
		case errors.Is(err, usecase.ErrOtpExpired):
			statusCode = http.StatusGone
		case errors.Is(err, usecase.ErrInvalidOtp):
			statusCode = http.StatusUnauthorized
		default:
			statusCode = http.StatusInternalServerError
		}
		response.ErrorResponse(ctx, statusCode, "Failed to verify otp", err, nil)
		return
	}
	fmt.Printf("userID: %d\n", userID)
	a.setupTokenAndResponse(ctx, token.Admin, userID)
}

// access and refresh token generating for user and admin is same so created
// a common function for it.(differentiate user by user type )
// customResponse is optional - if provided, it will be used instead of default success response
func (c *adminHandler) setupTokenAndResponse(ctx *gin.Context, tokenUser token.UserType, userID uint, customResponse ...interface{}) {

	tokenParams := usecaseInterface.GenerateTokenParams{
		UserID:   userID,
		UserType: tokenUser,
	}
	fmt.Printf("Generating tokens for userID: %d, userType: %s\n", userID, tokenUser)
	accessToken, err := c.adminUseCase.GenerateAccessToken(ctx, tokenParams)
	fmt.Printf("Access token generation result for userID: %d, userType: %s, accessToken: %s, error: %v\n", userID, tokenUser, accessToken, err)
	if err != nil {
		fmt.Printf("Error generating access token for userID: %d, userType: %s, error: %v\n", userID, tokenUser, err)
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to generate access token", err, nil)
		return
	}

	refreshToken, err := c.adminUseCase.GenerateRefreshToken(ctx, usecaseInterface.GenerateTokenParams{
		UserID:   userID,
		UserType: tokenUser,
	})

	if err != nil {
		fmt.Printf("Error generating refresh token for userID: %d, userType: %s, error: %v\n", userID, tokenUser, err)
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to generate refresh token", err, nil)
		return
	}

	authorizationValue := authorizationType + " " + accessToken
	ctx.Header(authorizationHeaderKey, authorizationValue)

	ctx.Header("access_token", accessToken)
	ctx.Header("refresh_token", refreshToken)
	fmt.Printf("Set access and refresh tokens in headers for userID: %d, userType: %s\n", userID, tokenUser)

	tokenRes := response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	// Merge custom response with token response if provided
	var responseData interface{} = tokenRes
	var message string = "Successfully logged in"

	if len(customResponse) > 0 {
		// Create merged response combining tokenRes and custom data
		mergedData := map[string]interface{}{
			"tokens": tokenRes,
		}

		// Add custom data to the merged response
		if customData, ok := customResponse[0].(map[string]interface{}); ok {
			for key, value := range customData {
				mergedData[key] = value
			}
		} else {
			mergedData["data"] = customResponse[0]
		}

		responseData = mergedData
	}

	if len(customResponse) > 1 {
		if msg, ok := customResponse[1].(string); ok {
			message = msg
		}
	}
	fmt.Printf("Final response data for userID: %d, userType: %s, responseData: %+v\n", userID, tokenUser, responseData)
	response.SuccessResponse(ctx, http.StatusOK, message, responseData)
}

// GetAllUsers godoc
//
//	@Summary		Get all users
//	@Security		BearerAuth
//	@Description	API for admin to get all user details
//	@Id				GetAllUsers
//	@Tags			Admin User
//	@Param			page_number	query	int	false	"Page Number"
//	@Param			count		query	int	false	"Count"
//	@Router			/admin/users [get]
//	@Success		200	{object}	response.Response{}	"Successfully got all users"
//	@Success		204	{object}	response.Response{}	"No users found"
//	@Failure		500	{object}	response.Response{}	"Failed to find all users"
func (a *adminHandler) GetAllUsers(ctx *gin.Context) {

	pagination := request.GetPagination(ctx)

	users, err := a.adminUseCase.FindAllUser(ctx, pagination)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to find all users", err, nil)
		return
	}

	if len(users) == 0 {
		response.SuccessResponse(ctx, http.StatusNoContent, "No users found", []interface{}{})
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully found all users", users)
}

// BlockUser godoc
//
//	@summary 	api for admin to block or unblock user
//	@Security	BearerAuth
//	@id			BlockUser
//	@tags		Admin User
//	@Param		input	body	request.BlockUser{}	true	"inputs"
//	@Router		/admin/users/block [patch]
//	@Success	200	{object}	response.Response{}	"Successfully changed block status of user"
//	@Failure	400	{object}	response.Response{}	"invalid input"
func (a *adminHandler) BlockUser(ctx *gin.Context) {

	var body request.BlockUser

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	err := a.adminUseCase.BlockOrUnBlockUser(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to change block status of user", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully changed block status of user")
}

// GetFullSalesReport godoc
//
//	@Summary		Get full sales report (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to get all sales report for a specific period in csv form
//	@id				GetFullSalesReport
//	@tags			Admin Sales
//	@Param			start_date	query	string	false	"Sales report starting date"
//	@Param			end_date	query	string	false	"Sales report ending date"
//	@Param			page_number	query	int		false	"Page Number"
//	@Param			count		query	int		false	"Count"
//	@Router			/admin/sales [get]
//	@Success		200	{object}	response.Response{}	"ecommerce_sales_report.csv"
//	@Success		204	{object}	response.Response{}	"No sales report found"
//	@Failure		500	{object}	response.Response{}	"failed to get sales report"
func (c *adminHandler) GetFullSalesReport(ctx *gin.Context) {

	// time
	startDate, err1 := utils.StringToTime(ctx.Query("start_date"))
	endDate, err2 := utils.StringToTime(ctx.Query("end_date"))

	// join all error and send it if its not nil
	err := errors.Join(err1, err2)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindQueryFailMessage, err1, nil)
		return
	}

	pagination := request.GetPagination(ctx)

	reqData := request.SalesReport{
		StartDate:  startDate,
		EndDate:    endDate,
		Pagination: pagination,
	}

	salesReport, err := c.adminUseCase.GetFullSalesReport(ctx, reqData)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get full sales report", err, nil)
		return
	}

	if len(salesReport) == 0 {
		response.SuccessResponse(ctx, http.StatusNoContent, "No sales report found", []interface{}{})
		return
	}

	ctx.Header("Content-Type", "text/csv")
	ctx.Header("Content-Disposition", "attachment;filename=ecommerce_sales_report.csv")

	csvWriter := csv.NewWriter(ctx.Writer)
	headers := []string{
		"UserID", "FirstName", "Email",
		"ShopOrderID", "OrderDate", "OrderTotalPrice",
		"Discount", "OrderStatus", "PaymentType",
	}

	if err := csvWriter.Write(headers); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to write sales report on csv", err, nil)
		return
	}

	for _, sales := range salesReport {
		row := []string{
			fmt.Sprintf("%v", sales.UserID),
			sales.FirstName,
			sales.Email,
			fmt.Sprintf("%v", sales.ShopOrderID),
			sales.OrderDate.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%v", sales.OrderTotalPrice),
			fmt.Sprintf("%v", sales.Discount),
			sales.OrderStatus,
			sales.PaymentType,
		}

		if err := csvWriter.Write(row); err != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create write sales report to csv", err, nil)
			return
		}
	}

	csvWriter.Flush()

}

// VerifyShop godoc
//
//	@summary 	api for admin to verify shop
//	@Security	BearerAuth
//	@id			VerifyShop
//	@tags		Admin Shop
//	@Param		input	body	domain.ShopVerification{}	true	"inputs"
//	@Router		/admin/shops/verify [patch]
//	@Success	200	{object}	response.Response{}	"Successfully updated shop verification status"
//	@Failure	400	{object}	response.Response{}	"invalid input"
func (c *adminHandler) VerifyShop(ctx *gin.Context) {

	var body request.ShopVerification
	tokenString := ctx.GetHeader("Authorization")
	adminId := c.adminUseCase.DecodeTokenData(tokenString)

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	err := c.adminUseCase.VerifyShop(ctx, body, adminId)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update shop verification status", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated shop verification status", nil)
}

// CreateAdvertisement godoc
//
//	@summary 	api for admin to create advertisement
//	@Security	BearerAuth
//	@id			CreateAdvertisement
//	@tags		Advertisement Management
//	@Param		input	body	domain.Advertisement{}	true	"inputs"
//	@Router		/admin/advertisements [post]
//	@Success	200	{object}	response.Response{}	"Successfully created advertisement"
//	@Failure	400	{object}	response.Response{}	"invalid input"
func (c *adminHandler) CreateAdvertisement(ctx *gin.Context) {
	var body domain.Advertisement

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	_, err := c.adminUseCase.CreateAdvertisement(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create advertisement", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully created advertisement", nil)
}

// GetAllAdvertisements godoc
//
//	@summary		Get all advertisements
//	@Security		BearerAuth
//	@Description	API for admin to get all advertisements
//	@Id				GetAllAdvertisements
//	@Tags			Advertisement Management
//	@Param			page_number	query	int	false	"Page Number"
//	@Param			count		query	int	false	"Count"
//	@Router			/admin/advertisements [get]
//	@Success		200	{object}	response.Response{}	"Successfully got all advertisements"
//	@Success		204	{object}	response.Response{}	"No advertisements found"
//	@Failure		500	{object}	response.Response{}	"Failed to get all advertisements"
func (c *adminHandler) GetAllAdvertisements(ctx *gin.Context) {
	pagination := request.GetPagination(ctx)

	ads, err := c.adminUseCase.GetAllAdvertisements(ctx, pagination)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get all advertisements", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully got all advertisements", ads)
}

// UpdateAdvertisement godoc
//
//	@summary 	api for admin to update advertisement
//	@Security	BearerAuth
//	@id			UpdateAdvertisement
//	@tags		Advertisement Management
//	@Param		input	body	domain.Advertisement{}	true	"inputs"
//	@Router		/admin/advertisements [put]
//	@Success	200	{object}	response.Response{}	"Successfully updated advertisement"
//	@Failure	400	{object}	response.Response{}	"invalid input"
func (c *adminHandler) UpdateAdvertisement(ctx *gin.Context) {
	var body domain.Advertisement

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	_, err := c.adminUseCase.UpdateAdvertisement(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update advertisement", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated advertisement", nil)
}

// DeleteAdvertisement godoc
//
//	@summary 	api for admin to delete advertisement
//	@Security	BearerAuth
//	@id			DeleteAdvertisement
//	@tags		Advertisement Management
//	@Param		advertisement_id	path	string	true	"Advertisement ID"
//	@Router		/admin/advertisements/{advertisement_id} [delete]
//	@Success	200	{object}	response.Response{}	"Successfully deleted advertisement"
//	@Failure	400	{object}	response.Response{}	"invalid advertisement ID"
//	@Failure	500	{object}	response.Response{}	"Failed to delete advertisement"
func (c *adminHandler) DeleteAdvertisement(ctx *gin.Context) {
	advertisementIDStr := ctx.Param("advertisement_id")
	_, err := strconv.ParseUint(advertisementIDStr, 10, 64)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid advertisement ID", err, nil)
		return
	}

	err = c.adminUseCase.DeleteAdvertisement(ctx, advertisementIDStr)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete advertisement", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully deleted advertisement", nil)
}

// CreateShop godoc
//
//	@summary 	api for admin to create shop
//	@Security	BearerAuth
//	@id			CreateShop
//	@tags		Admin Shop
//	@Param		input	body	domain.ShopDetails{}	true	"inputs"
//	@Router		/admin/shops [post]
//	@Success	200	{object}	response.Response{}	"Successfully created shop"
//	@Failure	400	{object}	response.Response{}	"invalid input"
func (h *adminHandler) CreateShop(ctx *gin.Context) {
	var body domain.ShopDetails

	if err := ctx.ShouldBindJSON(&body); err != nil {
		fmt.Printf("JSON Binding Error: %v\n", err)
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	// get the adminId from authorization and add it in body
	tokenString := ctx.GetHeader("Authorization")
	adminId := h.adminUseCase.DecodeTokenData(tokenString)
	adminIdUint, err := strconv.ParseUint(adminId, 10, 64)
	if err != nil {
		fmt.Printf("Error parsing admin ID from token: %v\n", err)
	}
	body.AdminID = uint(adminIdUint)
	body.Country = "India"
	fmt.Printf("Decoded admin ID from token: %d\n", adminId)

	fmt.Printf("Received request to create shop with body: %+v\n", body)

	// Call use case to create shop
	res, err := h.adminUseCase.CreateShop(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create shop", err, nil)
		return
	}
	fmt.Printf("Shop created successfully with ID: %d\n", res.ID)

	response.SuccessResponse(ctx, http.StatusOK, "Successfully created shop", res)
}

// GetAllShops godoc
//
//	@summary		Get all shops
//	@Security		BearerAuth
//	@Description	API for admin to get all shop details
//	@Id				GetAllShops
//	@Tags			Admin Shop
//	@Param			page_number	query	int	false	"Page Number"
//	@Param			count		query	int	false	"Count"
//	@Router			/admin/shops [get]
//	@Success		200	{object}	response.Response{}	"Successfully got all shops"
//	@Success		204	{object}	response.Response{}	"No shops found"
//	@Failure		500	{object}	response.Response{}	"Failed to get all shops"
func (h *adminHandler) GetAllShops(ctx *gin.Context) {
	pagination := request.GetPagination(ctx)

	shops, err := h.adminUseCase.GetAllShops(ctx, pagination)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get all shops", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully got all shops", shops)
}

// GetShopByID godoc
//
//	@summary		Get shop by ID
//	@Security		BearerAuth
//	@Description	API for admin to get shop details by shop ID
//	@Id				GetShopByID
//	@Tags			Admin Shop
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/admin/shops/{shop_id} [get]
//	@Success		200	{object}	response.Response{}	"Successfully got shop by ID"
//	@Failure		400	{object}	response.Response{}	"Invalid shop ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get shop by ID"
func (h *adminHandler) GetShopByID(ctx *gin.Context) {
	shopIDStr := ctx.Param("shop_id")
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", err, nil)
		return
	}

	shop, err := h.adminUseCase.GetShopByID(ctx, uint(shopID))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get shop by ID", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully got shop by ID", shop)
}

// UpdateShop godoc
//
//	@summary 	api for admin to update shop
//	@Security	BearerAuth
//	@id			UpdateShop
//	@tags		Admin Shop
//	@Param		input	body	domain.ShopDetails{}	true	"inputs"
//	@Router		/admin/shops [put]
//	@Success	200	{object}	response.Response{}	"Successfully updated shop"
//	@Failure	400	{object}	response.Response{}	"invalid input"
func (h *adminHandler) UpdateShop(ctx *gin.Context) {
	// Accept a map for partial update
	var updateFields map[string]interface{}
	if err := ctx.ShouldBindJSON(&updateFields); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, updateFields)
		return
	}

	decodeToken := ctx.GetHeader("Authorization")
	shopId := h.adminUseCase.DecodeTokenData(decodeToken)

	// Remove empty or nil fields (optional, can be handled in usecase/repo)
	for k, v := range updateFields {
		if v == nil || v == "" {
			delete(updateFields, k)
		}
	}

	// Call use case to update only changed fields
	updatedData, err := h.adminUseCase.UpdateShop(ctx, updateFields, shopId)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update shop", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated shop", updatedData)
}

// UploadShopById godoc
//
//	@summary 	API for admin to update shop by ID
//	@Security	BearerAuth
//	@id			UploadShopById
//	@tags		Admin Shop
//	@Param		shop_id	path	string	true	"Shop ID"
//	@Param		input	body	map[string]interface{}	true	"Shop details (single or multiple attributes)"
//	@Router		/admin/shops/{shop_id} [put]
//	@Success	200	{object}	response.Response{}	"Successfully updated shop"
//	@Failure	400	{object}	response.Response{}	"Invalid input"
func (h *adminHandler) UploadShopById(ctx *gin.Context) {
	shopId := ctx.Param("shop_id")

	// Accept a map for partial update - can be single or multiple attributes
	var shopDetails map[string]interface{}
	if err := ctx.ShouldBindJSON(&shopDetails); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid input", err, nil)
		return
	}

	// Remove empty or nil fields
	for k, v := range shopDetails {
		if v == nil || v == "" {
			delete(shopDetails, k)
		}
	}

	updatedData, err := h.adminUseCase.UpdateShop(ctx, shopDetails, shopId)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update shop", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated shop", updatedData)
}

// GetShopByOwnerID godoc
//
//	@summary		Get shop by owner ID
//	@Security		BearerAuth
//	@Description	API for admin to get shop details by owner ID
//	@Id				GetShopByOwnerID
//	@Tags			Admin Shop
//	@Param			owner_id	path	int	true	"Owner ID"
//	@Router			/admin/shops/owner/{owner_id} [get]
//	@Success		200	{object}	response.Response{}	"Successfully got shop by owner ID"
//	@Failure		400	{object}	response.Response{}	"Invalid owner ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get shop by owner ID"
func (h *adminHandler) GetShopByOwnerID(ctx *gin.Context) {
	tokenString := ctx.GetHeader("Authorization")
	adminId := h.adminUseCase.DecodeTokenData(tokenString)
	ownerID, err := strconv.ParseUint(adminId, 10, 64)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid owner ID", err, nil)
		return
	}

	shop, err := h.adminUseCase.GetShopByOwnerID(ctx, uint(ownerID))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get shop by owner ID", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully got shop by owner ID", shop)
}

// SendNotificationToUsersInRadius godoc
//
//	@summary 	api for admin to send notification to users in radius
//	@Security	BearerAuth
//	@id			SendNotificationToUsersInRadius
//	@tags		Admin Notification
//	@Param		input	body	request.NotificationRadiusRequest{}	true	"inputs"
//	@Router		/admin/notifications/radius [get]
//	@Success	200	{object}	response.Response{}	"Successfully sent notification to users in radius"
//	@Failure	400	{object}	response.Response{}	"invalid input"
func (c *adminHandler) SendNotificationToUsersInRadius(ctx *gin.Context) {

	var body request.NotificationRadiusRequest

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	// Call use case to send notification to users in radius
	err := c.adminUseCase.SendNotificationToUsersInRadius(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to send notification to users in radius", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully sent notification to users in radius", nil)
}

// SendNotificationToUser godoc
//
//	@summary 	api for admin to send notification to a user
//	@Security	BearerAuth
//	@id			SendNotificationToUser
//	@tags		Admin Notification
//	@Param		user_id	path	int	true	"User ID"
//	@Param		message	query	string	true	"Notification Message"
//	@Router		/admin/notifications/user/{user_id} [get]
//	@Success	200	{object}	response.Response{}	"Successfully sent notification to user"
//	@Failure	400	{object}	response.Response{}	"invalid input"
func (c *adminHandler) SendNotificationToUser(ctx context.Context, userID uint, message string) error {
	err := c.adminUseCase.SendNotificationToUser(ctx, userID, message)
	if err != nil {
		return fmt.Errorf("failed to send notification to user \nerror:%v", err.Error())
	}
	return nil
}

// UploadAdminProfileImage godoc
// @summary api for admin to upload profile image
// @id UploadAdminProfileImage
// @tags Admin Account
// @Param profile_image formData file true "Profile Image"
// @Router /admin/account/upload-profile-image [post]
// @Success 200 {object} response.Response{} "Successfully uploaded profile image"
// @Failure 400 {object} response.Response{} "invalid input"
// @Failure 500 {object} response.Response{} "failed to upload profile image"
func (a *adminHandler) UploadAdminProfileImage(ctx *gin.Context) {
	var shopId = ctx.Param("shop_id")
	// Implementation goes here
	var req request.AdminUploadImageRequest

	// Parse the multipart form first to access files
	if err := ctx.Request.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to parse multipart form", err, nil)
		return
	}

	// Check what files are available
	if ctx.Request.MultipartForm != nil {
		fmt.Printf("Multipart form files: %+v\n", ctx.Request.MultipartForm.File)
	} else {
		fmt.Println("No multipart form files")
	}

	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Image file is required", err, nil)
		return
	}

	// Additional validation for image file
	if req.Image == nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "No image file provided", fmt.Errorf("image field is nil in request"), nil)
		return
	}

	//get token from and send to decode and get the data
	tokenString := ctx.GetHeader("Authorization")
	adminId := a.adminUseCase.DecodeTokenData(tokenString)

	if adminId == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid token data", fmt.Errorf("failed to decode admin ID from token"), nil)
		return
	}

	// Open and validate the image file
	file, err := req.Image.Open()
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to open image file", err, nil)
		return
	}
	defer file.Close()

	// Validate file type using both content type and filename extension
	var contentType string
	if req.Image.Header != nil {
		contentType = req.Image.Header.Get("Content-Type")
	}

	// Check if it's a valid image content type
	validContentTypes := []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif",
		"application/octet-stream", // Allow octet-stream as fallback
	}

	contentTypeValid := false
	if contentType == "" {
		contentTypeValid = true // Allow empty content type
	} else {
		for _, validType := range validContentTypes {
			if contentType == validType {
				contentTypeValid = true
				break
			}
		}
	}

	// If content type is not valid or is octet-stream, validate by filename extension
	if !contentTypeValid || contentType == "application/octet-stream" {
		filename := req.Image.Filename

		validExtensions := []string{".jpg", ".jpeg", ".png", ".gif"}
		extensionValid := false

		// Convert filename to lowercase for case-insensitive comparison
		filenameLower := strings.ToLower(filename)

		for _, ext := range validExtensions {
			if strings.HasSuffix(filenameLower, ext) {
				extensionValid = true
				break
			}
		}

		if !extensionValid {
			response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid file type. Only JPEG, PNG, and GIF images are allowed", fmt.Errorf("unsupported file extension for: %s", filename), nil)
			return
		}
	}

	// Save the file to local storage (you can modify this to use AWS S3 or other cloud storage)
	uploadDir := "uploads/admin-profiles"

	// Create upload directory if it doesn't exist
	if err := ensureDir(uploadDir); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create upload directory", err, nil)
		return
	}

	// Generate unique filename to avoid conflicts
	fileExt := getFileExtension(req.Image.Filename)
	newFileName := fmt.Sprintf("admin_%s_%d%s", adminId, time.Now().Unix(), fileExt)
	filePath := fmt.Sprintf("%s/%s", uploadDir, newFileName)

	// Save the uploaded file
	if err := ctx.SaveUploadedFile(req.Image, filePath); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to save uploaded file", err, nil)
		return
	}

	// Update database with the file path
	imageURL, err := a.adminUseCase.UploadAdminProfileImage(ctx, adminId, filePath, shopId)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update admin profile image", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully uploaded profile image", map[string]interface{}{
		"image_url": imageURL,
		"file_path": filePath,
	})

}

// AddAdminProfile godoc
// @summary api for admin to add profile
// @id AddAdminProfile
// @tags Admin Account
// @Param input body domain.AdminProfile{} true "inputs"
// @Router /admin/account [post]
// @Success 200 {object} response.Response{} "Successfully added admin profile"
// @Failure 400 {object} response.Response{} "invalid input"
// @Failure 500 {object} response.Response{} "failed to add admin profile"
func (a *adminHandler) AddAdminProfile(ctx *gin.Context) {
	// Implementation goes here
}

// GetAdminProfile godoc
// @summary api for admin to get profile
// @id GetAdminProfile
// @tags Admin Account
// @Router /admin/account [get]
// @Success 200 {object} response.Response{} "Successfully retrieved admin profile"
// @Failure 500 {object} response.Response{} "failed to retrieve admin profile"
func (a *adminHandler) GetAdminProfile(ctx *gin.Context) {
	// Implementation goes here
}

// UpdateAdminProfile godoc
// @summary api for admin to update profile
// @id UpdateAdminProfile
// @tags Admin Account
// @Param input body domain.AdminProfile{} true "inputs"
// @Router /admin/account [put]
// @Success 200 {object} response.Response{} "Successfully updated admin profile"
// @Failure 400 {object} response.Response{} "invalid input"
// @Failure 500 {object} response.Response{} "failed to update admin profile"
func (a *adminHandler) UpdateAdminProfile(ctx *gin.Context) {
	// Implementation goes here
}

// Helper functions for file upload
func ensureDir(dirName string) error {
	err := os.MkdirAll(dirName, 0755)
	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}

func getFileExtension(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return filename[idx:]
	}
	return ""
}

func (a *adminHandler) UploadShopDocument(ctx *gin.Context) {
	// Implementation goes here
	var req request.DocumentRequest
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	// Additional validation for document file
	if req.DocumentType != "" && req.DocumentValue == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "No document file provided", fmt.Errorf("document field is empty in request"), nil)
		return
	}

	//get token from and send to decode and get the data
	tokenString := ctx.GetHeader("Authorization")
	shopOwnerIdStr := a.adminUseCase.DecodeTokenData(tokenString)

	if shopOwnerIdStr == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid token data", fmt.Errorf("failed to decode shop owner ID from token"), nil)
		return
	}

	// Convert shopOwnerIdStr to uint
	shopOwnerId, err := strconv.ParseUint(shopOwnerIdStr, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid shop owner ID format", err, nil)
		return
	}

	// Call use case to upload shop document
	err = a.adminUseCase.UploadShopDocument(ctx.Request.Context(), uint(shopOwnerId), req.DocumentType, req.DocumentValue)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload shop document", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully uploaded shop document", nil)
}

func (a *adminHandler) UploadAddress(ctx *gin.Context) {
	// Implementation goes here
	var req request.AddressRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	//get token from and send to decode and get the data
	tokenString := ctx.GetHeader("Authorization")
	adminIdStr := a.adminUseCase.DecodeTokenData(tokenString)

	if adminIdStr == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid token data", fmt.Errorf("failed to decode admin ID from token"), nil)
		return
	}

	// Convert adminIdStr to uint
	adminId, err := strconv.ParseUint(adminIdStr, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid admin ID format", err, nil)
		return
	}
	// Call use case to upload address
	err = a.adminUseCase.UploadAddress(ctx.Request.Context(), strconv.FormatUint(adminId, 10), req)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload address", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully uploaded address", nil)
}

func (a *adminHandler) VerifyShopDocument(ctx *gin.Context) {
	var req request.VerifyShopDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	otp := req.OTP
	if otp == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "OTP is required", nil, nil)
		return
	}
	err := a.adminUseCase.VerifyShopDocument(ctx.Request.Context(), otp)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to verify shop document", err, nil)
		return
	}
}

func (a *adminHandler) AdminDocumentOtpSend(ctx *gin.Context) {
	var req request.DocumentRequest
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	// Additional validation for document file
	if req.DocumentType != "" && req.DocumentValue == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "No document file provided", fmt.Errorf("document field is empty in request"), nil)
		return
	}

	//get token from and send to decode and get the data
	tokenString := ctx.GetHeader("Authorization")
	shopOwnerIdStr := a.adminUseCase.DecodeTokenData(tokenString)

	if shopOwnerIdStr == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid token data", fmt.Errorf("failed to decode shop owner ID from token"), nil)
		return
	}

	// Call use case to upload shop document
	err := a.adminUseCase.UploadAdminDocumentOtpSend(ctx.Request.Context(), shopOwnerIdStr, req.DocumentType, req.DocumentValue)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload shop document", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully uploaded shop document", nil)
}

func (a *adminHandler) AdminDocumentOtpVerify(ctx *gin.Context) {
	var req request.VerifyShopDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	otp := req.OTP
	if otp == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "OTP is required", nil, nil)
		return
	}
	err := a.adminUseCase.UploadAdminDocumentOtpVerify(ctx.Request.Context(), otp, req.DocumentType, req.DocumentValue)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to verify admin document", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully verified admin document", nil)
}

func (a *adminHandler) GetVerificationStatus(ctx *gin.Context) {
	//get token from and send to decode and get the data
	tokenString := ctx.GetHeader("Authorization")
	shopOwnerIdStr := a.adminUseCase.DecodeTokenData(tokenString)

	if shopOwnerIdStr == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid token data", fmt.Errorf("failed to decode shop owner ID from token"), nil)
		return
	}

	admin, shopVerification, err := a.adminUseCase.GetVerificationStatus(ctx.Request.Context(), shopOwnerIdStr)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get verification status", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved verification status", struct {
		Admin            domain.Admin
		ShopVerification domain.ShopVerification
	}{
		Admin:            admin,
		ShopVerification: shopVerification,
	})
}

// GetAllProductDetails godoc
// @summary api for admin to get all product details from hardcoded JSON file
// @id GetAllProductDetails
// @tags Admin Product
// @Router /admin/products/all-details [get]
// @Success 200 {object} response.Response{} "Successfully retrieved all product details"
// @Failure 500 {object} response.Response{} "Failed to read product details"
func (a *adminHandler) GetAllProductDetails(ctx *gin.Context) {
	productDetails, err := a.adminUseCase.GetAllProductDetails(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get product details", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved all product details", productDetails)
}

func (a *ProductHandler) GetAllSubCategories(ctx *gin.Context) {
	subCategories, err := a.productUseCase.GetAllSubCategories(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get sub-categories", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved all sub-categories", subCategories)
}

func (a *adminHandler) GetShopProfileImageById(ctx *gin.Context) {
	shopId := ctx.Param("shop_id")

	imageURL, err := a.adminUseCase.GetShopProfileImageById(ctx, shopId)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get shop profile image", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved shop profile image", map[string]interface{}{
		"image_url": imageURL,
	})
}

// SetShopTime godoc
// @summary Set shop open/close status and times
// @id SetShopTime
// @tags Admin Shop
// @Param shop_id path uint true "Shop ID"
// @Param input body request.SetShopTimeRequest{} true "Shop Time details"
// @Router /admin/shop/time/{shop_id} [post]
// @Success 200 {object} response.Response{} "Successfully set shop time"
// @Failure 400 {object} response.Response{} "Invalid input"
func (a *adminHandler) SetShopTime(ctx *gin.Context) {
	shopIDStr := ctx.Param("shop_id")
	shopID, err := strconv.ParseUint(shopIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop_id", err, nil)
		return
	}

	var req request.SetShopTimeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	status := "close"
	if req.Status {
		status = "open"
	}

	shopTime := domain.ShopTime{
		Status:    status,
		OpenTime:  req.OpenTime,
		CloseTime: req.CloseTime,
	}

	err = a.shopTimeUseCase.SetShopTime(ctx, uint(shopID), shopTime)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to set shop time", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully set shop time", nil)
}

// GetShopTime godoc
// @summary Get shop open/close status and times
// @id GetShopTime
// @tags Admin Shop
// @Param shop_id query uint true "Shop ID"
// @Router /admin/shop-time [get]
// @Success 200 {object} response.Response{} "Successfully retrieved shop time"
// @Failure 400 {object} response.Response{} "Invalid shop_id"
func (a *adminHandler) GetShopTime(ctx *gin.Context) {
	shopIDStr := ctx.Param("shop_id")
	shopID, err := strconv.ParseUint(shopIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop_id", err, nil)
		return
	}

	shopTime, err := a.shopTimeUseCase.GetShopTime(ctx, uint(shopID))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get shop time", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved shop time", shopTime)
}

func (h *adminHandler) GetShopSocialDetails(ctx *gin.Context) {
	shopIDStr := ctx.Param("shop_id")
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop_id", err, nil)
		return
	}
	details, err := h.adminUseCase.GetShopSocialDetails(ctx, uint(shopID))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch shop social details", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Shop social details fetched", details)
}
