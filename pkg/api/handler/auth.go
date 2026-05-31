package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

const (
	authorizationHeaderKey = "Authorization"
	authorizationType      = "Bearer"
)

type AuthHandler struct {
	authUseCase  usecaseInterface.AuthUseCase
	config       config.Config
	tokenService token.TokenService
}

func NewAuthHandler(authUsecase usecaseInterface.AuthUseCase, config config.Config, tokenService token.TokenService) interfaces.AuthHandler {
	return &AuthHandler{
		authUseCase:  authUsecase,
		config:       config,
		tokenService: tokenService,
	}
}

// UserLogin godoc
//
//	@Summary		Login with password (User)
//	@Description	API for user to login with email | phone | user_name with password
//	@Id				UserLogin
//	@Tags			User Authentication
//	@Param			inputs	body	request.Login{}	true	"Login Details"
//	@Router			/auth/sign-in [post]
//	@Success		200	{object}	response.Response{data=response.TokenResponse}	"Successfully logged in"
//	@Failure		400	{object}	response.Response{}								"Invalid inputs"
//	@Failure		403	{object}	response.Response{}								"User blocked by admin"
//	@Failure		401	{object}	response.Response{}								"User not exist with given login credentials"
//	@Failure		500	{object}	response.Response{}								"Failed to login"
func (c *AuthHandler) UserLogin(ctx *gin.Context) {

	var body request.Login

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	userID, err := c.authUseCase.UserLogin(ctx, body)

	if err != nil {
		errResponse(ctx, "Failed to login", err)
		return
	}

	// common functionality for admin and user
	c.setupTokenAndResponse(ctx, token.User, userID)
}

// UserLoginOtpSend godoc
//
//	@Summary		Login with Otp send (User)
//	@Description	API for user to send otp for login enter email | phone | user_name : otp will send to user registered number
//	@Id				UserLoginOtpSend
//	@Tags			User Authentication
//	@Param			inputs	body	request.OTPLogin{}	true	"Login credentials"
//	@Router			/auth/sign-in/otp/send [post]
//	@Success		200	{object}	response.Response{response.OTPResponse{}}	"Successfully otp send to user's registered number"
//	@Failure		400	{object}	response.Response{}							"Invalid Otp"
//	@Failure		403	{object}	response.Response{}							"User blocked by admin"
//	@Failure		401	{object}	response.Response{}							"User not exist with given login credentials"
//	@Failure		500	{object}	response.Response{}							"Failed to send otp"
func (u *AuthHandler) UserLoginOtpSend(ctx *gin.Context) {

	var body request.OTPLogin
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	//check all input field is empty
	if body.Email == "" && body.Phone == "" {
		err := errors.New("enter at least user_name or email or phone")
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	otpID, err := u.authUseCase.UserLoginOtpSend(ctx, body)

	if err != nil {
		errResponse(ctx, "Failed to send otp", err)
		return
	}

	otpRes := response.OTPResponse{
		OtpID: otpID,
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully otp send to user's registered number", otpRes)
}

// UserLoginOtpVerify godoc
//
//	@summary		Login with Otp verify (User)
//	@description	API for user to verify otp
//	@id				UserLoginOtpVerify
//	@tags			User Authentication
//	@param			inputs	body	request.OTPVerify{}	true	"Otp Verify Details"
//	@Router			/auth/sign-in/otp/verify [post]
//	@Success		200	{object}	response.Response{data=response.TokenResponse}	"Successfully user logged in"
//	@Failure		400	{object}	response.Response{}								"Invalid inputs"
//	@Failure		401	{object}	response.Response{}								"Otp not matched"
//	@Failure		410	{object}	response.Response{}								"Otp Expired"
//	@Failure		500	{object}	response.Response{}								"Failed to verify otp
func (c *AuthHandler) UserLoginOtpVerify(ctx *gin.Context) {

	var body request.OTPVerify
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	// get the user using loginOtp useCase
	userID, err := c.authUseCase.LoginOtpVerify(ctx, body)
	if err != nil {
		errResponse(ctx, "Failed to verify otp", err)
		return
	}

	c.setupTokenAndResponse(ctx, token.User, userID)
}

// UserSignUp godoc
//
//	@Summary		Signup (User)
//	@Description	API for user to register a new account
//	@Id				UserSignUp
//	@Tags			User Authentication
//	@Param			input	body	request.UserSignUp{}	true	"Input Fields"
//	@Router			/auth/sign-up [post]
//	@Success		200	{object}	response.Response{data=response.OTPResponse}	"Successfully account created and otp send to registered number"
//	@Failure		400	{object}	response.Response{}								"Invalid input"
//	@Failure		409	{object}	response.Response{}								"A verified user already exist with given user credentials"
//	@Failure		500	{object}	response.Response{}								"Failed to signup"
func (c *AuthHandler) UserSignUp(ctx *gin.Context) {

	var body request.UserSignUp

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	var user domain.User
	if err := copier.Copy(&user, body); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "failed to copy details", err, nil)
		return
	}

	otpID, err := c.authUseCase.UserSignUp(ctx, user)

	if err != nil {
		errResponse(ctx, "Failed to signup", err)
		return
	}

	otpRes := response.OTPResponse{
		OtpID: otpID,
	}

	response.SuccessResponse(ctx, http.StatusCreated,
		"Successfully account created and otp send to registered number", otpRes)
}

