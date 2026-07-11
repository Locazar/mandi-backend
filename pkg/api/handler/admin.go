package handler

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
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
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
	"gorm.io/gorm"
)

type adminHandler struct {
	adminUseCase    usecaseInterface.AdminUseCase
	shopTimeUseCase usecaseInterface.ShopTimeUseCase
	cloudService    cloud.CloudService
}

// UserLogout implements the UserLogout method required by the AdminHandler interface.
func (a *adminHandler) UserLogout(ctx *gin.Context) {
	// UserLogout godoc
	//
	//	@Summary		Admin logout
	//	@Security		BearerAuth
	//	@Description	API for admin to logout
	//	@Id				AdminUserLogout
	//	@Tags			Admin Auth
	//	@Router			/admin/auth/logout [post]
	//	@Success		200	{object}	response.Response{}	"Successfully logged out"
	// Implementation goes here, or just a stub if not needed yet
	response.SuccessResponse(ctx, http.StatusOK, "Successfully logged out", nil)
}

func NewAdminHandler(adminUsecase usecaseInterface.AdminUseCase, shopTimeUseCase usecaseInterface.ShopTimeUseCase, cloudService cloud.CloudService) interfaces.AdminHandler {
	return &adminHandler{
		adminUseCase:    adminUsecase,
		shopTimeUseCase: shopTimeUseCase,
		cloudService:    cloudService,
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

	// Reject any create request that supplies the record's own primary key.
	if body.ID != "" {
		appErr := domain.ValidationError("id", "must not be provided when creating a new admin")
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr, nil)
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

// SignupOtpSend godoc
//
//	@Summary		Send signup/login OTP (Seller)
//	@Description	Sends an OTP to the seller's mobile. Creates the seller if new (full_name/password), otherwise resends for login. Returns otp_id used on verify.
//	@Id				SellerSignupOtpSend
//	@Tags			Seller Authentication
//	@Param			input	body	request.SignupOtpSend{}	true	"Phone (mobile/phone) and optional full_name/password"
//	@Router			/admin/auth/signup/otp/send [post]
//	@Success		200	{object}	response.Response{}	"OTP sent"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Failed to send OTP"
func (a *adminHandler) SignupOtpSend(ctx *gin.Context) {
	var body request.SignupOtpSend
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	phone := body.PhoneNumber()
	if phone == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Mobile number is required", nil, nil)
		return
	}
	if matched, _ := regexp.MatchString(`^\d{10}$`, phone); !matched {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid mobile number. Must be 10 digits.", nil, nil)
		return
	}

	otpID, err := a.adminUseCase.SignupOtpSend(ctx, domain.Admin{
		Mobile:   phone,
		FullName: body.FullName,
		Password: body.Password,
	})
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to send OTP", err, nil)
		return
	}

	responseData := map[string]interface{}{
		"otp_id":  otpID,
		"message": "OTP sent to mobile number for verification",
	}
	response.SuccessResponse(ctx, http.StatusOK, "OTP sent successfully", responseData)
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
	// AdminSignUpVerify godoc
	//
	//	@Summary		Verify admin signup OTP
	//	@Description	API for admin to verify OTP and complete signup
	//	@Id				AdminSignUpVerify
	//	@Tags			Admin Auth
	//	@Param			input	body	request.OTPVerify{}	true	"OTP verification input"
	//	@Router			/admin/auth/signup/verify [post]
	//	@Success		200	{object}	response.Response{}	"Successfully verified OTP and logged in"
	//	@Failure		400	{object}	response.Response{}	"Invalid input"
	//	@Failure		401	{object}	response.Response{}	"Invalid OTP"
	//	@Failure		410	{object}	response.Response{}	"OTP expired"
	//	@Failure		500	{object}	response.Response{}	"Failed to verify OTP"
	var body request.OTPVerify
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	// get the user using loginOtp useCase
	userID, shop, err := a.adminUseCase.AdminSignUpOtpVerify(ctx, body)
	println("userID:", userID)
	if err != nil {
		errResponse(ctx, "Failed to verify otp", err)
		return
	}
	log.Printf("userID: %s", userID)
	a.setupTokenAndResponse(ctx, token.Admin, userID, shop)
}

// access and refresh token generating for user and admin is same so created
// a common function for it.(differentiate user by user type )
// customResponse is optional - if provided, it will be used instead of default success response
func (c *adminHandler) setupTokenAndResponse(ctx *gin.Context, tokenUser token.UserType, userID string, customResponse ...interface{}) {

	tokenParams := usecaseInterface.GenerateTokenParams{
		UserID:   userID,
		UserType: tokenUser,
	}
	log.Printf("Generating tokens for userID: %s, userType: %s", userID, tokenUser)
	accessToken, err := c.adminUseCase.GenerateAccessToken(ctx, tokenParams)
	log.Printf("Access token generation result for userID: %s, userType: %s, accessToken: %s, error: %v", userID, tokenUser, accessToken, err)
	if err != nil {
		log.Printf("Error generating access token for userID: %s, userType: %s, error: %v", userID, tokenUser, err)
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to generate access token", err, nil)
		return
	}

	refreshToken, err := c.adminUseCase.GenerateRefreshToken(ctx, usecaseInterface.GenerateTokenParams{
		UserID:   userID,
		UserType: tokenUser,
	})

	if err != nil {
		log.Printf("Error generating refresh token for userID: %s, userType: %s, error: %v", userID, tokenUser, err)
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to generate refresh token", err, nil)
		return
	}

	authorizationValue := authorizationType + " " + accessToken
	ctx.Header(authorizationHeaderKey, authorizationValue)

	ctx.Header("access_token", accessToken)
	ctx.Header("refresh_token", refreshToken)
	log.Printf("Set access and refresh tokens in headers for userID: %s, userType: %s", userID, tokenUser)

	tokenRes := response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       userID,
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

	log.Printf("Custom response for userID: %s, userType: %s, customResponse: %+v", userID, tokenUser, customResponse)

	if len(customResponse) > 1 {
		if msg, ok := customResponse[1].(string); ok {
			message = msg
		}
	}
	log.Printf("Final response data for userID: %s, userType: %s, responseData: %+v", userID, tokenUser, responseData)
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
		errResponse(ctx, "Failed to change block status of user", err)
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

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	err := c.adminUseCase.VerifyShop(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update shop verification status", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated shop verification status", nil)
}

// CreateAdvertisement godoc
//
//	@summary		Create advertisement (Admin)
//	@Security		BearerAuth
//	@id				CreateAdvertisement
//	@tags			Advertisement Management
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			title				formData	string	true	"Title"
//	@Param			content				formData	string	true	"Content / description"
//	@Param			target_url			formData	string	false	"Click-through URL"
//	@Param			start_date			formData	string	true	"Start date (RFC3339)"
//	@Param			end_date			formData	string	true	"End date (RFC3339)"
//	@Param			status				formData	string	false	"active | inactive | expired"
//	@Param			priority			formData	string	false	"high | medium | low"
//	@Param			area_targeted		formData	string	false	"Target area name"
//	@Param			pincode_targeted	formData	string	false	"Target pincode"
//	@Param			latitude			formData	number	false	"Latitude"
//	@Param			longitude			formData	number	false	"Longitude"
//	@Param			distance_km			formData	number	false	"Radius in km"
//	@Param			photo				formData	file	false	"Banner image"
//	@Router			/admin/advertisements [post]
//	@Success		201	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
//	@Failure		500	{object}	response.Response{}
func (c *adminHandler) CreateAdvertisement(ctx *gin.Context) {
	adminID, exists := ctx.Get("userId")
	if !exists {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Unauthorized", nil, nil)
		return
	}

	title := ctx.PostForm("title")
	if title == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "title is required", nil, nil)
		return
	}
	content := ctx.PostForm("content")
	if content == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "content is required", nil, nil)
		return
	}

	startDate, err := time.Parse(time.RFC3339, ctx.PostForm("start_date"))
	if err != nil {
		startDate, err = time.Parse("2006-01-02", ctx.PostForm("start_date"))
		if err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, "start_date must be RFC3339 or YYYY-MM-DD", nil, nil)
			return
		}
	}
	endDate, err := time.Parse(time.RFC3339, ctx.PostForm("end_date"))
	if err != nil {
		endDate, err = time.Parse("2006-01-02", ctx.PostForm("end_date"))
		if err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, "end_date must be RFC3339 or YYYY-MM-DD", nil, nil)
			return
		}
	}

	lat, _ := strconv.ParseFloat(ctx.PostForm("latitude"), 64)
	lng, _ := strconv.ParseFloat(ctx.PostForm("longitude"), 64)
	dist, _ := strconv.ParseFloat(ctx.PostForm("distance_km"), 64)

	status := domain.AdvertisementStatus(ctx.PostForm("status"))
	if status == "" {
		status = domain.AdvertisementStatusActive
	}
	priority := domain.AdvertisementPriority(ctx.PostForm("priority"))
	if priority == "" {
		priority = domain.AdvertisementPriorityMedium
	}

	var imageURL string
	if fh, err := ctx.FormFile("photo"); err == nil && fh != nil {
		objectKey, uploadErr := c.cloudService.SaveFile(ctx, fh, cloud.SaveOptions{
			Namespace:  "advertisements",
			Visibility: cloud.VisibilityPublic,
		})
		if uploadErr != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload banner image", uploadErr, nil)
			return
		}
		imageURL = objectKey
	}

	audience := domain.AdvertisementAudience(ctx.PostForm("audience"))
	if audience != domain.AdvertisementAudienceSeller {
		audience = domain.AdvertisementAudienceCustomer
	}

	adminIDStr := fmt.Sprintf("%v", adminID)
	ad := domain.Advertisement{
		Title:           title,
		Content:         content,
		ImageURL:        imageURL,
		TargetURL:       ctx.PostForm("target_url"),
		StartDate:       startDate,
		EndDate:         endDate,
		Status:          status,
		Priority:        priority,
		AreaTargeted:    ctx.PostForm("area_targeted"),
		PincodeTargeted: ctx.PostForm("pincode_targeted"),
		Latitude:        lat,
		Longitude:       lng,
		DistanceKM:      dist,
		AdminID:         adminIDStr,
		CreatedByAdmin:  adminIDStr,
		Audience:        audience,
		DepartmentID:    ctx.PostForm("department_id"),
		CategoryID:      ctx.PostForm("category_id"),
	}

	created, err := c.adminUseCase.CreateAdvertisement(ctx, ad)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create advertisement", err, nil)
		return
	}

	if created.ImageURL != "" {
		created.ImageURL = c.cloudService.PublicURL(created.ImageURL)
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Advertisement created successfully", created)
}

