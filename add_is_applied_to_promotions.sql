-- Add is_applied column to promotions
-- is_applied = false: deal just created, not yet explicitly applied by the seller
-- is_applied = true: seller has actively applied (or resumed) this deal
-- Synced with is_active: true when activated, false when paused
ALTER TABLE promotions ADD COLUMN IF NOT EXISTS is_applied BOOLEAN NOT NULL DEFAULT false;
