package handler

import (
	"fmt"
	"net/http"

	services "github.com/rohit221990/mandi-backend/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type JobHandler struct {
	Service *services.JobService
}

func NewJobHandler(svc *services.JobService) *JobHandler {
	return &JobHandler{Service: svc}
}

func (h *JobHandler) GetAllJobs() gin.HandlerFunc {
	return func(c *gin.Context) {
		jobs, err := h.Service.GetAllJobs(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"jobs": jobs})
	}
}

func (h *JobHandler) ApplyToJob() gin.HandlerFunc {
	return func(c *gin.Context) {
		jobIDStr := c.Param("job_id")
		jobID, err := uuid.Parse(jobIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
			return
		}

		userIDStr := c.GetString("user_id") // assume middleware sets user_id
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		err = h.Service.ApplyToJob(c, userID, jobID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "application successful"})
	}
}

func (h *JobHandler) GetUserJobApplications() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetString("user_id")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		apps, err := h.Service.GetUserJobApplications(c, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"applications": apps})
	}
}

func (h *JobHandler) DeleteJobApplication() gin.HandlerFunc {
	return func(c *gin.Context) {
		applicationIDStr := c.Param("application_id")
		applicationID, err := uuid.Parse(applicationIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application_id"})
			return
		}

		err = h.Service.DeleteJobApplication(c, applicationID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "application deleted"})
	}
}

func (h *JobHandler) GetJobSearchSuggestions() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		suggestions, err := h.Service.GetJobSearchSuggestions(c, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
	}
}

func (h *JobHandler) GetJobSearchFilters() gin.HandlerFunc {
	return func(c *gin.Context) {
		filters, err := h.Service.GetJobSearchFilters(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"filters": filters})
	}
}