// GetAllAdvertisements godoc
//
//	@summary		List advertisements (Admin)
//	@Security		BearerAuth
//	@Description	Returns all advertisements with optional filters. All filters are optional and can be combined.
//	@Id				GetAllAdvertisements
//	@Tags			Advertisement Management
//	@Param			page_number			query	int		false	"Page number"
//	@Param			count				query	int		false	"Items per page"
//	@Param			department_id		query	string	false	"Filter by department ID"
//	@Param			category_id			query	string	false	"Filter by category ID"
//	@Param			pincode_targeted	query	string	false	"Filter by targeted pincode"
//	@Param			latitude			query	number	false	"Latitude for geo proximity filter (requires longitude and distance_km)"
//	@Param			longitude			query	number	false	"Longitude for geo proximity filter (requires latitude and distance_km)"
//	@Param			distance_km			query	number	false	"Radius in km for geo proximity filter (requires latitude and longitude)"
//	@Param			status				query	string	false	"Filter by status (active|inactive|expired)"
//	@Param			audience			query	string	false	"Filter by audience (customer|seller)"
//	@Param			priority			query	string	false	"Filter by priority (high|medium|low)"
//	@Param			admin_id			query	string	false	"Filter by admin ID"
//	@Param			start_date			query	string	false	"Filter ads whose start_date >= this date (YYYY-MM-DD or RFC3339)"
//	@Param			end_date			query	string	false	"Filter ads whose end_date <= this date (YYYY-MM-DD or RFC3339)"
//	@Router			/admin/advertisements [get]
//	@Success		200	{object}	response.Response{}
//	@Failure		500	{object}	response.Response{}
func (c *adminHandler) GetAllAdvertisements(ctx *gin.Context) {
	pagination := request.GetPagination(ctx)

	var filter domain.AdvertisementFilter

	filter.DepartmentID = ctx.Query("department_id")
	filter.CategoryID = ctx.Query("category_id")
	filter.PincodeTargeted = ctx.Query("pincode_targeted")
	filter.Status = domain.AdvertisementStatus(ctx.Query("status"))
	filter.Audience = domain.AdvertisementAudience(ctx.Query("audience"))
	filter.Priority = domain.AdvertisementPriority(ctx.Query("priority"))
	filter.AdminID = ctx.Query("admin_id")

	if lat, err := strconv.ParseFloat(ctx.Query("latitude"), 64); err == nil {
		filter.FilterLatitude = lat
	}
	if lng, err := strconv.ParseFloat(ctx.Query("longitude"), 64); err == nil {
		filter.FilterLongitude = lng
	}
	if dist, err := strconv.ParseFloat(ctx.Query("distance_km"), 64); err == nil {
		filter.DistanceKM = dist
	}
	if sd := ctx.Query("start_date"); sd != "" {
		for _, layout := range []string{time.RFC3339, "2006-01-02"} {
			if t, err := time.Parse(layout, sd); err == nil {
				filter.StartDateFrom = t
				break
			}
		}
	}
	if ed := ctx.Query("end_date"); ed != "" {
		for _, layout := range []string{time.RFC3339, "2006-01-02"} {
			if t, err := time.Parse(layout, ed); err == nil {
				// end_date <= end of that day
				filter.EndDateTo = t.Add(24*time.Hour - time.Second)
				break
			}
		}
	}

	ads, err := c.adminUseCase.GetAllAdvertisements(ctx, pagination, filter)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get advertisements", err, nil)
		return
	}

	for i := range ads {
		if ads[i].ImageURL != "" {
			ads[i].ImageURL = c.cloudService.PublicURL(ads[i].ImageURL)
		}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully got all advertisements", ads)
}

// UpdateAdvertisement godoc
//
//	@summary		Update advertisement (Admin)
//	@Security		BearerAuth
//	@id				UpdateAdvertisement
//	@tags			Advertisement Management
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			advertisement_id	path		string	true	"Advertisement ID"
//	@Param			title				formData	string	false	"Title"
//	@Param			content				formData	string	false	"Content"
//	@Param			target_url			formData	string	false	"Click-through URL"
//	@Param			start_date			formData	string	false	"Start date"
//	@Param			end_date			formData	string	false	"End date"
//	@Param			status				formData	string	false	"active | inactive | expired"
//	@Param			priority			formData	string	false	"high | medium | low"
//	@Param			area_targeted		formData	string	false	"Target area"
//	@Param			pincode_targeted	formData	string	false	"Target pincode"
//	@Param			latitude			formData	number	false	"Latitude"
//	@Param			longitude			formData	number	false	"Longitude"
//	@Param			distance_km			formData	number	false	"Radius km"
//	@Param			photo				formData	file	false	"New banner image (omit to keep existing)"
//	@Router			/admin/advertisements/{advertisement_id} [put]
//	@Success		200	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
//	@Failure		500	{object}	response.Response{}
func (c *adminHandler) UpdateAdvertisement(ctx *gin.Context) {
	advertisementID := ctx.Param("advertisement_id")
	if advertisementID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "advertisement_id is required", nil, nil)
		return
	}

	// Fetch existing record so we can preserve fields the caller didn't send.
	existing, err := c.adminUseCase.GetAdvertisementByID(ctx, advertisementID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "Advertisement not found", err, nil)
		return
	}

	if v := ctx.PostForm("title"); v != "" {
		existing.Title = v
	}
	if v := ctx.PostForm("content"); v != "" {
		existing.Content = v
	}
	if v := ctx.PostForm("target_url"); v != "" {
		existing.TargetURL = v
	}
	if v := ctx.PostForm("status"); v != "" {
		existing.Status = domain.AdvertisementStatus(v)
	}
	if v := ctx.PostForm("priority"); v != "" {
		existing.Priority = domain.AdvertisementPriority(v)
	}
	if v := ctx.PostForm("area_targeted"); v != "" {
		existing.AreaTargeted = v
	}
	if v := ctx.PostForm("pincode_targeted"); v != "" {
		existing.PincodeTargeted = v
	}
	if v := ctx.PostForm("latitude"); v != "" {
		if f, e := strconv.ParseFloat(v, 64); e == nil {
			existing.Latitude = f
		}
	}
	if v := ctx.PostForm("longitude"); v != "" {
		if f, e := strconv.ParseFloat(v, 64); e == nil {
			existing.Longitude = f
		}
	}
	if v := ctx.PostForm("distance_km"); v != "" {
		if f, e := strconv.ParseFloat(v, 64); e == nil {
			existing.DistanceKM = f
		}
	}
	if v := ctx.PostForm("start_date"); v != "" {
		if t, e := time.Parse(time.RFC3339, v); e == nil {
			existing.StartDate = t
		} else if t, e := time.Parse("2006-01-02", v); e == nil {
			existing.StartDate = t
		}
	}
	if v := ctx.PostForm("end_date"); v != "" {
		if t, e := time.Parse(time.RFC3339, v); e == nil {
			existing.EndDate = t
		} else if t, e := time.Parse("2006-01-02", v); e == nil {
			existing.EndDate = t
		}
	}

	if v := ctx.PostForm("audience"); v == string(domain.AdvertisementAudienceCustomer) || v == string(domain.AdvertisementAudienceSeller) {
		existing.Audience = domain.AdvertisementAudience(v)
	}
	if v := ctx.PostForm("department_id"); v != "" {
		existing.DepartmentID = v
	}
	if v := ctx.PostForm("category_id"); v != "" {
		existing.CategoryID = v
	}

	// Upload new banner only when the caller provides one.
	if fh, ferr := ctx.FormFile("photo"); ferr == nil && fh != nil {
		objectKey, uploadErr := c.cloudService.SaveFile(ctx, fh, cloud.SaveOptions{
			Namespace:  "advertisements",
			Visibility: cloud.VisibilityPublic,
		})
		if uploadErr != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload banner image", uploadErr, nil)
			return
		}
		existing.ImageURL = objectKey
	}

	updated, err := c.adminUseCase.UpdateAdvertisement(ctx, existing)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update advertisement", err, nil)
		return
	}

	if updated.ImageURL != "" {
		updated.ImageURL = c.cloudService.PublicURL(updated.ImageURL)
	}
	response.SuccessResponse(ctx, http.StatusOK, "Advertisement updated successfully", updated)
}

// DeleteAdvertisement godoc
//
//	@summary		Delete advertisement (Admin)
//	@Security		BearerAuth
//	@id				DeleteAdvertisement
//	@tags			Advertisement Management
//	@Param			advertisement_id	path	string	true	"Advertisement ID"
//	@Router			/admin/advertisements/{advertisement_id} [delete]
//	@Success		200	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
//	@Failure		500	{object}	response.Response{}
func (c *adminHandler) DeleteAdvertisement(ctx *gin.Context) {
	advertisementID := ctx.Param("advertisement_id")
	if advertisementID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "advertisement_id is required", nil, nil)
		return
	}

	if err := c.adminUseCase.DeleteAdvertisement(ctx, advertisementID); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete advertisement", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Advertisement deleted successfully", nil)
}

