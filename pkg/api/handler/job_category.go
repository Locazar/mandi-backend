package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Local interface describing the methods the handler expects.
// This allows the handler to work with any implementation provided
// by the usecase layer without depending on a concrete type name.
type JobCategoryService interface {
	GetAllJobCategories(*gin.Context) (interface{}, error)
	GetJobsByCategory(*gin.Context, uuid.UUID, int, int) (interface{}, error)
	GetJobSubCategories(*gin.Context, uuid.UUID) (interface{}, error)
	GetJobsBySubCategory(*gin.Context, uuid.UUID, int, int) (interface{}, error)
	GetJobCategoryFilters(*gin.Context) (interface{}, error)
}

type JobCategoryHandler struct {
	Service JobCategoryService
}

func NewJobCategoryHandler(svc JobCategoryService) *JobCategoryHandler {
	return &JobCategoryHandler{Service: svc}
}

func (h *JobCategoryHandler) GetAllJobCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, err := h.Service.GetAllJobCategories(c)
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

		jobs, err := h.Service.GetJobsByCategory(c, categoryID, limit, offset)
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

		subcategories, err := h.Service.GetJobSubCategories(c, categoryID)
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

		jobs, err := h.Service.GetJobsBySubCategory(c, subcategoryID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"jobs": jobs})
	}
}

func (h *JobCategoryHandler) GetJobCategoryFilters() gin.HandlerFunc {
	return func(c *gin.Context) {
		filters, err := h.Service.GetJobCategoryFilters(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"filters": filters})
	}
}
