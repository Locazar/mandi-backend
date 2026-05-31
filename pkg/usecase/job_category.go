package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"gorm.io/gorm"
)

type JobCategoryService struct {
	DB *gorm.DB
}

func NewJobCategoryService(db *gorm.DB) *JobCategoryService {
	return &JobCategoryService{DB: db}
}

func (s *JobCategoryService) GetAllJobCategories(ctx context.Context) ([]domain.JobCategory, error) {
	var categories []domain.JobCategory
	result := s.DB.WithContext(ctx).Order("name").Find(&categories)
	return categories, result.Error
}

func (s *JobCategoryService) GetJobsByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset int) ([]domain.Job, error) {
	var jobs []domain.Job
	result := s.DB.WithContext(ctx).
		Where("category_id = ?", categoryID.String()).
		Order("posted_date DESC").
		Limit(limit).Offset(offset).
		Find(&jobs)
	return jobs, result.Error
}

func (s *JobCategoryService) GetJobSubCategories(ctx context.Context, categoryID uuid.UUID) ([]domain.JobSubCategory, error) {
	var subs []domain.JobSubCategory
	result := s.DB.WithContext(ctx).Where("category_id = ?", categoryID.String()).Find(&subs)
	return subs, result.Error
}

func (s *JobCategoryService) GetJobsBySubCategory(ctx context.Context, subCategoryID uuid.UUID, limit, offset int) ([]domain.Job, error) {
	// Jobs don't have a direct sub_category_id column in the domain model;
	// return empty slice until the schema is extended.
	return []domain.Job{}, nil
}

func (s *JobCategoryService) GetJobCategoryFilters(ctx context.Context) ([]domain.JobFilter, error) {
	var filters []domain.JobFilter
	result := s.DB.WithContext(ctx).Find(&filters)
	return filters, result.Error
}
