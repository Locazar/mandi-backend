package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type JobCategoryService struct {
	DB *pgxpool.Pool
}

func NewJobCategoryService(db *pgxpool.Pool) *JobCategoryService {
	return &JobCategoryService{DB: db}
}

func (s *JobCategoryService) GetAllJobCategories(ctx context.Context) ([]domain.JobCategory, error) {
	query := `SELECT category_id, name, parent_id FROM job_categories ORDER BY name`
	rows, err := s.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.JobCategory
	for rows.Next() {
		var c domain.JobCategory
		if err := rows.Scan(&c.CategoryID, &c.Name, &c.ParentID); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (s *JobCategoryService) GetJobsByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset int) ([]domain.Job, error) {
	query := `SELECT job_id, title, description, category_id, location_id, company, posted_date, expiry_date 
              FROM jobs WHERE category_id = $1 ORDER BY posted_date DESC LIMIT $2 OFFSET $3`
	rows, err := s.DB.Query(ctx, query, categoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var j domain.Job
		if err := rows.Scan(&j.ID, &j.Title, &j.Description, &j.CategoryID, &j.LocationID, &j.Company, &j.PostedDate, &j.ExpiryDate); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

// Similarly implement GetJobSubCategories, GetJobsBySubCategory, GetJobCategoryFilters, GetJobCategoryLocations, SearchJobsInCategory
