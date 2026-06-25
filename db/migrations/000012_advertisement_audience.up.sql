-- Add audience segmentation columns to advertisements.
-- audience: 'customer' (default) or 'seller'
-- department_id / category_id: optional targeting for customer ads.

ALTER TABLE advertisements
    ADD COLUMN IF NOT EXISTS audience      VARCHAR(20)  NOT NULL DEFAULT 'customer',
    ADD COLUMN IF NOT EXISTS department_id VARCHAR(32)  DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS category_id   VARCHAR(32)  DEFAULT NULL;

-- Backfill existing rows so the NOT NULL constraint is satisfied.
UPDATE advertisements SET audience = 'customer' WHERE audience IS NULL OR audience = '';
