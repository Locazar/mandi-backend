package usecase

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

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

func (u *platformUserUseCase) ListAdmins(ctx context.Context, callerID string, pagination request.Pagination) ([]domain.Admin, error) {
	caller, err := u.adminUC.GetAdminByID(ctx, callerID)
	if err != nil {
		return nil, domain.NotFoundError("caller admin")
	}
	if caller.Role != domain.AdminRoleSuperAdmin {
		return nil, domain.ForbiddenError("only super admins can list platform users")
	}
	return u.repo.ListAdmins(ctx, pagination)
}

func (u *platformUserUseCase) CreateAdmin(ctx context.Context, callerID string, body domain.Admin) (string, error) {
	caller, err := u.adminUC.GetAdminByID(ctx, callerID)
	if err != nil {
		return "", domain.NotFoundError("caller admin")
	}
	if caller.Role != domain.AdminRoleSuperAdmin {
		return "", domain.ForbiddenError("only super admins can create platform users")
	}
	otpID, err := u.adminUC.SignUp(ctx, body)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return "", domain.ConflictError("phone number already registered", "an admin account with this phone number already exists")
		}
		return "", err
	}
	return otpID, nil
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

	if err := u.repo.UpdateAdminRole(ctx, targetID, role); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.NotFoundError("admin")
		}
		return err
	}
	return nil
}

func (u *platformUserUseCase) UpdateAdmin(ctx context.Context, callerID, targetID string, body domain.Admin) error {
	caller, err := u.adminUC.GetAdminByID(ctx, callerID)
	if err != nil {
		return domain.NotFoundError("caller admin")
	}
	if caller.Role != domain.AdminRoleSuperAdmin {
		return domain.ForbiddenError("only super admins can update platform users")
	}

	updates := map[string]interface{}{}
	if body.FullName != "" {
		updates["full_name"] = body.FullName
	}
	if body.Email != "" {
		updates["email"] = body.Email
	}
	if body.Mobile != "" {
		updates["mobile"] = body.Mobile
	}
	if body.Role != "" {
		if callerID == targetID {
			return domain.ForbiddenError("you cannot change your own role")
		}
		if !body.Role.IsValid() || body.Role == domain.AdminRoleSeller {
			return domain.ValidationError("role", "invalid or disallowed role; seller accounts are managed separately")
		}
		updates["role"] = body.Role
	}
	if len(updates) == 0 {
		return domain.ValidationError("body", "no updatable fields provided")
	}

	if err := u.repo.UpdateAdminDetails(ctx, targetID, updates); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.NotFoundError("admin")
		}
		return err
	}
	return nil
}

func (u *platformUserUseCase) DeleteAdmin(ctx context.Context, callerID, targetID string) error {
	if callerID == targetID {
		return domain.ForbiddenError("you cannot delete your own account")
	}

	caller, err := u.adminUC.GetAdminByID(ctx, callerID)
	if err != nil {
		return domain.NotFoundError("caller admin")
	}
	if caller.Role != domain.AdminRoleSuperAdmin {
		return domain.ForbiddenError("only super admins can delete accounts")
	}

	if err := u.repo.DeleteAdmin(ctx, targetID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.NotFoundError("admin")
		}
		return err
	}
	return nil
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

	if err := u.repo.DeactivateAdmin(ctx, targetID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.NotFoundError("admin")
		}
		return err
	}
	return nil
}
