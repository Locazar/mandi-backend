package domain

import "time"

// PermissionKey enumerates the fixed set of gate-able admin-portal
// capabilities. The set of *keys* is fixed — each one corresponds to a real
// feature area / API surface — but which *roles* carry which keys (and at
// what AccessLevel) is fully admin-managed via the Role/RolePermission
// tables below. Keys intentionally match the admin-portal's existing
// Permissions interface 1:1 so the frontend's nav-gating logic needs no
// redesign, only a dynamic data source.
type PermissionKey string

const (
	PermCanManageCatalog      PermissionKey = "canManageCatalog"
	PermCanManageMarketing    PermissionKey = "canManageMarketing"
	PermCanManageOrders       PermissionKey = "canManageOrders"
	PermCanManageUsers        PermissionKey = "canManageUsers"
	PermCanManageSettings     PermissionKey = "canManageSettings"
	PermCanSendNotifications  PermissionKey = "canSendNotifications"
	PermCanManageAdmins       PermissionKey = "canManageAdmins"
	PermCanBlockUsers         PermissionKey = "canBlockUsers"
	PermCanManageVerification PermissionKey = "canManageVerification"
)

func (p PermissionKey) IsValid() bool {
	switch p {
	case PermCanManageCatalog, PermCanManageMarketing, PermCanManageOrders,
		PermCanManageUsers, PermCanManageSettings, PermCanSendNotifications,
		PermCanManageAdmins, PermCanBlockUsers, PermCanManageVerification:
		return true
	}
	return false
}

// AllPermissionKeys is the canonical, ordered catalog of every permission
// key — used to validate role-permission payloads and to give the
// admin-portal a fixed list to render controls from.
func AllPermissionKeys() []PermissionKey {
	return []PermissionKey{
		PermCanManageCatalog, PermCanManageMarketing, PermCanManageOrders,
		PermCanManageUsers, PermCanManageSettings, PermCanSendNotifications,
		PermCanManageAdmins, PermCanBlockUsers, PermCanManageVerification,
	}
}

// AccessLevel is how much a role can do within a permission key's area.
// Read sees list/detail views only; Write additionally allows create/
// update/delete actions. There's no explicit "none" stored — a role simply
// has no RolePermission row for a key it can't access at all.
type AccessLevel string

const (
	AccessLevelRead  AccessLevel = "read"
	AccessLevelWrite AccessLevel = "write"
)

func (a AccessLevel) IsValid() bool {
	switch a {
	case AccessLevelRead, AccessLevelWrite:
		return true
	}
	return false
}

// CanWrite reports whether this level permits mutating actions.
func (a AccessLevel) CanWrite() bool { return a == AccessLevelWrite }

// RoleNameSuperAdmin and RoleNameSeller are reserved role names with special
// handling: super_admin implicitly carries every permission at write level
// (never stored/edited as rows — see AdminUseCase.GetPermissionsForRole)
// and can't be deleted or renamed; seller is a customer-facing account
// type, never offered as an assignable admin-portal role.
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
	Permissions []RolePermissionGrant `json:"permissions" gorm:"-"`
}

// RolePermissionGrant is one key+level pair, as exposed in the Role API
// (both request payloads and responses) and in the resolved permission set
// returned by GET /api/admin/me.
type RolePermissionGrant struct {
	PermissionKey PermissionKey `json:"permission_key" binding:"required"`
	AccessLevel   AccessLevel   `json:"access_level" binding:"required"`
}

// RolePermission is a single granted-permission row for a role, persisted
// in the role_permissions table.
type RolePermission struct {
	ID            string        `json:"id" gorm:"primaryKey;type:varchar(32)"`
	RoleID        string        `json:"role_id" gorm:"type:varchar(32);index;not null"`
	PermissionKey PermissionKey `json:"permission_key" gorm:"size:50;not null"`
	AccessLevel   AccessLevel   `json:"access_level" gorm:"size:10;not null;default:'write'"`
}
