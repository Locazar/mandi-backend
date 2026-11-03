-- 000002_seller_job_postings.up.sql
-- Seller-posted job requirements.
-- A seller uses the "Post Job" feature in the seller-app to advertise an open
-- position at their shop (e.g. "need 2 clothing helpers, age 18-30").
-- This is separate from the generic job-marketplace `jobs` table.

CREATE TABLE IF NOT EXISTS seller_job_postings (
    id             VARCHAR(32)  PRIMARY KEY,
    shop_id        VARCHAR(32)  REFERENCES shop_details (id) ON DELETE SET NULL,
    shop_name      VARCHAR(255) NOT NULL,
    department     VARCHAR(255) NOT NULL,
    persons_needed INTEGER      NOT NULL DEFAULT 1 CHECK (persons_needed >= 1),
    age_min        INTEGER      CHECK (age_min >= 14),
    age_max        INTEGER      CHECK (age_max >= 14),
    city           VARCHAR(255) NOT NULL DEFAULT '',
    state          VARCHAR(255) NOT NULL DEFAULT '',
    show_phone     BOOLEAN      NOT NULL DEFAULT TRUE,
    phone          VARCHAR(50),
    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_seller_job_postings_shop_id    ON seller_job_postings (shop_id);
CREATE INDEX IF NOT EXISTS idx_seller_job_postings_is_active  ON seller_job_postings (is_active);
CREATE INDEX IF NOT EXISTS idx_seller_job_postings_created_at ON seller_job_postings (created_at DESC);
