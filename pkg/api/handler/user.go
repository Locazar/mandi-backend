package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type UserHandler struct {
	userUseCase usecaseInterface.UserUseCase
}

func NewUserHandler(userUsecase usecaseInterface.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUsecase,
	}
}

// // Logout godoc
// // @summary api for user to logout
// // @description user can logout
// // @security ApiKeyAuth
// // @id UserLogout
// // @tags User Logout
// // @Router /logout [post]
// // @Success 200 "successfully logged out"
func (u *UserHandler) UserLogout(ctx *gin.Context) {

	ctx.SetCookie("user-auth", "", -1, "", "", false, true)

	response.SuccessResponse(ctx, http.StatusOK, "Successfully logged out", nil)
}

// // CheckOutCart godoc
// // @summary api for cart checkout
// // @description user can checkout user cart items
// // @Security BearerAuth
// // @id CheckOutCart
// // @tags User Cart
// // @Router /carts/checkout [get]
// // @Success 200 {object} response.Response{} "successfully got checkout data"
// // @Failure 401 {object} res.Response{} "cart is empty so user can't call this api"
// // @Failure 500 {object} res.Response{} "failed to get checkout items"
// func (c *UserHandler) CheckOutCart(ctx *gin.Context) {

// 	userId := utils.GetUserIdFromContext(ctx)

// 	resCheckOut, err := c.userUseCase.CheckOutCart(ctx, userId)

// 	if err != nil {
// 		response.ErrorResponse(500, "failed to get checkout items", err.Error(), nil)
// 		ctx.AbortWithStatusJSON(http.StatusInternalServerError, response)
// 		return
// 	}

// 	if resCheckOut.ProductItems == nil {
// 		response.ErrorResponse(401, "cart is empty can't checkout cart", "", nil)
// 		ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
// 		return
// 	}

// 	responser := res.SuccessResponse(200, "successfully got checkout data", resCheckOut)
// 	ctx.JSON(http.StatusOK, responser)
// }

// GetProfile godoc
//
//	@Summary		Get User Profile (User)
//	@Security		BearerAuth
//	@Description	API for user to get all user details
//	@Id				GetProfile
//	@Tags			User Profile
//	@Router			/account [get]
//	@Success		200	"Successfully retrieved user details"
//	@Failure		500	{object}	response.Response{}	"Failed to retrieve user details"
func (u *UserHandler) GetProfile(ctx *gin.Context) {

	userID := utils.GetUserIdFromContext(ctx)

	user, err := u.userUseCase.FindProfile(ctx, userID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve user details", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved user details", user)
}

// UpdateProfile godoc
//
//	@Summary		Edit profile (User)
//	@Security		BearerAuth
//	@Description	API for user to edit user details
//	@Id				UpdateProfile
//	@Tags			User Profile
//	@Param			input	body	request.EditUser{}	true	"User details input"
//	@Router			/account [put]
//	@Success		200	{object}	response.Response{}	"Successfully profile updated"
//	@Failure		400	{object}	response.Response{}	"Invalid inputs"
//	@Failure		500	{object}	response.Response{}	"Failed to update profile"
func (u *UserHandler) UpdateProfile(ctx *gin.Context) {

	userID := utils.GetUserIdFromContext(ctx)

	var body request.EditUser

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	var user domain.User
	copier.Copy(&user, &body)
	user.ID = userID

	err := u.userUseCase.UpdateProfile(ctx, user)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update profile", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully profile updated", nil)
}

