package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	service "github.com/rohit221990/mandi-backend/pkg/service/ai"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

// semaphore limits concurrent image processing operations
var semaphore = make(chan struct{}, runtime.NumCPU())

type ModerationResponse struct {
	Status string `json:"status"`
	Nudity struct {
		SexualActivity   float64 `json:"sexual_activity"`
		SexualDisplay    float64 `json:"sexual_display"`
		Erotica          float64 `json:"erotica"`
		VerySuggestive   float64 `json:"very_suggestive"`
		Suggestive       float64 `json:"suggestive"`
		MildlySuggestive float64 `json:"mildly_suggestive"`
		None             float64 `json:"none"`
	} `json:"nudity"`
}

type EmbeddingResponse struct {
	Vector []float32 `json:"vector"`
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type ProductHandler struct {
	productUseCase usecaseInterface.ProductUseCase
	tokenService   token.TokenService
	aiClient       service.Client
	cloudService   cloud.CloudService
}

func NewProductHandler(productUsecase usecaseInterface.ProductUseCase, tokenService token.TokenService, aiClient *service.Client, cloudService cloud.CloudService) interfaces.ProductHandler {
	return &ProductHandler{
		productUseCase: productUsecase,
		tokenService:   tokenService,
		aiClient:       *aiClient,
		cloudService:   cloudService,
	}
}

// callCompareImages validates that two images match by category using the AI service.
// It delegates to the injected aiClient (which honors AI_SERVICE_URL) and returns nil if
// the images match, or an AppError if they don't match or the service call fails.
func (p *ProductHandler) callCompareImages(imagePath1, imagePath2 string) *domain.AppError {
	result, err := p.aiClient.CompareImages(imagePath1, imagePath2)
	if err != nil {
		return domain.ExternalServiceError("compare-images", "failed to compare images via AI service", err)
	}

	// If categories don't match, return a user-friendly mismatch error.
	if !result.SameCategory {
		return domain.ImageMismatchError(result.Reason)
	}

	return nil
}

// GetAllCategories godoc
//
//	@Summary		Get all categories (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to get all categories and their subcategories
//	@Tags			Admin Category
//	@ID				GetAllCategories
//	@Accept			json
//	@Produce		json
//	@Param			page_number	query	int	false	"Page number"
//	@Param			count		query	int	false	"Count"
//	@Router			/admin/categories [get]
//	@Success		200	{object}	response.Response{}	"Successfully retrieved all categories"
//	@Failure		500	{object}	response.Response{}	"Failed to retrieve categories"
func (p *ProductHandler) GetAllCategories(ctx *gin.Context) {

	pagination := request.GetPagination(ctx)

	categories, err := p.productUseCase.FindAllCategories(ctx, pagination)

	if err != nil {
		appErr := utils.ConvertDBError(err, "categories")
		if appErr == nil {
			appErr = domain.InternalError("failed to retrieve categories", err)
		}
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr.Err, nil)
		return
	}

	if len(categories) == 0 {
		response.SuccessResponse(ctx, http.StatusOK, "No categories found", []interface{}{})
		return
	}

	response.ResolveCategoriesImages(p.cloudService, categories)
	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved all categories", categories)
}

