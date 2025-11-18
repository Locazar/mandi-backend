package handler

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type adminHandler struct {
	adminUseCase usecaseInterface.AdminUseCase
}

func NewAdminHandler(adminUsecase usecaseInterface.AdminUseCase) interfaces.AdminHandler {
	return &adminHandler{
		adminUseCase: adminUsecase,
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

	err := a.adminUseCase.SignUp(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to create account for admin", err, nil)
		return
	}

	response.SuccessResponse(ctx, 200, "Successfully account created for admin", nil)
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
		response.SuccessResponse(ctx, http.StatusNoContent, "No users found", nil)
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
		response.SuccessResponse(ctx, http.StatusNoContent, "No sales report found", nil)
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

	var body domain.ShopVerification

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
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	// Call use case to create shop
	_, err := h.adminUseCase.CreateShop(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create shop", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully created shop", nil)
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
	var body domain.ShopDetails

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	// Call use case to update shop
	_, err := h.adminUseCase.UpdateShop(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update shop", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated shop", nil)
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
	ownerIDStr := ctx.Param("owner_id")
	ownerID, err := strconv.ParseUint(ownerIDStr, 10, 64)
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
