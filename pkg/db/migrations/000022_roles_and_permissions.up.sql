-- Custom roles + per-role sidebar/feature permissions, replacing the
-- previously-hardcoded 4-role enum. Seeded with the same 4 roles and the
-- same permission matrix the admin-portal already hardcoded in
-- src/lib/permissions.ts, so existing admins' access is unchanged.

CREATE TABLE IF NOT EXISTS roles (
    id         VARCHAR(32) PRIMARY KEY,
    name       VARCHAR(50) NOT NULL UNIQUE,
    label      VARCHAR(100) NOT NULL,
    is_system  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS role_permissions (
    id             VARCHAR(32) PRIMARY KEY,
    role_id        VARCHAR(32) NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_key VARCHAR(50) NOT NULL
        CHECK (permission_key IN (
            'canManageCatalog', 'canManageMarketing', 'canManageOrders',
            'canManageUsers', 'canManageSettings', 'canSendNotifications',
            'canManageAdmins', 'canBlockUsers'
        )),
    UNIQUE (role_id, permission_key)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);

INSERT INTO roles (id, name, label, is_system) VALUES
    ('role_super_admin',      'super_admin',      'Super Admin',      TRUE),
    ('role_support_staff',    'support_staff',    'Support Staff',    TRUE),
    ('role_catalog_manager',  'catalog_manager',  'Catalog Manager',  TRUE),
    ('role_marketing_manager','marketing_manager','Marketing Manager',TRUE)
ON CONFLICT (name) DO NOTHING;

-- super_admin is intentionally seeded with NO permission rows: it's treated
-- as "all permissions, always" in application code and its rows are never
-- read for that role (see AdminUseCase.GetPermissionsForRole), so it's
-- never editable/removable via the roles UI either.

INSERT INTO role_permissions (id, role_id, permission_key) VALUES
    ('rolp_support_orders',   'role_support_staff',     'canManageOrders'),
    ('rolp_support_users',    'role_support_staff',     'canManageUsers'),
    ('rolp_support_block',    'role_support_staff',     'canBlockUsers'),
    ('rolp_catalog_catalog',  'role_catalog_manager',   'canManageCatalog'),
    ('rolp_marketing_mkt',    'role_marketing_manager', 'canManageMarketing')
ON CONFLICT (role_id, permission_key) DO NOTHING;