// SaveCategory godoc
//
//	@Summary		Add a new category (Admin)
//	@Security		BearerAuth
//	@Description	API for Admin to save new category
//	@Tags			Admin Category
//	@ID				SaveCategory
//	@Accept			json
//	@Produce		json
//	@Param			input	body	request.Category{}	true	"Category details"
//	@Router			/admin/categories [post]
//	@Success		201	{object}	response.Response{}	"Successfully added category"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		409	{object}	response.Response{}	"Category already exist"
//	@Failure		409	{object}	response.Response{}	"Failed to save category"
func (p *ProductHandler) SaveCategory(ctx *gin.Context) {
	var department_id string = ctx.Param("department_id")
	print("department id in handler", department_id)
	var body request.Category

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err := p.productUseCase.SaveCategory(ctx, body, department_id)

	if err != nil {
		errResponse(ctx, "Failed to add category", err)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully added category")
}

// SaveSubCategory godoc
//
//	@Summary		Add a new subcategory (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to add a new sub category for a existing category
//	@Tags			Admin Category
//	@ID				SaveSubCategory
//	@Accept			json
//	@Produce		json
//	@Param			input	body	request.SubCategory{}	true	"Subcategory details"
//	@Router			/admin/categories/sub-categories [post]
//	@Success		201	{object}	response.Response{}	"Successfully added subcategory"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		409	{object}	response.Response{}	"Sub category already exist"
//	@Failure		500	{object}	response.Response{}	"Failed to add subcategory"
func (p *ProductHandler) SaveSubCategory(ctx *gin.Context) {
	var department_id string = ctx.Param("department_id")
	var category_id string = ctx.Param("category_id")
	var body request.SubCategory
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err := p.productUseCase.SaveSubCategory(ctx, body, department_id, category_id)

	if err != nil {
		errResponse(ctx, "Failed to add sub category", err)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully sub category added")
}

// SaveVariation godoc
//
//	@Summary		Add new variations (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to add new variations for a category
//	@Tags			Admin Category
//	@ID				SaveVariation
//	@Accept			json
//	@Produce		json
//	@Param			category_id	path	int					true	"Category ID"
//	@Param			input		body	request.Variation{}	true	"Variation details"
//	@Router			/admin/categories/{category_id}/variations [post]
//	@Success		201	{object}	response.Response{}	"Successfully added variations"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Failed to add variation"
func (p *ProductHandler) SaveVariation(ctx *gin.Context) {

	categoryID := ctx.Param("category_id")

	var body request.Variation

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err := p.productUseCase.SaveVariation(ctx, categoryID, body.Names)

	if err != nil {
		errResponse(ctx, "Failed to add variation", err)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully added variations")
}

// SaveVariationOption godoc
//
//	@Summary		Add new variation options (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to add variation options for a variation
//	@Tags			Admin Category
//	@ID				SaveVariationOption
//	@Accept			json
//	@Produce		json
//	@Param			category_id		path	int							true	"Category ID"
//	@Param			variation_id	path	int							true	"Variation ID"
//	@Param			input			body	request.VariationOption{}	true	"Variation option details"
//	@Router			/admin/categories/{category_id}/variations/{variation_id}/options [post]
//	@Success		201	{object}	response.Response{}	"Successfully added variation options"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Failed to add variation options"
func (p *ProductHandler) SaveVariationOption(ctx *gin.Context) {

	variationID := ctx.Param("variation_id")

	var body request.VariationOption

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err := p.productUseCase.SaveVariationOption(ctx, variationID, body.Values)
	if err != nil {
		errResponse(ctx, "Failed to add variation options", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Successfully added variation options")
}

// GetAllVariations godoc
//
//	@Summary		Get all variations (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to get all variation and its values of a category
//	@Tags			Admin Category
//	@ID				GetAllVariations
//	@Accept			json
//	@Produce		json
//	@Param			category_id	path	int	true	"Category ID"
//	@Router			/admin/categories/{category_id}/variations [get]
//	@Success		200	{object}	response.Response{}	"Successfully retrieved all variations and its values"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Failed to Get variations and its values"
func (c *ProductHandler) GetAllVariations(ctx *gin.Context) {

	categoryID := ctx.Param("category_id")

	variations, err := c.productUseCase.FindAllVariationsAndItsValues(ctx, categoryID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to Get variations and its values", err, nil)
		return
	}

	if len(variations) == 0 {
		response.SuccessResponse(ctx, http.StatusOK, "No variations found", []interface{}{})
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved all variations and its values", variations)
}

// SaveProduct godoc
//
//	@Summary		Add a new product with image (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to add a new product with image file upload (multipart/form-data)
//	@ID				SaveProduct
//	@Tags			Admin Products
//	@Accept			mpfd
//	@Produce		json
//	@Param			product_name	formData	string				true	"Product Name"
//	@Param			description		formData	string				true	"Product Description"
//	@Param			department		formData	string				true	"Department Name"
//	@Param			department_id	formData	int					true	"Department ID"
//	@Param			category_id		formData	int					true	"Category ID"
//	@Param			brand_id		formData	int					true	"Brand ID"
//	@Param			price			formData	int					true	"Product Price"
//	@Param			condition		formData	string				false	"Product Condition"
//	@Param			specification	formData	string				false	"Product Specification"
//	@Param			highlights		formData	string				false	"Product Highlights"
//	@Param			image			formData	file				true	"Product Image"
//	@Success		201				{object}	response.Response{}	"successfully product added"
//	@Router			/admin/products [post]
//	@Failure		400	{object}	response.Response{}	"invalid input"
//	@Failure		409	{object}	response.Response{}	"Product name already exist"
func (p *ProductHandler) SaveProduct(ctx *gin.Context) {

	tokenString := ctx.GetHeader("Authorization")

	adminID := p.tokenService.DecodeTokenData(tokenString)
	// Check if this is a JSON request (without file upload)
	contentType := ctx.GetHeader("Content-Type")
	if contentType == "application/json" || contentType == "" {
		p.SaveProductJSON(ctx, adminID)
		return
	}

	// Handle multipart/form-data request
	name, err1 := request.GetFormValuesAsString(ctx, "category")
	departmentID, errDeptID := request.GetFormValuesAsString(ctx, "department_id")
	description, err2 := request.GetFormValuesAsString(ctx, "description")
	categoryID, err3 := request.GetFormValuesAsString(ctx, "category_id")
	fileHeader, err6 := ctx.FormFile("image")

	// Only check required fields
	err := errors.Join(err1, err2, err3, err6, errDeptID)
	if err != nil {
		appErr := domain.ValidationError("product fields", "missing or invalid required fields")
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr.Err, nil)
		return
	}

	product := request.Product{
		Name:            name,
		DepartmentID:    departmentID,
		Description:     description,
		CategoryID:      categoryID,
		ImageFileHeader: fileHeader,
	}

	productID, err := p.productUseCase.SaveProduct(ctx, product, adminID)

	if err != nil {
		appErr := domain.InternalError("failed to save product", err)
		if errors.Is(err, usecase.ErrProductAlreadyExist) {
			appErr = domain.AlreadyExistsError("product")
		}
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr.Err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully added product", map[string]string{"product_id": productID})
}

// SaveProductJSON handles JSON requests without image uploa
func (p *ProductHandler) SaveProductJSON(ctx *gin.Context, adminID string) {
	// Debug: Log raw request body
	rawData, _ := ctx.GetRawData()

	// Check if the body starts with a quote (indicating it's a string-wrapped JSON)
	if len(rawData) > 0 && rawData[0] == '"' {
		// Unwrap the double-encoded JSON string
		var unwrappedJSON string
		if err := json.Unmarshal(rawData, &unwrappedJSON); err != nil {
			appErr := domain.ValidationError("request body", "invalid JSON format (double-encoded)")
			response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr.Err, nil)
			return
		}
		rawData = []byte(unwrappedJSON)
	}

	// Additional validation: Check if it's valid JSON before binding
	var testJSON interface{}
	if err := json.Unmarshal(rawData, &testJSON); err != nil {
		appErr := domain.ValidationError("request body", "invalid JSON syntax")
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr.Err, nil)
		return
	}

	// Re-bind the body since we read it
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(rawData))

	var body struct {
		Name           string      `json:"category" binding:"min=3,max=50"`
		Description    string      `json:"description"`
		CategoryID     string      `json:"category_id"`
		DepartmentID   string      `json:"department_id"`
		Condition      string      `json:"condition" binding:"omitempty"`
		Specifications interface{} `json:"specifications" binding:"omitempty"` // Can be string or object
		Highlights     interface{} `json:"highlights" binding:"omitempty"`     // Can be string or array
		ImageURL       string      `json:"image_url" binding:"omitempty"`      // For existing image URL
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		appErr := domain.ValidationError("request body", "failed to parse JSON request")
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr.Err, nil)
		return
	}

	// For JSON requests without file upload, you need to provide image_url or handle it differently
	if body.ImageURL == "" {
		appErr := domain.ValidationError("image_url", "image_url is required for JSON requests")
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr.Err, nil)
		return
	}

	// Convert highlights to string

	product := request.Product{
		Name:         body.Name,
		DepartmentID: body.DepartmentID,
		Description:  body.Description,
		CategoryID:   body.CategoryID,
		// Note: ImageFileHeader is nil, you'll need to handle this in the usecase
	}

	productID, err := p.productUseCase.SaveProduct(ctx, product, adminID)

	if err != nil {
		appErr := domain.InternalError("failed to save product", err)
		if errors.Is(err, usecase.ErrProductAlreadyExist) {
			appErr = domain.AlreadyExistsError("product")
		}
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr.Err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully added product", map[string]string{"product_id": productID})
}

// GetAllProductsAdmin godoc
//
//	@Summary		Get all products (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to get all products
//	@ID				GetAllProductsAdmin
//	@Tags			Admin Products
//	@Param			page_number	query	int		false	"Page Number"
//	@Param			count		query	int		false	"Count"
//	@Param			search		query	string	false	"Search term to filter products by name"
//	@Router			/admin/products [get]
//	@Success		200	{object}	response.Response{}	"Successfully found all products"
//	@Failure		500	{object}	response.Response{}	"Failed to Get all products"
func (p *ProductHandler) GetAllProductsAdmin(ctx *gin.Context) {
	p.getAllProducts()(ctx)
}

// GetAllProductsUser godoc
//
//	@Summary		Get all products (User)
//	@Security		BearerAuth
//	@Description	API for user to get all products
//	@ID				GetAllProductsUser
//	@Tags			User Products
//	@Param			page_number	query	int		false	"Page Number"
//	@Param			count		query	int		false	"Count"
//	@Param			search		query	string	false	"Search term to filter products by name"
//	@Router			/products [get]
//	@Success		200	{object}	response.Response{}	"Successfully found all products"
//	@Failure		500	{object}	response.Response{}	"Failed to get all products"
func (p *ProductHandler) GetAllProductsUser(ctx *gin.Context) {
	p.getAllProducts()(ctx)
}

// Get products is common for user and admin so this function is to get the common Get all products func for them
func (p *ProductHandler) getAllProducts() func(ctx *gin.Context) {

	return func(ctx *gin.Context) {

		pagination := request.GetPagination(ctx)
		search := ctx.Query("search") // Get optional search parameter

		products, err := p.productUseCase.FindAllProducts(ctx, pagination, search)

		if err != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to Get all products", err, nil)
			return
		}

		if len(products) == 0 {
			response.SuccessResponse(ctx, http.StatusOK, "No products found", []interface{}{})
			return
		}

		response.ResolveProductsImages(p.cloudService, products)
		response.SuccessResponse(ctx, http.StatusOK, "Successfully found all products", products)
	}

}

// GetProductByID godoc
//
//	@Summary		Get product by ID (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to get a single product by ID with all details
//	@ID				GetProductByID
//	@Tags			Admin Products
//	@Param			product_id	path	int	true	"Product ID"
//	@Router			/admin/products/{product_id} [get]
//	@Success		200	{object}	response.Response{}	"Successfully found product"
//	@Failure		400	{object}	response.Response{}	"Invalid product ID"
//	@Failure		404	{object}	response.Response{}	"Product not found"
//	@Failure		500	{object}	response.Response{}	"Failed to get product"
func (p *ProductHandler) GetProductByID(ctx *gin.Context) {
	productID := ctx.Param("product_id")

	product, err := p.productUseCase.FindProductByID(ctx, productID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get product", err, nil)
		return
	}

	product.Image = cloud.ResolveURL(p.cloudService, product.Image)
	response.SuccessResponse(ctx, http.StatusOK, "Successfully found product", product)
}

// UpdateProduct godoc
//
//	@Summary		Update a product (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to update a product
//	@ID				UpdateProduct
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			input	body	request.UpdateProduct{}	true	"Product update input"
//	@Router			/admin/products [put]
//	@Success		200	{object}	response.Response{}	"successfully product updated"
//	@Failure		400	{object}	response.Response{}	"invalid input"
//	@Failure		409	{object}	response.Response{}	"Failed to update product"
//	@Failure		500	{object}	response.Response{}	"Product name already exist for another product"
func (c *ProductHandler) UpdateProduct(ctx *gin.Context) {

	var body request.UpdateProduct

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	var product domain.Product
	copier.Copy(&product, &body)

	err := c.productUseCase.UpdateProduct(ctx, product)
	if err != nil {
		errResponse(ctx, "Failed to update product", err)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully product updated", nil)
}

// aiValidationRejectConfidence is the confidence above which the AI service's
// "wrong category" verdict is trusted enough to reject a seller's upload.
const aiValidationRejectConfidence = 0.1

// validateProductImage runs the optional AI category check for one uploaded
// image. It returns a non-empty rejection message ONLY when the AI service
// actually judged the image to be the wrong category.
//
// Infrastructure failures — service down, timeout, non-200, malformed body —
// are logged and treated as "no opinion" rather than blocking the seller.
// This is a product-quality gate, not a security control, and a seller must
// never be locked out of adding products because an optional ML dependency is
// unavailable.
//
// Previously every one of those failures returned 400 Bad Request, so a single
// AI-service outage halted ALL product uploads across the platform and blamed
// the client for what was a server-side fault. The AI client has a 60s timeout,
// so an unresponsive service also held the request open that long.
func (p *ProductHandler) validateProductImage(localPath, categoryName string) string {
	if strings.TrimSpace(categoryName) == "" {
		return ""
	}

	result, err := p.aiClient.ValidateProduct(localPath, categoryName)
	if err != nil {
		log.Printf("WARN product image validation unavailable, allowing upload: category=%q path=%s err=%v",
			categoryName, localPath, err)
		return ""
	}
	// Defensive: a nil result with a nil error would otherwise panic below.
	if result == nil {
		log.Printf("WARN product image validation returned no result, allowing upload: category=%q path=%s",
			categoryName, localPath)
		return ""
	}

	if !result.Valid && result.Confidence > aiValidationRejectConfidence {
		log.Printf("INFO product image rejected by validation: category=%q confidence=%.2f reason=%q path=%s",
			categoryName, result.Confidence, result.Reason, localPath)
		return fmt.Sprintf("Product image does not match '%s' category. Reason: %s", categoryName, result.Reason)
	}

	return ""
}

// SaveProductItem godoc
//
//	@Summary		Add a product item (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to add a product item for a specific product(should select at least one variation option from each variations)
//	@ID				SaveProductItem
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			product_id				path		int		true	"Product ID"
//	@Param			price					formData	int		true	"Price"
//	@Param			qty_in_stock			formData	int		true	"Quantity In Stock"
//	@Param			variation_option_ids	formData	[]int	true	"Variation Option IDs"
//	@Param			images					formData	file	true	"Images"
//	@Router			/admin/products/{product_id}/items [post]
//	@Success		200	{object}	response.Response{}	"Successfully product item added"
//	@Failure		400	{object}	response.Response{}	"invalid input"
//	@Failure		409	{object}	response.Response{}	"Product have already this configured product items exist"
func (p *ProductHandler) SaveProductItem(ctx *gin.Context) {

	tokenString := ctx.GetHeader("Authorization")

	adminID := p.tokenService.DecodeTokenData(tokenString)

	shopID := ctx.PostForm("shop_id")
	if shopID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "shop_id is required", nil, nil)
		return
	}

	subCategoryID := ctx.PostForm("sub_category_id")
	categoryID := ctx.PostForm("category_id")
	departmentID := ctx.PostForm("department_id")

	subCategoryName := ctx.PostForm("sub_category_name")
	categoryName := ctx.PostForm("category_name")
	dynamicFieldsStr := ctx.PostForm("dynamic_fields")
	// Try to resolve subcategory image URL (if subCategoryID provided)
	// var subCatImageURL string
	// if subCategoryID != 0 && categoryID != 0 {
	// 	subs, err := p.productUseCase.GetAllSubCategoriesByCategoryID(ctx, categoryID)
	// 	if err == nil {
	// 		for _, sc := range subs {
	// 			if sc.ID == subCategoryID {
	// 				subCatImageURL = sc.ImageUrl
	// 				break
	// 			}
	// 		}
	// 	}
	// }
	files := ctx.Request.MultipartForm.File["images[]"]

	var imagePaths []string
	for _, fileHeader := range files {

		localPath, err := handleUpload(fileHeader)
		if err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to process image", err, nil)
			return
		}

		// Optional AI category check. Rejects only on a genuine mismatch verdict;
		// an unavailable AI service is logged and allowed through.
		if rejection := p.validateProductImage(localPath, categoryName); rejection != "" {
			response.ErrorResponse(ctx, http.StatusBadRequest, rejection, nil, nil)
			return
		}

		objectKey, err := uploadProcessedToCloud(ctx, p.cloudService, localPath)
		if err != nil {
			log.Printf("SaveProductItem processed image upload failed: localPath=%s originalFilename=%s size=%d category=%s err=%v", localPath, fileHeader.Filename, fileHeader.Size, categoryName, err)
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload processed image", err, nil)
			return
		}
		imagePaths = append(imagePaths, objectKey)

		// If we have a subcategory image, ask the compare-images service to compare
		// the uploaded product image with the subcategory reference image so external
		// services can validate or index similarity.
		// if subCatImageURL != "" {
		// 	// Build publicly accessible URLs that other services can fetch
		// 	prodURL := buildPublicURL(localPath)
		// 	subURL := buildPublicURL(subCatImageURL)
		// 	if err := p.callCompareImages(subURL, prodURL); err != nil {
		// 		// Return error if image doesn't match category
		// 		response.ErrorResponseAppError(ctx, err)
		// 		return
		// 	}
		// }
	}

	var dynamicFields map[string]interface{}
	if err := json.Unmarshal([]byte(dynamicFieldsStr), &dynamicFields); err != nil {
		// handle error
	}
	description := ctx.PostForm("description")
	highlightsStr := ctx.PostForm("highlights")
	var highlights []string
	if highlightsStr != "" {
		if err := json.Unmarshal([]byte(highlightsStr), &highlights); err != nil {
			highlights = []string{highlightsStr}
		}
	}

	productItem := request.ProductItem{
		SubCategoryName:   subCategoryName,
		SubCategoryID:     subCategoryID,
		DynamicFields:     dynamicFields,
		CategoryID:        categoryID,
		DepartmentID:      departmentID,
		ProductItemImages: imagePaths,
		Description:       description,
		Highlights:        highlights,
	}

	err := p.productUseCase.SaveProductItem(ctx, productItem, adminID, shopID)

	if err != nil {
		errResponse(ctx, "Failed to add product item", err)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully product item added", nil)

	// if err := ctx.ShouldBindJSON(&body); err != nil {
	// 	response.ErrorResponse(ctx, http.StatusBadRequest, "invalid request body", err, nil)
	// 	return
	// }

	// Map request to domain model
	// (Removed redeclaration of imageFileHeader)

	// 	productItem = request.ProductItem{
	// 		SubCategoryName:  subCategoryName,
	// 		SubCategoryID:    subCategoryID,
	// 		DynamicFields:    dynamicFields,
	// 		ProductItemImage: imagePaths,
	// 	}

	// 	// Convert request.ProductItem to domain.ProductItem
	// 	var domainProductItem2 domain.ProductItem
	// 	if err := copier.Copy(&domainProductItem2, &productItem); err != nil {
	// 		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to map product item", err, nil)
	// 		return
	// 	}

	// 	err = p.productUseCase.SaveProductItem(ctx, domainProductItem2, productID)

	// 	if err != nil {

	// 		var statusCode int

	// 		switch {
	// 		case errors.Is(err, usecase.ErrProductItemAlreadyExist):
	// 			statusCode = http.StatusConflict
	// 		case errors.Is(err, usecase.ErrNotEnoughVariations):
	// 			statusCode = http.StatusBadRequest
	// 		default:
	// 			statusCode = http.StatusInternalServerError
	// 		}

	// 		response.ErrorResponse(ctx, statusCode, "Failed to add product item", err, nil)
	// 		return
	// 	}

	// 	response.SuccessResponse(ctx, http.StatusCreated, "Successfully product item added", nil)
	// }

}