// GetAdvertisementPricePlans godoc
//
//	@summary		Get advertisement price plans for a date range
//	@Security		BearerAuth
//	@Id				GetAdvertisementPricePlans
//	@Tags			Advertisement Requests
//	@Param			start_date	query	string	true	"Start date (YYYY-MM-DD)"
//	@Param			end_date	query	string	true	"End date (YYYY-MM-DD)"
//	@Router			/admin/advertisement-requests/pricing [get]
//	@Success		200	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (c *adminHandler) GetAdvertisementPricePlans(ctx *gin.Context) {
	startDate, err := time.Parse("2006-01-02", ctx.Query("start_date"))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "invalid start_date (want YYYY-MM-DD)", err, nil)
		return
	}
	endDate, err := time.Parse("2006-01-02", ctx.Query("end_date"))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "invalid end_date (want YYYY-MM-DD)", err, nil)
		return
	}

	plans, err := c.adminUseCase.GetAdvertisementPricePlans(ctx, startDate, endDate)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to get price plans", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Advertisement price plans", plans)
}

// CreateAdvertisementRequest godoc
//
//	@summary		Create an advertisement request (seller)
//	@Security		BearerAuth
//	@Id				CreateAdvertisementRequest
//	@Tags			Advertisement Requests
//	@Router			/admin/advertisement-requests [post]
//	@Success		201	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (c *adminHandler) CreateAdvertisementRequest(ctx *gin.Context) {
	var body struct {
		Title     string `json:"title"`
		Content   string `json:"content"`
		StartDate string `json:"start_date" binding:"required"`
		EndDate   string `json:"end_date" binding:"required"`
		PlanKey   string `json:"plan_key" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	startDate, err := time.Parse("2006-01-02", body.StartDate)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "invalid start_date (want YYYY-MM-DD)", err, nil)
		return
	}
	endDate, err := time.Parse("2006-01-02", body.EndDate)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "invalid end_date (want YYYY-MM-DD)", err, nil)
		return
	}

	adminID := c.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	if adminID == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "invalid token", nil, nil)
		return
	}

	req := domain.AdvertisementRequest{
		AdminID:   adminID,
		Title:     body.Title,
		Content:   body.Content,
		StartDate: startDate,
		EndDate:   endDate,
		PlanKey:   body.PlanKey,
	}
	created, err := c.adminUseCase.CreateAdvertisementRequest(ctx, req)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to create advertisement request", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Advertisement request created", created)
}

// GetAllAdvertisementRequests godoc
//
//	@summary		List advertisement requests
//	@Security		BearerAuth
//	@Id				GetAllAdvertisementRequests
//	@Tags			Advertisement Requests
//	@Param			mine	query	bool	false	"Only the caller's own requests"
//	@Router			/admin/advertisement-requests [get]
//	@Success		200	{object}	response.Response{}
//	@Failure		500	{object}	response.Response{}
func (c *adminHandler) GetAllAdvertisementRequests(ctx *gin.Context) {
	pagination := request.GetPagination(ctx)

	// mine=true restricts the list to the caller's own requests (seller view);
	// otherwise all requests are returned (admin portal view).
	adminID := ""
	if ctx.Query("mine") == "true" {
		adminID = c.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	}

	requests, err := c.adminUseCase.GetAllAdvertisementRequests(ctx, pagination, adminID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get advertisement requests", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Advertisement requests", requests)
}

// GetAdvertisementRequestByID godoc
//
//	@summary		Get one advertisement request
//	@Security		BearerAuth
//	@Id				GetAdvertisementRequestByID
//	@Tags			Advertisement Requests
//	@Param			request_id	path	string	true	"Advertisement request ID"
//	@Router			/admin/advertisement-requests/{request_id} [get]
//	@Success		200	{object}	response.Response{}
//	@Failure		404	{object}	response.Response{}
func (c *adminHandler) GetAdvertisementRequestByID(ctx *gin.Context) {
	requestID := ctx.Param("request_id")
	req, err := c.adminUseCase.GetAdvertisementRequestByID(ctx, requestID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "Advertisement request not found", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Advertisement request", req)
}

// UpdateAdvertisementRequest godoc
//
//	@summary		Update an advertisement request (status/comment or details)
//	@Security		BearerAuth
//	@Id				UpdateAdvertisementRequest
//	@Tags			Advertisement Requests
//	@Param			request_id	path	string	true	"Advertisement request ID"
//	@Router			/admin/advertisement-requests/{request_id} [put]
//	@Success		200	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (c *adminHandler) UpdateAdvertisementRequest(ctx *gin.Context) {
	requestID := ctx.Param("request_id")
	existing, err := c.adminUseCase.GetAdvertisementRequestByID(ctx, requestID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "Advertisement request not found", err, nil)
		return
	}

	var body struct {
		Title        *string `json:"title"`
		Content      *string `json:"content"`
		Status       *string `json:"status"`
		AdminComment *string `json:"admin_comment"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	if body.Title != nil {
		existing.Title = *body.Title
	}
	if body.Content != nil {
		existing.Content = *body.Content
	}
	if body.Status != nil {
		existing.Status = domain.AdvertisementRequestStatus(*body.Status)
	}
	if body.AdminComment != nil {
		existing.AdminComment = *body.AdminComment
	}

	updated, err := c.adminUseCase.UpdateAdvertisementRequest(ctx, existing)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update advertisement request", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Advertisement request updated", updated)
}

// DeleteAdvertisementRequest godoc
//
//	@summary		Delete an advertisement request
//	@Security		BearerAuth
//	@Id				DeleteAdvertisementRequest
//	@Tags			Advertisement Requests
//	@Param			request_id	path	string	true	"Advertisement request ID"
//	@Router			/admin/advertisement-requests/{request_id} [delete]
//	@Success		200	{object}	response.Response{}
//	@Failure		404	{object}	response.Response{}
func (c *adminHandler) DeleteAdvertisementRequest(ctx *gin.Context) {
	requestID := ctx.Param("request_id")
	if err := c.adminUseCase.DeleteAdvertisementRequest(ctx, requestID); err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "Failed to delete advertisement request", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Advertisement request deleted", nil)
}

// GetAdvertisementRequestInvoice godoc
//
//	@summary		Get invoice (GST + charges breakdown) for an approved request
//	@Security		BearerAuth
//	@Id				GetAdvertisementRequestInvoice
//	@Tags			Advertisement Requests
//	@Param			request_id	path	string	true	"Advertisement request ID"
//	@Router			/admin/advertisement-requests/{request_id}/invoice [get]
//	@Success		200	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (c *adminHandler) GetAdvertisementRequestInvoice(ctx *gin.Context) {
	invoice, err := c.adminUseCase.GetAdvertisementRequestInvoice(ctx, ctx.Param("request_id"))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to get invoice", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Advertisement invoice", invoice)
}

// CreateAdvertisementPaymentOrder godoc
//
//	@summary		Create a Razorpay order for an approved advertisement request
//	@Security		BearerAuth
//	@Id				CreateAdvertisementPaymentOrder
//	@Tags			Advertisement Requests
//	@Param			request_id	path	string	true	"Advertisement request ID"
//	@Router			/admin/advertisement-requests/{request_id}/create-order [post]
//	@Success		200	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (c *adminHandler) CreateAdvertisementPaymentOrder(ctx *gin.Context) {
	adminID := c.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	if adminID == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "invalid token", nil, nil)
		return
	}

	orderID, keyID, amount, err := c.adminUseCase.CreateAdvertisementPaymentOrder(ctx, adminID, ctx.Param("request_id"))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to create payment order", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Payment order created", map[string]interface{}{
		"order_id": orderID,
		"key_id":   keyID,
		"amount":   amount,
		"currency": "INR",
	})
}

// VerifyAdvertisementPayment godoc
//
//	@summary		Verify a Razorpay payment for an advertisement request
//	@Security		BearerAuth
//	@Id				VerifyAdvertisementPayment
//	@Tags			Advertisement Requests
//	@Router			/admin/advertisement-requests/verify-payment [post]
//	@Success		200	{object}	response.Response{}
//	@Failure		402	{object}	response.Response{}
func (c *adminHandler) VerifyAdvertisementPayment(ctx *gin.Context) {
	adminID := c.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	if adminID == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "invalid token", nil, nil)
		return
	}

	var body struct {
		OrderID   string `json:"razorpay_order_id" binding:"required"`
		PaymentID string `json:"razorpay_payment_id" binding:"required"`
		Signature string `json:"razorpay_signature" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	if err := c.adminUseCase.VerifyAdvertisementPayment(ctx, adminID, body.OrderID, body.PaymentID, body.Signature); err != nil {
		response.ErrorResponse(ctx, http.StatusPaymentRequired, "Payment verification failed", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Payment verified", map[string]string{"status": "success"})
}

// AdvertisementPaymentFailed godoc
//
//	@summary		Record a failed advertisement payment attempt
//	@Security		BearerAuth
//	@Id				AdvertisementPaymentFailed
//	@Tags			Advertisement Requests
//	@Router			/admin/advertisement-requests/payment-failed [post]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) AdvertisementPaymentFailed(ctx *gin.Context) {
	adminID := c.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	if adminID == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "invalid token", nil, nil)
		return
	}

	var body struct {
		OrderID string `json:"order_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	if err := c.adminUseCase.AdvertisementPaymentFailed(ctx, adminID, body.OrderID); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to record payment failure", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Payment failure recorded", nil)
}

// GetAdvertisementPricing godoc
//
//	@summary		Get advertisement pricing configuration (plans + charges)
//	@Security		BearerAuth
//	@Id				GetAdvertisementPricing
//	@Tags			Advertisement Pricing
//	@Router			/admin/advertisement-pricing [get]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) GetAdvertisementPricing(ctx *gin.Context) {
	plans, err := c.adminUseCase.ListAdvertisementPlanConfigs(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to load price plans", err, nil)
		return
	}
	cfg, err := c.adminUseCase.GetAdvertisementPricingConfig(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to load pricing config", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Advertisement pricing", map[string]interface{}{
		"plans":  plans,
		"config": cfg,
	})
}

