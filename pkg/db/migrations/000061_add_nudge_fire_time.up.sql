ALTER TABLE subscription_expiry_nudge_settings
  ADD COLUMN IF NOT EXISTS fire_hour_ist   integer NOT NULL DEFAULT 9,
  ADD COLUMN IF NOT EXISTS fire_minute_ist integer NOT NULL DEFAULT 0;