// SaveAddress godoc
//
//	@Summary		Add a new address (User)
//	@Security		BearerAuth
//	@Description	API for user to add a new address
//	@Id				SaveAddress
//	@Tags			User Profile
//	@Param			inputs	body	request.Address{}	true	"Address input"
//	@Router			/account/address [post]
//	@Success		200	{object}	response.Response{}	"Successfully address added"
//	@Failure		400	{object}	response.Response{}	"invalid input"
//	@Failure		500	{object}	response.Response{}	"Failed to save address"
func (u *UserHandler) SaveAddress(ctx *gin.Context) {

	var body request.Address
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	userID := utils.GetUserIdFromContext(ctx)

	var address domain.Address
	address.LandMark = body.LandMark
	address.City = body.City
	address.Pincode = body.Pincode
	address.CountryID = body.CountryID
	address.Latitude = body.Latitude
	address.Longitude = body.Longitude
	address.PhoneNumber = body.PhoneNumber
	address.AddressType = body.AddressType
	address.AddressLine1 = body.AddressLine1
	address.AddressLine2 = body.AddressLine2
	address.IsDefault = body.IsDefault

	// check is default is null
	if body.IsDefault == nil {
		body.IsDefault = new(bool)
	}

	err := u.userUseCase.SaveAddress(ctx, userID, address, *body.IsDefault)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to save address", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully address saved")
}

// GetAllAddresses godoc
//
//	@Summary		Get all addresses (User)
//	@Security		BearerAuth
//	@Description	API for user to get all user addresses
//	@Id				GetAllAddresses
//	@Tags			User Profile
//	@Router			/account/address [get]
//	@Success		200	{object}	response.Response{}	"successfully retrieved all user addresses"
//	@Failure		500	{object}	response.Response{}	"failed to show user addresses"
func (u *UserHandler) GetAllAddresses(ctx *gin.Context) {

	userID := utils.GetUserIdFromContext(ctx)

	addresses, err := u.userUseCase.FindAddresses(ctx, userID)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get user addresses", err, nil)
		return
	}

	if addresses == nil {
		response.SuccessResponse(ctx, http.StatusOK, "No addresses found")
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved all user addresses", addresses)
}

// UpdateAddress godoc
//
//	@Summary		Update address (User)
//	@Security		BearerAuth
//	@Description	API for user to update user address
//	@Id				UpdateAddress
//	@Tags			User Profile
//	@Param			input	body	request.EditAddress{}	true	"Address input"
//	@Router			/account/address [put]
//	@Success		200	{object}	response.Response{}	"successfully addresses updated"
//	@Failure		400	{object}	response.Response{}	"can't update the address"
func (u *UserHandler) UpdateAddress(ctx *gin.Context) {

	userID := utils.GetUserIdFromContext(ctx)
	var body request.EditAddress

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	// address is_default reference pointer need to change in future
	if body.IsDefault == nil {
		body.IsDefault = new(bool)
	}

	fmt.Printf("Update address request body: %+v\n", body)

	err := u.userUseCase.UpdateAddress(ctx, body, userID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update user address", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "successfully addresses updated", body)

}

// SaveToWishList godoc
//
//	@Summary		Add to whish list (User)
//	@Security		BearerAuth
//	@Descriptions	API for user to add product item to wish list
//	@Id				SaveToWishList
//	@Tags			User Profile
//	@Param			product_item_id	path	int	true	"Product Item ID"
//	@Router			/account/wishlist/{product_item_id} [post]
//	@Success		200	{object}	response.Response{}	"Successfully product items added to whish list"
//	@Failure		400	{object}	response.Response{}	"invalid input"
//	@Failure		409	{object}	response.Response{}	"Product item already exist on wish list"
//	@Failure		500	{object}	response.Response{}	"Failed to add product item to wishlist"
func (u *UserHandler) SaveToWishList(ctx *gin.Context) {

	productItemID, err := request.GetParamAsUint(ctx, "product_item_id")

	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	userID := utils.GetUserIdFromContext(ctx)

	var wishList = domain.WishList{
		ProductItemID: productItemID,
		UserID:        userID,
	}

	err = u.userUseCase.SaveToWishList(ctx, wishList)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrExistWishListProductItem) {
			statusCode = http.StatusConflict
		}
		response.ErrorResponse(ctx, statusCode, "Failed to add product item to wishlist", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Successfully product items added to whish list", nil)
}

// RemoveFromWishList godoc
//
//	@Summary		Remove from whish list (User)
//	@Security		BearerAuth
//	@Descriptions	API for user to remove a product item from whish list
//	@Id				RemoveFromWishList
//	@Tags			User Profile
//	@Param			product_item_id	path	int	true	"Product Item ID"
//	@Router			/account/wishlist/{product_item_id} [delete]
//	@Success		200	{object}	response.Response{}	"successfully removed product item from wishlist"
//	@Failure		400	{object}	response.Response{}	"invalid input"
func (u *UserHandler) RemoveFromWishList(ctx *gin.Context) {

	productItemID, err := request.GetParamAsUint(ctx, "product_item_id")

	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	userID := utils.GetUserIdFromContext(ctx)

	// remove form wishlist
	if err := u.userUseCase.RemoveFromWishList(ctx, userID, productItemID); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to remove product item from wishlist", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully removed product item from wishlist", nil)
}