// CreateAdvertisementPlan godoc
//
//	@summary		Create an advertisement price plan
//	@Security		BearerAuth
//	@Id				CreateAdvertisementPlan
//	@Tags			Advertisement Pricing
//	@Router			/admin/advertisement-pricing/plans [post]
//	@Success		201	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (c *adminHandler) CreateAdvertisementPlan(ctx *gin.Context) {
	var body domain.AdvertisementPlanConfig
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	body.ID = ""
	created, err := c.adminUseCase.CreateAdvertisementPlanConfig(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to create price plan", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Price plan created", created)
}

// UpdateAdvertisementPlan godoc
//
//	@summary		Update an advertisement price plan
//	@Security		BearerAuth
//	@Id				UpdateAdvertisementPlan
//	@Tags			Advertisement Pricing
//	@Param			plan_id	path	string	true	"Plan ID"
//	@Router			/admin/advertisement-pricing/plans/{plan_id} [put]
//	@Success		200	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (c *adminHandler) UpdateAdvertisementPlan(ctx *gin.Context) {
	var body domain.AdvertisementPlanConfig
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	body.ID = ctx.Param("plan_id")
	updated, err := c.adminUseCase.UpdateAdvertisementPlanConfig(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update price plan", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Price plan updated", updated)
}

// DeleteAdvertisementPlan godoc
//
//	@summary		Delete an advertisement price plan
//	@Security		BearerAuth
//	@Id				DeleteAdvertisementPlan
//	@Tags			Advertisement Pricing
//	@Param			plan_id	path	string	true	"Plan ID"
//	@Router			/admin/advertisement-pricing/plans/{plan_id} [delete]
//	@Success		200	{object}	response.Response{}
//	@Failure		404	{object}	response.Response{}
func (c *adminHandler) DeleteAdvertisementPlan(ctx *gin.Context) {
	if err := c.adminUseCase.DeleteAdvertisementPlanConfig(ctx, ctx.Param("plan_id")); err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "Failed to delete price plan", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Price plan deleted", nil)
}

// UpdateAdvertisementPricingConfig godoc
//
//	@summary		Update GST/platform-fee percentages
//	@Security		BearerAuth
//	@Id				UpdateAdvertisementPricingConfig
//	@Tags			Advertisement Pricing
//	@Router			/admin/advertisement-pricing/config [put]
//	@Success		200	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (c *adminHandler) UpdateAdvertisementPricingConfig(ctx *gin.Context) {
	var body domain.AdvertisementPricingConfig
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	updated, err := c.adminUseCase.UpdateAdvertisementPricingConfig(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update pricing config", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Pricing config updated", updated)
}

// GetFeatureFlagsObject godoc
//
//	@summary		Get feature flags as a key→enabled object (clients)
//	@Id				GetFeatureFlagsObject
//	@Tags			Feature Flags
//	@Router			/feature-flags [get]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) GetFeatureFlagsObject(ctx *gin.Context) {
	flags, err := c.adminUseCase.GetFeatureFlagsObject(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get feature flags", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Feature flags", map[string]interface{}{
		"featureFlag": flags,
	})
}

// ListFeatureFlags godoc
//
//	@summary		List feature flags with metadata (admin panel)
//	@Security		BearerAuth
//	@Id				ListFeatureFlags
//	@Tags			Feature Flags
//	@Router			/admin/feature-flags [get]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) ListFeatureFlags(ctx *gin.Context) {
	flags, err := c.adminUseCase.ListFeatureFlags(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list feature flags", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Feature flags", flags)
}

// CreateFeatureFlag godoc
//
//	@summary		Create a feature flag
//	@Security		BearerAuth
//	@Id				CreateFeatureFlag
//	@Tags			Feature Flags
//	@Router			/admin/feature-flags [post]
//	@Success		201	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (c *adminHandler) CreateFeatureFlag(ctx *gin.Context) {
	var body domain.FeatureFlag
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	body.ID = ""
	created, err := c.adminUseCase.CreateFeatureFlag(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to create feature flag", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Feature flag created", created)
}

// UpdateFeatureFlag godoc
//
//	@summary		Update a feature flag
//	@Security		BearerAuth
//	@Id				UpdateFeatureFlag
//	@Tags			Feature Flags
//	@Param			flag_id	path	string	true	"Feature flag ID"
//	@Router			/admin/feature-flags/{flag_id} [put]
//	@Success		200	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (c *adminHandler) UpdateFeatureFlag(ctx *gin.Context) {
	var body domain.FeatureFlag
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	body.ID = ctx.Param("flag_id")
	updated, err := c.adminUseCase.UpdateFeatureFlag(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update feature flag", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Feature flag updated", updated)
}

// DeleteFeatureFlag godoc
//
//	@summary		Delete a feature flag
//	@Security		BearerAuth
//	@Id				DeleteFeatureFlag
//	@Tags			Feature Flags
//	@Param			flag_id	path	string	true	"Feature flag ID"
//	@Router			/admin/feature-flags/{flag_id} [delete]
//	@Success		200	{object}	response.Response{}
//	@Failure		404	{object}	response.Response{}
func (c *adminHandler) DeleteFeatureFlag(ctx *gin.Context) {
	if err := c.adminUseCase.DeleteFeatureFlag(ctx, ctx.Param("flag_id")); err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "Failed to delete feature flag", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Feature flag deleted", nil)
}

// ListAppConfigs godoc
//
//	@summary		List app configs
//	@Security		BearerAuth
//	@Id				ListAppConfigs
//	@Tags			App Configs
//	@Router			/admin/app-configs [get]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) ListAppConfigs(ctx *gin.Context) {
	cfgs, err := c.adminUseCase.ListAppConfigs(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list app configs", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "App configs", cfgs)
}

// GetAppConfigByKey godoc
//
//	@summary		Get app config by key
//	@Security		BearerAuth
//	@Id				GetAppConfigByKey
//	@Tags			App Configs
//	@Router			/admin/app-configs/{config_key} [get]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) GetAppConfigByKey(ctx *gin.Context) {
	cfg, err := c.adminUseCase.GetAppConfigByKey(ctx, ctx.Param("config_key"))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "App config not found", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "App config", cfg)
}

// CreateAppConfig godoc
//
//	@summary		Create app config
//	@Security		BearerAuth
//	@Id				CreateAppConfig
//	@Tags			App Configs
//	@Router			/admin/app-configs [post]
//	@Success		201	{object}	response.Response{}
func (c *adminHandler) CreateAppConfig(ctx *gin.Context) {
	var body domain.AppConfig
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	body.ID = ""
	created, err := c.adminUseCase.CreateAppConfig(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to create app config", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "App config created", created)
}

// UpdateAppConfig godoc
//
//	@summary		Update app config
//	@Security		BearerAuth
//	@Id				UpdateAppConfig
//	@Tags			App Configs
//	@Router			/admin/app-configs/{config_id} [put]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) UpdateAppConfig(ctx *gin.Context) {
	var body domain.AppConfig
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	body.ID = ctx.Param("config_id")
	updated, err := c.adminUseCase.UpdateAppConfig(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update app config", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "App config updated", updated)
}

// DeleteAppConfig godoc
//
//	@summary		Delete app config
//	@Security		BearerAuth
//	@Id				DeleteAppConfig
//	@Tags			App Configs
//	@Router			/admin/app-configs/{config_id} [delete]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) DeleteAppConfig(ctx *gin.Context) {
	if err := c.adminUseCase.DeleteAppConfig(ctx, ctx.Param("config_id")); err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "Failed to delete app config", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "App config deleted", nil)
}

// GetGlobalConfig godoc
//
//	@summary		Get the structured global app config (CDN, feature flags, HTTP, image upload, AI)
//	@Security		BearerAuth
//	@Id				GetGlobalConfig
//	@Tags			App Configs
//	@Router			/admin/config [get]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) GetGlobalConfig(ctx *gin.Context) {
	cfg, err := c.adminUseCase.GetGlobalConfig(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch config", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Config fetched", cfg)
}

// UpdateGlobalConfig godoc
//
//	@summary		Update the structured global app config
//	@Security		BearerAuth
//	@Id				UpdateGlobalConfig
//	@Tags			App Configs
//	@Router			/admin/config [put]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) UpdateGlobalConfig(ctx *gin.Context) {
	var body domain.GlobalAppConfig
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	updated, err := c.adminUseCase.UpdateGlobalConfig(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update config", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Config updated", updated)
}

// GetHelpSettings godoc
//
//	@summary		Get help center contact settings (admin panel)
//	@Security		BearerAuth
//	@Id				GetHelpSettings
//	@Tags			Help Center
//	@Router			/admin/help/settings [get]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) GetHelpSettings(ctx *gin.Context) {
	settings, err := c.adminUseCase.GetHelpSettings(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get help settings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Help settings", settings)
}

// UpdateHelpSettings godoc
//
//	@summary		Update help center contact settings (admin panel)
//	@Security		BearerAuth
//	@Id				UpdateHelpSettings
//	@Tags			Help Center
//	@Router			/admin/help/settings [put]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) UpdateHelpSettings(ctx *gin.Context) {
	var body domain.HelpSettings
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	updated, err := c.adminUseCase.UpdateHelpSettings(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update help settings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Help settings updated", updated)
}

// ListHelpFAQs godoc
//
//	@summary		List all FAQs, including inactive (admin panel)
//	@Security		BearerAuth
//	@Id				ListHelpFAQs
//	@Tags			Help Center
//	@Router			/admin/help/faqs [get]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) ListHelpFAQs(ctx *gin.Context) {
	faqs, err := c.adminUseCase.ListHelpFAQs(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list FAQs", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Help FAQs", faqs)
}

