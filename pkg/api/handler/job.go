package handler

import (
	"fmt"
	"net/http"
	"strings"

	services "github.com/rohit221990/mandi-backend/pkg/usecase"
	"github.com/rohit221990/mandi-backend/pkg/utils"

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
		jobs, err := h.Service.GetAllJobs(c.Request.Context())
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
		suggestions, _ := h.Service.GetJobSearchSuggestions(c, query)
		c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
	}
}

func (h *JobHandler) GetJobSearchFilters() gin.HandlerFunc {
	return func(c *gin.Context) {
		filters, _ := h.Service.GetJobSearchFilters(c)
		c.JSON(http.StatusOK, gin.H{"filters": filters})
	}
}

// postJobBody is the JSON body expected from the seller-app.
type postJobBody struct {
	ShopID        *string `json:"shop_id"`
	ShopName      string  `json:"shop_name"`
	Department    string  `json:"department"`
	PersonsNeeded int     `json:"persons_needed"`
	AgeMin        *int    `json:"age_min"`
	AgeMax        *int    `json:"age_max"`
	City          string  `json:"city"`
	State         string  `json:"state"`
	ShowPhone     bool    `json:"show_phone"`
	Phone         *string `json:"phone"`
}

// PostSellerJob handles POST /api/seller/jobs.
// Requires a valid seller JWT (AuthenticateUser middleware).
func (h *JobHandler) PostSellerJob() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body postJobBody
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if strings.TrimSpace(body.ShopName) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "shop_name is required"})
			return
		}
		if strings.TrimSpace(body.Department) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "department is required"})
			return
		}
		if body.PersonsNeeded < 1 {
			body.PersonsNeeded = 1
		}

		// Resolve the caller's shop_id from their JWT user ID when the client
		// doesn't send one (safer than trusting a client-supplied shop_id).
		shopID := body.ShopID
		if shopID == nil || strings.TrimSpace(*shopID) == "" {
			if uid := utils.GetUserIdFromContext(c); uid != "" {
				shopID = &uid
			}
		}

		req := services.CreateSellerJobPostingRequest{
			ShopID:        shopID,
			ShopName:      body.ShopName,
			Department:    body.Department,
			PersonsNeeded: body.PersonsNeeded,
			AgeMin:        body.AgeMin,
			AgeMax:        body.AgeMax,
			City:          body.City,
			State:         body.State,
			ShowPhone:     body.ShowPhone,
			Phone:         body.Phone,
		}

		posting, err := h.Service.CreateSellerJobPosting(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"success": true, "data": posting})
	}
}

// GetSellerJobs handles GET /api/seller/jobs.
// Returns active job postings for the authenticated seller's shop.
func (h *JobHandler) GetSellerJobs() gin.HandlerFunc {
	return func(c *gin.Context) {
		shopID := utils.GetUserIdFromContext(c)
		postings, err := h.Service.GetSellerJobPostings(c.Request.Context(), shopID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": postings})
	}
}
