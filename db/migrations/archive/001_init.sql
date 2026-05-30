-- SQL migration for auth service tables
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(30) UNIQUE,
    password_hash VARCHAR(255),
    pin_hash VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS otp_requests (
    id SERIAL PRIMARY KEY,
    target VARCHAR(255),
    otp_hash VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE,
    attempts INTEGER DEFAULT 0,
    max_attempt INTEGER DEFAULT 3,
    status VARCHAR(50) DEFAULT 'pending',
    ip_address VARCHAR(50),
    user_agent VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS login_audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER,
    event_type VARCHAR(50),
    ip_address VARCHAR(50),
    user_agent VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
