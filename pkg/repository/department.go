package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// CreateDepartment inserts a new department with image_url, icon, sort_order, and is_active.
func (c *productDatabase) CreateDepartment(ctx context.Context, name, imageURL, iconURL string, sortOrder int, isActive bool) error {
	query := `INSERT INTO departments (id, name, image_url, icon, sort_order, is_active)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	return c.DB.Exec(query,
		domain.NewID(domain.PrefixDepartment),
		name,
		imageURL,
		iconURL,
		sortOrder,
		isActive,
	).Error
}

// UpdateDepartment updates an existing department. image_url and icon are only
// written when a replacement is supplied; an empty value preserves what is
// already stored, so saving the form without re-picking a file is not
// destructive.
func (c *productDatabase) UpdateDepartment(ctx context.Context, departmentID, name, imageURL, iconURL string, sortOrder int, isActive bool) error {
	query := `UPDATE departments
	          SET name = $1,
	              sort_order = $2,
	              is_active = $3,
	              image_url = COALESCE(NULLIF($4, ''), image_url),
	              icon = COALESCE(NULLIF($5, ''), icon)
	          WHERE id = $6`
	return c.DB.Exec(query, name, sortOrder, isActive, imageURL, iconURL, departmentID).Error
}

func (c *productDatabase) DeleteDepartment(ctx context.Context, departmentID string) error {
	return c.DB.Exec(`DELETE FROM departments WHERE id = $1`, departmentID).Error
}