// GetWishList godoc
//
//	@Summary		Get whish list product items (User)
//	@Security		BearerAuth
//	@Descriptions	API for user to get product items in the wish list
//	@Id				GetWishList
//	@Tags			User Profile
//	@Router			/account/wishlist [get]
//	@Success		200	"Successfully retrieved all product items in th wish list"
//	@Failure		500	"Failed to retrieve product items from the wish list"
func (u *UserHandler) GetWishList(ctx *gin.Context) {

	userID := utils.GetUserIdFromContext(ctx)

	wishListItems, err := u.userUseCase.FindAllWishListItems(ctx, userID)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve product items from the wish list", err, nil)
		return
	}

	if len(wishListItems) == 0 {
		response.SuccessResponse(ctx, http.StatusOK, "No wishlist items found", nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved all product items in th wish list", wishListItems)
}

// UploadProfileImage godoc
//
//	@Summary		Upload profile image to S3
//	@Description	API endpoint to upload a user profile image file and save it to AWS S3 storage.
//	@Tags			User Profile, File Upload
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			image	formData	file	true	"Profile image file to upload"
//	@Success		200	{object}	map[string]string	"Returns URL of the uploaded image"
//	@Failure		400	{object}	map[string]string	"Image file is required or invalid request"
//	@Failure		500	{object}	map[string]string	"Failed to upload image to storage"
//	@Router			/api/upload-profile-image [post]
// func UploadProfileImage(c *gin.Context) {
// 	file, err := c.FormFile("image")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
// 		return
// 	}

// 	openedFile, err := file.Open()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
// 		return
// 	}
// 	defer openedFile.Close()

// 	ctx := context.Background()
// 	region := "ap-south-1" // Replace with your AWS region
// 	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load AWS config"})
// 		return
// 	}

// 	s3Client := s3.NewFromConfig(cfg)
// 	uploader := manager.NewUploader(s3Client)

// 	// Unique filename with timestamp and original extension
// 	fileName := fmt.Sprintf("user-profile/%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))

// 	bucketName := "s3-mandi-bucket" // Replace with your actual bucket name
// 	contentType := file.Header.Get("Content-Type")
// 	result, err := uploader.Upload(ctx, &s3.PutObjectInput{
// 		Bucket:      &bucketName,
// 		Key:         &fileName,
// 		Body:        openedFile,
// 		ContentType: &contentType,
// 		ACL:         "public-read", // Adjust according to your bucket policy
// 	})

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload to S3 failed"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message":   "Upload successful",
// 		"image_url": result.Location,
// 	})
// }

//// UploadProfileImage godoc

// @Summary		Upload profile image (User)
// @Security		BearerAuth
// @Description	API for user to upload profile image
// @Id				UploadProfileImage
// @Tags			User Profile
// @Accept			multipart/form-data
// @Param			image	formData	file	true	"Profile image file to upload"
// @Router			/account/profile-image [post]
// @Success		200	{object}	response.Response{}	"Successfully uploaded profile image"
// @Failure		400	{object}	response.Response{}	"Image file is required or invalid request"
// @Failure		500	{object}	response.Response{}	"Failed to upload image"
func (h *UserHandler) UploadProfileImage(c *gin.Context) {
	var req request.UploadImageRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}

	file, err := req.Image.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
		return
	}
	defer file.Close()

	// Generate unique filename
	fileName := fmt.Sprintf("user-profile/%d%s", time.Now().UnixNano(), filepath.Ext(req.Image.Filename))

	ctx := context.Background()
	imageURL, err := h.userUseCase.UploadProfileImage(ctx, req.UserID, req.Image, req.Image.Size, fileName, req.Image.Header.Get("Content-Type"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Upload successful",
		"image_url": imageURL,
	})
}