// CreateHelpFAQ godoc
//
//	@summary		Create a help FAQ (admin panel)
//	@Security		BearerAuth
//	@Id				CreateHelpFAQ
//	@Tags			Help Center
//	@Router			/admin/help/faqs [post]
//	@Success		201	{object}	response.Response{}
func (c *adminHandler) CreateHelpFAQ(ctx *gin.Context) {
	var body domain.HelpFAQ
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	body.ID = ""
	created, err := c.adminUseCase.CreateHelpFAQ(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to create FAQ", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "FAQ created", created)
}

// UpdateHelpFAQ godoc
//
//	@summary		Update a help FAQ (admin panel)
//	@Security		BearerAuth
//	@Id				UpdateHelpFAQ
//	@Tags			Help Center
//	@Router			/admin/help/faqs/{faq_id} [put]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) UpdateHelpFAQ(ctx *gin.Context) {
	var body domain.HelpFAQ
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	body.ID = ctx.Param("faq_id")
	updated, err := c.adminUseCase.UpdateHelpFAQ(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update FAQ", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "FAQ updated", updated)
}

// DeleteHelpFAQ godoc
//
//	@summary		Delete a help FAQ (admin panel)
//	@Security		BearerAuth
//	@Id				DeleteHelpFAQ
//	@Tags			Help Center
//	@Router			/admin/help/faqs/{faq_id} [delete]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) DeleteHelpFAQ(ctx *gin.Context) {
	if err := c.adminUseCase.DeleteHelpFAQ(ctx, ctx.Param("faq_id")); err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "Failed to delete FAQ", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "FAQ deleted", nil)
}

// GetHelpCenter godoc
//
//	@summary		Get combined help center content (public — no auth)
//	@Id				GetHelpCenter
//	@Tags			Help Center
//	@Router			/help [get]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) GetHelpCenter(ctx *gin.Context) {
	helpCenter, err := c.adminUseCase.GetHelpCenter(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get help center", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Help center", helpCenter)
}

// ListSubscriptionPlans godoc
//
//	@summary		List subscription plans (admin panel)
//	@Security		BearerAuth
//	@Id				ListSubscriptionPlans
//	@Tags			Subscription Plans
//	@Router			/admin/subscription-plans [get]
//	@Success		200	{object}	response.Response{}
func (c *adminHandler) ListSubscriptionPlans(ctx *gin.Context) {
	plans, err := c.adminUseCase.ListSubscriptionPlans(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list subscription plans", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Subscription plans", plans)
}

// CreateSubscriptionPlan godoc
//
//	@summary		Create a subscription plan
//	@Security		BearerAuth
//	@Id				CreateSubscriptionPlan
//	@Tags			Subscription Plans
//	@Router			/admin/subscription-plans [post]
//	@Success		201	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (c *adminHandler) CreateSubscriptionPlan(ctx *gin.Context) {
	var body domain.SubscriptionPlan
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	body.ID = ""
	created, err := c.adminUseCase.CreateSubscriptionPlan(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to create subscription plan", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Subscription plan created", created)
}

// UpdateSubscriptionPlan godoc
//
//	@summary		Update a subscription plan
//	@Security		BearerAuth
//	@Id				UpdateSubscriptionPlan
//	@Tags			Subscription Plans
//	@Param			plan_id	path	string	true	"Plan ID"
//	@Router			/admin/subscription-plans/{plan_id} [put]
//	@Success		200	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (c *adminHandler) UpdateSubscriptionPlan(ctx *gin.Context) {
	var body domain.SubscriptionPlan
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	body.ID = ctx.Param("plan_id")
	updated, err := c.adminUseCase.UpdateSubscriptionPlan(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update subscription plan", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Subscription plan updated", updated)
}

// DeleteSubscriptionPlan godoc
//
//	@summary		Delete a subscription plan (fails if referenced by orders; deactivate instead)
//	@Security		BearerAuth
//	@Id				DeleteSubscriptionPlan
//	@Tags			Subscription Plans
//	@Param			plan_id	path	string	true	"Plan ID"
//	@Router			/admin/subscription-plans/{plan_id} [delete]
//	@Success		200	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (c *adminHandler) DeleteSubscriptionPlan(ctx *gin.Context) {
	if err := c.adminUseCase.DeleteSubscriptionPlan(ctx, ctx.Param("plan_id")); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest,
			"Failed to delete subscription plan (plans already purchased cannot be deleted — deactivate it instead)", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Subscription plan deleted", nil)
}

// GetActiveAdvertisements godoc
//
//	@summary		Get active advertisements (public)
//	@Description	Returns currently active advertisements for display inside the app.
//	@Id				GetActiveAdvertisements
//	@Tags			Advertisement Management
//	@Router			/advertisements/active [get]
//	@Success		200	{object}	response.Response{}
//	@Failure		500	{object}	response.Response{}
func (c *adminHandler) GetActiveAdvertisements(ctx *gin.Context) {
	ads, err := c.adminUseCase.GetActiveAdvertisements(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get active advertisements", err, nil)
		return
	}

	for i := range ads {
		if ads[i].ImageURL != "" {
			ads[i].ImageURL = c.cloudService.PublicURL(ads[i].ImageURL)
		}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Active advertisements", ads)
}

// GetActiveAdvertisementsFiltered godoc
//
//	@summary		Get active advertisements with location/audience filters
//	@Description	Returns active ads filtered by lat/lng/radius, pincode, and app type (seller|customer).
//	@Id				GetActiveAdvertisementsFiltered
//	@Tags			Advertisement Management
//	@Param			lat      query number false "Latitude of the user"
//	@Param			lng      query number false "Longitude of the user"
//	@Param			radius   query number false "Search radius in km (default 10)"
//	@Param			pincode  query string false "User pincode"
//	@Param			app_type query string false "App type: seller or customer"
//	@Router			/advertisements/active/filter [get]
//	@Success		200	{object}	response.Response{}
//	@Failure		500	{object}	response.Response{}
func (c *adminHandler) GetActiveAdvertisementsFiltered(ctx *gin.Context) {
	lat, _ := strconv.ParseFloat(ctx.Query("lat"), 64)
	lng, _ := strconv.ParseFloat(ctx.Query("lng"), 64)
	radiusKM, _ := strconv.ParseFloat(ctx.Query("radius"), 64)
	if radiusKM == 0 {
		radiusKM = 10
	}
	pincode := ctx.Query("pincode")
	appType := domain.AdvertisementAudience(ctx.Query("app_type"))

	filter := domain.AdvertisementFilter{
		Latitude:  lat,
		Longitude: lng,
		RadiusKM:  radiusKM,
		Pincode:   pincode,
		AppType:   appType,
	}

	log.Printf("[AdsFilter] lat=%v lng=%v radius=%v pincode=%q app_type=%q", lat, lng, radiusKM, pincode, appType)

	ads, err := c.adminUseCase.GetActiveAdvertisementsFiltered(ctx, filter)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get advertisements", err, nil)
		return
	}

	log.Printf("[AdsFilter] query returned %d ads", len(ads))
	for i := range ads {
		log.Printf("[AdsFilter]  → id=%v audience=%v pincode=%q lat=%v lng=%v radius=%v image=%q start=%v end=%v",
			ads[i].ID, ads[i].Audience, ads[i].PincodeTargeted, ads[i].Latitude, ads[i].Longitude, ads[i].DistanceKM, ads[i].ImageURL, ads[i].StartDate, ads[i].EndDate)
		if ads[i].ImageURL != "" {
			ads[i].ImageURL = c.cloudService.PublicURL(ads[i].ImageURL)
		}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Active advertisements", ads)
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
		log.Printf("JSON Binding Error: %v", err)
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	// Reject any create request that supplies the record's own primary key.
	if body.ID != "" {
		appErr := domain.ValidationError("id", "must not be provided when creating a new shop")
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr, nil)
		return
	}

	// get the adminId from authorization and add it in body
	tokenString := ctx.GetHeader("Authorization")
	adminId := h.adminUseCase.DecodeTokenData(tokenString)
	body.AdminID = adminId
	body.Country = "India"
	fmt.Printf("Decoded admin ID from token: %s\n", adminId)

	// Fetch admin mobile and assign to shop phone
	admin, err := h.adminUseCase.GetAdminByID(ctx, adminId)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch admin details", err, nil)
		return
	}
	body.Phone = admin.Mobile

	log.Printf("Received request to create shop with body: %+v", body)

	// Call use case to create shop
	res, err := h.adminUseCase.CreateShop(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create shop", err, nil)
		return
	}
	log.Printf("Shop created successfully with ID: %s", res.ID)

	res.Shop_Image_URL = cloud.ResolveURL(h.cloudService, res.Shop_Image_URL)
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

	for i := range shops {
		shops[i].Shop_Image_URL = cloud.ResolveURL(h.cloudService, shops[i].Shop_Image_URL)
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully got all shops", shops)
}

