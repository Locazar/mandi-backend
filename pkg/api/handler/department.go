package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
)

// CreateDepartment godoc
//
//	@Summary		Create a new department with optional image
//	@Description	Accepts multipart/form-data with department_name, sort_order, is_active, and an optional photo file.
//	@Tags			Admin Products
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			department_name	formData	string	true	"Department name"
//	@Param			sort_order		formData	int		false	"Sort order"
//	@Param			is_active		formData	bool	false	"Is active"
//	@Param			photo			formData	file	false	"Department image"
//	@Success		201				{object}	response.Response{}
//	@Failure		400				{object}	response.Response{}
//	@Failure		500				{object}	response.Response{}
//	@Router			/admin/departments [post]
func (a *ProductHandler) CreateDepartment(ctx *gin.Context) {
	name := ctx.PostForm("department_name")
	if len(name) < 3 || len(name) > 25 {
		response.ErrorResponse(ctx, http.StatusBadRequest, "department_name must be between 3 and 25 characters", nil, nil)
		return
	}

	sortOrder, _ := strconv.Atoi(ctx.PostForm("sort_order"))
	isActive := ctx.PostForm("is_active") != "false"

	var imageURL string
	fileHeader, err := ctx.FormFile("photo")
	if err == nil && fileHeader != nil {
		objectKey, uploadErr := a.cloudService.SaveFile(ctx, fileHeader, cloud.SaveOptions{
			Namespace:  "departments",
			Visibility: cloud.VisibilityPublic,
		})
		if uploadErr != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload image", uploadErr, nil)
			return
		}
		imageURL = objectKey
	}

	if err := a.productUseCase.CreateDepartment(ctx, name, imageURL, sortOrder, isActive); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create department", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully created department", nil)
}

// UpdateDepartment godoc
//
//	@Summary		Update an existing department
//	@Description	Accepts multipart/form-data with department_name, sort_order, is_active, and an optional photo file.
//	@Tags			Admin Products
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			department_id	path		string	true	"Department ID"
//	@Param			department_name	formData	string	true	"Department name"
//	@Param			sort_order		formData	int		false	"Sort order"
//	@Param			is_active		formData	bool	false	"Is active"
//	@Param			photo			formData	file	false	"Department image"
//	@Success		200				{object}	response.Response{}
//	@Failure		400				{object}	response.Response{}
//	@Failure		500				{object}	response.Response{}
//	@Router			/admin/departments/{department_id} [put]
func (a *ProductHandler) UpdateDepartment(ctx *gin.Context) {
	departmentID := ctx.Param("department_id")
	if departmentID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "department_id is required", nil, nil)
		return
	}

	name := ctx.PostForm("department_name")
	if len(name) < 3 || len(name) > 25 {
		response.ErrorResponse(ctx, http.StatusBadRequest, "department_name must be between 3 and 25 characters", nil, nil)
		return
	}

	sortOrder, _ := strconv.Atoi(ctx.PostForm("sort_order"))
	isActive := ctx.PostForm("is_active") != "false"

	// Upload new photo if provided; otherwise imageURL stays empty and the existing one is preserved in DB.
	var imageURL string
	fileHeader, err := ctx.FormFile("photo")
	if err == nil && fileHeader != nil {
		objectKey, uploadErr := a.cloudService.SaveFile(ctx, fileHeader, cloud.SaveOptions{
			Namespace:  "departments",
			Visibility: cloud.VisibilityPublic,
		})
		if uploadErr != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload image", uploadErr, nil)
			return
		}
		imageURL = objectKey
	}

	if err := a.productUseCase.UpdateDepartment(ctx, departmentID, name, imageURL, sortOrder, isActive); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update department", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated department", nil)
}

func (a *ProductHandler) DeleteDepartment(ctx *gin.Context) {
	departmentID := ctx.Param("department_id")
	if departmentID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "department_id is required", nil, nil)
		return
	}
	if err := a.productUseCase.DeleteDepartment(ctx, departmentID); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete department", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully deleted department", nil)
}
