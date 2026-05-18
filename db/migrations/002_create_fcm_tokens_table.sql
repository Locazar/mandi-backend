-- +migrate Up
CREATE TABLE fcm_tokens (
    id SERIAL PRIMARY KEY,
    token TEXT UNIQUE NOT NULL,
    device TEXT,
    platform TEXT,
    owner_id VARCHAR(100) NOT NULL,
    owner_type VARCHAR(10) NOT NULL CHECK (owner_type IN ('user', 'seller')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fcm_tokens_owner_active
    ON fcm_tokens(owner_id, owner_type)
    WHERE is_active = TRUE;

-- +migrate Down
DROP TABLE fcm_tokens;
