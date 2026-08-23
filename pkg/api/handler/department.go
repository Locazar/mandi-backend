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
//	@Description	Accepts multipart/form-data with department_name, sort_order, is_active, and optional photo and icon files.
//	@Tags			Admin Products
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			department_name	formData	string	true	"Department name"
//	@Param			sort_order		formData	int		false	"Sort order"
//	@Param			is_active		formData	bool	false	"Is active"
//	@Param			photo			formData	file	false	"Department card image"
//	@Param			icon			formData	file	false	"Department icon"
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

	imageURL, err := a.saveDepartmentUpload(ctx, "photo", "departments")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload image", err, nil)
		return
	}
	iconURL, err := a.saveDepartmentUpload(ctx, "icon", "departments/icons")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload icon", err, nil)
		return
	}

	if err := a.productUseCase.CreateDepartment(ctx, name, imageURL, iconURL, sortOrder, isActive); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create department", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Successfully created department", nil)
}

// UpdateDepartment godoc
//
//	@Summary		Update an existing department
//	@Description	Accepts multipart/form-data with department_name, sort_order, is_active, and optional photo and icon files.
//	@Tags			Admin Products
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			department_id	path		string	true	"Department ID"
//	@Param			department_name	formData	string	true	"Department name"
//	@Param			sort_order		formData	int		false	"Sort order"
//	@Param			is_active		formData	bool	false	"Is active"
//	@Param			photo			formData	file	false	"Department card image"
//	@Param			icon			formData	file	false	"Department icon"
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

	// Upload replacements only when the admin picked new files; an empty key
	// leaves the stored image_url / icon untouched.
	imageURL, err := a.saveDepartmentUpload(ctx, "photo", "departments")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload image", err, nil)
		return
	}
	iconURL, err := a.saveDepartmentUpload(ctx, "icon", "departments/icons")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload icon", err, nil)
		return
	}

	if err := a.productUseCase.UpdateDepartment(ctx, departmentID, name, imageURL, iconURL, sortOrder, isActive); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update department", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated department", nil)
}

// saveDepartmentUpload stores the file posted under formField, if any, and
// returns its object key. A missing field is not an error: it yields an empty
// key, which callers treat as "leave the stored value alone".
func (a *ProductHandler) saveDepartmentUpload(ctx *gin.Context, formField, namespace string) (string, error) {
	fileHeader, err := ctx.FormFile(formField)
	if err != nil || fileHeader == nil {
		return "", nil
	}
	return a.cloudService.SaveFile(ctx, fileHeader, cloud.SaveOptions{
		Namespace:  namespace,
		Visibility: cloud.VisibilityPublic,
	})
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
