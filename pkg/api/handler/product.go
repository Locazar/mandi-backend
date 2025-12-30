package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

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
}

func NewProductHandler(productUsecase usecaseInterface.ProductUseCase, tokenService token.TokenService) interfaces.ProductHandler {
	return &ProductHandler{
		productUseCase: productUsecase,
		tokenService:   tokenService,
	}
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
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve categories", err, nil)
		return
	}

	if len(categories) == 0 {
		response.SuccessResponse(ctx, http.StatusOK, "No categories found", nil)
		return
	}

	ctx.JSON(http.StatusOK, response.Response{
		Status:  true,
		Message: "Successfully get all categories",
		Data:    categories,
	})
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

		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrCategoryAlreadyExist) {
			statusCode = http.StatusConflict
		}

		response.ErrorResponse(ctx, statusCode, "Failed to add category", err, nil)
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

		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrCategoryAlreadyExist) {
			statusCode = http.StatusConflict
		}

		response.ErrorResponse(ctx, statusCode, "Failed to add sub category", err, nil)
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

	categoryID, err := request.GetParamAsUint(ctx, "category_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	var body request.Variation

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err = p.productUseCase.SaveVariation(ctx, categoryID, body.Names)

	if err != nil {
		var statusCode = http.StatusInternalServerError
		if errors.Is(err, usecase.ErrVariationAlreadyExist) {
			statusCode = http.StatusConflict
		}
		response.ErrorResponse(ctx, statusCode, "Failed to add variation", err, nil)
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

	variationID, err := request.GetParamAsUint(ctx, "variation_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	var body request.VariationOption

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err = p.productUseCase.SaveVariationOption(ctx, variationID, body.Values)
	if err != nil {
		var statusCode = http.StatusInternalServerError
		if errors.Is(err, usecase.ErrVariationOptionAlreadyExist) {
			statusCode = http.StatusConflict
		}
		response.ErrorResponse(ctx, statusCode, "Failed to add variation options", err, nil)
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

	categoryID, err := request.GetParamAsUint(ctx, "category_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	variations, err := c.productUseCase.FindAllVariationsAndItsValues(ctx, categoryID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to Get variations and its values", err, nil)
		return
	}

	if len(variations) == 0 {
		response.SuccessResponse(ctx, http.StatusOK, "No variations found")
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
	fmt.Printf("tokenString: %v\n", tokenString)

	adminID := p.tokenService.DecodeTokenData(tokenString)
	// Check if this is a JSON request (without file upload)
	contentType := ctx.GetHeader("Content-Type")
	if contentType == "application/json" || contentType == "" {
		p.SaveProductJSON(ctx, adminID)
		return
	}

	// Handle multipart/form-data request
	name, err1 := request.GetFormValuesAsString(ctx, "category")
	departmentID, errDeptID := request.GetFormValuesAsUint(ctx, "department_id")
	description, err2 := request.GetFormValuesAsString(ctx, "description")
	categoryID, err3 := request.GetFormValuesAsUint(ctx, "category_id")

	fileHeader, err6 := ctx.FormFile("image")

	fmt.Printf("Received form data - Name: %s, DepartmentID: %d, Description: %s, CategoryID: %d\n", name, departmentID, description, categoryID)
	// Only check required fields
	err := errors.Join(err1, err2, err3, err6, errDeptID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindFormValueMessage, err, nil)
		return
	}

	product := request.Product{
		Name:            name,
		DepartmentID:    departmentID,
		Description:     description,
		CategoryID:      categoryID,
		ImageFileHeader: fileHeader,
	}

	fmt.Printf("Product to be saved: %+v\n", product)

	productID, err := p.productUseCase.SaveProduct(ctx, product, adminID)

	fmt.Printf("Result of SaveProduct - productID: %d, err: %v\n", productID, err)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrProductAlreadyExist) {
			statusCode = http.StatusConflict
		}
		fmt.Printf("Successfully product added: %v\n", map[string]uint{"product_id": productID})
		response.ErrorResponse(ctx, statusCode, "Failed to add product", err, map[string]uint{"product_id": productID})
		return
	}

	fmt.Printf("Successfully product added: %v\n", map[string]uint{"product_id": productID})

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully product added", map[string]uint{"product_id": productID})
}

