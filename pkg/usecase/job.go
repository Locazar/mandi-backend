package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"gorm.io/gorm"
)

type JobService struct {
	DB *gorm.DB
}

func NewJobService(db *gorm.DB) *JobService {
	return &JobService{DB: db}
}

func (s *JobService) GetAllJobs(ctx context.Context) ([]domain.Job, error) {
	var jobs []domain.Job
	result := s.DB.WithContext(ctx).
		Order("posted_date DESC").
		Find(&jobs)
	return jobs, result.Error
}

func (s *JobService) SearchJobs(ctx context.Context, keyword string, categoryID, locationID uuid.UUID, limit, offset int) ([]domain.Job, error) {
	query := s.DB.WithContext(ctx).Model(&domain.Job{})

	if keyword != "" {
		query = query.Where("to_tsvector('english', title || ' ' || description) @@ plainto_tsquery('english', ?)", keyword)
	}
	if categoryID != uuid.Nil {
		query = query.Where("category_id = ?", categoryID.String())
	}
	if locationID != uuid.Nil {
		query = query.Where("location_id = ?", locationID.String())
	}

	var jobs []domain.Job
	result := query.Order("posted_date DESC").Limit(limit).Offset(offset).Find(&jobs)
	return jobs, result.Error
}

func (s *JobService) ApplyToJob(c *gin.Context, userID uuid.UUID, jobID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (s *JobService) GetUserJobApplications(c *gin.Context, userID uuid.UUID) (any, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *JobService) DeleteJobApplication(c *gin.Context, applicationID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (s *JobService) GetJobSearchSuggestions(c *gin.Context, query string) (any, any) {
	return nil, nil
}

func (s *JobService) GetJobSearchFilters(c *gin.Context) (any, any) {
	return nil, nil
}

// CreateSellerJobPosting persists a new job requirement posted by a seller.
func (s *JobService) CreateSellerJobPosting(ctx context.Context, req CreateSellerJobPostingRequest) (*domain.SellerJobPosting, error) {
	// Resolve the correct shop_details.id.
	// The client sends its locally-cached shop_id when available.
	// If not, we look it up via the JWT admin_id (admins.id → shop_details.admin_id).
	shopID := req.ShopID
	if (shopID == nil || strings.TrimSpace(*shopID) == "") && req.AdminID != "" {
		var shopRow struct{ ID string }
		if err := s.DB.WithContext(ctx).Table("shop_details").Select("id").
			Where("admin_id = ? AND deleted_at IS NULL", req.AdminID).
			First(&shopRow).Error; err == nil && shopRow.ID != "" {
			shopID = &shopRow.ID
		}
	}

	now := time.Now()
	id := strings.ReplaceAll(uuid.New().String(), "-", "")
	posting := &domain.SellerJobPosting{
		ID:            id,
		ShopID:        shopID,
		ShopName:      strings.TrimSpace(req.ShopName),
		ShopType:      strings.TrimSpace(req.ShopType),
		Department:    strings.TrimSpace(req.Department),
		PersonsNeeded: req.PersonsNeeded,
		AgeMin:        req.AgeMin,
		AgeMax:        req.AgeMax,
		City:          strings.TrimSpace(req.City),
		State:         strings.TrimSpace(req.State),
		ShowPhone:     req.ShowPhone,
		Phone:         req.Phone,
		IsActive:      false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if posting.PersonsNeeded < 1 {
		posting.PersonsNeeded = 1
	}
	if err := s.DB.WithContext(ctx).Create(posting).Error; err != nil {
		return nil, fmt.Errorf("create seller job posting: %w", err)
	}
	return posting, nil
}

// GetSellerJobPostings returns job postings, newest first.
// shopID filters to a specific seller's posts (pass "" for all).
// onlyActive when true restricts results to is_active = true (customer-app view).
func (s *JobService) GetSellerJobPostings(ctx context.Context, shopID string, onlyActive bool) ([]domain.SellerJobPosting, error) {
	var postings []domain.SellerJobPosting
	q := s.DB.WithContext(ctx).Order("created_at DESC")
	if onlyActive {
		q = q.Where("is_active = true")
	}
	if shopID != "" {
		q = q.Where("shop_id = ?", shopID)
	}
	return postings, q.Find(&postings).Error
}

// GetSellerJobPostingsByAdmin returns a seller's own postings looked up via their
// admin_id (JWT userId). It resolves shop_details.id first so the filter is correct.
func (s *JobService) GetSellerJobPostingsByAdmin(ctx context.Context, adminID string, onlyActive bool) ([]domain.SellerJobPosting, error) {
	shopID := ""
	if adminID != "" {
		var shopRow struct{ ID string }
		if err := s.DB.WithContext(ctx).Table("shop_details").Select("id").
			Where("admin_id = ? AND deleted_at IS NULL", adminID).
			First(&shopRow).Error; err == nil {
			shopID = shopRow.ID
		}
	}
	return s.GetSellerJobPostings(ctx, shopID, onlyActive)
}

// GetAllSellerJobPostings returns all seller job postings for admin review, newest first.
func (s *JobService) GetAllSellerJobPostings(ctx context.Context) ([]domain.SellerJobPosting, error) {
	var postings []domain.SellerJobPosting
	return postings, s.DB.WithContext(ctx).Order("created_at DESC").Find(&postings).Error
}

// SetSellerJobPostingActive enables or disables a seller job posting's visibility
// on the customer app. Returns the updated record.
func (s *JobService) SetSellerJobPostingActive(ctx context.Context, id string, active bool) (*domain.SellerJobPosting, error) {
	var posting domain.SellerJobPosting
	if err := s.DB.WithContext(ctx).First(&posting, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("seller job posting not found: %w", err)
	}
	posting.IsActive = active
	posting.UpdatedAt = time.Now()
	if err := s.DB.WithContext(ctx).Save(&posting).Error; err != nil {
		return nil, fmt.Errorf("update seller job posting: %w", err)
	}
	return &posting, nil
}

// CreateSellerJobPostingRequest holds the validated fields from the POST body.
type CreateSellerJobPostingRequest struct {
	AdminID       string  // JWT user_id (admins.id) — used to look up shop_details.id when ShopID is blank
	ShopID        *string // shop_details.id sent by client (preferred over AdminID lookup)
	ShopName      string
	ShopType      string
	Department    string
	PersonsNeeded int
	AgeMin        *int
	AgeMax        *int
	City          string
	State         string
	ShowPhone     bool
	Phone         *string
}
