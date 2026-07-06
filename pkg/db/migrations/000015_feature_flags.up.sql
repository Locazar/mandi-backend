CREATE TABLE IF NOT EXISTS feature_flags (
    id          VARCHAR(32)  PRIMARY KEY,
    flag_key    VARCHAR(64)  NOT NULL UNIQUE,
    enabled     BOOLEAN      NOT NULL DEFAULT FALSE,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- The advertisement feature ships disabled.
INSERT INTO feature_flags (id, flag_key, enabled, description)
VALUES ('ff_seed_advertisement', 'advertisement', FALSE, 'Seller advertisement requests (drawer entry, request & payment flow)')
ON CONFLICT (flag_key) DO NOTHING;