// GetAllProductItemsAdmin godoc
//
//	@Summary		Get all product items (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to get all product items for a specific product
//	@ID				GetAllProductItemsAdmin
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			product_id	path	int	true	"Product ID"
//	@Router			/admin/products/{product_id}/items [get]
//	@Success		200	{object}	response.Response{}	"Successfully get all product items"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		400	{object}	response.Response{}	"Failed to get all product items"
func (p *ProductHandler) GetAllProductItemsAdmin() func(*gin.Context) {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")

		adminID := p.tokenService.DecodeTokenData(tokenString)
		p.getAllProductItems(adminID, false)(ctx)
	}
}

// GetAllProductItemsUser godoc
//
//	@Summary		Get all product items (User)
//	@Security		BearerAuth
//	@Description	API for user to get all product items for a specific product
//	@ID				GetAllProductItemsUser
//	@Tags			User Products
//	@Accept			json
//	@Produce		json
//	@Param			product_id	path	int	true	"Product ID"
//	@Router			/products/{product_id}/items [get]
//	@Success		200	{object}	response.Response{}	"Successfully get all product items"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		400	{object}	response.Response{}	"Failed to get all product items"
func (p *ProductHandler) GetAllProductItemsUser() func(*gin.Context) {
	return func(ctx *gin.Context) {
		p.getAllProductItems("", true)(ctx)
	}
}

// same functionality of get all product items for admin and user.
// customerView=true applies the shop-status and subscription visibility gates
// (customer-facing); admin/seller callers pass false so a seller still sees
// their own items while awaiting approval or with a lapsed subscription.
func (p *ProductHandler) getAllProductItems(adminID string, customerView bool) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {

		shopID := ctx.Param("shop_id")

		// Parse optional query params
		keyword := ctx.Query("q")
		var catIDPtr, brandIDPtr, locIDPtr *string
		// category_id/brand_id/location_id are typed-prefix string IDs (e.g.
		// "cat_00003"), never numeric — a strconv.ParseUint check here would
		// always fail for every real ID and silently drop the filter.
		if cid := ctx.Query("category_id"); cid != "" {
			s := cid
			catIDPtr = &s
		}
		if bid := ctx.Query("brand_id"); bid != "" {
			s := bid
			brandIDPtr = &s
		}
		if lid := ctx.Query("location_id"); lid != "" {
			s := lid
			locIDPtr = &s
		}

		sortby := ctx.Query("sortby")
		if sortby == "" {
			sortby = ctx.Query("sortBy")
		}

		var pagination *request.Pagination
		limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "0"))
		offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
		if limit > 0 {
			if limit <= 0 {
				limit = 20
			}
			if offset < 0 {
				offset = 0
			}
			limitUint64, err := SafeIntToUint64(limit)
			if err != nil {
				response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid limit", err, nil)
				return
			}
			offsetUint64, err := SafeIntToUint64(offset)
			if err != nil {
				response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid offset", err, nil)
				return
			}
			pagination = &request.Pagination{
				Limit:  limitUint64,
				Offset: offsetUint64,
			}
		}

		// Define offer with a default value
		offer := ctx.Query("offers")

		// Concurrent search strategies: Elasticsearch + Database
		type searchResult struct {
			items   []response.ProductItems
			err     error
			source  string
			latency time.Duration
		}

		resultChan := make(chan searchResult, 2)

		// Goroutine 1: Primary search (could be Elasticsearch if available, otherwise database)
		go func() {
			start := time.Now()
			items, err := p.productUseCase.FindAllProductItems(ctx, adminID, keyword, catIDPtr, brandIDPtr, locIDPtr, offer, sortby, pagination, shopID, customerView)
			resultChan <- searchResult{
				items:   items,
				err:     err,
				source:  "primary",
				latency: time.Since(start),
			}
		}()

		// Goroutine 2: Fallback search (database)
		go func() {
			start := time.Now()
			// Add small delay to prefer primary search
			time.Sleep(50 * time.Millisecond)
			items, err := p.productUseCase.FindAllProductItems(ctx, adminID, keyword, catIDPtr, brandIDPtr, locIDPtr, offer, sortby, pagination, shopID, customerView)
			resultChan <- searchResult{
				items:   items,
				err:     err,
				source:  "fallback",
				latency: time.Since(start),
			}
		}()

		// Wait for first successful result with timeout
		var productItems []response.ProductItems
		var err error

		timeout := time.After(5 * time.Second) // 5 second timeout

	selectLoop:
		for i := 0; i < 2; i++ {
			select {
			case result := <-resultChan:
				if result.err == nil {
					productItems = result.items
					break selectLoop
				} else {
				}
			case <-timeout:
				err = fmt.Errorf("search timeout after 5 seconds")
				break selectLoop
			}
		}

		// If both searches failed or timed out, return error
		if err != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get all product items", err, nil)
			return
		}

		// If no results found
		if len(productItems) == 0 {
			response.SuccessResponse(ctx, http.StatusOK, "No product items found", []interface{}{})
			return
		}

		// Log search query asynchronously (non-blocking)
		go func() {
			// This runs in background without blocking the response

			// Here you could also:
			// - Save to analytics database
			// - Update search popularity metrics
			// - Log to external monitoring service
		}()

		// Background analytics and caching
		go func() {
			if keyword != "" && len(productItems) > 0 {
				// Implement caching logic here
			}
		}()

		go func() {
			// Implement analytics logic here
		}()

		response.ResolveProductItemsImages(p.cloudService, productItems)
		response.SuccessResponse(ctx, http.StatusOK, "Successfully get all product items", productItems)
	}
}

// GetProductItemsByShopID godoc
//
//	@Summary		Get product items by shop ID (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to get product items for a specific shop, including category names
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			shop_id	path	int	true	"Shop ID"
//	@Router			/admin/items/shop/{shop_id} [get]
//	@Success		200	{object}	response.Response{}	"Successfully get product items for shop"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		400	{object}	response.Response{}	"Failed to get product items"
func (p *ProductHandler) GetProductItemsByShopID() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")

		adminID := p.tokenService.DecodeTokenData(tokenString)

		shopID := ctx.Param("shop_id")

		// Parse optional query params
		keyword := ctx.Query("q")
		var catIDPtr, brandIDPtr, locIDPtr *string
		var offer *string
		if offerStr := ctx.Query("offers"); offerStr != "" {
			offer = &offerStr
		}
		// category_id/brand_id/location_id are typed-prefix string IDs (e.g.
		// "cat_00003", domain.NewID(domain.PrefixCategory)), never numeric —
		// the previous strconv.ParseUint validation here always failed for
		// every real ID, silently dropping the filter param no matter what
		// the caller sent. Only an empty string should be treated as "no
		// filter"; any non-empty value is passed through as-is.
		if cid := ctx.Query("category_id"); cid != "" {
			s := cid
			catIDPtr = &s
		}
		if bid := ctx.Query("brand_id"); bid != "" {
			s := bid
			brandIDPtr = &s
		}
		if lid := ctx.Query("location_id"); lid != "" {
			s := lid
			locIDPtr = &s
		}

		sortby := ctx.Query("sortby")
		if sortby == "" {
			sortby = ctx.Query("sortBy")
		}

		var pagination *request.Pagination
		limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "0"))
		offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
		if limit > 0 {
			if limit <= 0 {
				limit = 20
			}
			if offset < 0 {
				offset = 0
			}
			limitUint64, err := SafeIntToUint64(limit)
			if err != nil {
				response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid limit", err, nil)
				return
			}
			offsetUint64, err := SafeIntToUint64(offset)
			if err != nil {
				response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid offset", err, nil)
				return
			}
			pagination = &request.Pagination{
				Limit:  limitUint64,
				Offset: offsetUint64,
			}
		}

		var offerVal string
		if offer != nil {
			offerVal = *offer
		}
		productItems, err := p.productUseCase.FindAllProductItems(ctx, adminID, keyword, catIDPtr, brandIDPtr, locIDPtr, offerVal, sortby, pagination, shopID, false)

		if err != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get product items", err, nil)
			return
		}

		// check the product have productItem exist or not
		if len(productItems) == 0 {
			response.SuccessResponse(ctx, http.StatusOK, "No product items found for this shop", []response.ProductItems{})
			return
		}

		response.ResolveProductItemsImages(p.cloudService, productItems)
		response.SuccessResponse(ctx, http.StatusOK, "Successfully get product items for shop", productItems)
	}
}