// UserSignUpVerify godoc
//
//	@summary		UserSingUp verify OTP  (User)
//	@description	API for user to verify otp on sign up
//	@id				UserSignUpVerify
//	@tags			User Authentication
//	@param			inputs	body	request.OTPVerify{}	true	"Otp Verify Details"
//	@Router			/auth/sign-up/verify [post]
//	@Success		200	{object}	response.Response{data=response.TokenResponse}	"Successfully otp verified for user sign up"
//	@Failure		400	{object}	response.Response{}								"Invalid inputs"
//	@Failure		401	{object}	response.Response{}								"Otp not matched"
//	@Failure		410	{object}	response.Response{}								"Otp Expired"
//	@Failure		500	{object}	response.Response{}								"Failed to verify otp"
func (c *AuthHandler) UserSignUpVerify(ctx *gin.Context) {

	var body request.OTPVerify
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	// get the user using loginOtp useCase
	userID, err := c.authUseCase.SingUpOtpVerify(ctx, body)
	if err != nil {
		errResponse(ctx, "Failed to verify otp", err)
		return
	}

	c.setupTokenAndResponse(ctx, token.User, userID)
}

// AdminLogin godoc
//
//	@Summary		Login with password (Admin)
//	@Description	API for admin to login with password
//	@Id				AdminLogin
//	@Tags			Admin Authentication
//	@Param			input	body	request.Login{}	true	"Login credentials"
//	@Router			/admin/auth/sign-in [post]
//	@Success		200	{object}	response.Response{data=response.TokenResponse}	"Successfully logged in"
//	@Failure		400	{object}	response.Response{}								"Invalid input"
//	@Failure		401	{object}	response.Response{}								"Wrong password"
//	@Failure		404	{object}	response.Response{}								"Admin not exist with this details"
//	@Failure		500	{object}	response.Response{}								"Failed to login"
func (c *AuthHandler) AdminLogin(ctx *gin.Context) {

	var body request.Login

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	admin, err := c.authUseCase.AdminLogin(ctx, body)
	if err != nil {
		errResponse(ctx, "Failed to login", err)
		return
	}

	// setup token common part
	role := string(admin.Role)
	if role == "" {
		role = string(domain.AdminRoleSuperAdmin)
	}
	c.setupTokenAndResponse(ctx, token.Admin, admin.ID, map[string]interface{}{
		"role": role,
	})
}

// access and refresh token generating for user and admin is same so created
// a common function for it.(differentiate user by user type )
// customResponse is optional - if provided, it will be used instead of default success response
func (c *AuthHandler) setupTokenAndResponse(ctx *gin.Context, tokenUser token.UserType, userID string, customResponse ...interface{}) {

	tokenParams := usecaseInterface.GenerateTokenParams{
		UserID:   userID,
		UserType: tokenUser,
	}

	accessToken, err := c.authUseCase.GenerateAccessToken(ctx, tokenParams)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to generate access token", err, nil)
		return
	}

	refreshToken, err := c.authUseCase.GenerateRefreshToken(ctx, usecaseInterface.GenerateTokenParams{
		UserID:   userID,
		UserType: tokenUser,
	})
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to generate refresh token", err, nil)
		return
	}

	authorizationValue := authorizationType + " " + accessToken
	ctx.Header(authorizationHeaderKey, authorizationValue)

	ctx.Header("access_token", accessToken)
	ctx.Header("refresh_token", refreshToken)

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
		fmt.Printf("Merged Data: %+v\n", mergedData)
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

	response.SuccessResponse(ctx, http.StatusOK, message, responseData)
}

