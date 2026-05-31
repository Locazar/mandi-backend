package usecase

import (
	"context"
	"fmt"

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
		Where("is_active = ?", true).
		Order("posted_date DESC").
		Find(&jobs)
	return jobs, result.Error
}

func (s *JobService) SearchJobs(ctx context.Context, keyword string, categoryID, locationID uuid.UUID, limit, offset int) ([]domain.Job, error) {
	query := s.DB.WithContext(ctx).Where("is_active = ?", true)

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