// SearchShops godoc
//
//	@summary		Search shops with filters
//	@Security		BearerAuth
//	@Description	API for admin to search shops by phone, pincode, city, name, or radius around a point
//	@Id				SearchShops
//	@Tags			Admin Shop
//	@Param			phone		query	string	false	"Phone number (partial match)"
//	@Param			pincode		query	string	false	"Pincode (exact match)"
//	@Param			city		query	string	false	"City (partial match)"
//	@Param			search		query	string	false	"Shop or owner name (partial match)"
//	@Param			lat			query	number	false	"Latitude for radius filter"
//	@Param			lng			query	number	false	"Longitude for radius filter"
//	@Param			radius_km	query	number	false	"Radius in km around lat/lng"
//	@Param			limit		query	int		false	"Max results (default 50, max 200)"
//	@Router			/admin/shops/search [get]
//	@Success		200	{object}	response.Response{}	"Successfully searched shops"
//	@Failure		500	{object}	response.Response{}	"Failed to search shops"
func (h *adminHandler) SearchShops(ctx *gin.Context) {
	var filter request.ShopSearch
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid search filters", err, nil)
		return
	}

	shops, err := h.adminUseCase.SearchShops(ctx, filter)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to search shops", err, nil)
		return
	}

	for i := range shops {
		shops[i].Shop_Image_URL = cloud.ResolveURL(h.cloudService, shops[i].Shop_Image_URL)
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully searched shops", shops)
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
	if shopIDStr == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop ID", nil, nil)
		return
	}

	shop, err := h.adminUseCase.GetShopByID(ctx, shopIDStr)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get shop by ID", err, nil)
		return
	}

	shop.Shop_Image_URL = cloud.ResolveURL(h.cloudService, shop.Shop_Image_URL)
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

	if raw, ok := updateFields["shop_type"]; ok {
		if st, ok := raw.(string); !ok || !domain.ShopType(st).IsValid() {
			response.ErrorResponse(ctx, http.StatusBadRequest, "invalid shop_type: must be retail, wholesale, or service", nil, nil)
			return
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

	if raw, ok := shopDetails["shop_type"]; ok {
		if st, ok := raw.(string); !ok || !domain.ShopType(st).IsValid() {
			response.ErrorResponse(ctx, http.StatusBadRequest, "invalid shop_type: must be retail, wholesale, or service", nil, nil)
			return
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
	if adminId == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid owner ID", nil, nil)
		return
	}

	shop, err := h.adminUseCase.GetShopByOwnerID(ctx, adminId)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get shop by owner ID", err, nil)
		return
	}

	shop.Shop_Image_URL = cloud.ResolveURL(h.cloudService, shop.Shop_Image_URL)
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
func (c *adminHandler) SendNotificationToUser(ctx context.Context, userID string, message string) error {
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
		log.Printf("Multipart form files: %+v", ctx.Request.MultipartForm.File)
	} else {
		log.Println("No multipart form files")
	}

	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Image file is required", err, nil)
		return
	}

	fileHeader := req.Image
	processedPath, err := handleUpload(fileHeader)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to process image", err, nil)
		return
	}
	defer func() {
		if rerr := os.Remove(processedPath); rerr != nil && !os.IsNotExist(rerr) {
			log.Printf("failed to remove temp file %s: %v", processedPath, rerr)
		}
	}()

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
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("failed to close file: %v", cerr)
		}
	}()

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

	fileExt := getFileExtension(req.Image.Filename)
	newFileName := fmt.Sprintf("admin_%s_%d%s", adminId, time.Now().Unix(), fileExt)

	objectKey, err := a.cloudService.SaveFile(ctx, req.Image, cloud.SaveOptions{
		Namespace:   "admin-profiles",
		Visibility:  cloud.VisibilityPublic,
		ContentType: contentType,
		Filename:    newFileName,
	})
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload profile image", err, nil)
		return
	}

	if _, err := a.adminUseCase.UploadAdminProfileImage(ctx, adminId, objectKey, shopId); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update admin profile image", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully uploaded profile image", map[string]interface{}{
		"image_url": a.cloudService.PublicURL(objectKey),
		"file_path": objectKey,
	})

}

// UploadBusinessDocument godoc
// @summary api for admin to upload a business document
// @id UploadBusinessDocument
// @tags Admin Shop
// @Param document formData file true "Business Document"
// @Router /admin/shops/business-document/upload [post]
// @Success 200 {object} response.Response{} "Successfully uploaded business document"
// @Failure 400 {object} response.Response{} "invalid input"
// @Failure 401 {object} response.Response{} "unauthorized"
// @Failure 500 {object} response.Response{} "failed to upload business document"
func (a *adminHandler) UploadBusinessDocument(ctx *gin.Context) {
	// Parse the multipart form (10 MB limit)
	if err := ctx.Request.ParseMultipartForm(10 << 20); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to parse multipart form", err, nil)
		return
	}

	fileHeader, err := ctx.FormFile("document")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Document file is required", err, nil)
		return
	}

	// Validate file size (max 10 MB)
	if fileHeader.Size > 10<<20 {
		response.ErrorResponse(ctx, http.StatusBadRequest, "File size exceeds 10MB limit", fmt.Errorf("file size %d exceeds limit", fileHeader.Size), nil)
		return
	}

	// Extract owner ID from token
	tokenString := ctx.GetHeader("Authorization")
	ownerID := a.adminUseCase.DecodeTokenData(tokenString)
	if ownerID == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid token data", fmt.Errorf("failed to decode owner ID from token"), nil)
		return
	}

	// Build unique filename and namespace for S3
	timestamp := time.Now().Unix()
	uniqueFilename := fmt.Sprintf("%d-%s", timestamp, fileHeader.Filename)
	namespace := fmt.Sprintf("business-docs/%s", ownerID)
	// Final key will be: business-docs/{ownerID}/{timestamp}-{filename}

	// Determine content type
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	savedKey, err := a.cloudService.SaveFile(ctx, fileHeader, cloud.SaveOptions{
		Namespace:   namespace,
		Visibility:  cloud.VisibilityPublic,
		ContentType: contentType,
		Filename:    uniqueFilename,
	})
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload business document", err, nil)
		return
	}

	fileURL := a.cloudService.PublicURL(savedKey)
	response.SuccessResponse(ctx, http.StatusOK, "Successfully uploaded business document", map[string]interface{}{
		"document_url": fileURL,
	})
}

// AddAdminProfile godoc
// @summary api for admin to add profile
// @id AddAdminProfile
// @tags Admin Account
// @Param input body object true "inputs"
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
// @Param input body object true "inputs"
// @Router /admin/account [put]
// @Success 200 {object} response.Response{} "Successfully updated admin profile"
// @Failure 400 {object} response.Response{} "invalid input"
// @Failure 500 {object} response.Response{} "failed to update admin profile"
func (a *adminHandler) UpdateAdminProfile(ctx *gin.Context) {
	// Implementation goes here
}

func getFileExtension(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return filename[idx:]
	}
	return ""
}

func (a *adminHandler) UploadShopDocument(ctx *gin.Context) {
	// UploadShopDocument godoc
	//
	//	@Summary		Send shop business document OTP (Admin)
	//	@Security		BearerAuth
	//	@Description	API for admin to send OTP for shop business document verification
	//	@Id				UploadShopDocument
	//	@Tags			Admin Shop
	//	@Param			input	body	request.DocumentRequest{}	true	"Document details"
	//	@Router			/admin/shops/business-document/send-otp [post]
	//	@Success		200	{object}	response.Response{}	"Successfully uploaded shop document OTP sent"
	//	@Failure		400	{object}	response.Response{}	"Invalid input or document missing"
	//	@Failure		401	{object}	response.Response{}	"Invalid token"
	//	@Failure		500	{object}	response.Response{}	"Failed to upload shop document"
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

	// Call use case to upload shop document
	if err := a.adminUseCase.UploadShopDocument(ctx.Request.Context(), shopOwnerIdStr, req.DocumentType, req.DocumentValue); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload shop document", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully uploaded shop document", nil)
}

func (a *adminHandler) UploadAddress(ctx *gin.Context) {
	// UploadAddress godoc
	//
	//	@Summary		Save shop address (Admin)
	//	@Security		BearerAuth
	//	@Description	API for admin to save or update the shop address
	//	@Id				UploadAddress
	//	@Tags			Admin Shop
	//	@Param			input	body	request.AddressRequest{}	true	"Address details"
	//	@Router			/admin/shops/address-details/save [post]
	//	@Success		200	{object}	response.Response{}	"Successfully uploaded address"
	//	@Failure		400	{object}	response.Response{}	"Invalid input"
	//	@Failure		401	{object}	response.Response{}	"Invalid token"
	//	@Failure		500	{object}	response.Response{}	"Failed to upload address"
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
	// VerifyShopDocument godoc
	//
	//	@Summary		Verify shop business document OTP (Admin)
	//	@Security		BearerAuth
	//	@Description	API for admin to verify OTP for shop business document
	//	@Id				VerifyShopDocumentOtp
	//	@Tags			Admin Shop
	//	@Param			input	body	request.VerifyShopDocumentRequest{}	true	"OTP to verify"
	//	@Router			/admin/shops/business-document/verify-otp [post]
	//	@Success		200	{object}	response.Response{}	"Successfully verified shop document"
	//	@Failure		400	{object}	response.Response{}	"OTP is required or invalid"
	//	@Failure		500	{object}	response.Response{}	"Failed to verify shop document"
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
	// AdminDocumentOtpSend godoc
	//
	//	@Summary		Send admin identity document OTP (Admin)
	//	@Security		BearerAuth
	//	@Description	API for admin to send OTP for identity document verification
	//	@Id				AdminDocumentOtpSend
	//	@Tags			Admin Shop
	//	@Param			input	body	request.DocumentRequest{}	true	"Document details"
	//	@Router			/admin/shops/identity-document/send-otp [post]
	//	@Success		200	{object}	response.Response{}	"Successfully sent OTP for admin document"
	//	@Failure		400	{object}	response.Response{}	"Invalid input or document missing"
	//	@Failure		401	{object}	response.Response{}	"Invalid token"
	//	@Failure		500	{object}	response.Response{}	"Failed to send OTP for admin document"
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
	// AdminDocumentOtpVerify godoc
	//
	//	@Summary		Verify admin identity document OTP (Admin)
	//	@Security		BearerAuth
	//	@Description	API for admin to verify OTP for identity document
	//	@Id				AdminDocumentOtpVerify
	//	@Tags			Admin Shop
	//	@Param			input	body	request.VerifyShopDocumentRequest{}	true	"OTP and document details"
	//	@Router			/admin/shops/identity-document/verify-otp [post]
	//	@Success		200	{object}	response.Response{}	"Successfully verified admin document"
	//	@Failure		400	{object}	response.Response{}	"OTP is required or invalid"
	//	@Failure		500	{object}	response.Response{}	"Failed to verify admin document"
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
	// GetVerificationStatus godoc
	//
	//	@Summary		Get shop verification status (Admin)
	//	@Security		BearerAuth
	//	@Description	API for admin to retrieve their shop verification status and admin profile
	//	@Id				GetVerificationStatus
	//	@Tags			Admin Shop
	//	@Router			/admin/shops/verify-status [get]
	//	@Success		200	{object}	response.Response{}	"Successfully retrieved verification status"
	//	@Failure		401	{object}	response.Response{}	"Invalid token"
	//	@Failure		500	{object}	response.Response{}	"Failed to get verification status"
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
	// GetAllSubCategories godoc
	//
	//	@Summary		Get all sub-categories (Admin)
	//	@Security		BearerAuth
	//	@Description	API for admin to retrieve all sub-categories across all categories
	//	@Id				GetAllSubCategories
	//	@Tags			Admin Category
	//	@Router			/admin/sub-categories [get]
	//	@Success		200	{object}	response.Response{}	"Successfully retrieved all sub-categories"
	//	@Failure		500	{object}	response.Response{}	"Failed to get sub-categories"
	subCategories, err := a.productUseCase.GetAllSubCategories(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get sub-categories", err, nil)
		return
	}

	response.ResolveSubCategoriesImages(a.cloudService, subCategories)

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved all sub-categories", subCategories)
}