// FindLowViewProductItems godoc
// @Summary Find low view product items (less than 5 views in days 10-20 after creation)
// @Description Get product items with low views for a specific shop in the 10-20 day period after creation
// @Tags Product Items
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param q query string false "Search keyword"
// @Param category_id query string false "Category ID"
// @Param brand_id query string false "Brand ID"
// @Param location_id query string false "Location ID"
// @Param sortby query string false "Sort by: created_at, updated_at, name, views"
// @Param limit query integer false "Pagination limit"
// @Param offset query integer false "Pagination offset"
// @Success 200 {object} response.ProductItems "Successfully retrieved low view product items"
// @Failure 400 {object} response.Response{} "Bad request"
// @Failure 500 {object} response.Response{} "Internal server error"
// @Router /api/admin/items/lowViewproductitems [get]
func (p *ProductHandler) FindLowViewProductItems(ctx *gin.Context) {
	tokenString := ctx.GetHeader("Authorization")
	adminID := p.tokenService.DecodeTokenData(tokenString)

	// Parse optional query params
	keyword := ctx.Query("q")
	var catIDPtr, brandIDPtr, locIDPtr *string

	// category_id/brand_id/location_id are typed-prefix string IDs (e.g.
	// "cat_00003"), never numeric — a strconv.ParseUint check here would
	// always fail for every real ID and silently drop the filter.
	if cid := ctx.Query("category_id"); cid != "" {
		s := cid
		catIDPtr = &s
	}
	if bid := ctx.Query("brand_id"); bid != "" {
		s := bid
		brandIDPtr = &s
	}
	if lid := ctx.Query("location_id"); lid != "" {
		s := lid
		locIDPtr = &s
	}

	sortby := ctx.Query("sortby")
	if sortby == "" {
		sortby = ctx.Query("sortBy")
	}

	var pagination *request.Pagination
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "25"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	if limit <= 0 {
		limit = 25
	}
	if offset < 0 {
		offset = 0
	}

	limitUint64, err := SafeIntToUint64(limit)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid limit", err, nil)
		return
	}
	offsetUint64, err := SafeIntToUint64(offset)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid offset", err, nil)
		return
	}
	pagination = &request.Pagination{
		Limit:  limitUint64,
		Offset: offsetUint64,
	}

	// Get shopID from token/admin details
	var shopID *string
	if sid := ctx.Query("shop_id"); sid != "" {
		shopID = &sid
	}

	productItems, err := p.productUseCase.FindLowViewProductItems(ctx, adminID, keyword, catIDPtr, brandIDPtr, locIDPtr, sortby, pagination, shopID)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get low view product items", err, nil)
		return
	}

	if len(productItems) == 0 {
		response.SuccessResponse(ctx, http.StatusOK, "No low view product items found", []response.ProductItems{})
		return
	}

	response.ResolveProductItemsImages(p.cloudService, productItems)
	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved low view product items", productItems)
}

// Helper to convert *uuid.UUID to *string
func uuidToStringPtr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}
func SafeIntToUint64(i int) (uint64, error) {
	if i < 0 {
		return 0, fmt.Errorf("cannot convert negative int (%d) to uint64", i)
	}
	return uint64(i), nil
}

// SearchProducts godoc
//
//	@Summary		Search products
//	@Security		BearerAuth
//	@Description	API for user to search products with filters. Supports: 1) Search by name + lat + lng + radius, 2) Search by name + pincode
//	@ID				SearchProducts
//	@Tags			User Products
//	@Accept			json
//	@Produce		json
//	@Param			q			query	string	false	"Search keyword"
//	@Param			category_id	query	string	false	"Category ID"
//	@Param			department_id	query	string	false	"Department ID"
//	@Param			brand_id	query	string	false	"Brand ID"
//	@Param			location_id	query	string	false	"Location ID"
//	@Param			lat			query	float64	false	"Latitude for geolocation search"
//	@Param			long		query	float64	false	"Longitude for geolocation search"
//	@Param			radius		query	float64	false	"Radius in kilometers for geolocation search"
//	@Param			pincode		query	uint	false	"Pincode for location-based search"
//	@Param			limit		query	int		false	"Limit"
//	@Param			offset		query	int		false	"Offset"
//	@Router			/products/search [get]
//	@Success		200	{object}	response.Response{}	"Successfully searched products"
//	@Failure		500	{object}	response.Response{}	"Failed to search products"
func (h *ProductHandler) SearchProducts(c *gin.Context) {

	// Support both 'q' and 'name' as search keyword parameter
	keyword := c.Query("q")
	if keyword == "" {
		keyword = c.Query("name")
	}

	// category_id/brand_id/location_id/department_id/shop_id are typed-prefix
	// string IDs (e.g. "cat_00003"), never numeric — pass them through as-is.
	var catIDPtr, brandIDPtr, locIDPtr, shopIDPtr, deptIDPtr *string
	if cid := c.Query("category_id"); cid != "" {
		s := cid
		catIDPtr = &s
	}
	if did := c.Query("department_id"); did != "" {
		s := did
		deptIDPtr = &s
	}
	if bid := c.Query("brand_id"); bid != "" {
		s := bid
		brandIDPtr = &s
	}
	if lid := c.Query("location_id"); lid != "" {
		s := lid
		locIDPtr = &s
	}
	if sid := c.Query("shop_id"); sid != "" {
		s := sid
		shopIDPtr = &s
	}

	// Parse geolocation parameters (optional)
	// Support both 'lng' and 'long' for longitude
	var latitude, longitude, radius float64
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	if lngStr == "" {
		lngStr = c.Query("long")
	}
	radiusStr := c.Query("radius")

	if latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			latitude = lat
		}
	}
	if lngStr != "" {
		if lng, err := strconv.ParseFloat(lngStr, 64); err == nil {
			longitude = lng
		}
	}
	if radiusStr != "" {
		if r, err := strconv.ParseFloat(radiusStr, 64); err == nil {
			radius = r
		}
	}

	// Parse pincode parameter (optional)
	var pincode *uint
	pincodeStr := c.Query("pincode")
	if pincodeStr != "" {
		if p, err := strconv.ParseUint(pincodeStr, 10, 32); err == nil {
			pincodeVal := uint(p)
			pincode = &pincodeVal
		}
	}

	// trending=true sorts most-viewed products first (overrides default
	// relevance/geo/recency ordering); trending=false (or omitted) changes nothing.
	trending := c.Query("trending") == "true"

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Convert limit/offset into pageNumber for request.Pagination
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	limitUint64, err := SafeIntToUint64(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	offsetUint64, err := SafeIntToUint64(offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pagination := request.Pagination{
		Limit:  limitUint64,
		Offset: offsetUint64,
	}

	products, err := h.productUseCase.SearchProducts(c, keyword, catIDPtr, deptIDPtr, brandIDPtr, locIDPtr, shopIDPtr, latitude, longitude, radius, pincode, trending, int(pagination.Limit), int(pagination.Offset))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response.ResolveProductItemsImages(h.cloudService, products)
	c.JSON(http.StatusOK, gin.H{"products": products})
}

// GetProductSearchSuggestions godoc
//
//	@Summary		Get product search suggestions
//	@Security		BearerAuth
//	@Description	API for user to get product name suggestions based on a prefix
//	@ID				GetProductSearchSuggestions
//	@Tags			User Products
//	@Accept			json
//	@Produce		json
//	@Param			q	query	string	true	"Search prefix"
//	@Router			/products/search/suggestions [get]
//	@Success		200	{object}	response.Response{}	"Successfully retrieved product search suggestions"
//	@Failure		400	{object}	response.Response{}	"Query parameter q is required"
//	@Failure		500	{object}	response.Response{}	"Failed to retrieve product search suggestions"
func (h *ProductHandler) GetProductSearchSuggestions(c *gin.Context) {
	prefix := c.Query("q")
	if prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query param q is required"})
		return
	}

	suggestions, err := h.productUseCase.GetProductNameSuggestions(c, prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})

}

// GetProductSearchFilters godoc
//
//	@Summary		Get product search filters
//	@Security		BearerAuth
//	@Description	API for user to get available filters for product search
//	@ID				GetProductSearchFilters
//	@Tags			User Products
//	@Accept			json
//	@Produce		json
//	@Router			/products/search/filters [get]
//	@Success		200	{object}	response.Response{}	"Successfully retrieved product search filters"
//	@Failure		500	{object}	response.Response{}	"Failed to retrieve product search filters"
func (h *ProductHandler) GetProductSearchFilters(c *gin.Context) {
	// userID := h.tokenService.DecodeTokenData(c.GetHeader("Authorization"))
	filters, err := h.productUseCase.GetProductFilters(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, filters)

}

// GetProductSearchLocations godoc
//
//	@Summary		Get product search locations
//	@Security		BearerAuth
//	@Description	API for user to get available locations for product search
//	@ID				GetProductSearchLocations
//	@Tags			User Products
//	@Accept			json
//	@Produce		json
//	@Router			/products/search/locations [get]
//	@Success		200	{object}	response.Response{}	"Successfully retrieved product search locations"
//	@Failure		500	{object}	response.Response{}	"Failed to retrieve product search locations"
func (h *ProductHandler) GetProductSearchLocations(c *gin.Context) {
	locations, err := h.productUseCase.GetProductLocations(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"locations": locations})

}

// GetProductsByCategory godoc
//
//	@Summary		Get products by category ID
//	@Description	Retrieve a paginated list of products filtered by the given category ID.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			category_id	path		int	true	"Category ID"
//	@Param			limit		query		int	false	"Limit number of results"	default(20)
//	@Param			offset		query		int	false	"Offset for pagination"		default(0)
//	@Success		200			{object}	map[string]interface{}	"List of products matching category"
//	@Failure		400			{object}	map[string]string		"Invalid category ID"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Router			/products/category/{category_id} [get]
func (h *ProductHandler) GetProductsByCategory(c *gin.Context) {
	cid := c.Param("category_id")
	categoryID, err := strconv.Atoi(cid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	products, err := h.productUseCase.GetProductsByCategory(c, categoryID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response.ResolveProductsImages(h.cloudService, products)
	c.JSON(http.StatusOK, gin.H{"products": products})

}

// GetAllBrands godoc
//
//	@Summary		Get all brands
//	@Description	API endpoint to retrieve the list of all product brands.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{} "List of brands"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Router			/brands [get]
func (h *ProductHandler) GetAllBrands(c *gin.Context) {

	brands, err := h.productUseCase.GetAllBrands(c)
	c.JSON(http.StatusOK, gin.H{"brands": brands})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"brands": brands})

}

