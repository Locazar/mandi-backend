package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// CreateDepartment inserts a new department with image_url, sort_order, and is_active.
func (c *productDatabase) CreateDepartment(ctx context.Context, name, imageURL string, sortOrder int, isActive bool) error {
	query := `INSERT INTO departments (id, name, image_url, sort_order, is_active)
	          VALUES ($1, $2, $3, $4, $5)`
	return c.DB.Exec(query,
		domain.NewID(domain.PrefixDepartment),
		name,
		imageURL,
		sortOrder,
		isActive,
	).Error
}

// UpdateDepartment updates an existing department. If imageURL is empty, the existing image_url is preserved.
func (c *productDatabase) UpdateDepartment(ctx context.Context, departmentID, name, imageURL string, sortOrder int, isActive bool) error {
	if imageURL != "" {
		query := `UPDATE departments
		          SET name = $1, image_url = $2, sort_order = $3, is_active = $4
		          WHERE id = $5`
		return c.DB.Exec(query, name, imageURL, sortOrder, isActive, departmentID).Error
	}
	query := `UPDATE departments
	          SET name = $1, sort_order = $2, is_active = $3
	          WHERE id = $4`
	return c.DB.Exec(query, name, sortOrder, isActive, departmentID).Error
}

func (c *productDatabase) DeleteDepartment(ctx context.Context, departmentID string) error {
	return c.DB.Exec(`DELETE FROM departments WHERE id = $1`, departmentID).Error
}
