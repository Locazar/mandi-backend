-- Initialize database for e-commerce application
-- This script runs automatically when the PostgreSQL container starts

-- Create database extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create sequences for auto-incrementing IDs
-- These will be used by GORM automatically, but we can pre-create them

-- You can add custom indexes here for better performance
-- CREATE INDEX idx_users_email ON users(email);
-- CREATE INDEX idx_users_phone ON users(phone);
-- CREATE INDEX idx_products_name ON products(name);

-- Insert default countries if needed
INSERT INTO countries (country_name, iso_code) VALUES 
    ('India', 'IN'),
    ('United States', 'US'),
    ('United Kingdom', 'GB'),
    ('Canada', 'CA')
ON CONFLICT (iso_code) DO NOTHING;

-- Create any custom functions or triggers here if needed

-- Grant permissions (usually not needed with the default setup)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO rohitjangid;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO rohitjangid;