// GetProductsByBrand godoc
//
//	@Summary		Get products by brand ID
//	@Description	Retrieve a paginated list of products filtered by the given brand ID.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			brand_id	path		int		true	"Brand ID"
//	@Param			limit		query		int		false	"Limit number of results"	default(20)
//	@Param			offset		query		int		false	"Offset for pagination"		default(0)
//	@Success		200			{object}	map[string]interface{}	"List of products filtered by brand"
//	@Failure		400			{object}	map[string]string		"Invalid brand ID"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Router			/products/brand/{brand_id} [get]
func (h *ProductHandler) GetProductsByBrand(c *gin.Context) {

	bid := c.Param("brand_id")
	brandID, err := strconv.Atoi(bid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid brand_id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	products, err := h.productUseCase.GetProductsByBrand(c, brandID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response.ResolveProductsImages(h.cloudService, products)
	c.JSON(http.StatusOK, gin.H{"products": products})

}

// GetCategoryFilters godoc
//
//	@Summary		Get category filters
//	@Description	API endpoint to retrieve product category filters.
//	@Tags			Filters
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of product categories"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/filters/categories [get]
func (h *ProductHandler) GetCategoryFilters(c *gin.Context) {
	categories, err := h.productUseCase.GetCategoryFilters(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response.ResolveCategoriesImages(h.cloudService, categories)
	c.JSON(http.StatusOK, gin.H{"categories": categories})

}

// GetBrandFilters godoc
//
//	@Summary		Get brand filters
//	@Description	API endpoint to retrieve product brand filters.
//	@Tags			Filters
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of product brands"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/filters/brands [get]
func (h *ProductHandler) GetBrandFilters(c *gin.Context) {
	brands, err := h.productUseCase.GetBrandFilters(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"brands": brands})

}

// GetLocationFilter godoc
//
//	@Summary		Get location filters
//	@Description	API endpoint to retrieve product location filters.
//	@Tags			Filters
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of product locations"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/filters/locations [get]
func (h *ProductHandler) GetLocationFilter(c *gin.Context) {
	locations, err := h.productUseCase.GetLocationFilter(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"locations": locations})

}

// GetProductsByLocation godoc
//
//	@Summary		Get products by location ID
//	@Description	Retrieve a paginated list of products filtered by the given location ID.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			location_id	query		int		true	"Location ID"
//	@Param			limit		query		int		false	"Limit number of results"	default(20)
//	@Param			offset		query		int		false	"Offset for pagination"		default(0)
//	@Success		200			{object}	map[string]interface{}	"List of products filtered by location"
//	@Failure		400			{object}	map[string]string		"Invalid location ID"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Router			/products/location [get]
func (h *ProductHandler) GetProductsByLocation(c *gin.Context) {

	locationIDStr := c.Query("location_id")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location_id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	products, err := h.productUseCase.GetProductsByLocation(c, locationID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response.ResolveProductsImages(h.cloudService, products)
	c.JSON(http.StatusOK, gin.H{"products": products})

}

// GetAllAreas godoc
//
//	@Summary		Get all areas
//	@Description	API endpoint to retrieve all available areas for filtering or display.
//	@Tags			Filters
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of all areas"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/filters/areas [get]
func (h *ProductHandler) GetAllAreas(c *gin.Context) {
	areas, err := h.productUseCase.GetAllAreas(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"areas": areas})
}

// GetAllCities godoc
//
//	@Summary		Get all cities
//	@Description	API endpoint to retrieve all available cities for filtering or display.
//	@Tags			Filters
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of all cities"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/filters/cities [get]
func (h *ProductHandler) GetAllCities(c *gin.Context) {
	cities, err := h.productUseCase.GetAllCities(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cities": cities})

}

// GetAllStates godoc
//
//	@Summary		Get all states
//	@Description	API endpoint to retrieve all available states for filtering or display.
//	@Tags			Filters
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of all states"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/filters/states [get]
func (h *ProductHandler) GetAllStates(c *gin.Context) {
	states, err := h.productUseCase.GetAllStates(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"states": states})

}

// GetAllCountries godoc
//
//	@Summary		Get all countries
//	@Description	API endpoint to retrieve all available countries for filtering or display.
//	@Tags			Filters
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of all countries"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/filters/countries [get]
func (h *ProductHandler) GetAllCountries(c *gin.Context) {
	countries, err := h.productUseCase.GetAllCountries(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"countries": countries})

}

// GetAllPincodes godoc
//
//	@Summary		Get all pincodes
//	@Description	API endpoint to retrieve all available pincodes for filtering or display.
//	@Tags			Filters
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of all pincodes"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/filters/pincodes [get]
func (h *ProductHandler) GetAllPincodes(c *gin.Context) {
	pincodes, err := h.productUseCase.GetAllPincodes(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pincodes": pincodes})

}

// GetCitiesByState godoc
//
//	@Summary		Get cities by state ID
//	@Description	API endpoint to retrieve all cities corresponding to a specific state ID.
//	@Tags			Filters
//	@Accept			json
//	@Produce		json
//	@Param			state_id	path		string	true	"State ID"
//	@Success		200		{object}	map[string]interface{}	"List of cities for the given state"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Router			/filters/states/{state_id}/cities [get]
func (h *ProductHandler) GetCitiesByState(c *gin.Context) {
	state := c.Param("state_id") // Adjust type depending on state_id format

	cities, err := h.productUseCase.GetCitiesByState(c, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cities": cities})

}

func (h *ProductHandler) GetAreasByCity(c *gin.Context) {
	city := c.Param("city_id")

	areas, err := h.productUseCase.GetAreasByCity(c, city)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"areas": areas})

}

// GetAreasByCity godoc
//
//	@Summary		Get areas by city ID
//	@Description	API endpoint to retrieve all areas corresponding to a specific city ID.
//	@Tags			Filters
//	@Accept			json
//	@Produce		json
//	@Param			city_id		path		string	true	"City ID"
//	@Success		200			{object}	map[string]interface{}	"List of areas for the given city"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Router			/filters/cities/{city_id}/areas [get]
func (h *ProductHandler) GetPincodesByArea(c *gin.Context) {
	area := c.Param("area_id")

	pincodes, err := h.productUseCase.GetPincodesByArea(c, area)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pincodes": pincodes})

}

// GetLocationByPincode godoc
//
//	@Summary		Get location by pincode
//	@Description	API endpoint to retrieve a location corresponding to a specific pincode.
//	@Tags			Filters
//	@Accept			json
//	@Produce		json
//	@Param			pincode_id	path		string	true	"Pincode ID"
//	@Success		200			{object}	response.Location	"Location matching the pincode"
//	@Failure		404			{object}	map[string]string	"Location not found"
//	@Failure		500			{object}	map[string]string	"Internal server error"
//	@Router			/filters/pincode/{pincode_id}/location [get]
func (h *ProductHandler) GetLocationByPincode(c *gin.Context) {
	pincode := c.Param("pincode_id")
	loc, err := h.productUseCase.GetLocationByPincode(c, pincode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if loc == (response.Location{}) {
		c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"location": loc})

}

// GetNearbyProductsByPincode godoc
//
//	@Summary		Get nearby products by pincode within a radius
//	@Description	API endpoint to retrieve a paginated list of products available near a specified pincode within a given radius (in kilometers).
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			pincode		query		string	true	"Pincode to search around"
//	@Param			radius_km	query		number	false	"Radius in kilometers for nearby search"	default(10)
//	@Param			limit		query		int		false	"Limit number of results"			default(20)
//	@Param			offset		query		int		false	"Offset for pagination"				default(0)
//	@Success		200			{object}	map[string]interface{}	"List of nearby products"
//	@Failure		400			{object}	map[string]string		"Invalid input parameters"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Router			/products/nearby [get]
func (h *ProductHandler) GetNearbyProductsByPincode(c *gin.Context) {
	pincode := c.Query("pincode")
	latStr := c.Query("lat")
	lngStr := c.Query("lng")

	radiusKmStr := c.DefaultQuery("radius_km", "10")
	radiusKm, err := strconv.ParseFloat(radiusKmStr, 64)
	if err != nil || radiusKm <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid radius_km"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// If latitude and longitude are provided, use them for search
	if latStr != "" && lngStr != "" {
		latitude, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid latitude"})
			return
		}
		longitude, err := strconv.ParseFloat(lngStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid longitude"})
			return
		}
		// Call radius-based search
		products, err := h.productUseCase.GetProductsByRadius(c, latitude, longitude, radiusKm, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response.ResolveProductItemsImages(h.cloudService, products)
		c.JSON(http.StatusOK, gin.H{"products": products})
		return
	} else if pincode != "" {
		products, err := h.productUseCase.GetNearbyProductsByPincode(c, pincode, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response.ResolveProductItemsImages(h.cloudService, products)
		c.JSON(http.StatusOK, gin.H{"products": products})
		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter pincode or latitude/longitude is required"})
		return
	}
}

// GetProductsByRadius godoc
//
//	@Summary		Get products by geographic radius
//	@Description	Retrieve a paginated list of products available within a specified radius (in kilometers) from given latitude and longitude coordinates.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			lat		query		number	true	"Latitude coordinate"
//	@Param			lng		query		number	true	"Longitude coordinate"
//	@Param			radius	query		number	true	"Radius in kilometers to search within"
//	@Param			limit	query		int		false	"Limit number of results"	default(20)
//	@Param			offset	query		int		false	"Offset for pagination"		default(0)
//	@Success		200		{object}	map[string]interface{}	"List of products within the specified radius"
//	@Failure		400		{object}	map[string]string		"Invalid input parameters"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Router			/products/radius [get]
func (h *ProductHandler) GetProductsByRadius(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	radiusStr := c.Query("radius")

	if latStr == "" || lngStr == "" || radiusStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat, lng, and radius query parameters are required"})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat"})
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lng"})
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid radius"})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
		return
	}

	products, err := h.productUseCase.GetProductsByRadius(c, lat, lng, radius, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response.ResolveProductItemsImages(h.cloudService, products)
	c.JSON(http.StatusOK, gin.H{"products": products})

}

// SaveDepartment godoc
//
//	@Summary		Save a new department
//	@Description	API endpoint to create a new product department.
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			department	body		request.Department	true	"Department to be created"
//	@Success		201			{object}	response.Response{}	"Successfully department saved"
//	@Failure		400			{object}	response.Response{}	"Invalid input"
//	@Failure		500			{object}	response.Response{}	"Failed to save department"
//	@Router			/admin/departments [post]
func (a *ProductHandler) SaveDepartment(ctx *gin.Context) {
	var body request.Department

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err := a.productUseCase.SaveDepartment(ctx, body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to save department", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully department saved", nil)
}

// GetAllDepartments godoc
//
//	@Summary		Get all departments
//	@Description	API endpoint to retrieve all product departments.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{}	"Successfully retrieved all departments"
//	@Failure		500	{object}	response.Response{}	"Failed to get departments"
//	@Router			/departments [get]

func (a *ProductHandler) GetAllDepartments(ctx *gin.Context) {

	tokenString := ctx.GetHeader("Authorization")
	adminId := a.tokenService.DecodeTokenData(tokenString)
	fmt.Printf("Admin ID from token: %s\n", adminId)
	departments, err := a.productUseCase.GetAllDepartments(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get departments", err, nil)
		return
	}

	response.ResolveDepartmentsImages(a.cloudService, departments)
	ctx.JSON(http.StatusOK, response.Response{
		Status:  true,
		Message: "Successfully get all departments",
		Data:    departments,
	})
}

// GetDepartmentByID godoc
//
//	@Summary		Get department by ID
//	@Description	API endpoint to retrieve a product department by its ID.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint	true	"Department ID"
//	@Success		200	{object}	response.Response{}	"Successfully retrieved department"
//	@Failure		400	{object}	response.Response{}	"Invalid department ID"
//	@Failure		500	{object}	response.Response{}	"Failed to get department"
//	@Router			/departments/{id} [get]
func (a *ProductHandler) GetDepartmentByID(ctx *gin.Context) {
	departmentID := ctx.Param("department_id")

	department, err := a.productUseCase.GetDepartmentByID(ctx, departmentID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get department", err, nil)
		return
	}

	department.ResolveImages(a.cloudService)
	response.SuccessResponse(ctx, http.StatusOK, "Successfully get department", department)
}