// UserRenewAccessToken godoc
//
//	@Summary		Renew Access Token (User)
//	@Description	API for user to renew access token using refresh token
//	@Security		ApiKeyAuth
//	@Id				UserRenewAccessToken
//	@Tags			User Authentication
//	@Param			input	body	request.RefreshToken{}	true	"Refresh token"
//	@Router			/auth/renew-access-token [post]
//	@Success		200	{object}	response.Response{}	"Successfully generated access token using refresh token"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		401	{object}	response.Response{}	"Invalid refresh token"
//	@Failure		404	{object}	response.Response{}	"No session found for the given refresh token"
//	@Failure		410	{object}	response.Response{}	"Refresh token expired"
//	@Failure		403	{object}	response.Response{}	"Refresh token blocked"
//	@Failure		500	{object}	response.Response{}	"Failed generate access token"
func (c *AuthHandler) UserRenewAccessToken() gin.HandlerFunc {
	return c.renewAccessToken(token.User)
}

// AdminRenewAccessToken godoc
//
//	@Summary		Renew Access Token (Admin)
//	@Description	API for admin to renew access token using refresh token
//	@Security		ApiKeyAuth
//	@Id				AdminRenewAccessToken
//	@Tags			Admin Authentication
//	@Param			input	body	request.RefreshToken{}	true	"Refresh token"
//	@Router			/admin/auth/renew-access-token [post]
//	@Success		200	{object}	response.Response{}	"Successfully generated access token using refresh token"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		401	{object}	response.Response{}	"Invalid refresh token"
//	@Failure		404	{object}	response.Response{}	"No session found for the given refresh token"
//	@Failure		410	{object}	response.Response{}	"Refresh token expired"
//	@Failure		403	{object}	response.Response{}	"Refresh token blocked"
//	@Failure		500	{object}	response.Response{}	"Failed generate access token"
func (c *AuthHandler) AdminRenewAccessToken() gin.HandlerFunc {
	return c.renewAccessToken(token.Admin)
}

// common functionality of renewing access token for user and admin
func (c *AuthHandler) renewAccessToken(tokenUser token.UserType) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var body request.RefreshToken

		if err := ctx.ShouldBindJSON(&body); err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
			return
		}

		refreshSession, err := c.authUseCase.VerifyAndGetRefreshTokenSession(ctx, body.RefreshToken, tokenUser)

		if err != nil {
			errResponse(ctx, "Failed verify refresh token", err)
			return
		}

		accessTokenParams := usecaseInterface.GenerateTokenParams{
			UserID:   refreshSession.UserID,
			UserType: tokenUser,
		}

		accessToken, err := c.authUseCase.GenerateAccessToken(ctx, accessTokenParams)

		if err != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed generate access token", err, nil)
			return
		}
		cookieName := "auth-" + string(tokenUser)
		ctx.SetCookie(cookieName, accessToken, 15*60, "", "", false, true)

		accessTokenRes := response.TokenResponse{
			AccessToken: accessToken,
		}
		c.setCookie(ctx, cookieName, accessToken, 15*60)
		response.SuccessResponse(ctx, http.StatusOK, "Successfully generated access token using refresh token", accessTokenRes)
	}
}

func (c *AuthHandler) setCookie(ctx *gin.Context, name, value string, maxAge int) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: c.config.Security.CookieHTTPOnly,
		Secure:   c.config.Security.SecureCookies,
		SameSite: parseSameSite(c.config.Security.CookieSameSite),
		MaxAge:   maxAge,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
	}

	http.SetCookie(ctx.Writer, cookie)
}

func parseSameSite(mode string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteLaxMode
	}
}