func (a *adminHandler) GetShopProfileImageById(ctx *gin.Context) {
	// GetShopProfileImageById godoc
	//
	//	@Summary		Get shop profile image by ID (Admin)
	//	@Security		BearerAuth
	//	@Description	API for admin to get the profile image URL for a specific shop
	//	@Id				GetShopProfileImageById
	//	@Tags			Admin Shop
	//	@Param			shop_id	path	string	true	"Shop ID"
	//	@Router			/admin/shops/shop-profile-image/{shop_id} [get]
	//	@Success		200	{object}	response.Response{}	"Successfully retrieved shop profile image"
	//	@Failure		500	{object}	response.Response{}	"Failed to get shop profile image"
	shopId := ctx.Param("shop_id")

	imageURL, err := a.adminUseCase.GetShopProfileImageById(ctx, shopId)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get shop profile image", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved shop profile image", map[string]interface{}{
		"image_url": cloud.ResolveURL(a.cloudService, imageURL),
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
	if shopIDStr == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop_id", nil, nil)
		return
	}

	log.Printf("Received request to set shop time for shop_id: %s", shopIDStr)

	var req request.SetShopTimeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	log.Printf("Parsed request body: %+v", req)

	status := domain.ShopTimeClose
	if req.Status {
		status = domain.ShopTimeOpen
	}

	shopTime := domain.ShopTime{
		Status:    status,
		OpenTime:  req.OpenTime,
		CloseTime: req.CloseTime,
	}

	err := a.shopTimeUseCase.SetShopTime(ctx, shopIDStr, shopTime)
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
	if shopIDStr == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop_id", nil, nil)
		return
	}

	shopTime, err := a.shopTimeUseCase.GetShopTime(ctx, shopIDStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ErrorResponse(ctx, http.StatusNotFound, "Shop time not configured", domain.NotFoundError("shop time not found for this shop"), nil)
			return
		}
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get shop time", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved shop time", shopTime)
}

func (h *adminHandler) GetShopSocialDetails(ctx *gin.Context) {
	// GetShopSocialDetails godoc (Admin)
	//
	//	@Summary		Get shop social details (Admin)
	//	@Security		BearerAuth
	//	@Description	API for admin to get social media and contact details of a shop
	//	@Id				GetShopSocialDetailsAdmin
	//	@Tags			Admin Shop
	//	@Param			shop_id	path	int	true	"Shop ID"
	//	@Router			/admin/shops/{shop_id}/social [get]
	//	@Success		200	{object}	response.Response{}	"Shop social details fetched"
	//	@Failure		400	{object}	response.Response{}	"Invalid shop_id"
	//	@Failure		500	{object}	response.Response{}	"Failed to fetch shop social details"
	shopIDStr := ctx.Param("shop_id")
	if shopIDStr == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop_id", nil, nil)
		return
	}
	details, err := h.adminUseCase.GetShopSocialDetails(ctx, shopIDStr)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch shop social details", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Shop social details fetched", details)
}

func (a *adminHandler) GetDashboardStats(ctx *gin.Context) {
	stats, err := a.adminUseCase.GetDashboardStats(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch dashboard stats", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "dashboard stats fetched", stats)
}

// notificationImageNamespace is the object-storage prefix under which images
// uploaded for use in push notifications (welcome banners, frames, etc.) are kept.
const notificationImageNamespace = "notifications"

// UploadNotificationImage godoc
//
//	@Summary		Upload an image for use in push notifications
//	@Security		BearerAuth
//	@Id				UploadNotificationImage
//	@Tags			Notification
//	@Accept			mpfd
//	@Produce		json
//	@Param			photo	formData	file	true	"Image to upload"
//	@Router			/admin/notifications/images [post]
//	@Success		201	{object}	response.Response{}	"Image uploaded"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Upload failed"
func (a *adminHandler) UploadNotificationImage(ctx *gin.Context) {
	fh, err := ctx.FormFile("photo")
	if err != nil || fh == nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "photo file is required", err, nil)
		return
	}

	objectKey, uploadErr := a.cloudService.SaveFile(ctx, fh, cloud.SaveOptions{
		Namespace:  notificationImageNamespace,
		Visibility: cloud.VisibilityPublic,
	})
	if uploadErr != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload image", uploadErr, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Image uploaded", gin.H{
		"object_key": objectKey,
		"url":        a.cloudService.PublicURL(objectKey),
	})
}

// ListNotificationImages godoc
//
//	@Summary		List previously uploaded notification images
//	@Security		BearerAuth
//	@Id				ListNotificationImages
//	@Tags			Notification
//	@Produce		json
//	@Router			/admin/notifications/images [get]
//	@Success		200	{object}	response.Response{}	"Images fetched"
//	@Failure		500	{object}	response.Response{}	"Failed to list images"
func (a *adminHandler) ListNotificationImages(ctx *gin.Context) {
	keys, err := a.cloudService.ListObjects(ctx, notificationImageNamespace+"/")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list images", err, nil)
		return
	}

	images := make([]gin.H, 0, len(keys))
	for _, key := range keys {
		images = append(images, gin.H{
			"object_key": key,
			"url":        a.cloudService.PublicURL(key),
		})
	}
	response.SuccessResponse(ctx, http.StatusOK, "Images fetched", images)
}

// ── Subscription GST configuration ─────────────────────────────────────────

// subscriptionGSTResponse mirrors the rate in both representations so the
// admin-portal UI can work in friendly percent while the backend keeps the
// basis-points precision used by SubscriptionOrder snapshots.
type subscriptionGSTResponse struct {
	GSTRateBasisPoints int     `json:"gst_rate_basis_points"`
	GSTRatePercent     float64 `json:"gst_rate_percent"`
}

// GetSubscriptionGSTConfig godoc
//
//	@summary		Get the current subscription GST rate
//	@Security		BearerAuth
//	@Id				GetSubscriptionGSTConfig
//	@Tags			Subscription Pricing
//	@Router			/admin/subscription-pricing/gst [get]
//	@Success		200	{object}	response.Response{}
func (a *adminHandler) GetSubscriptionGSTConfig(ctx *gin.Context) {
	cfg, err := a.adminUseCase.GetSubscriptionGSTConfig(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to load GST config", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Subscription GST config", subscriptionGSTResponse{
		GSTRateBasisPoints: cfg.GSTRateBasisPoints,
		GSTRatePercent:     float64(cfg.GSTRateBasisPoints) / 100,
	})
}

