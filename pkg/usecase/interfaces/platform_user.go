package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

//go:generate mockgen -destination=../../mock/mockusecase/platform_user_mock.go -package=mockusecase . PlatformUserUseCase
type PlatformUserUseCase interface {
	ListAdmins(ctx context.Context, pagination request.Pagination) ([]domain.Admin, error)
	CreateAdmin(ctx context.Context, body domain.Admin) (string, error)
	UpdateAdminRole(ctx context.Context, callerID, targetID string, role domain.AdminRole) error
	DeactivateAdmin(ctx context.Context, callerID, targetID string) error
}
