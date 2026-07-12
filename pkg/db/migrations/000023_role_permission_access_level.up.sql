-- Adds a read/write access level to each role_permission grant. Existing
-- rows default to 'write' so every previously-created role keeps exactly
-- the same (full) access it had before this migration.

ALTER TABLE role_permissions
    ADD COLUMN IF NOT EXISTS access_level VARCHAR(10) NOT NULL DEFAULT 'write'
        CHECK (access_level IN ('read', 'write'));