// GetAllCategoriesByDepartmentID godoc
//
//	@Summary		Get all categories by department ID
//	@Description	API endpoint to retrieve all product categories under a specific department.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			department_id	path		uint	true	"Department ID"
//	@Success		200				{object}	response.Response{}	"Successfully retrieved all categories"
//	@Failure		400				{object}	response.Response{}	"Invalid department ID"
//	@Failure		500				{object}	response.Response{}	"Failed to get categories"
//	@Router			/departments/{department_id}/categories [get]
func (a *ProductHandler) GetAllCategoriesByDepartmentID(ctx *gin.Context) {
	departmentID := ctx.Param("department_id")

	categories, err := a.productUseCase.GetAllCategoriesByDepartmentID(ctx, departmentID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get categories", err, nil)
		return
	}

	response.ResolveCategoriesImages(a.cloudService, categories)
	ctx.JSON(http.StatusOK, response.Response{
		Status:  true,
		Message: "Successfully get all categories",
		Data:    categories,
	})
}

// GetAllSubCategoriesByCategoryID godoc
//
//	@Summary		Get all sub-categories by category ID
//	@Description	API endpoint to retrieve all product sub-categories under a specific category.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			category_id	path		uint	true	"Category ID"
//	@Success		200			{object}	response.Response{}	"Successfully retrieved all sub-categories"
//	@Failure		400			{object}	response.Response{}	"Invalid category ID"
//	@Failure		500			{object}	response.Response{}	"Failed to get sub-categories"
//	@Router			/categories/{category_id}/sub-categories [get]
func (a *ProductHandler) GetAllSubCategoriesByCategoryID(ctx *gin.Context) {
	categoryID := ctx.Param("category_id")

	subCategories, err := a.productUseCase.GetAllSubCategoriesByCategoryID(ctx, categoryID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get sub-categories", err, nil)
		return
	}

	response.ResolveSubCategoriesImages(a.cloudService, subCategories)
	ctx.JSON(http.StatusOK, response.Response{
		Status:  true,
		Message: "Successfully get all sub-categories",
		Data:    subCategories,
	})
}

// SaveSubTypeAttribute godoc
//
//	@Summary		Save a new sub type attribute
//	@Description	API endpoint to create a new sub type attribute for a specific subcategory.
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			sub_category_id	path		uint					true	"Subcategory ID"
//	@Param			attribute		body		request.SubTypeAttribute	true	"Sub type attribute to be created"
//	@Success		201				{object}	response.Response{}	"Successfully sub type attribute created"
//	@Failure		400				{object}	response.Response{}	"Invalid input"
//	@Failure		500				{object}	response.Response{}	"Failed to save sub type attribute"
//	@Router			/admin/sub-categories/{sub_category_id}/attributes [post]
func (p *ProductHandler) SaveSubTypeAttribute(ctx *gin.Context) {
	subCategoryID := ctx.Param("sub_category_id")

	var body request.SubTypeAttribute
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	if err := p.productUseCase.SaveSubTypeAttribute(ctx, subCategoryID, body); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to save sub type attribute", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Sub type attribute created successfully", nil)
}

// GetAllSubTypeAttributes godoc
//
//	@Summary		Get all sub type attributes
//	@Description	API endpoint to retrieve all sub type attributes for a specific subcategory.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			sub_category_id	path		uint	true	"Subcategory ID"
//	@Success		200				{object}	response.Response{}	"Successfully retrieved sub type attributes"
//	@Failure		400				{object}	response.Response{}	"Invalid subcategory ID"
//	@Failure		500				{object}	response.Response{}	"Failed to get sub type attributes"
//	@Router			/sub-categories/{sub_category_id}/attributes [get]
func (p *ProductHandler) GetAllSubTypeAttributes(ctx *gin.Context) {
	subCategoryID := ctx.Param("sub_category_id")

	attributes, err := p.productUseCase.GetAllSubTypeAttributes(ctx, subCategoryID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get sub type attributes", err, nil)
		return
	}

	ctx.JSON(http.StatusOK, response.Response{
		Status:  true,
		Message: "Successfully retrieved sub type attributes",
		Data:    attributes,
	})
}

// GetSubTypeAttributeByID godoc
//
//	@Summary		Get sub type attribute by ID
//	@Description	API endpoint to retrieve a sub type attribute by its ID.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			attribute_id	path		uint	true	"Sub type attribute ID"
//	@Success		200				{object}	response.Response{}	"Successfully retrieved sub type attribute"
//	@Failure		400				{object}	response.Response{}	"Invalid attribute ID"
//	@Failure		500				{object}	response.Response{}	"Failed to get sub type attribute"
//	@Router			/attributes/{attribute_id} [get]
func (p *ProductHandler) GetSubTypeAttributeByID(ctx *gin.Context) {
	attributeID := ctx.Param("attribute_id")

	attribute, err := p.productUseCase.GetSubTypeAttributeByID(ctx, attributeID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get sub type attribute", err, nil)
		return
	}

	ctx.JSON(http.StatusOK, response.Response{
		Status:  true,
		Message: "Successfully retrieved sub type attribute",
		Data:    attribute,
	})
}

// SaveSubTypeAttributeOption godoc
//
//	@Summary		Save a new sub type attribute option
//	@Description	API endpoint to create a new option for a specific sub type attribute.
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			attribute_id	path		uint						true	"Sub type attribute ID"
//	@Param			option			body		request.SubTypeAttributeOption	true	"Sub type attribute option to be created"
//	@Success		201				{object}	response.Response{}	"Successfully sub type attribute option created"
//	@Failure		400				{object}	response.Response{}	"Invalid input"
//	@Failure		500				{object}	response.Response{}	"Failed to save sub type attribute option"
//	@Router			/admin/attributes/{attribute_id}/options [post]
func (p *ProductHandler) SaveSubTypeAttributeOption(ctx *gin.Context) {
	attributeID := ctx.Param("attribute_id")

	var body request.SubTypeAttributeOption
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	if err := p.productUseCase.SaveSubTypeAttributeOption(ctx, attributeID, body); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to save sub type attribute option", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Sub type attribute option created successfully", nil)
}

// GetAllSubTypeAttributeOptions godoc
//
//	@Summary		Get all sub type attribute options
//	@Description	API endpoint to retrieve all options for a specific sub type attribute.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			attribute_id	path		uint	true	"Sub type attribute ID"
//	@Success		200				{object}	response.Response{}	"Successfully retrieved sub type attribute options"
//	@Failure		400				{object}	response.Response{}	"Invalid attribute ID"
//	@Failure		500				{object}	response.Response{}	"Failed to get sub type attribute options"
//	@Router			/attributes/{attribute_id}/options [get]
func (p *ProductHandler) GetAllSubTypeAttributeOptions(ctx *gin.Context) {
	attributeID := ctx.Param("attribute_id")

	options, err := p.productUseCase.GetAllSubTypeAttributeOptions(ctx, attributeID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get sub type attribute options", err, nil)
		return
	}

	ctx.JSON(http.StatusOK, response.Response{
		Status:  true,
		Message: "Successfully retrieved sub type attribute options",
		Data:    options,
	})
}

// GetSubTypeAttributeOptionByID godoc
//
//	@Summary		Get sub type attribute option by ID
//	@Description	API endpoint to retrieve a sub type attribute option by its ID.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			option_id	path		uint	true	"Sub type attribute option ID"
//	@Success		200			{object}	response.Response{}	"Successfully retrieved sub type attribute option"
//	@Failure		400			{object}	response.Response{}	"Invalid option ID"
//	@Failure		500			{object}	response.Response{}	"Failed to get sub type attribute option"
//	@Router			/options/{option_id} [get]
func (p *ProductHandler) GetSubTypeAttributeOptionByID(ctx *gin.Context) {
	optionID := ctx.Param("option_id")

	option, err := p.productUseCase.GetSubTypeAttributeOptionByID(ctx, optionID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get sub type attribute option", err, nil)
		return
	}

	ctx.JSON(http.StatusOK, response.Response{
		Status:  true,
		Message: "Successfully retrieved sub type attribute option",
		Data:    option,
	})
}

// UpdateSubTypeAttribute godoc
//
//	@Summary		Update sub type attribute
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			sub_category_id	path		string					true	"Sub Category ID"
//	@Param			attribute_id	path		string					true	"Attribute ID"
//	@Param			attribute		body		request.SubTypeAttribute	true	"Updated attribute"
//	@Success		200				{object}	response.Response{}
//	@Failure		400				{object}	response.Response{}
//	@Router			/admin/catalog/categories/{category_id}/sub-categories/{sub_category_id}/attributes/{attribute_id} [put]
func (p *ProductHandler) UpdateSubTypeAttribute(ctx *gin.Context) {
	attributeID := ctx.Param("attribute_id")
	var body request.SubTypeAttribute
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}
	if err := p.productUseCase.UpdateSubTypeAttribute(ctx, attributeID, body); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update sub type attribute", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Sub type attribute updated successfully", nil)
}

// DeleteSubTypeAttribute godoc
//
//	@Summary		Delete sub type attribute
//	@Tags			Admin Products
//	@Produce		json
//	@Param			attribute_id	path		string	true	"Attribute ID"
//	@Success		200				{object}	response.Response{}
//	@Failure		500				{object}	response.Response{}
//	@Router			/admin/catalog/categories/{category_id}/sub-categories/{sub_category_id}/attributes/{attribute_id} [delete]
func (p *ProductHandler) DeleteSubTypeAttribute(ctx *gin.Context) {
	attributeID := ctx.Param("attribute_id")
	if err := p.productUseCase.DeleteSubTypeAttribute(ctx, attributeID); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete sub type attribute", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Sub type attribute deleted successfully", nil)
}

// UpdateSubTypeAttributeOption godoc
//
//	@Summary		Update sub type attribute option
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			option_id	path		string							true	"Option ID"
//	@Param			option		body		request.SubTypeAttributeOption	true	"Updated option"
//	@Success		200			{object}	response.Response{}
//	@Failure		400			{object}	response.Response{}
//	@Router			/admin/catalog/categories/{category_id}/sub-categories/{sub_category_id}/attributes/{attribute_id}/options/{option_id} [put]
func (p *ProductHandler) UpdateSubTypeAttributeOption(ctx *gin.Context) {
	optionID := ctx.Param("option_id")
	var body request.SubTypeAttributeOption
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}
	if err := p.productUseCase.UpdateSubTypeAttributeOption(ctx, optionID, body); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update sub type attribute option", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Sub type attribute option updated successfully", nil)
}

// DeleteSubTypeAttributeOption godoc
//
//	@Summary		Delete sub type attribute option
//	@Tags			Admin Products
//	@Produce		json
//	@Param			option_id	path		string	true	"Option ID"
//	@Success		200			{object}	response.Response{}
//	@Failure		500			{object}	response.Response{}
//	@Router			/admin/catalog/categories/{category_id}/sub-categories/{sub_category_id}/attributes/{attribute_id}/options/{option_id} [delete]
func (p *ProductHandler) DeleteSubTypeAttributeOption(ctx *gin.Context) {
	optionID := ctx.Param("option_id")
	if err := p.productUseCase.DeleteSubTypeAttributeOption(ctx, optionID); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete sub type attribute option", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Sub type attribute option deleted successfully", nil)
}

