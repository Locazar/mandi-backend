package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type ProductHandler struct {
	productUseCase usecaseInterface.ProductUseCase
}

func NewProductHandler(productUsecase usecaseInterface.ProductUseCase) interfaces.ProductHandler {
	return &ProductHandler{
		productUseCase: productUsecase,
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

	var body request.Category

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err := p.productUseCase.SaveCategory(ctx, body.Name)

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

	var body request.SubCategory
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err := p.productUseCase.SaveSubCategory(ctx, body)

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
//	@Summary		Add a new product (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to add a new product
//	@ID				SaveProduct
//	@Tags			Admin Products
//	@Produce		json
//	@Param			name		formData	string				true	"Product Name"
//	@Param			description	formData	string				true	"Product Description"
//	@Param			category_id	formData	int					true	"Category Id"
//	@Param			brand_id	formData	int					true	"Brand Id"
//	@Param			price		formData	int					true	"Product Price"
//	@Param			image		formData	file				true	"Product Description"
//	@Success		200			{object}	response.Response{}	"successfully product added"
//	@Router			/admin/products [post]
//	@Failure		400	{object}	response.Response{}	"invalid input"
//	@Failure		409	{object}	response.Response{}	"Product name already exist"
func (p *ProductHandler) SaveProduct(ctx *gin.Context) {

	name, err1 := request.GetFormValuesAsString(ctx, "name")
	description, err2 := request.GetFormValuesAsString(ctx, "description")
	categoryID, err3 := request.GetFormValuesAsUint(ctx, "category_id")
	price, err4 := request.GetFormValuesAsUint(ctx, "price")
	brandID, err5 := request.GetFormValuesAsUint(ctx, "brand_id")

	fileHeader, err6 := ctx.FormFile("image")

	err := errors.Join(err1, err2, err3, err4, err5, err6)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindFormValueMessage, err, nil)
		return
	}

	product := request.Product{
		Name:            name,
		Description:     description,
		CategoryID:      categoryID,
		BrandID:         brandID,
		Price:           price,
		ImageFileHeader: fileHeader,
	}

	err = p.productUseCase.SaveProduct(ctx, product)

	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrProductAlreadyExist) {
			statusCode = http.StatusConflict
		}
		response.ErrorResponse(ctx, statusCode, "Failed to add product", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully product added")
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

	productID, err := request.GetParamAsUint(ctx, "product_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
	}

	fmt.Printf("Product ID: %d\n", productID)

	price, err1 := request.GetFormValuesAsUint(ctx, "price")
	qtyInStock, err2 := request.GetFormValuesAsUint(ctx, "qty_in_stock")
	variationOptionIDS, err3 := request.GetArrayFormValueAsUint(ctx, "variation_option_ids")
	imageFileHeaders, err4 := request.GetArrayOfFromFiles(ctx, "images")

	err = errors.Join(err1, err2, err3, err4)

	fmt.Printf("Variation Option IDs: %+v\n", err)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindFormValueMessage, err, nil)
		return
	}

	productItem := request.ProductItem{
		Price:              price,
		VariationOptionIDs: variationOptionIDS,
		QtyInStock:         qtyInStock,
		ImageFileHeaders:   imageFileHeaders,
	}

	fmt.Println(productItem, productID)

	err = p.productUseCase.SaveProductItem(ctx, productID, productItem)

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

		productID, err := request.GetParamAsUint(ctx, "product_id")
		if err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		}

		productItems, err := p.productUseCase.FindAllProductItems(ctx, productID)

		if err != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get all product items", err, nil)
			return
		}

		// check the product have productItem exist or not
		if len(productItems) == 0 {
			response.SuccessResponse(ctx, http.StatusOK, "No product items found")
			return
		}

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

	limitUint64, err := SafeIntToUint64(limit)

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