// UserLoginOtpSendEmail godoc
//
//	@Summary		Login with Otp send email (User)
//	@Description	API for user to send otp for login enter email : otp will send to user registered email
//	@Id				UserLoginOtpSendEmail
//	@Tags			User Authentication
//	@Param			inputs	body	request.OTPLoginEmail{}	true	"Login credentials"
//	@Router			/auth/sign-in/otp/send-email [post]
//	@Success		200	{object}	response.Response{response.OTPResponse{}}	"Successfully otp send to user's registered email"
//	@Failure		400	{object}	response.Response{}							"Invalid Otp"
//	@Failure		403	{object}	response.Response{}							"User blocked by admin"
//	@Failure		401	{object}	response.Response{}							"User not exist with given login credentials"
//	@Failure		500	{object}	response.Response{}							"Failed to send otp"
func (c *AuthHandler) UserLoginOtpSendEmail(ctx *gin.Context) {

	var body request.OTPLoginEmail
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	//check all input field is empty
	if body.Email == "" {
		err := errors.New("enter at least user_name or email or phone")
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	otpID, err := c.authUseCase.UserLoginOtpSendEmail(ctx, body)

	if err != nil {
		errResponse(ctx, "Failed to send otp", err)
		return
	}

	otpRes := response.OTPResponse{
		OtpID: otpID,
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully otp send to user's registered email", otpRes)
}

// UserLoginOtpVerifyEmail godoc
//
//	@summary		Login with Otp verify email (User)
//	@description	API for user to verify otp
//	@id				UserLoginOtpVerifyEmail
//	@tags			User Authentication
//	@param			inputs	body	request.OTPVerify{}	true	"Otp Verify Details"
//	@Router			/auth/sign-in/otp/verify-email [post]
//	@Success		200	{object}	response.Response{data=response.TokenResponse}	"Successfully user logged in"
//	@Failure		400	{object}	response.Response{}								"Invalid inputs"
//	@Failure		401	{object}	response.Response{}								"Otp not matched"
//	@Failure		410	{object}	response.Response{}								"Otp Expired"
//	@Failure		500	{object}	response.Response{}								"Failed to verify otp
func (c *AuthHandler) UserLoginOtpVerifyEmail(ctx *gin.Context) {
	var body request.OTPVerify
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	// get the user using loginOtp useCase
	userID, err := c.authUseCase.LoginOtpVerify(ctx, body)
	if err != nil {
		errResponse(ctx, "Failed to verify otp", err)
		return
	}

	c.setupTokenAndResponse(ctx, token.User, userID)
}

// AdminSignUpOtpSend godoc
// @Summary		Send OTP for Admin Signup
// @Description	API for admin to send OTP for signup
// @Id				AdminSignUpOtpSend
// @Tags			Admin Authentication
// @Param			inputs	body	request.OTPLogin{}	true	"Mobile number to send OTP"
// @Router			/admin/auth/sign-up/otp/send [post]
// @Success		200	{object}	response.Response{response.OTPResponse{}}	"Successfully sent OTP to admin's mobile number"
// @Failure		400	{object}	response.Response{}							"Invalid input"
// @Failure		409	{object}	response.Response{}							"Admin already exists"
// @Failure		500	{object}	response.Response{}							"Failed to send OTP"
func (c *AuthHandler) AdminSignUpOtpSend(ctx *gin.Context) {
	var body request.OTPLogin
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	if body.Phone == "" {
		err := errors.New("mobile number is required")
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	otpID, err := c.authUseCase.AdminSignUpOtpSend(ctx, body.Phone)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to send OTP", err, nil)
		return
	}

	otpRes := response.OTPResponse{
		OtpID: otpID,
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully sent OTP to admin's mobile number", otpRes)
}

// AdminSignUpOtpVerify godoc
// @Summary		Verify OTP for Admin Signup
// @Description	API for admin to verify OTP for signup and either login or register
// @Id				AdminSignUpOtpVerify
// @Tags			Admin Authentication
// @Param			inputs	body	request.OTPVerify{}	true	"OTP verification details"
// @Router			/admin/auth/sign-up/otp/verify [post]
// @Success		200	{object}	response.Response{data=response.TokenResponse}	"Successfully verified OTP and logged in/registered admin"
// @Failure		400	{object}	response.Response{}								"Invalid inputs"
// @Failure		401	{object}	response.Response{}								"OTP not matched"
// @Failure		410	{object}	response.Response{}								"OTP Expired"
// @Failure		500	{object}	response.Response{}								"Failed to verify OTP"
func (c *AuthHandler) AdminSignUpOtpVerify(ctx *gin.Context) {
	var body request.OTPVerify
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	adminID, err := c.authUseCase.AdminSignUpOtpVerify(ctx, body)
	if err != nil {
		errResponse(ctx, "Failed to verify OTP", err)
		return
	}

	c.setupTokenAndResponse(ctx, token.Admin, adminID)
}

func (c *AuthHandler) AdminLogout(ctx *gin.Context) {
	// Get token from Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Authorization header is required", nil, nil)
		return
	}

	// Decode token to get user data
	// Fix: DecodeTokenDataToGetData does not return any values, so just call it without assignment.
	userID, userType, err := c.tokenService.DecodeTokenDataToGetData(authHeader)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid token", err, nil)
		return
	}

	// If you need userID and userType, you should implement or use a method that returns them.
	// For demonstration, we'll just proceed with logout logic.

	// TODO: Replace "0" and "" with actual userID and userType if needed.
	err = c.authUseCase.UserLogout(ctx, userID, string(userType))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to logout admin", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully logged out", nil)
}
