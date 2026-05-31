-- Migration: Add geo-targeting fields to banners table
-- Run this against your PostgreSQL database

ALTER TABLE banners
    ADD COLUMN IF NOT EXISTS department_id  INTEGER      DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS category_id    INTEGER      DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS app_type       VARCHAR(50)  DEFAULT '',
    ADD COLUMN IF NOT EXISTS latitude       DECIMAL(10,7) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS longitude      DECIMAL(10,7) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS pincode        VARCHAR(20)  DEFAULT '',
    ADD COLUMN IF NOT EXISTS radius_km      DECIMAL(10,2) DEFAULT 0;

-- Optional indexes for faster filtering
CREATE INDEX IF NOT EXISTS idx_banners_department_id ON banners(department_id);
CREATE INDEX IF NOT EXISTS idx_banners_category_id   ON banners(category_id);
CREATE INDEX IF NOT EXISTS idx_banners_app_type       ON banners(app_type);
CREATE INDEX IF NOT EXISTS idx_banners_pincode        ON banners(pincode);

-- Existing banners: set radius_km = 0 (global — no geo restriction)
-- and app_type = 'all' so they still appear for all users
UPDATE banners
SET radius_km = 0,
    app_type  = 'customer'
WHERE radius_km IS NULL OR radius_km = 0;