// SaveProductJSON handles JSON requests without image uploa
func (p *ProductHandler) SaveProductJSON(ctx *gin.Context, adminID string) {
	// Debug: Log raw request body
	rawData, _ := ctx.GetRawData()
	fmt.Printf("\n=== DEBUG SaveProductJSON ===\n")
	fmt.Printf("Raw request body: %s\n", string(rawData))
	fmt.Printf("Content-Type: %s\n", ctx.GetHeader("Content-Type"))
	fmt.Printf("Body length: %d bytes\n", len(rawData))

	// Check if the body starts with a quote (indicating it's a string-wrapped JSON)
	if len(rawData) > 0 && rawData[0] == '"' {
		fmt.Printf("WARNING: Request body appears to be a JSON string (double-encoded)\n")
		// Unwrap the double-encoded JSON string
		var unwrappedJSON string
		if err := json.Unmarshal(rawData, &unwrappedJSON); err != nil {
			fmt.Printf("Failed to unwrap double-encoded JSON: %v\n", err)
			response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid JSON format (double-encoded)", err, nil)
			return
		}
		rawData = []byte(unwrappedJSON)
		fmt.Printf("Unwrapped JSON: %s\n", string(rawData))
		fmt.Printf("First 100 chars of unwrapped: %s\n", string(rawData[:min(100, len(rawData))]))
	}

	// Additional validation: Check if it's valid JSON before binding
	var testJSON interface{}
	if err := json.Unmarshal(rawData, &testJSON); err != nil {
		fmt.Printf("JSON validation failed: %v\n", err)
		fmt.Printf("Problematic JSON (first 200 chars): %s\n", string(rawData[:min(200, len(rawData))]))
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid JSON syntax", err, nil)
		return
	}
	fmt.Printf("JSON validation passed\n")
	fmt.Printf("=============================\n\n")

	// Re-bind the body since we read it
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(rawData))

	var body struct {
		Name           string      `json:"category" binding:"min=3,max=50"`
		Description    string      `json:"description"`
		CategoryID     uint        `json:"category_id"`
		DepartmentID   uint        `json:"department_id"`
		Condition      string      `json:"condition" binding:"omitempty"`
		Specifications interface{} `json:"specifications" binding:"omitempty"` // Can be string or object
		Highlights     interface{} `json:"highlights" binding:"omitempty"`     // Can be string or array
		ImageURL       string      `json:"image_url" binding:"omitempty"`      // For existing image URL
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		fmt.Printf("JSON binding error: %v\n", err)
		fmt.Printf("Error type: %T\n", err)
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	// For JSON requests without file upload, you need to provide image_url or handle it differently
	if body.ImageURL == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "image_url is required for JSON requests", errors.New("image_url field is missing or empty"), nil)
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
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrProductAlreadyExist) {
			statusCode = http.StatusConflict
		}
		response.ErrorResponse(ctx, statusCode, "Failed to add product", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully product added", map[string]uint{"product_id": productID})
}

// GetAllProductsAdmin godoc
//
//	@Summary		Get all products (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to get all products
//	@ID				GetAllProductsAdmin
//	@Tags			Admin Products
//	@Param			page_number	query	int	false	"Page Number"
//	@Param			count		query	int	false	"Count"
//	@Router			/admin/products [get]
//	@Success		200	{object}	response.Response{}	"Successfully found all products"
//	@Failure		500	{object}	response.Response{}	"Failed to Get all products"
func (p *ProductHandler) GetAllProductsAdmin() func(ctx *gin.Context) {
	return p.getAllProducts()
}

// GetAllProductsUser godoc
//
//	@Summary		Get all products (User)
//	@Security		BearerAuth
//	@Description	API for user to get all products
//	@ID				GetAllProductsUser
//	@Tags			User Products
//	@Param			page_number	query	int	false	"Page Number"
//	@Param			count		query	int	false	"Count"
//	@Router			/products [get]
//	@Success		200	{object}	response.Response{}	"Successfully found all products"
//	@Failure		500	{object}	response.Response{}	"Failed to get all products"
func (p *ProductHandler) GetAllProductsUser() func(ctx *gin.Context) {
	return p.getAllProducts()
}

