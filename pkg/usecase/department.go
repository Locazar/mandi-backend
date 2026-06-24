package usecase

import "context"

// CreateDepartment inserts a new department row. imageURL is the S3 object key (may be empty).
func (c *productUseCase) CreateDepartment(ctx context.Context, name, imageURL string, sortOrder int, isActive bool) error {
	return c.productRepo.CreateDepartment(ctx, name, imageURL, sortOrder, isActive)
}

// UpdateDepartment updates name, sort_order, is_active, and optionally image_url for a department.
// If imageURL is empty the existing image_url in the database is preserved.
func (c *productUseCase) UpdateDepartment(ctx context.Context, departmentID, name, imageURL string, sortOrder int, isActive bool) error {
	return c.productRepo.UpdateDepartment(ctx, departmentID, name, imageURL, sortOrder, isActive)
}
