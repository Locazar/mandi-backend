-- Self-healing migration: the live database was found missing
-- shop_socials.comments even though it's defined in both
-- 000001_baseline.up.sql and 000003_add_shop_socials.up.sql (likely
-- migration drift from an earlier deploy). ADD COLUMN IF NOT EXISTS is a
-- no-op wherever the column already exists, and fixes it wherever it
-- doesn't.
ALTER TABLE shop_socials
    ADD COLUMN IF NOT EXISTS comments TEXT NOT NULL DEFAULT '';
