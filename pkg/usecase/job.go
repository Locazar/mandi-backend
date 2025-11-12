package usecase

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type JobService struct {
	DB *pgxpool.Pool
}

func (s *JobService) GetJobSearchFilters(c *gin.Context) (any, any) {
	panic("unimplemented")
}

func (s *JobService) GetJobSearchSuggestions(c *gin.Context, query string) (any, any) {
	panic("unimplemented")
}

func (s *JobService) DeleteJobApplication(c *gin.Context, applicationID uuid.UUID) error {
	panic("unimplemented")
}

func (s *JobService) GetUserJobApplications(c *gin.Context, userID uuid.UUID) (any, error) {
	panic("unimplemented")
}

func (s *JobService) ApplyToJob(c *gin.Context, userID uuid.UUID, jobID uuid.UUID) error {
	panic("unimplemented")
}

func NewJobService(db *pgxpool.Pool) *JobService {
	return &JobService{DB: db}
}

func (s *JobService) GetAllJobs(ctx context.Context) ([]domain.Job, error) {
	query := `SELECT job_id, title, description, category_id, location_id, company, posted_date, expiry_date FROM jobs WHERE is_active = true ORDER BY posted_date DESC`
	rows, err := s.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var j domain.Job
		err := rows.Scan(&j.ID, &j.Title, &j.Description, &j.CategoryID, &j.LocationID, &j.Company, &j.PostedDate, &j.ExpiryDate)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

// Implement ApplyToJob, GetUserJobApplications, DeleteJobApplication etc. in a similar way

func (s *JobService) SearchJobs(ctx context.Context, keyword string, categoryID, locationID uuid.UUID, limit, offset int) ([]domain.Job, error) {
	baseQuery := `SELECT job_id, title, description, category_id, location_id, company, posted_date, expiry_date FROM jobs WHERE is_active = true`
	params := []interface{}{}
	paramIdx := 1

	if keyword != "" {
		baseQuery += fmt.Sprintf(" AND to_tsvector('english', title || ' ' || description) @@ plainto_tsquery('english', $%d)", paramIdx)
		params = append(params, keyword)
		paramIdx++
	}
	if categoryID != uuid.Nil {
		baseQuery += fmt.Sprintf(" AND category_id = $%d", paramIdx)
		params = append(params, categoryID)
		paramIdx++
	}
	if locationID != uuid.Nil {
		baseQuery += fmt.Sprintf(" AND location_id = $%d", paramIdx)
		params = append(params, locationID)
		paramIdx++
	}

	baseQuery += fmt.Sprintf(" ORDER BY posted_date DESC LIMIT $%d OFFSET $%d", paramIdx, paramIdx+1)
	params = append(params, limit, offset)

	rows, err := s.DB.Query(ctx, baseQuery, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var j domain.Job
		err := rows.Scan(&j.ID, &j.Title, &j.Description, &j.CategoryID, &j.LocationID, &j.Company, &j.PostedDate, &j.ExpiryDate)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

// Similarly implement other service methods.
