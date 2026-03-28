CREATE TABLE IF NOT EXISTS fcm_tokens (
    id          BIGSERIAL PRIMARY KEY,
    token       TEXT        NOT NULL,
    device      VARCHAR(50) NOT NULL DEFAULT 'android',
    platform    VARCHAR(50) NOT NULL DEFAULT 'android',
    owner_id    VARCHAR(100) NOT NULL DEFAULT '',
    owner_type  VARCHAR(10)  NOT NULL DEFAULT 'user',
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Legacy compatibility: older deployments used shop_id/admin_id without
-- owner_id/owner_type/is_active. Convert in-place so existing tokens keep working.
ALTER TABLE fcm_tokens ADD COLUMN IF NOT EXISTS owner_id VARCHAR(100);
ALTER TABLE fcm_tokens ADD COLUMN IF NOT EXISTS owner_type VARCHAR(10);
ALTER TABLE fcm_tokens ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

-- Backfill owner fields from legacy columns when needed.
UPDATE fcm_tokens
SET owner_id = CAST(shop_id AS TEXT)
WHERE (owner_id IS NULL OR owner_id = '')
  AND shop_id IS NOT NULL;

UPDATE fcm_tokens
SET owner_type = CASE
    WHEN admin_id IS NOT NULL AND admin_id > 0 THEN 'seller'
    ELSE 'user'
END
WHERE owner_type IS NULL OR owner_type = '';

ALTER TABLE fcm_tokens ALTER COLUMN owner_id SET NOT NULL;
ALTER TABLE fcm_tokens ALTER COLUMN owner_type SET NOT NULL;

-- Enforce valid owner_type values expected by application code.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fcm_tokens_owner_type_check'
    ) THEN
        ALTER TABLE fcm_tokens
            ADD CONSTRAINT fcm_tokens_owner_type_check
            CHECK (owner_type IN ('user', 'seller'));
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_fcm_tokens_token ON fcm_tokens (token);
CREATE INDEX IF NOT EXISTS idx_fcm_tokens_owner_active ON fcm_tokens (owner_id, owner_type) WHERE is_active = TRUE;