// UpdateSubscriptionGSTConfig godoc
//
//	@summary		Update the subscription GST rate
//	@Description	Accepts either gst_rate_percent (e.g. 18) or gst_rate_basis_points (e.g. 1800); percent wins if both are sent. Only affects subscriptions purchased after the change — existing orders keep their snapshotted rate.
//	@Security		BearerAuth
//	@Id				UpdateSubscriptionGSTConfig
//	@Tags			Subscription Pricing
//	@Router			/admin/subscription-pricing/gst [put]
//	@Success		200	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}
func (a *adminHandler) UpdateSubscriptionGSTConfig(ctx *gin.Context) {
	var body struct {
		GSTRatePercent     *float64 `json:"gst_rate_percent"`
		GSTRateBasisPoints *int     `json:"gst_rate_basis_points"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	var basisPoints int
	switch {
	case body.GSTRatePercent != nil:
		basisPoints = int(*body.GSTRatePercent*100 + 0.5) // round to nearest basis point
	case body.GSTRateBasisPoints != nil:
		basisPoints = *body.GSTRateBasisPoints
	default:
		appErr := domain.ValidationError("body", "gst_rate_percent or gst_rate_basis_points is required")
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr, nil)
		return
	}

	updated, err := a.adminUseCase.UpdateSubscriptionGSTConfig(ctx, domain.SubscriptionGSTConfig{GSTRateBasisPoints: basisPoints})
	if err != nil {
		a.handleAppErr(ctx, "Failed to update GST config", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Subscription GST rate updated", subscriptionGSTResponse{
		GSTRateBasisPoints: updated.GSTRateBasisPoints,
		GSTRatePercent:     float64(updated.GSTRateBasisPoints) / 100,
	})
}

// ── Shop phone number change (OTP-gated) ────────────────────────────────────

// SendShopPhoneChangeOtp godoc
//
//	@summary		Send an OTP to a new phone number before it can be saved to the shop profile
//	@Security		BearerAuth
//	@Id				SendShopPhoneChangeOtp
//	@Tags			Admin Shop
//	@Router			/admin/shops/phone/send-otp [post]
//	@Success		200	{object}	response.Response{}
func (a *adminHandler) SendShopPhoneChangeOtp(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	var body struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	otpID, err := a.adminUseCase.SendShopPhoneChangeOtp(ctx, callerID, body.Phone)
	if err != nil {
		a.handleAppErr(ctx, "Failed to send OTP", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "OTP sent", gin.H{"otp_id": otpID})
}

// VerifyShopPhoneChangeOtp godoc
//
//	@summary		Verify the OTP for a new phone number and save it to the shop profile
//	@Security		BearerAuth
//	@Id				VerifyShopPhoneChangeOtp
//	@Tags			Admin Shop
//	@Router			/admin/shops/phone/verify-otp [post]
//	@Success		200	{object}	response.Response{}
//	@Failure		400	{object}	response.Response{}	"Invalid or expired OTP"
func (a *adminHandler) VerifyShopPhoneChangeOtp(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	var body struct {
		OtpID string `json:"otp_id" binding:"required"`
		Otp   string `json:"otp" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	phone, err := a.adminUseCase.VerifyShopPhoneChangeOtp(ctx, callerID, body.OtpID, body.Otp)
	if err != nil {
		a.handleAppErr(ctx, "Invalid or expired OTP", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Phone number updated", gin.H{"phone": phone})
}

// ── Roles & permissions ─────────────────────────────────────────────────────

// meResponse is what the admin-portal reads right after login (and on
// session restore) to know its own identity, role, and — critically — its
// resolved permission set, since permissions are now DB-driven per role
// rather than a static frontend lookup table.
type meResponse struct {
	ID          string                `json:"id"`
	FullName    string                `json:"full_name"`
	Email       string                `json:"email"`
	Role        string                `json:"role"`
	RoleLabel   string                `json:"role_label"`
	Permissions []domain.PermissionKey `json:"permissions"`
}

// GetMyPermissions godoc
//
//	@summary		Get the authenticated admin's identity, role, and resolved permission set
//	@Security		BearerAuth
//	@Id				GetMyPermissions
//	@Tags			Admin Roles
//	@Router			/admin/me [get]
//	@Success		200	{object}	response.Response{}
func (a *adminHandler) GetMyPermissions(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	caller, err := a.adminUseCase.GetAdminByID(ctx, callerID)
	if err != nil || caller.ID == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid or expired session", err, nil)
		return
	}

	perms, err := a.adminUseCase.GetPermissionsForRole(ctx, string(caller.Role))
	if err != nil {
		a.handleAppErr(ctx, "Failed to resolve permissions", err)
		return
	}

	roleLabel := string(caller.Role)
	if role, err := a.adminUseCase.GetRoleByName(ctx, string(caller.Role)); err == nil && role.ID != "" {
		roleLabel = role.Label
	} else if caller.Role == domain.AdminRoleSuperAdmin {
		roleLabel = "Super Admin"
	}

	response.SuccessResponse(ctx, http.StatusOK, "Fetched current admin", meResponse{
		ID:          caller.ID,
		FullName:    caller.FullName,
		Email:       caller.Email,
		Role:        string(caller.Role),
		RoleLabel:   roleLabel,
		Permissions: perms,
	})
}

// CreateRole godoc
//
//	@summary		Create a custom admin-portal role
//	@Security		BearerAuth
//	@Id				CreateRole
//	@Tags			Admin Roles
//	@Router			/admin/roles [post]
//	@Success		201	{object}	response.Response{}
func (a *adminHandler) CreateRole(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	var body domain.Role
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	role, err := a.adminUseCase.CreateRole(ctx, callerID, body)
	if err != nil {
		a.handleAppErr(ctx, "Failed to create role", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Role created successfully", role)
}

// ListRoles godoc
//
//	@summary		List every admin-portal role with its permissions
//	@Security		BearerAuth
//	@Id				ListRoles
//	@Tags			Admin Roles
//	@Router			/admin/roles [get]
//	@Success		200	{object}	response.Response{}
func (a *adminHandler) ListRoles(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	roles, err := a.adminUseCase.ListRoles(ctx, callerID)
	if err != nil {
		a.handleAppErr(ctx, "Failed to list roles", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Roles fetched successfully", roles)
}

// GetRole godoc
//
//	@summary		Get a single role with its permissions
//	@Security		BearerAuth
//	@Id				GetRole
//	@Tags			Admin Roles
//	@Router			/admin/roles/{role_id} [get]
//	@Success		200	{object}	response.Response{}
func (a *adminHandler) GetRole(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	role, err := a.adminUseCase.GetRole(ctx, callerID, ctx.Param("role_id"))
	if err != nil {
		a.handleAppErr(ctx, "Failed to get role", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Role fetched successfully", role)
}

// UpdateRole godoc
//
//	@summary		Update a role's label and/or permission set
//	@Security		BearerAuth
//	@Id				UpdateRole
//	@Tags			Admin Roles
//	@Router			/admin/roles/{role_id} [put]
//	@Success		200	{object}	response.Response{}
func (a *adminHandler) UpdateRole(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	var body domain.Role
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	role, err := a.adminUseCase.UpdateRole(ctx, callerID, ctx.Param("role_id"), body)
	if err != nil {
		a.handleAppErr(ctx, "Failed to update role", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Role updated successfully", role)
}

// DeleteRole godoc
//
//	@summary		Delete a custom role (built-in roles and roles still assigned to admins can't be deleted)
//	@Security		BearerAuth
//	@Id				DeleteRole
//	@Tags			Admin Roles
//	@Router			/admin/roles/{role_id} [delete]
//	@Success		200	{object}	response.Response{}
func (a *adminHandler) DeleteRole(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	if err := a.adminUseCase.DeleteRole(ctx, callerID, ctx.Param("role_id")); err != nil {
		a.handleAppErr(ctx, "Failed to delete role", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Role deleted successfully")
}

// ── Sales Executive referral program ───────────────────────────────────────

func (a *adminHandler) handleAppErr(ctx *gin.Context, fallbackMsg string, err error) {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr, nil)
		return
	}
	response.ErrorResponse(ctx, http.StatusInternalServerError, fallbackMsg, err, nil)
}

// CreateShopReferral godoc
//
//	@Summary		Manually attach a shop to a Sales Executive
//	@Security		BearerAuth
//	@Description	Super-admin-only. Sellers normally get attached automatically by submitting a referral_coupon_id when creating their shop.
//	@Id				CreateShopReferral
//	@Tags			Admin Referrals
//	@Router			/admin/referrals [post]
//	@Success		201	{object}	response.Response{}	"Referral created"
func (a *adminHandler) CreateShopReferral(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	var body domain.ShopReferral
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	if body.ID != "" {
		appErr := domain.ValidationError("id", "must not be provided when creating a new referral")
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr, nil)
		return
	}
	referral, err := a.adminUseCase.CreateShopReferral(ctx, callerID, body)
	if err != nil {
		a.handleAppErr(ctx, "Failed to create referral", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Referral created successfully", referral)
}

// ListShopReferrals godoc
//
//	@Summary		List all shop referral attachments
//	@Security		BearerAuth
//	@Id				ListShopReferrals
//	@Tags			Admin Referrals
//	@Param			page_number	query	int	false	"Page Number"
//	@Param			count		query	int	false	"Count"
//	@Router			/admin/referrals [get]
//	@Success		200	{object}	response.Response{}	"Referrals fetched"
func (a *adminHandler) ListShopReferrals(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	pagination := request.GetPagination(ctx)
	referrals, err := a.adminUseCase.ListShopReferrals(ctx, callerID, pagination)
	if err != nil {
		a.handleAppErr(ctx, "Failed to list referrals", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Referrals fetched successfully", referrals)
}

// GetShopReferral godoc
//
//	@Summary		Get a single referral attachment
//	@Security		BearerAuth
//	@Id				GetShopReferral
//	@Tags			Admin Referrals
//	@Router			/admin/referrals/{referral_id} [get]
//	@Success		200	{object}	response.Response{}	"Referral fetched"
func (a *adminHandler) GetShopReferral(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	referralID := ctx.Param("referral_id")
	referral, err := a.adminUseCase.GetShopReferral(ctx, callerID, referralID)
	if err != nil {
		a.handleAppErr(ctx, "Failed to get referral", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Referral fetched successfully", referral)
}

// UpdateShopReferral godoc
//
//	@Summary		Update a referral attachment (status or reassignment)
//	@Security		BearerAuth
//	@Id				UpdateShopReferral
//	@Tags			Admin Referrals
//	@Router			/admin/referrals/{referral_id} [put]
//	@Success		200	{object}	response.Response{}	"Referral updated"
func (a *adminHandler) UpdateShopReferral(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	referralID := ctx.Param("referral_id")
	var body domain.ShopReferral
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	if err := a.adminUseCase.UpdateShopReferral(ctx, callerID, referralID, body); err != nil {
		a.handleAppErr(ctx, "Failed to update referral", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Referral updated successfully")
}

// DeleteShopReferral godoc
//
//	@Summary		Delete a referral attachment
//	@Security		BearerAuth
//	@Id				DeleteShopReferral
//	@Tags			Admin Referrals
//	@Router			/admin/referrals/{referral_id} [delete]
//	@Success		200	{object}	response.Response{}	"Referral deleted"
func (a *adminHandler) DeleteShopReferral(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	referralID := ctx.Param("referral_id")
	if err := a.adminUseCase.DeleteShopReferral(ctx, callerID, referralID); err != nil {
		a.handleAppErr(ctx, "Failed to delete referral", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Referral deleted successfully")
}

// ListShopsForSalesExecutive godoc
//
//	@Summary		List shops attached to a Sales Executive (a platform user)
//	@Security		BearerAuth
//	@Description	A Sales Executive may only view their own shops; a super admin may view any executive's shops.
//	@Id				ListShopsForSalesExecutive
//	@Tags			Admin Referrals
//	@Param			admin_id	path	string	true	"Platform user (Sales Executive) ID"
//	@Router			/admin/platform-users/{admin_id}/shops [get]
//	@Success		200	{object}	response.Response{}	"Attached shops fetched"
func (a *adminHandler) ListShopsForSalesExecutive(ctx *gin.Context) {
	callerID := a.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
	platformUserID := ctx.Param("admin_id")
	referrals, err := a.adminUseCase.ListShopsForSalesExecutive(ctx, callerID, platformUserID, request.GetPagination(ctx))
	if err != nil {
		a.handleAppErr(ctx, "Failed to list attached shops", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Attached shops fetched successfully", referrals)
}
