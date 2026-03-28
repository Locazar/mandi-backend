-- Create support users table for support care staff
CREATE TABLE IF NOT EXISTS supports (
    id BIGSERIAL PRIMARY KEY,
    full_name VARCHAR(120) NOT NULL,
    email VARCHAR(120) NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    department VARCHAR(80),
    role VARCHAR(50) NOT NULL DEFAULT 'support_care',
    can_access_admin BOOLEAN NOT NULL DEFAULT TRUE,
    can_access_user BOOLEAN NOT NULL DEFAULT TRUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create support refresh sessions table
CREATE TABLE IF NOT EXISTS support_refresh_sessions (
    token_id VARCHAR(255) PRIMARY KEY,
    support_id BIGINT,
    user_type VARCHAR(20),
    refresh_token TEXT NOT NULL,
    expire_at TIMESTAMP NOT NULL,
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_support_refresh_sessions_support_id ON support_refresh_sessions(support_id);
CREATE INDEX IF NOT EXISTS idx_support_refresh_sessions_expire_at ON support_refresh_sessions(expire_at);
