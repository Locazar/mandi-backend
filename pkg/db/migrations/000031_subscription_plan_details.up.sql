-- Admin-editable marketing copy shown on the seller-app plan-details bottom
-- sheet; neither column affects billing logic.
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS features TEXT[] NOT NULL DEFAULT '{}';
