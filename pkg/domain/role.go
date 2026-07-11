package domain

import "time"

// PermissionKey enumerates the fixed set of gate-able admin-portal
// capabilities. The set of *keys* is fixed — each one corresponds to a real
// feature area / API surface — but which *roles* carry which keys is fully
// admin-managed via the Role/RolePermission tables below. Keys intentionally
// match the admin-portal's existing Permissions interface 1:1 so the
// frontend's nav-gating logic needs no redesign, only a dynamic data source.
type PermissionKey string

const (
	PermCanManageCatalog     PermissionKey = "canManageCatalog"
	PermCanManageMarketing   PermissionKey = "canManageMarketing"
	PermCanManageOrders      PermissionKey = "canManageOrders"
	PermCanManageUsers       PermissionKey = "canManageUsers"
	PermCanManageSettings    PermissionKey = "canManageSettings"
	PermCanSendNotifications PermissionKey = "canSendNotifications"
	PermCanManageAdmins      PermissionKey = "canManageAdmins"
	PermCanBlockUsers        PermissionKey = "canBlockUsers"
)

func (p PermissionKey) IsValid() bool {
	switch p {
	case PermCanManageCatalog, PermCanManageMarketing, PermCanManageOrders,
		PermCanManageUsers, PermCanManageSettings, PermCanSendNotifications,
		PermCanManageAdmins, PermCanBlockUsers:
		return true
	}
	return false
}

// AllPermissionKeys is the canonical, ordered catalog of every permission
// key — used to validate role-permission payloads and to give the
// admin-portal a fixed list to render checkboxes from.
func AllPermissionKeys() []PermissionKey {
	return []PermissionKey{
		PermCanManageCatalog, PermCanManageMarketing, PermCanManageOrders,
		PermCanManageUsers, PermCanManageSettings, PermCanSendNotifications,
		PermCanManageAdmins, PermCanBlockUsers,
	}
}

// RoleNameSuperAdmin and RoleNameSeller are reserved role names with special
// handling: super_admin implicitly carries every permission (never stored/
// edited as rows — see RoleUseCase.GetPermissionsForRole) and can't be
// deleted or renamed; seller is a customer-facing account type, never
// offered as an assignable admin-portal role.
const (
	RoleNameSuperAdmin = "super_admin"
	RoleNameSeller     = "seller"
)

// Role is an admin-managed named permission set assigned to platform users
// (Admin.Role references Role.Name by string, not by ID, to keep the
// existing admins.role column and every AdminRole-typed comparison in the
// codebase working unchanged). The roles that existed before this feature —
// super_admin, support_staff, catalog_manager, marketing_manager — are
// seeded with IsSystem=true so they can't be deleted and existing admins'
// access is preserved; their permissions (except super_admin's) can still
// be edited like any custom role.
type Role struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	Name      string    `json:"name" gorm:"size:50;uniqueIndex;not null" binding:"required"`
	Label     string    `json:"label" gorm:"size:100;not null" binding:"required"`
	IsSystem  bool      `json:"is_system" gorm:"not null;default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Permissions is populated by the repository from role_permissions; not
	// a persisted column on this table.
	Permissions []PermissionKey `json:"permissions" gorm:"-"`
}

// RolePermission is a single granted permission row for a role.
type RolePermission struct {
	ID            string        `json:"id" gorm:"primaryKey;type:varchar(32)"`
	RoleID        string        `json:"role_id" gorm:"type:varchar(32);index;not null"`
	PermissionKey PermissionKey `json:"permission_key" gorm:"size:50;not null"`
}