// GetSellerByRadius godoc
//
//	@Summary		Get sellers by radius (User)
//	@Security		BearerAuth
//	@Description	API for user to get sellers within a specified radius from given latitude and longitude
//	@Id				GetSellerByRadius
//	@Tags			User Profile
//	@Param			latitude	query	float64	true	"Latitude"
//	@Param			longitude	query	float64	true	"Longitude"
//	@Param			radius_km	query	float64	true	"Radius in kilometers"
//	@Router			/shop/search/radius [get]
//	@Success		200	{object}	response.Response{}	"Successfully found sellers in the given radius"
//	@Success		204	{object}	response.Response{}	"No sellers found in the given radius"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Failed to get sellers by radius"
func (c *UserHandler) GetSellerByRadius(ctx *gin.Context) {
	// get latitude, longitude and radius from query params
	latitudeStr := ctx.Query("lat")
	longitudeStr := ctx.Query("lng")
	radiusStr := ctx.Query("radius")

	latitude, err1 := strconv.ParseFloat(latitudeStr, 64)
	longitude, err2 := strconv.ParseFloat(longitudeStr, 64)
	radius, err3 := strconv.ParseFloat(radiusStr, 64)

	// join all error and send it if its not nil
	err := errors.Join(err1, err2, err3)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindQueryFailMessage, err, nil)
		return
	}

	pagination := request.GetPagination(ctx)

	reqData := request.SellerRadiusRequest{
		Latitude:   latitude,
		Longitude:  longitude,
		RadiusKm:   radius,
		Pagination: pagination,
	}

	sellers, err := c.userUseCase.GetSellersByRadius(ctx, reqData)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get sellers by radius", err, nil)
		return
	}

	if len(sellers) == 0 {
		response.SuccessResponse(ctx, http.StatusNoContent, "No sellers found in the given radius", nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully found sellers in the given radius", sellers)
}

// GetSellerByPincode godoc
//
//	@Summary		Get sellers by pincode (User)
//	@Security		BearerAuth
//	@Description	API for user to get sellers in a specified pincode
//	@Id				GetSellerByPincode
//	@Tags			User Profile
//	@Param			pincode	query	uint	true	"Pincode"
//	@Router			/shop/search/pincode [get]
//	@Success		200	{object}	response.Response{}	"Successfully found sellers in the given pincode"
//	@Success		204	{object}	response.Response{}	"No sellers found in the given pincode"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Failed to get sellers by pincode"
func (c *UserHandler) GetSellerByPincode(ctx *gin.Context) {
	// get pincode from query params
	pincodeStr := ctx.Query("pincode")

	pincode, err := strconv.ParseUint(pincodeStr, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid pincode", err, nil)
		return
	}

	pagination := request.GetPagination(ctx)

	reqData := request.SellerPincodeRequest{
		Pincode:    uint(pincode),
		Pagination: pagination,
	}

	sellers, err := c.userUseCase.GetSellersByPincode(ctx, reqData)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get sellers by pincode", err, nil)
		return
	}

	if len(sellers) == 0 {
		response.SuccessResponse(ctx, http.StatusNoContent, "No sellers found in the given pincode", nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully found sellers in the given pincode", sellers)
}

// SearchShopList godoc
//
//	@Summary		Search shops (User)
//	@Security		BearerAuth
//	@Description	API for user to search shops by query, location, and radius (all optional)
//	@Id				SearchShopList
//	@Tags			User Profile
//	@Param			q		query	string	false	"Search query for shop name"
//	@Param			lat		query	float64	false	"Latitude"
//	@Param			lng		query	float64	false	"Longitude"
//	@Param			radius	query	float64	false	"Radius in kilometers"
//	@Router			/shop/search [get]
//	@Success		200	{object}	response.Response{}	"Successfully found shops"
//	@Success		204	{object}	response.Response{}	"No shops found"
//	@Failure		500	{object}	response.Response{}	"Failed to search shops"
func (c *UserHandler) SearchShopList(ctx *gin.Context) {
	// get optional query parameters
	query := ctx.Query("q")
	latStr := ctx.Query("lat")
	lngStr := ctx.Query("long")
	radiusStr := ctx.Query("radius")
	pincodeStr := ctx.Query("pincode")

	var latitude, longitude, radius float64
	var pincode *uint
	var err error

	// parse latitude if provided
	if latStr != "" {
		latitude, err = strconv.ParseFloat(latStr, 64)
		if err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid latitude", err, nil)
			return
		}
	}

	// parse longitude if provided
	if lngStr != "" {
		longitude, err = strconv.ParseFloat(lngStr, 64)
		if err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid longitude", err, nil)
			return
		}
	}

	// parse radius if provided
	if radiusStr != "" {
		radius, err = strconv.ParseFloat(radiusStr, 64)
		if err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid radius", err, nil)
			return
		}
	}

	// parse pincode if provided
	if pincodeStr != "" {
		if p, err := strconv.ParseUint(pincodeStr, 10, 32); err == nil {
			pincodeVal := uint(p)
			pincode = &pincodeVal
		}
	}

	pagination := request.GetPagination(ctx)

	reqData := request.SearchShopListRequest{
		Query:      query,
		Latitude:   latitude,
		Longitude:  longitude,
		Radius:     radius,
		Pincode:    pincode,
		Pagination: pagination,
	}

	shops, err := c.userUseCase.SearchShopList(ctx, reqData)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to search shops", err, nil)
		return
	}

	if len(shops) == 0 {
		response.SuccessResponse(ctx, http.StatusNoContent, "No shops found", nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully found shops", shops)
}

func (c *UserHandler) GetProductItemsByDepartment(ctx *gin.Context) {
	// Route may provide department_id or document_id depending on routes setup.
	idStr := ctx.Param("department_id")
	if idStr == "" {
		idStr = ctx.Param("document_id")
	}

	if idStr == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid document ID", fmt.Errorf("missing id param"), nil)
		return
	}

	documentID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid document ID", err, nil)
		return
	}

	products, err := c.userUseCase.GetProductItemsByDepartment(ctx, uint(documentID))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get product items by document", err, nil)
		return
	}

	// Ensure we return an empty array instead of null when no products found
	if products == nil {
		products = []response.ProductItems{}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved product items by document", products)
}