// SaveCategoryImage godoc
//
//	@Summary		Save category image
//	@Description	API endpoint to save an image for a specific product category.
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			category_id	path		uint				true	"Category ID"
//	@Param			image		body		request.CategoryImage	true	"Category image to be saved"
//	@Success		201			{object}	response.Response{}	"Successfully saved category image"
//	@Failure		400			{object}	response.Response{}	"Invalid input"
//	@Failure		500			{object}	response.Response{}	"Failed to save category image"
//	@Router			/admin/categories/{category_id}/images [post]
func (p *ProductHandler) SaveCategoryImage(ctx *gin.Context) {
	categoryID := ctx.Param("category_id")

	var image request.CategoryImage
	if err := ctx.ShouldBindJSON(&image); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err := p.productUseCase.SaveCategoryImage(ctx, categoryID, image)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to save category image", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully saved category image", nil)
}

// GetAllCategoryImages godoc
//
//	@Summary		Get all category images
//	@Description	API endpoint to retrieve all images for a specific product category.
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			category_id	path		uint	true	"Category ID"
//	@Success		200			{object}	response.Response{}	"Successfully retrieved category images"
//	@Failure		400			{object}	response.Response{}	"Invalid category ID"
//	@Failure		500			{object}	response.Response{}	"Failed to get category images"
//	@Router			/admin/categories/{category_id}/images [get]
func (p *ProductHandler) GetAllCategoryImages(ctx *gin.Context) {
	categoryID := ctx.Param("category_id")

	images, err := p.productUseCase.GetAllCategoryImages(ctx, categoryID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get category images", err, nil)
		return
	}

	ctx.JSON(http.StatusOK, response.Response{
		Status:  true,
		Message: "Successfully retrieved category images",
		Data:    images,
	})
}

// GetCategoryImageByID godoc
//
//	@Summary		Get category image by ID
//	@Description	API endpoint to retrieve a category image by its ID.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			image_id	path		uint	true	"Category Image ID"
//	@Success		200			{object}	response.Response{}	"Successfully retrieved category image"
//	@Failure		400			{object}	response.Response{}	"Invalid category image ID"
//	@Failure		500			{object}	response.Response{}	"Failed to get category image"
//	@Router			/products/category/image/{image_id} [get]
func (p *ProductHandler) GetCategoryImageByID(ctx *gin.Context) {
	imageID := ctx.Param("image_id")

	image, err := p.productUseCase.GetCategoryImageByID(ctx, imageID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get category image", err, nil)
		return
	}

	ctx.JSON(http.StatusOK, response.Response{
		Status:  true,
		Message: "Successfully retrieved category image",
		Data:    image,
	})
}

// UpdateCategoryImage godoc
//
//	@Summary		Update category image
//	@Description	API endpoint to update a category image by its ID.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			image_id	path		int	true	"Category Image ID"
//	@Param			image		body		request.CategoryImage	true	"Updated category image data"
//	@Success		200			{object}	response.Response{}	"Successfully updated category image"
//	@Failure		400			{object}	response.Response{}	"Invalid input"
//	@Failure		500			{object}	response.Response{}	"Failed to update category image"
//	@Router			/products/category/image/{image_id} [put]
func (p *ProductHandler) UpdateCategoryImage(ctx *gin.Context) {
	imageID := ctx.Param("image_id")

	var image request.CategoryImage
	if err := ctx.ShouldBindJSON(&image); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err := p.productUseCase.UpdateCategoryImage(ctx, imageID, image)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update category image", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated category image", nil)
}

// DeleteCategoryImage godoc
//	@Summary		Delete category image
//	@Description	Delete a category image by its ID.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			image_id	path		int	true	"Category Image ID"
//	@Success		200			{object}	response.Response{}	"Successfully deleted category image"
//	@Failure		400			{object}	response.Response{}	"Invalid category image ID"
//	@Failure		500			{object}	response.Response{}	"Internal server error"
//	@Router			/products/category/image/{image_id} [delete]

func (p *ProductHandler) DeleteCategoryImage(ctx *gin.Context) {
	imageID := ctx.Param("image_id")

	err := p.productUseCase.DeleteCategoryImage(ctx, imageID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete category image", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully deleted category image", nil)
}

// GetProductItemByID godoc
// /	@Summary		Get product item by ID
//
//	@Description	Retrieve a product item by its unique ID.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			product_item_id	path		int	true	"Product Item ID"
//	@Success		200				{object}	response.Response{}	"Successfully retrieved product item"
//	@Failure		400				{object}	response.Response{}	"Invalid product item ID"
//	@Failure		500				{object}	response.Response{}	"Internal server error"
//	@Router			/products/item/{product_item_id} [get]
//
// GetProductItemByID is the admin/seller entry point (not subscription-gated).
func (p *ProductHandler) GetProductItemByID(ctx *gin.Context) {
	p.getProductItemByID(ctx, false)
}

// GetProductItemByIDUser is the customer entry point: a product whose owning
// shop has no active subscription is hidden (404) when the gate is enabled.
func (p *ProductHandler) GetProductItemByIDUser(ctx *gin.Context) {
	p.getProductItemByID(ctx, true)
}

func (p *ProductHandler) getProductItemByID(ctx *gin.Context, customerView bool) {
	productItemID := ctx.Param("product_item_id")

	productItem, err := p.productUseCase.GetProductItemByID(ctx, productItemID, customerView)
	fmt.Printf("Fetched product item: %+v\n", productItem)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ErrorResponse(ctx, http.StatusNotFound, "Product item not found", err, nil)
			return
		}
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get product item", err, nil)
		return
	}

	// Increment view count
	tokenString := ctx.GetHeader("Authorization")
	adminId := p.tokenService.DecodeTokenData(tokenString)
	err = p.productUseCase.IncrementProductItemViewCount(ctx, productItemID, adminId)
	if err != nil {
		// Log the error but don't fail the request
	}

	productItem.ResolveImages(p.cloudService)

	response.SuccessResponse(ctx, http.StatusOK, "Successfully get product item", productItem)
}

func (p *ProductHandler) IncrementProductItemViewCount(ctx *gin.Context) {
	productItemID := ctx.Param("product_item_id")

	tokenString := ctx.GetHeader("Authorization")
	adminId := p.tokenService.DecodeTokenData(tokenString)

	err := p.productUseCase.IncrementProductItemViewCount(ctx, productItemID, adminId)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to increment view count", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully incremented view count", nil)
}

func (p *ProductHandler) GetProductItemViewCount(ctx *gin.Context) {
	productItemID := ctx.Param("product_item_id")

	tokenString := ctx.GetHeader("Authorization")
	adminId := p.tokenService.DecodeTokenData(tokenString)

	viewCount, err := p.productUseCase.GetProductItemViewCount(ctx, productItemID, adminId)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get view count", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved view count", map[string]interface{}{
		"product_item_id": productItemID,
		"view_count":      viewCount,
	})
}

// DeleteProductItem godoc
//
//	@Summary		Delete a product item
//	@Security		BearerAuth
//	@Description	API for admin to delete a product item by ID
//	@ID				DeleteProductItem
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			product_item_id	path	int	true	"Product Item ID"
//	@Success		200				{object}	response.Response{}	"Successfully deleted product item"
//	@Failure		400				{object}	response.Response{}	"Bad request"
//	@Failure		500				{object}	response.Response{}	"Internal server error"
//	@Router			/items/{product_item_id} [delete]
func (p *ProductHandler) DeleteProductItem(ctx *gin.Context) {
	productItemID := ctx.Param("product_item_id")

	err := p.productUseCase.DeleteProductItem(ctx, productItemID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete product item", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully deleted product item", nil)
}

func (p *ProductHandler) UpdateProductItem(ctx *gin.Context) {
	productItemID := ctx.Param("product_item_id")

	subCategoryID := ctx.PostForm("sub_category_id")
	categoryID := ctx.PostForm("category_id")
	departmentID := ctx.PostForm("department_id")
	subCategoryName := ctx.PostForm("sub_category_name")
	categoryName := ctx.PostForm("category_name")
	dynamicFieldsStr := ctx.PostForm("dynamic_fields")
	description := ctx.PostForm("description")
	highlightsStr := ctx.PostForm("highlights")
	retainedImagesStr := ctx.PostFormArray("retained_images[]")

	var highlights []string
	if highlightsStr != "" {
		if err := json.Unmarshal([]byte(highlightsStr), &highlights); err != nil {
			highlights = []string{highlightsStr}
		}
	}

	files := ctx.Request.MultipartForm.File["images[]"]
	newImagePaths := make([]string, len(files))

	// Upload new images (single path, no duplicate)
	for i, fileHeader := range files {
		localPath, err := handleUpload(fileHeader)
		if err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to process image", err, nil)
			return
		}

		if rejection := p.validateProductImage(localPath, categoryName); rejection != "" {
			response.ErrorResponse(ctx, http.StatusBadRequest, rejection, nil, nil)
			return
		}

		objectKey, err := uploadProcessedToCloud(ctx, p.cloudService, localPath)
		if err != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload processed image", err, nil)
			return
		}
		newImagePaths[i] = objectKey
	}

	var dynamicFields map[string]interface{}
	if dynamicFieldsStr != "" {
		if err := json.Unmarshal([]byte(dynamicFieldsStr), &dynamicFields); err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid dynamic_fields JSON", err, nil)
			return
		}
	}

	req := request.ProductItem{
		SubCategoryName:   subCategoryName,
		SubCategoryID:     subCategoryID,
		DynamicFields:     dynamicFields,
		CategoryID:        categoryID,
		DepartmentID:      departmentID,
		ProductItemImages: newImagePaths,
		Description:       description,
		Highlights:        highlights,
		RetainedImages:    retainedImagesStr,
	}

	if err := p.productUseCase.UpdateProductItem(ctx, productItemID, req); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update product item", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated product item", nil)
}

// UpdateProductItemStock godoc
//
//	@Summary		Update product item stock status
//	@Description	Toggle a product item's in-stock / out-of-stock status.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			product_item_id	path		string							true	"Product Item ID"
//	@Param			body			body		request.UpdateProductItemStock	true	"Stock status"
//	@Success		200				{object}	response.Response{}				"Successfully updated stock status"
//	@Failure		400				{object}	response.Response{}				"Invalid request"
//	@Failure		500				{object}	response.Response{}				"Failed to update stock status"
//	@Router			/admin/items/{product_item_id}/stock [patch]
func (p *ProductHandler) UpdateProductItemStock(ctx *gin.Context) {
	productItemID := ctx.Param("product_item_id")

	var body request.UpdateProductItemStock
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}

	if err := p.productUseCase.UpdateProductItemStock(ctx, productItemID, *body.Stock); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update stock status", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated stock status", nil)
}

// FindProductItemFilters godoc
//
//	@Summary		Find product item filters
//	@Description	API endpoint to retrieve available filters for product items based on provided criteria.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	response.Response{}	"Successfully retrieved product item filters"

func (p *ProductHandler) FindProductItemFilters(ctx *gin.Context) {
	shopID := ctx.Param("shop_id")

	tokenString := ctx.GetHeader("Authorization")
	adminID := p.tokenService.DecodeTokenData(tokenString)

	filters, err := p.productUseCase.FindProductItemFilters(ctx, adminID, shopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get product item filters", err, nil)
		return
	}

	ctx.JSON(http.StatusOK, response.Response{
		Status:  true,
		Message: "Successfully retrieved product item filters",
		Data:    filters,
	})
}

