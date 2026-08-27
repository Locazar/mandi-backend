package usecase

import "context"

// CreateDepartment inserts a new department row. imageURL and iconURL are the
// object keys of the uploaded card image and icon (either may be empty).
func (c *productUseCase) CreateDepartment(ctx context.Context, name, imageURL, iconURL string, sortOrder int, isActive bool) error {
	return c.productRepo.CreateDepartment(ctx, name, imageURL, iconURL, sortOrder, isActive)
}

// UpdateDepartment updates name, sort_order, is_active, and optionally image_url
// and icon for a department. An empty imageURL or iconURL preserves the value
// already in the database.
func (c *productUseCase) UpdateDepartment(ctx context.Context, departmentID, name, imageURL, iconURL string, sortOrder int, isActive bool) error {
	return c.productRepo.UpdateDepartment(ctx, departmentID, name, imageURL, iconURL, sortOrder, isActive)
}

func (c *productUseCase) DeleteDepartment(ctx context.Context, departmentID string) error {
	return c.productRepo.DeleteDepartment(ctx, departmentID)
}
