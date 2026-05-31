-- Align users table with the User domain struct.
-- Struct uses email_verified, phone_verified, date_of_birth.
-- Old schema had verified (single bool) and age (bigint).

ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_verified  BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS date_of_birth   DATE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at      TIMESTAMPTZ;

-- Migrate existing `verified` values into phone_verified (the closer semantic match).
DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'verified'
	) THEN
		EXECUTE 'UPDATE users SET phone_verified = verified WHERE verified IS NOT NULL';
	END IF;
END $$;

-- Drop legacy columns no longer in the struct.
ALTER TABLE users DROP COLUMN IF EXISTS verified;
ALTER TABLE users DROP COLUMN IF EXISTS age;

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