func (p *ProductHandler) GetProductItemsByOfferID(ctx *gin.Context) {
	// Get Offer ID, Category ID, Department ID, Sub category ID, Lattitde, Longitude, Radius, Pincode, Limit, Offset from query parameters
	offerID := ctx.Param("offer_id")
	var err error
	_ = err // suppress unused warning below
	CategoryId, _ := strconv.Atoi(ctx.Query("category_id"))
	DepartmentId, _ := strconv.Atoi(ctx.Query("department_id"))
	SubCategoryId, _ := strconv.Atoi(ctx.Query("sub_category_id"))
	latStr := ctx.Query("lat")
	lngStr := ctx.Query("lng")
	pincode := ctx.Query("pincode")

	// Accept both radius_km and radius parameters
	radiusKmStr := ctx.Query("radius_km")
	if radiusKmStr == "" {
		radiusKmStr = ctx.Query("radius")
	}

	var radiusKm float64
	if radiusKmStr != "" {
		var err error
		radiusKm, err = strconv.ParseFloat(radiusKmStr, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid radius parameter"})
			return
		}
		if radiusKm < 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "radius must be non-negative"})
			return
		}
	}
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	products, err := p.productUseCase.GetProductItemsByOfferID(ctx, offerID, CategoryId, DepartmentId, SubCategoryId, latStr, lngStr, pincode, radiusKm, limit, offset)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get product items by offer ID", err, nil)
		return
	}

	response.ResolveProductItemsImages(p.cloudService, products)

	ctx.JSON(http.StatusOK, response.Response{
		Status:  true,
		Message: "Successfully retrieved product items by offer ID",
		Data:    products,
	})

}

// uploadProcessedToCloud uploads the processed JPEG at processedPath to object
// storage under the products/ namespace and removes the temp file. Returns the
// bare object key suitable for DB storage.
func uploadProcessedToCloud(ctx context.Context, cs cloud.CloudService, processedPath string) (string, error) {
	data, err := os.ReadFile(processedPath)
	if err != nil {
		log.Printf("uploadProcessedToCloud read failed: path=%s err=%v", processedPath, err)
		return "", err
	}
	defer func() {
		if rerr := os.Remove(processedPath); rerr != nil && !os.IsNotExist(rerr) {
			log.Printf("failed to remove temp file %s: %v", processedPath, rerr)
		}
	}()
	log.Printf("uploadProcessedToCloud start: path=%s bytes=%d filename=%s", processedPath, len(data), filepath.Base(processedPath))
	objectKey, err := cs.SaveBytes(ctx, data, cloud.SaveOptions{
		Namespace:   "products",
		Visibility:  cloud.VisibilityPublic,
		ContentType: "image/jpeg",
		Filename:    filepath.Base(processedPath),
	})
	if err != nil {
		log.Printf("uploadProcessedToCloud save failed: path=%s bytes=%d err=%v", processedPath, len(data), err)
		return "", err
	}
	log.Printf("uploadProcessedToCloud success: path=%s objectKey=%s bytes=%d", processedPath, objectKey, len(data))
	return objectKey, nil
}

// handleUpload runs the image-processing pipeline and adult-content moderation.
// Returns the absolute path to a temp file holding the processed JPEG; the
// caller is responsible for cleaning it up (uploadProcessedToCloud does so).
func handleUpload(fileHeader *multipart.FileHeader) (string, error) {
	savedPath, err := handleSecureMagic(fileHeader)
	if err != nil {
		return "", err
	}
	isAdult, err := utils.CheckNudity(savedPath)
	if err != nil {
		_ = os.Remove(savedPath)
		return "", fmt.Errorf("failed to check nudity: %w", err)
	}
	if isAdult {
		_ = os.Remove(savedPath)
		return "", fmt.Errorf("image contains adult content and cannot be uploaded")
	}
	return savedPath, nil
}

// func handleAdultContent(fileName string) (bool, error) {
// 	isAdult, err := checkAdultContent(fileName)
// 	if err != nil {
// 		return false, err
// 	}

// 	if isAdult {
// 		return false, fmt.Errorf("content policy violation: adult image detected")
// 	}
// 	return isAdult, nil
// }

// handleSecureMagic validates and enhances the image, returns the file path as a string
func handleSecureMagic(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("failed to close file: %v", cerr)
		}
	}()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	contentType := http.DetectContentType(buffer)
	file.Seek(0, 0)

	if contentType != "image/jpeg" && contentType != "image/png" {
		return "", fmt.Errorf("unsupported file type: %s", contentType)
	}

	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	src, err := imaging.Decode(file)
	if err != nil {
		return "", err
	}

	// Resize first (important for file size)
	processed := imaging.Resize(src, 1080, 0, imaging.Lanczos)

	// Improve brightness slightly
	processed = imaging.AdjustBrightness(processed, 5)

	// Add contrast for punch
	processed = imaging.AdjustContrast(processed, 18)

	// Boost saturation carefully (avoid oversaturation)
	processed = imaging.AdjustSaturation(processed, 12)

	// Light sharpening for product clarity
	processed = imaging.Sharpen(processed, 0.6)

	filename := fmt.Sprintf("%s_tweak.jpg", uuid.New().String())
	savePath := filepath.Join(os.TempDir(), filename)
	outFile, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := outFile.Close(); cerr != nil {
			log.Printf("failed to close file: %v", cerr)
		}
	}()
	if err := imaging.Encode(outFile, processed, imaging.JPEG, imaging.JPEGQuality(20)); err != nil {
		return "", err
	}
	return savePath, nil
}

func (p *ProductHandler) CreateCategory(ctx *gin.Context) {
	departmentID := ctx.Param("department_id")
	if departmentID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "department_id is required", nil, nil)
		return
	}
	name := ctx.PostForm("category_name")
	if len(name) < 1 || len(name) > 30 {
		response.ErrorResponse(ctx, http.StatusBadRequest, "category_name must be between 1 and 30 characters", nil, nil)
		return
	}
	sortOrder, _ := strconv.Atoi(ctx.PostForm("sort_order"))
	isActive := ctx.PostForm("is_active") != "false"

	var imageURL string
	fileHeader, err := ctx.FormFile("photo")
	if err == nil && fileHeader != nil {
		objectKey, uploadErr := p.cloudService.SaveFile(ctx, fileHeader, cloud.SaveOptions{
			Namespace:  "categories",
			Visibility: cloud.VisibilityPublic,
		})
		if uploadErr != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload image", uploadErr, nil)
			return
		}
		imageURL = objectKey
	}

	if err := p.productUseCase.CreateCategory(ctx, departmentID, name, imageURL, sortOrder, isActive); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create category", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Successfully created category", nil)
}

func (p *ProductHandler) UpdateCategory(ctx *gin.Context) {
	categoryID := ctx.Param("category_id")
	if categoryID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "category_id is required", nil, nil)
		return
	}
	name := ctx.PostForm("category_name")
	if len(name) < 1 || len(name) > 30 {
		response.ErrorResponse(ctx, http.StatusBadRequest, "category_name must be between 1 and 30 characters", nil, nil)
		return
	}
	sortOrder, _ := strconv.Atoi(ctx.PostForm("sort_order"))
	isActive := ctx.PostForm("is_active") != "false"

	var imageURL string
	fileHeader, err := ctx.FormFile("photo")
	if err == nil && fileHeader != nil {
		objectKey, uploadErr := p.cloudService.SaveFile(ctx, fileHeader, cloud.SaveOptions{
			Namespace:  "categories",
			Visibility: cloud.VisibilityPublic,
		})
		if uploadErr != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload image", uploadErr, nil)
			return
		}
		imageURL = objectKey
	}

	if err := p.productUseCase.UpdateCategory(ctx, categoryID, name, imageURL, sortOrder, isActive); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update category", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated category", nil)
}

func (p *ProductHandler) CreateSubCategory(ctx *gin.Context) {
	departmentID := ctx.Param("department_id")
	categoryID := ctx.Param("category_id")
	if departmentID == "" || categoryID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "department_id and category_id are required", nil, nil)
		return
	}
	name := ctx.PostForm("sub_category_name")
	if len(name) < 1 || len(name) > 30 {
		response.ErrorResponse(ctx, http.StatusBadRequest, "sub_category_name must be between 1 and 30 characters", nil, nil)
		return
	}
	sortOrder, _ := strconv.Atoi(ctx.PostForm("sort_order"))
	isActive := ctx.PostForm("is_active") != "false"

	var imageURL string
	fileHeader, err := ctx.FormFile("photo")
	if err == nil && fileHeader != nil {
		objectKey, uploadErr := p.cloudService.SaveFile(ctx, fileHeader, cloud.SaveOptions{
			Namespace:  "sub-categories",
			Visibility: cloud.VisibilityPublic,
		})
		if uploadErr != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload image", uploadErr, nil)
			return
		}
		imageURL = objectKey
	}

	if err := p.productUseCase.CreateSubCategory(ctx, departmentID, categoryID, name, imageURL, sortOrder, isActive); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create sub-category", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Successfully created sub-category", nil)
}

func (p *ProductHandler) UpdateSubCategory(ctx *gin.Context) {
	subCategoryID := ctx.Param("sub_category_id")
	if subCategoryID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "sub_category_id is required", nil, nil)
		return
	}
	name := ctx.PostForm("sub_category_name")
	if len(name) < 1 || len(name) > 30 {
		response.ErrorResponse(ctx, http.StatusBadRequest, "sub_category_name must be between 1 and 30 characters", nil, nil)
		return
	}
	sortOrder, _ := strconv.Atoi(ctx.PostForm("sort_order"))
	isActive := ctx.PostForm("is_active") != "false"

	var imageURL string
	fileHeader, err := ctx.FormFile("photo")
	if err == nil && fileHeader != nil {
		objectKey, uploadErr := p.cloudService.SaveFile(ctx, fileHeader, cloud.SaveOptions{
			Namespace:  "sub-categories",
			Visibility: cloud.VisibilityPublic,
		})
		if uploadErr != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload image", uploadErr, nil)
			return
		}
		imageURL = objectKey
	}

	if err := p.productUseCase.UpdateSubCategory(ctx, subCategoryID, name, imageURL, sortOrder, isActive); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update sub-category", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated sub-category", nil)
}

func (p *ProductHandler) DeleteCategory(ctx *gin.Context) {
	categoryID := ctx.Param("category_id")
	if categoryID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "category_id is required", nil, nil)
		return
	}
	if err := p.productUseCase.DeleteCategory(ctx, categoryID); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete category", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully deleted category", nil)
}

func (p *ProductHandler) DeleteSubCategory(ctx *gin.Context) {
	subCategoryID := ctx.Param("sub_category_id")
	if subCategoryID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "sub_category_id is required", nil, nil)
		return
	}
	if err := p.productUseCase.DeleteSubCategory(ctx, subCategoryID); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete sub-category", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully deleted sub-category", nil)
}