// Get products is common for user and admin so this function is to get the common Get all products func for them
func (p *ProductHandler) getAllProducts() func(ctx *gin.Context) {

	return func(ctx *gin.Context) {

		pagination := request.GetPagination(ctx)

		products, err := p.productUseCase.FindAllProducts(ctx, pagination)

		if err != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to Get all products", err, nil)
			return
		}

		if len(products) == 0 {
			response.SuccessResponse(ctx, http.StatusOK, "No products found", nil)
			return
		}

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
	productID, err := request.GetParamAsUint(ctx, "product_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid product ID", err, nil)
		return
	}

	product, err := p.productUseCase.FindProductByID(ctx, productID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get product", err, nil)
		return
	}

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
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrProductAlreadyExist) {
			statusCode = http.StatusConflict
		}
		response.ErrorResponse(ctx, statusCode, "Failed to update product", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully product updated", nil)
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

	subCategoryIDStr := ctx.PostForm("sub_category_id")
	fmt.Printf("SubCategoryID from form: %s\n", subCategoryIDStr)
	var subCategoryID uint
	if subCategoryIDStr != "" {
		if n, err := strconv.Atoi(subCategoryIDStr); err == nil {
			subCategoryID = uint(n)
		} else {
			response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid sub_category_id", err, nil)
			return
		}
	}

	categoryIDStr := ctx.PostForm("category_id")
	var categoryID uint
	if categoryIDStr != "" {
		if n, err := strconv.Atoi(categoryIDStr); err == nil {
			categoryID = uint(n)
		} else {
			response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid category_id", err, nil)
			return
		}
	}

	departmentIDStr := ctx.PostForm("department_id")
	var departmentID uint
	if departmentIDStr != "" {
		if n, err := strconv.Atoi(departmentIDStr); err == nil {
			departmentID = uint(n)
		} else {
			response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid department_id", err, nil)
			return
		}
	}

	tokenString := ctx.GetHeader("Authorization")
	fmt.Printf("tokenString: %v\n", tokenString)

	adminID := p.tokenService.DecodeTokenData(tokenString)

	subCategoryName := ctx.PostForm("sub_category_name")
	dynamicFieldsStr := ctx.PostForm("dynamic_fields")
	files := ctx.Request.MultipartForm.File["images[]"]
	fmt.Printf("Admin ID from token: %s\n", adminID)
	fmt.Printf("SubCategory Name from form: %s\n", subCategoryName)
	fmt.Printf("Dynamic Fields from form: %s\n", dynamicFieldsStr)
	fmt.Printf("Image Files from form: %+v\n", files)

	var imagePaths []string
	for _, fileHeader := range files {
		localPath, err := utils.SaveFileLocally(fileHeader, "uploads/products")
		if err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, BindFormValueMessage, err, nil)
			return
		}
		imagePaths = append(imagePaths, localPath)
	}
	fmt.Printf("Saved image paths: %+v\n", imagePaths)

	var dynamicFields map[string]interface{}
	if err := json.Unmarshal([]byte(dynamicFieldsStr), &dynamicFields); err != nil {
		// handle error
	}
	productItem := request.ProductItem{
		SubCategoryName:   subCategoryName,
		SubCategoryID:     subCategoryID,
		DynamicFields:     dynamicFields,
		CategoryID:        categoryID,
		DepartmentID:      departmentID,
		ProductItemImages: imagePaths,
	}

	// // Convert request.ProductItem to domain.ProductItem
	// var domainProductItem domain.ProductItem
	// if err := copier.Copy(&domainProductItem, &productItem); err != nil {
	// 	response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to map product item", err, nil)
	// 	return
	// }

	err := p.productUseCase.SaveProductItem(ctx, productItem, adminID)

	if err != nil {

		var statusCode int

		switch {
		case errors.Is(err, usecase.ErrProductItemAlreadyExist):
			statusCode = http.StatusConflict
		case errors.Is(err, usecase.ErrNotEnoughVariations):
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusInternalServerError
		}

		response.ErrorResponse(ctx, statusCode, "Failed to add product item", err, nil)
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
func (p *ProductHandler) GetAllProductItemsAdmin() func(ctx *gin.Context) {
	return p.getAllProductItems()
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
func (p *ProductHandler) GetAllProductItemsUser() func(ctx *gin.Context) {
	return p.getAllProductItems()
}

// same functionality of get all product items for admin and user
func (p *ProductHandler) getAllProductItems() func(ctx *gin.Context) {

	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")
		fmt.Printf("tokenString: %v\n", tokenString)

		adminID := p.tokenService.DecodeTokenData(tokenString)
		productItems, err := p.productUseCase.FindAllProductItems(ctx, adminID)

		if err != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get all product items", err, nil)
			return
		}

		// check the product have productItem exist or not
		if len(productItems) == 0 {
			response.SuccessResponse(ctx, http.StatusOK, "No product items found")
			return
		}

		fmt.Printf("Product Items: %+v\n", productItems)

		response.SuccessResponse(ctx, http.StatusOK, "Successfully get all product items ", productItems)
	}
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
//	@Description	API for user to search products with filters
//	@ID				SearchProducts
//	@Tags			User Products
//	@Accept			json
//	@Produce		json
//	@Param			q			query	string	false	"Search keyword"
//	@Param			category_id	query	string	false	"Category ID"
//	@Param			brand_id	query	string	false	"Brand ID"
//	@Param			location_id	query	string	false	"Location ID"
//	@Param			limit		query	int		false	"Limit"
//	@Param			offset		query	int		false	"Offset"
//	@Router			/products/search [get]
//	@Success		200	{object}	response.Response{}	"Successfully searched products"
//	@Failure		500	{object}	response.Response{}	"Failed to search products
func (h *ProductHandler) SearchProducts(c *gin.Context) {

	keyword := c.Query("q")

	var categoryID, brandID, locationID *uuid.UUID
	if cid := c.Query("category_id"); cid != "" {
		id, err := uuid.Parse(cid)
		if err == nil {
			categoryID = &id
		}
	}
	if bid := c.Query("brand_id"); bid != "" {
		id, err := uuid.Parse(bid)
		if err == nil {
			brandID = &id
		}
	}
	if lid := c.Query("location_id"); lid != "" {
		id, err := uuid.Parse(lid)
		if err == nil {
			locationID = &id
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	catIDPtr := uuidToStringPtr(categoryID)
	brandIDPtr := uuidToStringPtr(brandID)
	locIDPtr := uuidToStringPtr(locationID)

	// Assuming request.Pagination looks like:
	pageNumber, err := SafeIntToUint64(offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limitUint64, err := SafeIntToUint64(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pagination := request.Pagination{
		PageNumber: pageNumber,
		Count:      limitUint64,
	}

	products, err := h.productUseCase.SearchProducts(c, keyword, catIDPtr, brandIDPtr, locIDPtr, int(pagination.Count), int(pagination.PageNumber))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
	if pincode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter pincode is required"})
		return
	}

	radiusKmStr := c.DefaultQuery("radius_km", "10")
	radiusKm, err := strconv.ParseFloat(radiusKmStr, 64)
	if err != nil || radiusKm <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid radius_km"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	products, err := h.productUseCase.GetNearbyProductsByPincode(c, pincode, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})

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

	products, err := h.productUseCase.GetProductsByRadius(c, int(lat), int(lng), int(radius), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

	err := a.productUseCase.SaveDepartment(ctx, body.Name)
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
	departments, err := a.productUseCase.GetAllDepartments(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get departments", err, nil)
		return
	}

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
	departmentID, err := request.GetParamAsUint(ctx, "id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	department, err := a.productUseCase.GetDepartmentByID(ctx, departmentID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get department", err, nil)
		return
	}

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
	departmentID, err := request.GetParamAsUint(ctx, "department_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	categories, err := a.productUseCase.GetAllCategoriesByDepartmentID(ctx, departmentID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get categories", err, nil)
		return
	}

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
	categoryID, err := request.GetParamAsUint(ctx, "category_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	subCategories, err := a.productUseCase.GetAllSubCategoriesByCategoryID(ctx, categoryID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get sub-categories", err, nil)
		return
	}

	fmt.Println("SubCategories:", subCategories)

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
	subCategoryID, err := request.GetParamAsUint(ctx, "sub_category_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

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
	subCategoryID, err := request.GetParamAsUint(ctx, "sub_category_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

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
	attributeID, err := request.GetParamAsUint(ctx, "attribute_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

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
	attributeID, err := request.GetParamAsUint(ctx, "attribute_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

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
	attributeID, err := request.GetParamAsUint(ctx, "attribute_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

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
	optionID, err := request.GetParamAsUint(ctx, "option_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

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
	categoryID, err := request.GetParamAsUint(ctx, "category_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	var image request.CategoryImage
	if err := ctx.ShouldBindJSON(&image); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err = p.productUseCase.SaveCategoryImage(ctx, categoryID, image)
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
	categoryID, err := request.GetParamAsUint(ctx, "category_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

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
	imageID, err := request.GetParamAsUint(ctx, "image_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

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
	imageID, err := request.GetParamAsUint(ctx, "image_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	var image request.CategoryImage
	if err := ctx.ShouldBindJSON(&image); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err = p.productUseCase.UpdateCategoryImage(ctx, imageID, image)
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
	imageID, err := request.GetParamAsUint(ctx, "image_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	err = p.productUseCase.DeleteCategoryImage(ctx, imageID)
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
func (p *ProductHandler) GetProductItemByID(ctx *gin.Context) {
	productItemID, err := request.GetParamAsUint(ctx, "product_item_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	productItem, err := p.productUseCase.GetProductItemByID(ctx, productItemID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get product item", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully get product item", productItem)
}

func (p *ProductHandler) IncrementProductItemViewCount(ctx *gin.Context) {
	productItemID, err := request.GetParamAsUint(ctx, "product_item_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	tokenString := ctx.GetHeader("Authorization")
	adminId := p.tokenService.DecodeTokenData(tokenString)

	err = p.productUseCase.IncrementProductItemViewCount(ctx, productItemID, adminId)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to increment view count", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully incremented view count", nil)
}

func (p *ProductHandler) GetProductItemViewCount(ctx *gin.Context) {
	productItemID, err := request.GetParamAsUint(ctx, "product_item_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

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
