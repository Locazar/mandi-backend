-- Note: the one-time data repair (nulling role for auto-generated seller
-- accounts) in the up migration is intentionally not reversed here — we
-- can't know what value those rows held before the original bug, and
-- restoring 'super_admin' on seller accounts would just reintroduce the
-- privilege-escalation bug this migration exists to fix.
ALTER TABLE admins ALTER COLUMN role SET DEFAULT 'super_admin';
