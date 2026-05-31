-- Add user_name column to admins table
ALTER TABLE admins ADD COLUMN IF NOT EXISTS user_name VARCHAR(255);
