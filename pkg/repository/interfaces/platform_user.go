package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

//go:generate mockgen -destination=../../mock/mockrepo/platform_user_mock.go -package=mockrepo . PlatformUserRepository
type PlatformUserRepository interface {
	// ListAdmins returns platform users (never sellers). roleFilter, when
	// non-empty, restricts the result to that exact role; empty returns
	// every non-seller role.
	ListAdmins(ctx context.Context, roleFilter string, pagination request.Pagination) ([]domain.Admin, error)
	UpdateAdminRole(ctx context.Context, adminID string, role domain.AdminRole) error
	UpdateAdminDetails(ctx context.Context, adminID string, updates map[string]interface{}) error
	DeactivateAdmin(ctx context.Context, adminID string) error
	DeleteAdmin(ctx context.Context, adminID string) error
}
