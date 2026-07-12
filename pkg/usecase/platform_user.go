package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	applogger "github.com/rohit221990/mandi-backend/pkg/logger"
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

// roleAssignable reports whether roleName is a real, currently-existing role
// (from the roles table) that isn't the reserved "seller" account type.
// Replaces the old closed AdminRole.IsValid() enum check now that roles are
// admin-managed instead of a fixed Go enum.
func (u *platformUserUseCase) roleAssignable(ctx context.Context, roleName string) (bool, error) {
	if roleName == "" || roleName == domain.RoleNameSeller {
		return false, nil
	}
	role, err := u.adminUC.GetRoleByName(ctx, roleName)
	if err != nil {
		return false, err
	}
	return role.ID != "", nil
}

func (u *platformUserUseCase) ListAdmins(ctx context.Context, callerID, roleFilter string, pagination request.Pagination) ([]domain.Admin, error) {
	caller, err := u.adminUC.GetAdminByID(ctx, callerID)
	if err != nil {
		return nil, domain.NotFoundError("caller admin")
	}
	if caller.Role != domain.AdminRoleSuperAdmin {
		return nil, domain.ForbiddenError("only super admins can list platform users")
	}
	// This endpoint is scoped to platform users, never sellers — reject an
	// explicit role=seller filter rather than silently ignoring it.
	if roleFilter == domain.RoleNameSeller {
		return nil, domain.ValidationError("role", "seller accounts are not platform users")
	}
	return u.repo.ListAdmins(ctx, roleFilter, pagination)
}

func (u *platformUserUseCase) CreateAdmin(ctx context.Context, callerID string, body domain.Admin) (string, error) {
	caller, err := u.adminUC.GetAdminByID(ctx, callerID)
	if err != nil {
		return "", domain.NotFoundError("caller admin")
	}
	if caller.Role != domain.AdminRoleSuperAdmin {
		return "", domain.ForbiddenError("only super admins can create platform users")
	}

	if body.Role != "" {
		assignable, err := u.roleAssignable(ctx, string(body.Role))
		if err != nil {
			return "", domain.InternalError("failed to validate role", err)
		}
		if !assignable {
			return "", domain.ValidationError("role", "unknown role; create it first under Roles")
		}
	}

	// Platform users log in with email or phone + password (not OTP), so
	// both must be set at creation time — unlike the self-service seller
	// SignUp path this usecase reuses, which allows either to be empty.
	if body.Email == "" {
		return "", domain.ValidationError("email", "email is required")
	}
	if body.Password == "" {
		return "", domain.ValidationError("password", "password is required")
	}
	existingByEmail, err := u.adminUC.FindAdminByEmail(ctx, body.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", domain.InternalError("failed to validate email", err)
	}
	if existingByEmail.ID != "" {
		return "", domain.ConflictError("email already registered", "an admin account with this email already exists")
	}

	if body.ReferralCouponID != "" {
		existing, err := u.adminUC.FindAdminByReferralCouponID(ctx, body.ReferralCouponID)
		if err != nil {
			return "", domain.InternalError("failed to validate referral coupon ID", err)
		}
		if existing.ID != "" {
			return "", domain.ConflictError("referral coupon ID already in use", "another platform user already holds this referral coupon ID")
		}
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

	assignable, err := u.roleAssignable(ctx, string(role))
	if err != nil {
		return domain.InternalError("failed to validate role", err)
	}
	if !assignable {
		return domain.ValidationError("role", "unknown or disallowed role; seller accounts are managed separately")
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
		assignable, err := u.roleAssignable(ctx, string(body.Role))
		if err != nil {
			return domain.InternalError("failed to validate role", err)
		}
		if !assignable {
			return domain.ValidationError("role", "unknown or disallowed role; seller accounts are managed separately")
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

	target, err := u.adminUC.GetAdminByID(ctx, targetID)
	if err != nil || target.ID == "" {
		return domain.NotFoundError("admin")
	}
	if target.Role == domain.AdminRoleSuperAdmin {
		return domain.ForbiddenError("super admin accounts can't be deleted")
	}

	// A Sales Executive with active shop attachments must be reassigned
	// first — deleting them out from under live referrals would sever
	// commission/reporting attribution with no way to trace it back.
	attachedShops, err := u.adminUC.CountShopsAttachedToPlatformUser(ctx, targetID)
	if err != nil {
		return domain.InternalError("failed to check attached shops", err)
	}
	if attachedShops > 0 {
		return domain.ConflictError(
			"platform user has attached shops",
			fmt.Sprintf("%d shop(s) are still attached to this user via referral; reassign or remove those referrals first", attachedShops),
		)
	}

	if err := u.repo.DeleteAdmin(ctx, targetID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.NotFoundError("admin")
		}
		return err
	}

	// Best-effort: kill any active session so an already-issued token can't
	// keep working against a now-deleted account. Deletion itself already
	// succeeded, so a failure here is logged-and-swallowed, not fatal.
	if err := u.adminUC.RevokeAdminSessions(ctx, targetID); err != nil {
		applogger.L().Warn("failed to revoke sessions after admin delete",
			zap.String("admin_id", targetID), zap.Error(err))
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
