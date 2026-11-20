ALTER TABLE subscription_expiry_nudge_settings
  DROP COLUMN IF EXISTS fire_hour_ist,
  DROP COLUMN IF EXISTS fire_minute_ist;