func (c *UserHandler) GetProductItemsByCategory(ctx *gin.Context) {
	idStr := ctx.Param("category_id")
	if idStr == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid category ID", fmt.Errorf("missing id param"), nil)
		return
	}

	categoryID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid category ID", err, nil)
		return
	}

	products, err := c.userUseCase.GetProductItemsByCategory(ctx, uint(categoryID))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get product items by category", err, nil)
		return
	}

	if products == nil {
		products = []response.ProductItems{}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved product items by category", products)
}

func (c *UserHandler) GetProductItemsBySubCategory(ctx *gin.Context) {
	idStr := ctx.Param("sub_category_id")
	if idStr == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid sub-category ID", fmt.Errorf("missing id param"), nil)
		return
	}

	subCategoryID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid sub-category ID", err, nil)
		return
	}

	products, err := c.userUseCase.GetProductItemsBySubCategory(ctx, uint(subCategoryID))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get product items by sub-category", err, nil)
		return
	}

	if products == nil {
		products = []response.ProductItems{}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved product items by sub-category", products)
}

func (c *UserHandler) GetProductItemsByShop(ctx *gin.Context) {
	idStr := ctx.Param("admin_id")
	if idStr == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid admin ID", fmt.Errorf("missing id param"), nil)
		return
	}

	adminID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid admin ID", err, nil)
		return
	}

	products, err := c.userUseCase.GetProductItemsByShop(ctx, uint(adminID))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get product items by shop", err, nil)
		return
	}

	if products == nil {
		products = []response.ProductItems{}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved product items by shop", products)
}
