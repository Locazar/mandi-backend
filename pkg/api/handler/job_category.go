package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// JobCategoryService describes the methods the handler expects from the usecase layer.
type JobCategoryService interface {
	GetAllJobCategories(context.Context) ([]domain.JobCategory, error)
	GetJobsByCategory(context.Context, uuid.UUID, int, int) ([]domain.Job, error)
	GetJobSubCategories(context.Context, uuid.UUID) ([]domain.JobSubCategory, error)
	GetJobsBySubCategory(context.Context, uuid.UUID, int, int) ([]domain.Job, error)
	GetJobCategoryFilters(context.Context) ([]domain.JobFilter, error)
}

type JobCategoryHandler struct {
	Service JobCategoryService
}

func NewJobCategoryHandler(svc JobCategoryService) *JobCategoryHandler {
	return &JobCategoryHandler{Service: svc}
}

func (h *JobCategoryHandler) GetAllJobCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, err := h.Service.GetAllJobCategories(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"categories": categories})
	}
}

func (h *JobCategoryHandler) GetJobsByCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		cid := c.Param("category_id")
		categoryID, err := uuid.Parse(cid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		jobs, err := h.Service.GetJobsByCategory(c.Request.Context(), categoryID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"jobs": jobs})
	}
}

func (h *JobCategoryHandler) GetJobSubCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		cid := c.Param("category_id")
		categoryID, err := uuid.Parse(cid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}

		subcategories, err := h.Service.GetJobSubCategories(c.Request.Context(), categoryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"subcategories": subcategories})
	}
}

func (h *JobCategoryHandler) GetJobsBySubCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("subcategory_id")
		subcategoryID, err := uuid.Parse(sid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subcategory_id"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		jobs, err := h.Service.GetJobsBySubCategory(c.Request.Context(), subcategoryID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"jobs": jobs})
	}
}

func (h *JobCategoryHandler) GetJobCategoryFilters() gin.HandlerFunc {
	return func(c *gin.Context) {
		filters, err := h.Service.GetJobCategoryFilters(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"filters": filters})
	}
}
