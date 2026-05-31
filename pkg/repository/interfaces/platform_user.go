package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

//go:generate mockgen -destination=../../mock/mockrepo/platform_user_mock.go -package=mockrepo . PlatformUserRepository
type PlatformUserRepository interface {
	ListAdmins(ctx context.Context, pagination request.Pagination) ([]domain.Admin, error)
	GetAdminByID(ctx context.Context, adminID string) (domain.Admin, error)
	UpdateAdminRole(ctx context.Context, adminID string, role domain.AdminRole) error
	DeactivateAdmin(ctx context.Context, adminID string) error
}
