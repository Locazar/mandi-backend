package usecase

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterfaces "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	ucInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type platformUserUseCase struct {
	repo    repoInterfaces.PlatformUserRepository
	adminUC ucInterfaces.AdminUseCase
}

func NewPlatformUserUseCase(repo repoInterfaces.PlatformUserRepository, adminUC ucInterfaces.AdminUseCase) ucInterfaces.PlatformUserUseCase {
	return &platformUserUseCase{repo: repo, adminUC: adminUC}
}

func (u *platformUserUseCase) ListAdmins(ctx context.Context, pagination request.Pagination) ([]domain.Admin, error) {
	return u.repo.ListAdmins(ctx, pagination)
}

func (u *platformUserUseCase) CreateAdmin(ctx context.Context, body domain.Admin) (string, error) {
	return u.adminUC.SignUp(ctx, body)
}

func (u *platformUserUseCase) UpdateAdminRole(ctx context.Context, callerID, targetID string, role domain.AdminRole) error {
	if callerID == targetID {
		return domain.ForbiddenError("you cannot change your own role")
	}

	caller, err := u.adminUC.GetAdminByID(ctx, callerID)
	if err != nil {
		return domain.NotFoundError("caller admin")
	}

	if caller.Role != domain.AdminRoleSuperAdmin {
		return domain.ForbiddenError("only super admins can update roles")
	}

	if !role.IsValid() || role == domain.AdminRoleSeller {
		return domain.ValidationError("role", "invalid or disallowed role; seller accounts are managed separately")
	}

	return u.repo.UpdateAdminRole(ctx, targetID, role)
}

func (u *platformUserUseCase) DeactivateAdmin(ctx context.Context, callerID, targetID string) error {
	if callerID == targetID {
		return domain.ForbiddenError("you cannot deactivate your own account")
	}

	caller, err := u.adminUC.GetAdminByID(ctx, callerID)
	if err != nil {
		return domain.NotFoundError("caller admin")
	}

	if caller.Role != domain.AdminRoleSuperAdmin {
		return domain.ForbiddenError("only super admins can deactivate accounts")
	}

	return u.repo.DeactivateAdmin(ctx, targetID)
}
